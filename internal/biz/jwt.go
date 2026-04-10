package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
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
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   req.UserID,
			Issuer:    IssuerName,
			ExpiresAt: jwtv5.NewNumericDate(now.Add(AccessTokenTTL)),
			IssuedAt:  jwtv5.NewNumericDate(now),
			ID:        uuid.New().String(),
			Audience:  []string{IssuerName},
		},
	}
	accessToken := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, accessClaims)
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
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   req.UserID,
			Issuer:    IssuerName,
			ExpiresAt: jwtv5.NewNumericDate(now.Add(RefreshTokenTTL)),
			IssuedAt:  jwtv5.NewNumericDate(now),
			ID:        uuid.New().String(),
			Audience:  []string{IssuerName},
		},
	}
	refreshToken := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, refreshClaims)
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

func FromContext(ctx context.Context) (string, bool) {
	if claims, ok := jwt.FromContext(ctx); ok {
		if c, ok := claims.(*CustomClaims); ok {
			return c.Subject, true
		}
	}
	return "", false
}
