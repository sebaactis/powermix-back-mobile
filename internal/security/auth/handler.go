package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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
}

func NewHTTPHandler(users *user.Service, tokens *token.Service, jwt *jwtx.JWT, validator validations.StructValidator) *HTTPHandler {
	return &HTTPHandler{users: users, tokens: tokens, jwt: jwt, validator: validator}
}

func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseLoginRequest(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	user, err := h.authenticateUser(r.Context(), req)
	if err != nil {
		h.handleLoginError(w, r.Context(), err, user)
		return
	}

	tokens, err := h.generateTokens(r.Context(), user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "cannot generate tokens", nil)
		return
	}
	h.setTokenCookie(w, "accessToken", tokens.AccessToken, jwtx.TokenTypeAccess)
	h.setTokenCookie(w, "refreshToken", tokens.RefreshToken, jwtx.TokenTypeRefresh)
	h.respondWithTokens(w, user, tokens)
}

func (h *HTTPHandler) OAuthGoogle(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AccessToken == "" {
		utils.WriteError(w, http.StatusBadRequest, "Access_token es requerido", nil)
		return
	}

	userInfo, err := oauth.GetGoogleUserInfo(r.Context(), body.AccessToken)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Token de google invalido", map[string]string{"error": err.Error()})
		return
	}

	user, err := h.users.FindOrCreateFromOAuth(r.Context(), userInfo)

	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "error al guardar el usuario", map[string]string{"error": err.Error()})
		return
	}

	accessToken, err := h.jwt.Sign(user.ID, user.Email, jwtx.TokenTypeAccess)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Error generando el access token", map[string]string{"error": err.Error()})
		return
	}

	refreshToken, err := h.jwt.Sign(user.ID, user.Email, jwtx.TokenTypeRefresh)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Error generando el refresh token", map[string]string{"error": err.Error()})
		return
	}

	h.tokens.Create(r.Context(), &token.TokenRequest{
		TokenType: string(jwtx.TokenTypeAccess),
		Token:     accessToken,
	})

	h.tokens.Create(r.Context(), &token.TokenRequest{
		TokenType: string(jwtx.TokenTypeRefresh),
		Token:     refreshToken,
	})

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (h *HTTPHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		utils.WriteError(w, http.StatusUnauthorized, "Refresh token vacio", nil)
		return
	}

	refreshToken := strings.TrimPrefix(authHeader, "Bearer ")

	userID, email, tokenType, err := h.jwt.Parse(refreshToken, jwtx.TokenTypeRefresh)
	if err != nil || tokenType != jwtx.TokenTypeRefresh {
		utils.WriteError(w, http.StatusUnauthorized, "Refresh token invalido", nil)
		return
	}

	// Generar nuevos tokens
	newAccessToken, err := h.jwt.Sign(userID, email, jwtx.TokenTypeAccess)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "No se pudo crear el access token", nil)
		return
	}

	newRefreshToken, err := h.jwt.Sign(userID, email, jwtx.TokenTypeRefresh)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "No se pudo crear el refresh token", nil)
		return
	}

	_ = h.tokens.RevokeToken(r.Context(), refreshToken)

	_, _ = h.tokens.Create(r.Context(), &token.TokenRequest{
		TokenType: string(jwtx.TokenTypeAccess),
		Token:     newAccessToken,
	})

	_, _ = h.tokens.Create(r.Context(), &token.TokenRequest{
		TokenType: string(jwtx.TokenTypeRefresh),
		Token:     newRefreshToken,
	})

	// Devolver al frontend
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"accessToken":  newAccessToken,
		"refreshToken": newRefreshToken,
	})
}

