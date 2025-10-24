package jwtx

type TokenType string

const (
	TokenTypeAccess        TokenType = "access"
	TokenTypeRefresh       TokenType = "refresh"
	TokenTypeResetPassword TokenType = "resetPassword"
)
