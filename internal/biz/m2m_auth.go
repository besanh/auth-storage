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

func (uc *M2MAuthUseCase) Login(ctx context.Context, clientID, clientSecret string) (string, int64, error) {
	// 1. Get client by ID
	client, err := uc.repo.GetMachineClientByID(ctx, clientID)
	if err != nil {
		return "", 0, fmt.Errorf("invalid client credentials")
	}

	// 2. Verify secret
	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(clientSecret)); err != nil {
		return "", 0, fmt.Errorf("invalid client credentials")
	}

	// 3. Generate token
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":   client.ClientID,
		"iss":   IssuerName,
		"exp":   now.Add(AccessTokenTTL).Unix(),
		"iat":   now.Unix(),
		"jti":   uuid.New().String(),
		"aud":   []string{"auth-service"},
		"scope": strings.Join(client.Scopes, " "),
		"role":  "service",
		"type":  "m2m",
		"spicedb": map[string]string{
			"type": "client",
			"id":   client.ClientID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = uc.conf.Kid

	signedToken, err := token.SignedString(uc.privateKey)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, int64(AccessTokenTTL.Seconds()), nil
}
