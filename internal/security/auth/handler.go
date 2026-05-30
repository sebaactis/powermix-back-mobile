package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sebaactis/powermix-back-mobile/internal/clients/mailer"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/token"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	jwtx "github.com/sebaactis/powermix-back-mobile/internal/security/jwt"
	"github.com/sebaactis/powermix-back-mobile/internal/security/oauth"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
	"github.com/sebaactis/powermix-back-mobile/internal/validations"
	"golang.org/x/crypto/bcrypt"
)

type HTTPHandler struct {
	users     *user.Service
	tokens    *token.Service
	jwt       *jwtx.JWT
	validator validations.StructValidator
	mailer    mailer.Mailer
}

func NewHTTPHandler(users *user.Service,
	tokens *token.Service,
	jwt *jwtx.JWT,
	validator validations.StructValidator,
	mailer mailer.Mailer) *HTTPHandler {
	return &HTTPHandler{users: users,
		tokens:    tokens,
		jwt:       jwt,
		validator: validator,
		mailer:    mailer}
}

func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseLoginRequest(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "No se pudo parsear el cuerpo de la solicitud",
		})
		return
	}

	user, err := h.authenticateUser(r.Context(), req)
	if err != nil {
		h.handleLoginError(w, r.Context(), err, user, req.Email)
		return
	}

	tokens, err := h.generateTokens(r.Context(), user)
	if err != nil {
		slog.ErrorContext(r.Context(), "error al generar tokens en login", "user_id", user.ID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "No se pueden generar los tokens",
		})
		return
	}

	h.respondWithTokens(w, user, tokens)
}

func (h *HTTPHandler) OAuthGoogle(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AccessToken == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "Access_token es requerido",
		})
		return
	}

	userInfo, err := oauth.GetGoogleUserInfo(r.Context(), body.AccessToken)
	if err != nil {
		slog.ErrorContext(r.Context(), "error al validar token de Google", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeExternalService,
			Message: "Token de Google inválido",
		})
		return
	}

	slog.InfoContext(r.Context(), "OAuth Google login", "email", userInfo.Email, "provider", userInfo.Provider)

	user, err := h.users.FindOrCreateFromOAuth(r.Context(), userInfo)
	if err != nil {
		slog.ErrorContext(r.Context(), "error al guardar usuario OAuth", "email", userInfo.Email, "error", err)
		utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
			Code:    utils.ErrCodeUnauthorized,
			Message: "No se pudo guardar el usuario",
		})
		return
	}

	accessToken, accessExpiration, err := h.jwt.Sign(user.ID, user.Email, jwtx.TokenTypeAccess)
	if err != nil {
		slog.ErrorContext(r.Context(), "error al generar access token OAuth", "user_id", user.ID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "Error generando el access token",
		})
		return
	}

	refreshToken, refreshExpiration, err := h.jwt.Sign(user.ID, user.Email, jwtx.TokenTypeRefresh)
	if err != nil {
		slog.ErrorContext(r.Context(), "error al generar refresh token OAuth", "user_id", user.ID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "Error generando el refresh token",
		})
		return
	}

	if _, err = h.tokens.CreateInitialRefreshToken(r.Context(), user.ID, refreshToken, refreshExpiration); err != nil {
		slog.ErrorContext(r.Context(), "error al persistir refresh token OAuth", "user_id", user.ID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "Error generando el refresh token",
		})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]any{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
		"accessToken":      accessToken,
		"refreshToken":     refreshToken,
		"accessExpiresAt":  accessExpiration,
		"refreshExpiresAt": refreshExpiration,
	})
}

