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
	data  *Data
	log   *log.Helper
	query *db.Queries
}

func NewUserRepo(data *Data, logger log.Logger) biz.AuthRepo {
	return &userRepo{
		data:  data,
		log:   log.NewHelper(logger),
		query: data.Query,
	}
}

func (r *userRepo) ExecTx(ctx context.Context, fn func(biz.AuthRepo) error) error {
	tx, err := r.data.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := db.New(tx)
	err = fn(&userRepo{
		data:  r.data,
		log:   r.log,
		query: q,
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *userRepo) InsertUser(ctx context.Context, in *db.User) (*db.User, error) {
	resp, err := r.query.InsertUser(ctx, db.InsertUserParams{
		Email:        in.Email,
		PasswordHash: in.PasswordHash,
	})
	if err == sql.ErrNoRows {
		return nil, biz.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	resp, err := r.query.GetUserByEmail(ctx, sql.NullString{String: email, Valid: true})
	if err == sql.ErrNoRows {
		return nil, biz.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *userRepo) UpdatePasswordHash(ctx context.Context, id string, passwordHash string) (*db.User, error) {
	resp, err := r.query.UpdatePasswordHash(ctx, db.UpdatePasswordHashParams{
		ID:           uuid.Must(uuid.Parse(id)),
		PasswordHash: sql.NullString{String: passwordHash, Valid: true},
	})
	if err == sql.ErrNoRows {
		return nil, biz.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id string) (*db.User, error) {
	resp, err := r.query.GetUserByID(ctx, uuid.Must(uuid.Parse(id)))
	if err == sql.ErrNoRows {
		return nil, biz.ErrUserNotFound
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
