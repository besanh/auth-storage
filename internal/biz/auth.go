package biz

import (
	"context"
	"crypto/rsa"
	"fmt"
	"server/internal/conf"
	db "server/internal/data/db/generated"
	"time"

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

	BlacklistToken(ctx context.Context, token string, expiresIn time.Duration) error
	IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

type AuthUseCase struct {
	authRepo   AuthRepo
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	permRepo   PermissionRepo
	conf       *conf.Server
	tm         Transaction
}

var ErrInvalidToken = fmt.Errorf("invalid or expired refresh token")

func NewAuthUseCase(authRepo AuthRepo, privatePEM PrivatePEM, publicPEM PublicPEM, permRepo PermissionRepo, conf *conf.Server, tm Transaction) (*AuthUseCase, error) {
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
		authRepo:   authRepo,
		privateKey: privateKey,
		publicKey:  publicKey,
		permRepo:   permRepo,
		conf:       conf,
		tm:         tm,
	}, nil
}

func (uc *AuthUseCase) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 1. Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user *db.User
	// 2 & 3. Save to database and Provision user in SpiceDB within a transaction
	if err = uc.tm.ExecTx(ctx, func(ctx context.Context) error {
		var err error
		user, err = uc.authRepo.InsertUser(ctx, &db.User{
			Email:        req.Email,
			PasswordHash: string(hash),
			Role:         "user",
			Scope:        "",
		})
		if err != nil {
			return err
		}

		// 3. Provision the user in SpiceDB (AuthZ)
		_, err = uc.permRepo.WriteRelationship(ctx, WriteRelationshipRequest{
			ResourceType: "platform",
			ResourceID:   "global",
			Relation:     "member",
			SubjectType:  "user",
			SubjectID:    user.ID.String(),
		})
		if err != nil {
			return fmt.Errorf("failed to provision user in spicedb: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		UserID: user.ID.String(),
	}, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 1. Get user by email
	user, err := uc.authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	// 2. Verify password
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	// 3. Generate JWT tokens
	tokenReply, err := uc.GenerateToken(&GenerateTokenRequest{
		UserID: user.ID.String(),
		Role:   user.Role,
		Scope:  user.Scope,
	})
	if err != nil {
		return nil, err
	}
	return &LoginResponse{
		AccessToken:  tokenReply.AccessToken,
		RefreshToken: tokenReply.RefreshToken,
		ExpiresIn:    tokenReply.ExpiresAt,
		UserID:       user.ID.String(),
	}, nil
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// 1. Validate the refresh token
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(t *jwt.Token) (any, error) {
		// 1. Verify it's actually an RSA signed token
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		// 2. Return the PUBLIC KEY to verify the signature
		return uc.publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// 2. Check if the token is blacklisted
	isBlacklisted, err := uc.authRepo.IsTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, ErrInvalidToken
	}

	// 3. Extract the User ID from the Subject claim
	userID := claims.Subject
	if userID == "" {
		return nil, ErrInvalidToken
	}

	// 4. Security Check:
	// Query the database to ensure the user still exists and hasn't been banned/deleted.
	// If they were banned, we want to reject the refresh attempt.
	user, err := uc.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// 5. Generate a brand new Access Token and a brand new Refresh Token (Token Rotation)
	tokenReply, err := uc.GenerateToken(&GenerateTokenRequest{
		UserID: user.ID.String(),
		Role:   user.Role,
		Scope:  user.Scope,
	})
	if err != nil {
		return nil, err
	}
	return &RefreshTokenResponse{
		AccessToken:  tokenReply.AccessToken,
		RefreshToken: tokenReply.RefreshToken,
		ExpiresIn:    tokenReply.ExpiresAt,
		UserID:       user.ID.String(),
	}, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	// 1. Parse the token (We don't need to verify the signature here
	// if your Middleware already verified it, but we parse it to get the claims)
	claims := &CustomClaims{}
	parser := jwt.NewParser() // Use an unverified parser just to read the claims safely

	_, _, err := parser.ParseUnverified(req.RefreshToken, claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token for logout: %w", err)
	}

	// 2. Extract the unique Token ID we generated earlier
	tokenID := claims.ID
	if tokenID == "" {
		return nil, fmt.Errorf("token does not have an ID")
	}

	// 3. Calculate how much time is left until the token naturally expires
	expirationTime := claims.ExpiresAt.Time
	timeLeft := expirationTime.Sub(time.Now().UTC())

	// 4. If it's already expired, we don't need to blacklist it!
	if timeLeft <= 0 {
		return nil, fmt.Errorf("token is already expired")
	}

	// 5. Send it to your Redis repository to be blacklisted
	if err := uc.authRepo.BlacklistToken(ctx, tokenID, timeLeft); err != nil {
		return nil, fmt.Errorf("failed to blacklist token: %w", err)
	}

	return &LogoutResponse{
		Success: true,
	}, nil
}

func (uc *AuthUseCase) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*CheckPermissionResponse, error) {
	resp, err := uc.permRepo.CheckPermission(ctx, CheckPermissionRequest{
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Relation:     req.Relation,
		SubjectType:  req.SubjectType,
		SubjectID:    req.SubjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check permission in spicedb: %w", err)
	}

	return &CheckPermissionResponse{
		Allowed: resp.Allowed,
	}, nil
}

func (uc *AuthUseCase) GetUser(ctx context.Context, id string) (*GetUserResponse, error) {
	user, err := uc.authRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &GetUserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Role:      user.Role,
		Scope:     user.Scope,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}
