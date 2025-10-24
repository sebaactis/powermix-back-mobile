package jwtx

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type JWT struct {
	secret       []byte
	reset_secret []byte
	ttlReset     time.Duration
	ttlNormal    time.Duration
	ttlRefresh   time.Duration
}

func NewJWT() *JWT {
	sec := os.Getenv("JWT_SECRET")

	if sec == "" {
		sec = "dev-secret"
	}

	resetSec := os.Getenv("JWT_RECOVERY_PASS_SECRET")

	if resetSec == "" {
		resetSec = "dev-reset-secret"
	}

	ttlMin := 60
	ttlMinReset := 15
	ttlMinRefresh := 1440

	if s := os.Getenv("JWT_TTL_MINUTES"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			ttlMin = n
		}
	}

	if s := os.Getenv("JWT_TTL_RECOVERY_MINUTES"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			ttlMinReset = n
		}
	}

	if s := os.Getenv("JWT_TTL_REFRESH_MINUTES"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			ttlMinRefresh = n
		}
	}

	return &JWT{
		secret:       []byte(sec),
		reset_secret: []byte(resetSec),
		ttlReset:     time.Duration(ttlMinReset) * time.Minute,
		ttlNormal:    time.Duration(ttlMin) * time.Minute,
		ttlRefresh:   time.Duration(ttlMinRefresh) * time.Minute}
}

func (j *JWT) Sign(userID uuid.UUID, email string, tokenType TokenType) (string, error) {
	now := time.Now()

	claims := Claims{
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(j.getExpiration(tokenType, now)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := j.secret

	if tokenType == TokenTypeResetPassword {
		secret = j.reset_secret
	}

	return token.SignedString(secret)
}

func (j *JWT) Parse(tokenIn string, tokenType TokenType) (uint, string, TokenType, error) {
	return j.parseWithSecret(tokenIn, j.secret, tokenType)
}

func (j *JWT) ParseResetPassword(tokenIn string) (uint, string, TokenType, error) {
	return j.parseWithSecret(tokenIn, j.reset_secret, TokenTypeResetPassword)
}

func (j *JWT) parseWithSecret(tokenIn string, secret []byte, expectedType TokenType) (uint, string, TokenType, error) {
	token, err := jwt.ParseWithClaims(
		tokenIn,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
	)

	if err != nil || !token.Valid {
		return 0, "", "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return 0, "", "", errors.New("invalid token claims")
	}

	if claims.TokenType != expectedType {
		return 0, "", "", errors.New("invalid token type")
	}

	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return 0, "", "", errors.New("token expired")
	}

	idU64, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, "", "", errors.New("invalid subject")
	}

	return uint(idU64), claims.Email, claims.TokenType, nil
}

func (j *JWT) getExpiration(tokenType TokenType, now time.Time) time.Time {
	if tokenType == TokenTypeAccess {
		return now.Add(j.ttlNormal)
	}

	if tokenType == TokenTypeResetPassword {
		return now.Add(j.ttlReset)
	}

	return now.Add(j.ttlRefresh)
}

func (j *JWT) GetTTL(tokenType TokenType) time.Duration {
	var ttl time.Duration

	if tokenType == TokenTypeAccess {
		ttl = j.ttlNormal
	} else {
		ttl = j.ttlRefresh
	}

	return ttl
}
