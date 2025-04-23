package cacheddatabase

import (
	"context"
	"fmt"
	"github.com/Innerworks-me/go-cached-database/queries_test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMigrateInsertSelectWithCache(t *testing.T) {
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
		Name: "Isaac Newton",
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Inserted new Author with ID: %d\n", newAuthor.ID)

	var dbCallCount = 0
	cacheKey := fmt.Sprintf("author:%d", newAuthor.ID)
	cachedAuthor1, err := WithCache(&adapter.Adapter, cacheKey, func() (queries_test.Author, error) {
		dbCallCount++
		return adapter.Queries.GetAuthor(ctx, newAuthor.ID)
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait for Redis SET to complete
	time.Sleep(20 * time.Millisecond)

	cachedAuthor2, err := WithCache(&adapter.Adapter, cacheKey, func() (queries_test.Author, error) {
		dbCallCount++
		return adapter.Queries.GetAuthor(ctx, newAuthor.ID)
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, dbCallCount, 1, "call count should be 1")
	assert.Equal(t, newAuthor.Name, cachedAuthor1.Name, "Must have the same name.")
	assert.Equal(t, newAuthor.Name, cachedAuthor2.Name, "Must have the same name.")
}