func (h *HTTPHandler) RecoveryPasswordRequest(w http.ResponseWriter, r *http.Request) {
	var req RecoveryPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json", nil)
		return
	}

	user, err := h.users.GetByEmail(r.Context(), req.Email)

	if err != nil {
		utils.WriteError(w, http.StatusOK, "if the mail exists, a recovery link will be sent you", nil)
		return
	}

	token, err := h.generateTokenRecovery(r.Context(), user)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Token generate failed", nil)
		return
	}

	response := &RecoveryPasswordRequestResponse{
		Email: req.Email,
		Token: *token,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *HTTPHandler) UnlockUser(w http.ResponseWriter, r *http.Request) {
	var req UnlockUserReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid json", nil)
		return
	}

	h.users.UnlockUser(r.Context(), req.UserId)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("User unlocked")
}

func (h *HTTPHandler) UpdatePasswordByRecovery(w http.ResponseWriter, r *http.Request) {
	var req user.UserRecoveryPassword

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	if _, _, _, err := h.jwt.ParseResetPassword(req.Token); err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	userRecovery, err := h.users.UpdatePasswordByRecovery(r.Context(), (req))

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, user.ToResponse(userRecovery))

}

// ==================== MÉTODOS PRIVADOS ====================

func (h *HTTPHandler) parseLoginRequest(r *http.Request) (*LoginRequest, error) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.New("invalid json")
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if fields, ok := h.validator.ValidateStruct(&req); !ok {
		return nil, fmt.Errorf("validation error: %v", fields)
	}

	return &req, nil
}

func (h *HTTPHandler) authenticateUser(ctx context.Context, req *LoginRequest) (*user.User, error) {

	user, err := h.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Locked_until.After(time.Now()) {
		return user, ErrAccountLocked
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return user, ErrInvalidCredentials
	}

	return user, nil
}

func (h *HTTPHandler) generateTokens(ctx context.Context, user *user.User) (*TokenPair, error) {
	accessToken, err := h.jwt.Sign(user.ID, user.Email, jwtx.TokenTypeAccess)
	if err != nil {
		return nil, err
	}

	refreshToken, err := h.jwt.Sign(user.ID, user.Email, jwtx.TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	if _, err = h.tokens.Create(ctx, &token.TokenRequest{
		TokenType: string(jwtx.TokenTypeAccess),
		Token:     accessToken,
	}); err != nil {
		return nil, err
	}

	if _, err = h.tokens.Create(ctx, &token.TokenRequest{
		TokenType: string(jwtx.TokenTypeRefresh),
		Token:     refreshToken,
	}); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *HTTPHandler) generateTokenRecovery(ctx context.Context, user *user.User) (*string, error) {

	recoveryToken, err := h.jwt.Sign(user.ID, user.Email, jwtx.TokenTypeResetPassword)
	if err != nil {
		return nil, err
	}

	if _, err = h.tokens.Create(ctx, &token.TokenRequest{
		TokenType: string(jwtx.TokenTypeResetPassword),
		Token:     recoveryToken,
	}); err != nil {
		return nil, err
	}

	return &recoveryToken, nil
}

func (h *HTTPHandler) handleLoginError(w http.ResponseWriter, ctx context.Context, err error, user *user.User) {
	switch err {
	case ErrAccountLocked:
		utils.WriteError(w, http.StatusLocked, "account temporarily locked", nil)

	case ErrInvalidCredentials:
		utils.WriteError(w, http.StatusUnauthorized, "invalid credentials", nil)

		// Incrementar intentos solo si el usuario existe
		if user != nil {
			h.users.IncrementLoginAttempt(ctx, user.ID)
		}

	default:
		utils.WriteError(w, http.StatusInternalServerError, "internal error", nil)
	}
}

func (h *HTTPHandler) setTokenCookie(w http.ResponseWriter, name string, token string, tokenType jwtx.TokenType) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		MaxAge:   int(h.jwt.GetTTL(tokenType).Seconds() * 2),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *HTTPHandler) respondWithTokens(w http.ResponseWriter, user *user.User, tokens *TokenPair) {
	response := LoginResponse{
		Email:        user.Email,
		Name:         user.Name,
		Token:        tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
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
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account locked")
)