func (h *HTTPHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
			Code:    utils.ErrCodeUnauthorized,
			Message: "Refresh token vacío",
		})
		return
	}

	refreshToken := strings.TrimPrefix(authHeader, "Bearer ")

	userID, email, tokenType, err := h.jwt.Parse(refreshToken, jwtx.TokenTypeRefresh)
	if err != nil || tokenType != jwtx.TokenTypeRefresh {
		utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
			Code:    utils.ErrCodeUnauthorized,
			Message: "Refresh token inválido",
		})
		return
	}

	ctx := r.Context()
	now := time.Now()

	var (
		newAccessToken  string
		newRefreshToken string
		accessExp       time.Time
		refreshExp      time.Time
	)

	err = h.tokens.Transaction(ctx, func(tokensTx *token.Service) error {
		var err error

		newAccessToken, accessExp, newRefreshToken, refreshExp, err =
			tokensTx.RotateRefresh(
				ctx,
				refreshToken,
				now,
				func() (string, time.Time, error) {
					return h.jwt.Sign(userID, email, jwtx.TokenTypeAccess)
				},
				func() (string, time.Time, error) {
					return h.jwt.Sign(userID, email, jwtx.TokenTypeRefresh)
				},
			)

		return err
	})

	if err != nil {
		switch {
		case errors.Is(err, token.ErrRefreshInvalid):
			utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
				Code:    utils.ErrCodeUnauthorized,
				Message: "Refresh token inválido o expirado",
			})
			return
		case errors.Is(err, token.ErrRefreshReuseDetected):
			utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
				Code:    utils.ErrCodeUnauthorized,
				Message: "Tu sesión ya no puede ser utilizada. Iniciá sesión nuevamente.",
			})
			return
		default:
			slog.ErrorContext(r.Context(), "error al refrescar token", "user_id", userID, "error", err)
			utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
				Code:    utils.ErrCodeInternal,
				Message: "Error al refrescar token",
			})
			return
		}
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]any{
		"accessToken":      newAccessToken,
		"refreshToken":     newRefreshToken,
		"accessExpiresAt":  accessExp,
		"refreshExpiresAt": refreshExp,
	})
}

func (h *HTTPHandler) RecoveryPasswordRequest(w http.ResponseWriter, r *http.Request) {
	var req RecoveryPasswordRequest
	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "No se pudo parsear el body de la request",
		})
		return
	}

	genericResponse := func() {
		utils.WriteSuccess(w, http.StatusOK, map[string]any{
			"email":   req.Email,
			"message": "Si cargó un email válido, recibira un link de recuperación",
		})
	}

	user, err := h.users.GetByEmail(ctx, req.Email)
	if err != nil {
		genericResponse()
		return
	}

	recoveryToken, err := h.generateTokenRecovery(ctx, user)
	if err != nil {
		slog.ErrorContext(r.Context(), "error al generar token de recuperación", "user_id", user.ID, "error", err)
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "Error al generar el token de recuperación",
		})
		return
	}

	tokenEscaped := url.QueryEscape(*recoveryToken)
	emailEscaped := url.QueryEscape(user.Email)
	resetURL := fmt.Sprintf("https://powermixstation.com.ar/reset-password?token=%s&email=%s", tokenEscaped, emailEscaped)

	if err := h.mailer.SendResetPasswordEmail(ctx, user.Email, resetURL); err != nil {
		genericResponse()
		slog.ErrorContext(r.Context(), "error al enviar email de recovery", "email", user.Email, "error", err)
		return
	}

	genericResponse()
}

func (h *HTTPHandler) UnlockUser(w http.ResponseWriter, r *http.Request) {
	var req UnlockUserReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "JSON inválido",
		})
		return
	}

	h.users.UnlockUser(r.Context(), req.UserId)
	utils.WriteSuccess(w, http.StatusOK, map[string]any{"message": "User unlocked"})
}

func (h *HTTPHandler) UpdatePasswordByRecovery(w http.ResponseWriter, r *http.Request) {
	var req user.UserRecoveryPassword

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
			Code:    utils.ErrCodeValidation,
			Message: "No se pudo parsear el cuerpo de la request",
		})
		return
	}

	userId, _, tokenType, err := h.jwt.ParseResetPassword(req.Token)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
			Code:    utils.ErrCodeUnauthorized,
			Message: "Token de recuperación inválido o expirado",
		})
		return
	}

	if tokenType != jwtx.TokenTypeResetPassword {
		utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
			Code:    utils.ErrCodeUnauthorized,
			Message: "Tipo de token inválido",
		})
		return
	}

	userRecovery, err := h.users.UpdatePasswordByRecovery(r.Context(), user.UserRecoveryPassword{
		UserID:          userId,
		Token:           req.Token,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	})

	if err != nil {
		if fields, ok := validations.AsValidationError(err); ok {
			utils.WriteError(w, http.StatusBadRequest, utils.WriteErrorOpts{
				Code:    utils.ErrCodeValidation,
				Message: "Error de validación",
				Fields:  fields,
			})
			return
		}
		// Token inválido/expirado/usado → 401 (mismo mensaje por privacidad)
		if errors.Is(err, token.ErrTokenInvalid) || errors.Is(err, user.ErrNotFound) {
			utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
				Code:    utils.ErrCodeUnauthorized,
				Message: "Token de recuperación inválido o expirado",
			})
			return
		}
		slog.ErrorContext(r.Context(), "error al actualizar contraseña por recuperación", "user_id", userId, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "Error en el servidor",
		})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, user.ToResponse(userRecovery))
}

