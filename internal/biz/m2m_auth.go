package biz

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"server/internal/conf"
	db "server/internal/data/db/generated"
)

type M2MAuthRepo interface {
	GetMachineClientByID(ctx context.Context, clientID string) (*db.MachineClient, error)
}

type M2MAuthUseCase struct {
	repo       M2MAuthRepo
	privateKey *rsa.PrivateKey
	conf       *conf.Server
	log        *log.Helper
}

func NewM2MAuthUseCase(repo M2MAuthRepo, privatePEM PrivatePEM, conf *conf.Server, logger log.Logger) (*M2MAuthUseCase, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &M2MAuthUseCase{
		repo:       repo,
		privateKey: privateKey,
		conf:       conf,
		log:        log.NewHelper(logger),
	}, nil
}

func (uc *M2MAuthUseCase) Login(ctx context.Context, req *M2MAuthRequest) (*M2MAuthResponse, error) {
	// 1. Get client by ID
	client, err := uc.repo.GetMachineClientByID(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client credentials")
	}

	// 2. Verify secret
	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(req.ClientSecret)); err != nil {
		return nil, fmt.Errorf("invalid client credentials")
	}

	// 3. Generate token
	now := time.Now().UTC()
	claims := CustomClaims{
		Role:  "service",
		Scope: strings.Join(client.Scopes, " "),
		Type:  "m2m",
		SpiceDB: SpiceDBClaim{
			Type: "client",
			ID:   client.ClientID,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   client.ClientID,
			Issuer:    IssuerName,
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
			Audience:  jwt.ClaimStrings{IssuerName}, // Ensure string array format
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = uc.conf.Kid

	signedToken, err := token.SignedString(uc.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &M2MAuthResponse{
		AccessToken: signedToken,
		ExpiresIn:   int64(AccessTokenTTL.Seconds()),
	}, nil
}
