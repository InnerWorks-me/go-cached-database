package cacheddatabase

import (
	"context"
	"github.com/InnerWorks-me/go-cached-database/queries_test"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrations(t *testing.T) {
	adapter := NewContainerBackedAdapter(t)
	defer adapter.Close()

	err := adapter.Migrate(context.Background(), MigrationConfig{
		MigrationsDir: "migrations_test",
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestMigrateInsertSelect(t *testing.T) {
	adapter := NewContainerBackedAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	err := adapter.Migrate(ctx, MigrationConfig{
		MigrationsDir: "migrations_test",
	})

	if err != nil {
		t.Fatal(err)
	}

	newAuthor, err := adapter.Queries.CreateAuthor(ctx, queries_test.CreateAuthorParams{
		Name: "Test1",
		Bio: pgtype.Text{
			String: "Test2",
			Valid:  true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	selectedAuthor, err := adapter.Queries.GetAuthor(ctx, newAuthor.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, newAuthor.Name, selectedAuthor.Name, "Must have the same name.")
}
