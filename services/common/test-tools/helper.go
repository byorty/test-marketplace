package testtools

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func NewTestDB(t *testing.T) *bun.DB {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("marketplace"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	testcontainers.CleanupContainer(t, container)

	connStr, err := container.ConnectionString(
		ctx,
		"sslmode=disable",
	)
	require.NoError(t, err)

	sqlDB, err := sql.Open("pg", connStr)
	require.NoError(t, err)

	db := bun.NewDB(
		sqlDB,
		pgdialect.New(),
	)

	require.NoError(t, db.Ping())

	runMigrations(t, connStr)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func runMigrations(t *testing.T, connStr string) {
	t.Helper()

	migrationsPath, err := filepath.Abs(
		"../../../../migrations",
	)
	require.NoError(t, err)

	m, err := migrate.New(
		"file://"+migrationsPath,
		connStr,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		srcErr, dbErr := m.Close()
		require.NoError(t, srcErr)
		require.NoError(t, dbErr)
	})

	err = m.Up()

	if errors.Is(err, migrate.ErrNoChange) {
		return
	}

	require.NoError(t, err)
}