// ==================== MÉTODOS PRIVADOS ====================

func (h *HTTPHandler) parseLoginRequest(r *http.Request) (*LoginRequest, error) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.New("invalid json")
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if fields, ok := h.validator.ValidateStruct(&req); !ok {
		msgs := make([]string, 0, len(fields))
		for _, m := range fields {
			msgs = append(msgs, m)
		}
		return nil, errors.New(strings.Join(msgs, ", "))
	}

	return &req, nil
}

func (h *HTTPHandler) authenticateUser(ctx context.Context, req *LoginRequest) (*user.User, error) {

	user, err := h.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.LockedUntil.After(time.Now()) {
		return user, ErrAccountLocked
	}

	wasUnlocked, err := h.users.CheckAndUnlockIfExpired(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if wasUnlocked {
		user, err = h.users.GetByID(ctx, user.ID)
		if err != nil {
			return nil, err
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return user, ErrInvalidCredentials
	}

	return user, nil
}

func (h *HTTPHandler) generateTokens(ctx context.Context, user *user.User) (*TokenPair, error) {
	accessToken, _, err := h.jwt.Sign(user.ID, user.Email, jwtx.TokenTypeAccess)
	if err != nil {
		return nil, err
	}

	refreshToken, expirationRefresh, err := h.jwt.Sign(user.ID, user.Email, jwtx.TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	if _, err = h.tokens.CreateInitialRefreshToken(ctx, user.ID, refreshToken, expirationRefresh); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *HTTPHandler) generateTokenRecovery(ctx context.Context, user *user.User) (*string, error) {

	recoveryToken, expirationRecovery, err := h.jwt.Sign(user.ID, user.Email, jwtx.TokenTypeResetPassword)
	if err != nil {
		return nil, err
	}

	if _, err = h.tokens.Create(ctx, &token.TokenRequest{
		TokenType: string(jwtx.TokenTypeResetPassword),
		Token:     recoveryToken,
		UserId:    user.ID,
		ExpiresAt: expirationRecovery,
	}); err != nil {
		return nil, err
	}

	return &recoveryToken, nil
}

func (h *HTTPHandler) handleLoginError(w http.ResponseWriter, ctx context.Context, err error, user *user.User, email string) {
	switch err {
	case ErrAccountLocked:
		utils.WriteError(w, http.StatusLocked, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInvalidCreds,
			Message: "Cuenta temporalmente bloqueada",
		})

	case ErrInvalidCredentials:
		utils.WriteError(w, http.StatusUnauthorized, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInvalidCreds,
			Message: "Credenciales inválidas",
		})

		if user == nil {
			user, _ = h.users.GetByEmail(ctx, email)
		}

		if user != nil {
			h.users.IncrementLoginAttempt(ctx, user.ID)
		}

	default:
		slog.ErrorContext(ctx, "error interno en login", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
			Code:    utils.ErrCodeInternal,
			Message: "Error interno del servidor",
		})
	}
}

func (h *HTTPHandler) respondWithTokens(w http.ResponseWriter, user *user.User, tokens *TokenPair) {
	response := LoginResponse{
		Email:         user.Email,
		Name:          user.Name,
		StampsCounter: user.StampsCounter,
		Token:         tokens.AccessToken,
		RefreshToken:  tokens.RefreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

var (
	ErrInvalidCredentials = errors.New("Credenciales inválidas")
	ErrAccountLocked      = errors.New("Cuenta bloqueada")
)
