package data

import (
	"context"
	"database/sql"
	"fmt"
	"server/internal/conf"
	db "server/internal/data/db/generated"
	"server/internal/util"
	"time"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewUserRepo, NewM2MAuthRepo, NewPermissionRepo, util.NewPrivatePEM, util.NewPublicPEM, NewTransactionManager)

// Data .
type Data struct {
	// TODO wrapped database client
	DB      *sql.DB
	Query   *db.Queries
	SpiceDB *authzed.Client
	Redis   *redis.Client
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	dbConn, err := sql.Open(
		c.Database.Driver,
		c.Database.Source,
	)
	if err != nil {
		return nil, nil, err
	}

	dbConn.SetMaxOpenConns(20)
	dbConn.SetMaxIdleConns(10)
	dbConn.SetConnMaxLifetime(time.Hour)

	if err := dbConn.Ping(); err != nil {
		return nil, nil, err
	}

	// Spicedb
	var spiceOpts []grpc.DialOption
	if c.Spicedb.GetInsecure() {
		// For insecure connections, WithInsecureBearerToken must be paired with
		// grpc.WithTransportCredentials(insecure.NewCredentials()) — order matters:
		// transport credentials must come first so gRPC knows plaintext is allowed.
		spiceOpts = append(spiceOpts,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpcutil.WithInsecureBearerToken(c.Spicedb.GetToken()),
		)
	} else {
		spiceOpts = append(spiceOpts,
			grpcutil.WithBearerToken(c.Spicedb.GetToken()),
		)
	}
	client, err := authzed.NewClient(c.Spicedb.GetEndpoint(), spiceOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to spicedb: %w", err)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	})

	// Check Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	cleanup := func() {
		log.Info("closing the data resources")
		_ = dbConn.Close()
		_ = client.Close()
		rdb.Close()
	}

	return &Data{
		DB:      dbConn,
		Query:   db.New(dbConn),
		SpiceDB: client,
		Redis:   rdb,
	}, cleanup, nil
}
