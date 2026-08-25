package storage

import (
	"context"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kyotomin/ab-router/migrations"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(connString, migrateConnString string, retries int) (*PostgresStorage, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(
		ctx,
		connString,
	)

	if err != nil {
		slog.Error(
			"error launching postgres database",
			"error", err,
		)
		return nil, err
	}

	retriesCount := 0
	for i := 0; i <= retries; i++ {
		if err := pool.Ping(ctx); err != nil {
			retriesCount++
		}
	}

	if retriesCount >= 5 {
		slog.Error(
			"error connecting to postgres database",
			"error", err,
		)
		return nil, err
	}

	store := &PostgresStorage{pool: pool}

	if err := run_migrations(migrateConnString); err != nil {
		return nil, err
	}

	return store, nil
}

func run_migrations(migrateConnString string) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		slog.Error(
			"error finding migrations directory destination",
			"error", err,
		)
		return err
	}

	mig, err := migrate.NewWithSourceInstance("iofs", src, migrateConnString)
	if err != nil {
		return err
	}

	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error(
			"error applying migrations",
			"error", err,
		)
		return err
	}

	return nil
}
