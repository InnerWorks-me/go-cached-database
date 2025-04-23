package cacheddatabase

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"time"
)

var defaultConnectionTimeout = time.Second * 30
var defaultCacheTTL = 90 * 24 * time.Hour // 90 days

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type DatabaseAdapter interface {
	Ping() error
	Close() error
	Migrate(ctx context.Context, config MigrationConfig) error
}

type Adapter[T any] struct {
	redisClient *redis.Client
	psqlConn    *pgxpool.Pool
	connTimeout time.Duration
	cacheTTL    time.Duration
	Queries     *T
}

func NewAdapter[T any](config Config[T]) (*Adapter[T], error) {
	// Default timeout
	var connectionTimeout = config.ConnectionTimeout
	if connectionTimeout == 0 {
		connectionTimeout = defaultConnectionTimeout
	}

	// Default TTL
	var cacheTTL = config.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = defaultCacheTTL
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Endpoint,
		Password: config.Redis.Password,
		DB:       config.Redis.Database,
	})

	connStr := fmt.Sprintf("postgresql://%s:%s@%s/%s",
		config.Postgres.User,
		config.Postgres.Password,
		config.Postgres.Endpoint,
		config.Postgres.Database,
	)

	ctx := context.Background()
	psqlConn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	adapter := &Adapter[T]{
		redisClient: redisClient,
		psqlConn:    psqlConn,
		connTimeout: connectionTimeout,
		cacheTTL:    cacheTTL,
		Queries:     config.QueryConstructor(psqlConn),
	}
	return adapter, adapter.Ping()
}

func (cda *Adapter[T]) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), cda.connTimeout)
	defer cancel()

	redisErr := cda.redisClient.Ping(ctx).Err()
	psqlErr := cda.psqlConn.Ping(ctx)

	if redisErr != nil && psqlErr != nil {
		return fmt.Errorf("redis connection error: %s. Postgres connection error %s",
			redisErr.Error(), psqlErr.Error(),
		)
	}

	if redisErr != nil {
		return redisErr
	}

	if psqlErr != nil {
		return psqlErr
	}

	return nil
}

func (cda *Adapter[T]) Close() error {
	cda.psqlConn.Close()
	return cda.redisClient.Close()
}
