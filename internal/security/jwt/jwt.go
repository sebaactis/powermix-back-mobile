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

func (j *JWT) Sign(userID uuid.UUID, email string, tokenType TokenType) (string, time.Time, error) {
	now := time.Now()
	exp := j.getExpiration(tokenType, now)

	claims := Claims{
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := j.secret

	if tokenType == TokenTypeResetPassword {
		secret = j.reset_secret
	}

	signed, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, exp, nil
}

func (j *JWT) Parse(tokenIn string, tokenType TokenType) (uuid.UUID, string, TokenType, error) {
	return j.parseWithSecret(tokenIn, j.secret, tokenType)
}

func (j *JWT) ParseResetPassword(tokenIn string) (uuid.UUID, string, TokenType, error) {
	return j.parseWithSecret(tokenIn, j.reset_secret, TokenTypeResetPassword)
}

func (j *JWT) parseWithSecret(tokenIn string, secret []byte, expectedType TokenType) (uuid.UUID, string, TokenType, error) {
	token, err := jwt.ParseWithClaims(
		tokenIn,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("firma invalida")
			}
			return secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
	)

	if err != nil || !token.Valid {
		return uuid.UUID{}, "", "", errors.New("token invalido")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return uuid.UUID{}, "", "", errors.New("credenciales en el token invalidas")
	}

	if claims.TokenType != expectedType {
		return uuid.UUID{}, "", "", errors.New("tipo de token invalido")
	}

	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return uuid.UUID{}, "", "", errors.New("token expirado")
	}

	uid, err := uuid.Parse(claims.Subject)

	if err != nil {
		return uuid.UUID{}, "", "", errors.New("subject inválido")
	}

	return uid, claims.Email, claims.TokenType, nil
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
