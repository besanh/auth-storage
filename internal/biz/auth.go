package biz

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	db "server/internal/data/db/generated"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/besanh/go-library/spicedb"

	"github.com/golang-jwt/jwt/v5"

	"golang.org/x/crypto/bcrypt"
)

// PrivatePEM and PublicPEM are named types so wire can distinguish them.
type PrivatePEM []byte
type PublicPEM []byte

type AuthRepo interface {
	InsertUser(ctx context.Context, user *db.User) (*db.User, error)
	GetUserByEmail(ctx context.Context, email string) (*db.User, error)
	UpdatePasswordHash(ctx context.Context, id string, passwordHash string) (*db.User, error)
	GetUserByID(ctx context.Context, id string) (*db.User, error)
	ExecTx(ctx context.Context, fn func(AuthRepo) error) error
}

type AuthUseCase struct {
	authRepo    AuthRepo
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
	spiceClient *spicedb.SpiceClient
}

func NewAuthUseCase(authRepo AuthRepo, privatePEM PrivatePEM, publicPEM PublicPEM, spiceClient *spicedb.SpiceClient) (*AuthUseCase, error) {
	// 1. Parse the Private Key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// 2. Parse the Public Key
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &AuthUseCase{
		authRepo:    authRepo,
		privateKey:  privateKey,
		publicKey:   publicKey,
		spiceClient: spiceClient,
	}, nil
}

func (uc *AuthUseCase) Register(ctx context.Context, email, password string) (string, string, int64, string, error) {
	// 1. Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", 0, "", err
	}

	var user *db.User
	// 2 & 3. Save to database and Provision user in SpiceDB within a transaction
	err = uc.authRepo.ExecTx(ctx, func(repo AuthRepo) error {
		var err error
		user, err = repo.InsertUser(ctx, &db.User{
			Email:        sql.NullString{String: email, Valid: true},
			PasswordHash: sql.NullString{String: string(hash), Valid: true},
		})
		if err != nil {
			return err
		}

		// 3. Provision the user in SpiceDB (AuthZ)
		request := &pb.WriteRelationshipsRequest{
			Updates: []*pb.RelationshipUpdate{
				{
					Operation: pb.RelationshipUpdate_OPERATION_TOUCH,
					Relationship: &pb.Relationship{
						Resource: &pb.ObjectReference{ObjectType: "platform", ObjectId: "global"},
						Relation: "member",
						Subject: &pb.SubjectReference{
							Object: &pb.ObjectReference{ObjectType: "user", ObjectId: user.ID.String()},
						},
					},
				},
			},
		}

		_, err = uc.spiceClient.Client.PermissionsServiceClient.WriteRelationships(ctx, request)
		if err != nil {
			return fmt.Errorf("failed to provision user in spicedb: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", "", 0, "", err
	}

	// 4. Generate JWT tokens
	accessToken, refreshToken, expiresIn, err := uc.GenerateToken(user.ID.String())
	if err != nil {
		return "", "", 0, "", err
	}
	return accessToken, refreshToken, expiresIn, user.ID.String(), nil
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (string, string, int64, string, error) {
	// 1. Get user by email
	user, err := uc.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", 0, "", err
	}

	// 2. Verify password
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password)); err != nil {
		return "", "", 0, "", fmt.Errorf("invalid password")
	}

	// 3. Generate JWT tokens
	accessToken, refreshToken, expiresIn, err := uc.GenerateToken(user.ID.String())
	if err != nil {
		return "", "", 0, "", err
	}
	return accessToken, refreshToken, expiresIn, user.ID.String(), nil
}
