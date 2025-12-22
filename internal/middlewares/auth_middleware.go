package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	jwtx "github.com/sebaactis/powermix-back-mobile/internal/security/jwt"
	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type ctxKey string

const (
	ctxUserID    ctxKey = "jwtx.user_id"
	ctxUserEmail ctxKey = "jwtx.email"
)

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ctxUserID)
	if v == nil {
		return uuid.Nil, false
	}

	id, ok := v.(uuid.UUID)
	return id, ok
}

func UserEmailFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxUserEmail)
	if v == nil {
		return "", false
	}

	email, ok := v.(string)
	return email, ok
}

type AuthMiddleware struct {
	jwt *jwtx.JWT
}

func NewAuthMiddleware(jwt *jwtx.JWT) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
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
			if !strings.HasPrefix(authHeader, prefix) {
				utils.WriteError(w, http.StatusUnauthorized, "El authorization header tiene un formato invalido", nil)
				return
			}

			accessToken := strings.TrimPrefix(authHeader, prefix)
			if accessToken == "" {
				utils.WriteError(w, http.StatusUnauthorized, "Token vacio", nil)
				return
			}

			userID, email, tokenType, err := a.jwt.Parse(accessToken, jwtx.TokenTypeAccess)
			if err != nil || tokenType != jwtx.TokenTypeAccess {
				utils.WriteError(w, http.StatusUnauthorized, "Token invalido o expirado", nil)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			ctx = context.WithValue(ctx, ctxUserEmail, email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
