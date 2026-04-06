package data

import (
	"context"
	"database/sql"
	"fmt"
	"server/internal/biz"
	db "server/internal/data/db/generated"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

var ErrUserNotFound = fmt.Errorf("user not found")

func NewUserRepo(data *Data, logger log.Logger) biz.AuthRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userRepo) InsertUser(ctx context.Context, in *db.User) (*db.User, error) {
	resp, err := r.data.Query.InsertUser(ctx, db.InsertUserParams{
		Email:        in.Email,
		PasswordHash: in.PasswordHash,
		Role:         in.Role,
		Scope:        in.Scope,
		Status:       in.Status,
	})
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	resp, err := r.data.Query.GetUserByEmail(ctx, sql.NullString{String: email, Valid: true})
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *userRepo) UpdatePasswordHash(ctx context.Context, id string, passwordHash string) (*db.User, error) {
	resp, err := r.data.Query.UpdatePasswordHash(ctx, db.UpdatePasswordHashParams{
		ID:           uuid.Must(uuid.Parse(id)),
		PasswordHash: sql.NullString{String: passwordHash, Valid: true},
	})
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id string) (*db.User, error) {
	resp, err := r.data.Query.GetUserByID(ctx, uuid.Must(uuid.Parse(id)))
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *userRepo) BlacklistToken(ctx context.Context, tokenID string, expiresIn time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	return r.data.Redis.Set(ctx, key, tokenID, expiresIn).Err()
}

func (r *userRepo) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	_, err := r.data.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
