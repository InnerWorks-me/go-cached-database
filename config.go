package cacheddatabase

import (
	"time"
)

type RedisConfig struct {
	Endpoint string // hostname:port
	Password string
	Database int
}

type PostgresConfig struct {
	Endpoint string // hostname:port
	User     string
	Password string
	Database string
}

type MigrationConfig struct {
	MigrationsDir string
}

type Config[T any] struct {
	Redis             RedisConfig
	Postgres          PostgresConfig
	Migrations        *MigrationConfig
	ConnectionTimeout time.Duration
	QueryConstructor  func(DBTX) *T
	CacheTTL          time.Duration
}
