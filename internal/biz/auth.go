package biz

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"server/internal/conf"
	db "server/internal/data/db/generated"
	"time"

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

	BlacklistToken(ctx context.Context, token string, expiresIn time.Duration) error
	IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

type AuthUseCase struct {
	authRepo    AuthRepo
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
	spiceClient *spicedb.SpiceClient
	conf        *conf.Server
}

var ErrInvalidToken = fmt.Errorf("invalid or expired refresh token")

func NewAuthUseCase(authRepo AuthRepo, privatePEM PrivatePEM, publicPEM PublicPEM, spiceClient *spicedb.SpiceClient, conf *conf.Server) (*AuthUseCase, error) {
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
		conf:        conf,
	}, nil
}

func (uc *AuthUseCase) Register(ctx context.Context, email, password string) (string, error) {
	// 1. Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
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
		return "", err
	}

	return user.ID.String(), nil
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

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, int64, string, error) {
	// 1. Validate the refresh token
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (any, error) {
		// 1. Verify it's actually an RSA signed token
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		// 2. Return the PUBLIC KEY to verify the signature
		return uc.publicKey, nil
	})
	if err != nil || !token.Valid {
		return "", "", 0, "", err
	}

	if !token.Valid {
		return "", "", 0, "", ErrInvalidToken
	}

	// 2. Check if the token is blacklisted
	isBlacklisted, err := uc.authRepo.IsTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return "", "", 0, "", err
	}
	if isBlacklisted {
		return "", "", 0, "", ErrInvalidToken
	}

	// 3. Extract the User ID from the Subject claim
	userID := claims.Subject
	if userID == "" {
		return "", "", 0, "", ErrInvalidToken
	}

	// 4. Security Check:
	// Query the database to ensure the user still exists and hasn't been banned/deleted.
	// If they were banned, we want to reject the refresh attempt.
	user, err := uc.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", 0, "", ErrInvalidToken
	}

	// 5. Generate a brand new Access Token and a brand new Refresh Token (Token Rotation)
	accessToken, refreshToken, expiresIn, err := uc.GenerateToken(user.ID.String())
	if err != nil {
		return "", "", 0, "", err
	}
	return accessToken, refreshToken, expiresIn, user.ID.String(), nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, tokenString string) error {
	// 1. Parse the token (We don't need to verify the signature here
	// if your Middleware already verified it, but we parse it to get the claims)
	claims := &jwt.RegisteredClaims{}
	parser := jwt.NewParser() // Use an unverified parser just to read the claims safely

	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return fmt.Errorf("failed to parse token for logout: %w", err)
	}

	// 2. Extract the unique Token ID we generated earlier
	tokenID := claims.ID
	if tokenID == "" {
		return fmt.Errorf("token does not have an ID")
	}

	// 3. Calculate how much time is left until the token naturally expires
	expirationTime := claims.ExpiresAt.Time
	timeLeft := expirationTime.Sub(time.Now().UTC())

	// 4. If it's already expired, we don't need to blacklist it!
	if timeLeft <= 0 {
		return nil
	}

	// 5. Send it to your Redis repository to be blacklisted
	if err := uc.authRepo.BlacklistToken(ctx, tokenID, timeLeft); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}
