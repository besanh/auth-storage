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

func (uc *AuthUseCase) GenerateToken(req *GenerateTokenRequest) (*GenerateTokenReply, error) {
	now := time.Now().UTC()

	// 1. Access Token
	accessClaims := CustomClaims{
		Role:  req.Role,
		Scope: req.Scope,
		SpiceDB: SpiceDBClaim{
			Type: "user",
			ID:   req.UserID,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   req.UserID,
			Issuer:    IssuerName,
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
			Audience:  []string{IssuerName},
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessToken.Header["kid"] = uc.conf.Kid
	signedAccess, err := accessToken.SignedString(uc.privateKey)
	if err != nil {
		return &GenerateTokenReply{}, err
	}

	// 2. Refresh Token
	refreshClaims := CustomClaims{
		Role:  req.Role,
		Scope: req.Scope,
		SpiceDB: SpiceDBClaim{
			Type: "user",
			ID:   req.UserID,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   req.UserID,
			Issuer:    IssuerName,
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
			Audience:  []string{IssuerName},
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshToken.Header["kid"] = uc.conf.Kid
	signedRefresh, err := refreshToken.SignedString(uc.privateKey)
	if err != nil {
		return &GenerateTokenReply{}, err
	}

	return &GenerateTokenReply{
		AccessToken:  signedAccess,
		RefreshToken: signedRefresh,
		ExpiresAt:    accessClaims.ExpiresAt.Time.Unix(),
	}, nil
}
