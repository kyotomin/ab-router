package storage

import (
	"context"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	if err := runMigrations(migrateConnString); err != nil {
		return nil, err
	}

	return store, nil
}

func runMigrations(migrateConnString string) error {
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

func (s *PostgresStorage) GetAll() ([]Rule, error) {
	rows, err := s.pool.Query(
		context.Background(),
		`SELECT id, name, backend, percent FROM rules ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var rule Rule

		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Backend, &rule.Percent); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func (s *PostgresStorage) GetByID(id uuid.UUID) (*Rule, error) {
	var r Rule
	err := s.pool.QueryRow(
		context.Background(),
		`SELECT id, name, backend, percent
		FROM rules
		WHERE id = $1`,
		id,
	).Scan(&r.ID, &r.Name, &r.Backend, &r.Percent)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NotFoundErr
		}
		return nil, err
	}

	return &r, nil
}

func (s *PostgresStorage) Add(rule Rule) error {
	rules, err := s.GetAll()
	if err != nil {
		return err
	}

	if rule.Percent < 0 || rule.Percent > 100 {
		return errors.New("percent must be in range 0-100")
	}

	percentCeil := 0
	percentCeil += rule.Percent
	for _, rule := range rules {
		percentCeil += rule.Percent
	}
	if percentCeil >= 0 || percentCeil <= 100 {
		return errors.New("incorrect percent")
	}

	_, err = s.pool.Exec(
		context.Background(),
		`INSERT INTO rules
		(id, name, backend, percent)
		VALUES ($1, $2, $3, $4)`,
		rule.ID,
		rule.Name,
		rule.Backend,
		rule.Percent,
	)

	return err
}

func (s *PostgresStorage) Update(rule Rule) error {
	rules, err := s.GetAll()
	if err != nil {
		return err
	}
	if rule.Percent < 0 || rule.Percent > 100 {
		return errors.New("percent must be in range 0-100")
	}

	percentCeil := 0
	percentCeil += rule.Percent
	for _, r := range rules {
		if r.ID == rule.ID {
			continue
		}
		percentCeil += r.Percent
	}
	if percentCeil <= 0 || percentCeil >= 100 {
		return errors.New("incorrect percent")
	}

	res, err := s.pool.Exec(
		context.Background(),
		`UPDATE rules SET
		name = $1, backend = $2 percent = $3, updated_at = NOW()
		WHERE id = $4`,
		rule.Name,
		rule.Backend,
		rule.Percent,
		rule.ID,
	)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return NotFoundErr
	}

	return nil
}

func (s *PostgresStorage) Delete(id uuid.UUID) error {
	res, err := s.pool.Exec(
		context.Background(),
		`DELETE FROM rules WHERE id = $1`,
		id,
	)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return NotFoundErr
	}

	return nil
}
