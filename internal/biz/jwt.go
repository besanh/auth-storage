package biz

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenTTL  = time.Hour * 1
	RefreshTokenTTL = time.Hour * 24
	IssuerName      = "auth-service"
)

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

var (
	ErrUserAlreadyExists  = fmt.Errorf("user with this email already exists")
	ErrInvalidCredentials = fmt.Errorf("invalid email or password")
)

func (uc *AuthUseCase) GenerateToken(userID, role, scope string) (string, string, int64, error) {
	now := time.Now().UTC()

	// 1. Access Token
	accessClaims := CustomClaims{
		Role:  role,
		Scope: scope,
		SpiceDB: SpiceDBClaim{
			Type: "user",
			ID:   userID,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    IssuerName,
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
			Audience:  []string{"auth-service"},
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessToken.Header["kid"] = uc.conf.Kid
	signedAccess, err := accessToken.SignedString(uc.privateKey)
	if err != nil {
		return "", "", 0, err
	}

	// 2. Refresh Token
	refreshClaims := CustomClaims{
		Role:  role,
		Scope: scope,
		SpiceDB: SpiceDBClaim{
			Type: "user",
			ID:   userID,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    IssuerName,
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
			Audience:  []string{"auth-service"},
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshToken.Header["kid"] = uc.conf.Kid
	signedRefresh, err := refreshToken.SignedString(uc.privateKey)
	if err != nil {
		return "", "", 0, err
	}

	return signedAccess, signedRefresh, accessClaims.ExpiresAt.Time.Unix(), nil
}
