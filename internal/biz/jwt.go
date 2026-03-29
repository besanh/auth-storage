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

var (
	ErrUserAlreadyExists  = fmt.Errorf("user with this email already exists")
	ErrInvalidCredentials = fmt.Errorf("invalid email or password")
)

func (uc *AuthUseCase) GenerateToken(userID string) (string, string, int64, error) {
	now := time.Now().UTC()

	// 1. Access Token
	accessClaims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    IssuerName,
		ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        uuid.New().String(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims).SignedString(uc.privateKey)
	if err != nil {
		return "", "", 0, err
	}

	// 2. Refresh Token
	refreshClaims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    IssuerName,
		ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        uuid.New().String(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims).SignedString(uc.privateKey)
	if err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, int64(AccessTokenTTL.Seconds()), nil
}
