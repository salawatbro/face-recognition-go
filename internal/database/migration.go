package database

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
)

// Migrator handles database migrations
type Migrator struct {
	migrate *migrate.Migrate
}

// NewMigrator creates a new Migrator instance
func NewMigrator(dbType DatabaseType, connectionString string) (*Migrator, error) {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create migration source: %w", err)
	}

	dbURL := buildDatabaseURL(dbType, connectionString)

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	if err := m.migrate.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("migration up failed: %w", err)
	}
	return nil
}

// Down rolls back all migrations
func (m *Migrator) Down() error {
	if err := m.migrate.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("migration down failed: %w", err)
	}
	return nil
}

// Steps runs n migrations (positive = up, negative = down)
func (m *Migrator) Steps(n int) error {
	if err := m.migrate.Steps(n); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("migration steps failed: %w", err)
	}
	return nil
}

// Version returns the current migration version
func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

// Close closes the migrator
func (m *Migrator) Close() error {
	srcErr, dbErr := m.migrate.Close()
	if srcErr != nil {
		return srcErr
	}
	return dbErr
}

// buildDatabaseURL constructs the database URL for migrations
func buildDatabaseURL(dbType DatabaseType, connectionString string) string {
	if dbType == DatabaseTypeSQLite {
		return fmt.Sprintf("sqlite://%s", connectionString)
	}

	dbURL := connectionString
	if !strings.HasPrefix(dbURL, "postgres://") {
		dbURL = "postgres://" + dbURL
	}
	return dbURL
}
