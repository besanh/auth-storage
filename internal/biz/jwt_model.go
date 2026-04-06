package biz

import "github.com/golang-jwt/jwt/v5"

type SpiceDBClaim struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CustomClaims struct {
	Role    string       `json:"role,omitempty"`
	Scope   string       `json:"scope,omitempty"`
	SpiceDB SpiceDBClaim `json:"spicedb,omitzero"`
	jwt.RegisteredClaims
}

type GenerateTokenRequest struct {
	UserID string
	Role   string
	Scope  string
}

type GenerateTokenReply struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}
