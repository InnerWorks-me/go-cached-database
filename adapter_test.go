package cacheddatabase

import (
	"context"
	"github.com/Innerworks-me/go-cached-database/queries_test"
	"github.com/testcontainers/testcontainers-go"
	"testing"
)

type ContainerBackedAdapter struct {
	Adapter[queries_test.Queries]
	t         *testing.T
	RedisC    testcontainers.Container
	PostgresC testcontainers.Container
}

func NewContainerBackedAdapter(t *testing.T) *ContainerBackedAdapter {
	ctx := context.Background()
	// Set up Redis Container
	redisC, err := redisContainer(ctx)
	defer testcontainers.CleanupContainer(t, redisC)
	if err != nil {
		t.Fatal(err)
	}

	// Set up Postgres
	postgresC, err := postgresContainer(ctx)
	defer testcontainers.CleanupContainer(t, postgresC)
	if err != nil {
		t.Fatal(err)
	}

	redisEndpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	postgresEndpoint, err := postgresC.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	adapter, err := NewAdapter[queries_test.Queries](Config[queries_test.Queries]{
		Redis: RedisConfig{
			Endpoint: redisEndpoint,
			Password: "",
			Database: 0,
		},
		Postgres: PostgresConfig{
			Endpoint: postgresEndpoint,
			User:     "postgres",
			Password: "postgres",
			Database: "postgres",
		},
		QueryConstructor: func(db DBTX) *queries_test.Queries {
			return queries_test.New(db)
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	return &ContainerBackedAdapter{
		Adapter:   *adapter,
		RedisC:    redisC,
		PostgresC: postgresC,
	}
}

func TestNewAdapterPing(t *testing.T) {
	adapter := NewContainerBackedAdapter(t)
	defer adapter.Close()

	err := adapter.Ping()
	if err != nil {
		t.Fatal(err)
	}
}
