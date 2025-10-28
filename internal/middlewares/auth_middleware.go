package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/token"
	"github.com/sebaactis/powermix-back-mobile/internal/domain/entities/user"
	jwtx "github.com/sebaactis/powermix-back-mobile/internal/security/jwt"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type ctxKey string

const ctxUserID ctxKey = "jwtx.user_id"

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ctxUserID)
	if v == nil {
		return uuid.Nil, false
	}

	id, ok := v.(uuid.UUID)
	return id, ok
}

type AuthMiddleware struct {
	jwt          *jwtx.JWT
	userService  *user.Service
	tokenService *token.Service
}

func NewAuthMiddleware(jwt *jwtx.JWT, userService *user.Service, tokenService *token.Service) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt, userService: userService, tokenService: tokenService}
}

func (a *AuthMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.WriteError(w, http.StatusUnauthorized, "El authorization header se encuentra vacio", nil)
				return
			}

			const prefix = "Bearer "
			if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
				utils.WriteError(w, http.StatusUnauthorized, "El authorization header tiene un formato invalido", nil)
				return
			}

			accessToken := authHeader[len(prefix):]

			userID, _, tokenType, err := a.jwt.Parse(accessToken, jwtx.TokenTypeAccess)
			if err != nil || tokenType != jwtx.TokenTypeAccess {
				_ = a.tokenService.RevokeToken(r.Context(), accessToken)
				utils.WriteError(w, http.StatusUnauthorized, "Token invalido o expirado", nil)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
