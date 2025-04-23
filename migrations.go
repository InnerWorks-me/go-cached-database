package cacheddatabase

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/stdlib"
)

func (cda *Adapter[T]) Migrate(ctx context.Context, config MigrationConfig) error {
	src := fmt.Sprintf("file://%s", config.MigrationsDir)

	stdDb := stdlib.OpenDBFromPool(cda.psqlConn)
	defer stdDb.Close()
	driver, err := postgres.WithInstance(stdDb, &postgres.Config{})
	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(
		src, cda.psqlConn.Config().ConnConfig.Database, driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
