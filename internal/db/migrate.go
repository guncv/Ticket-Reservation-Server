package db

import (
	"embed"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/guncv/ticket-reservation-server/internal/config"
)

type Migrate struct {
	*dbmate.DB
}

var (
	//go:embed migrations/*.sql
	fs         embed.FS
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func LoadMigrate(cfg *config.Config) (*Migrate, error) {
	connStr, err := connStr(cfg.DatabaseConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot create new migrations: %w", err)
	}
	u, err := url.Parse(connStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse connection string: %w", err)
	}

	db := dbmate.New(u)
	db.FS = fs
	db.MigrationsDir = []string{"migrations"}
	db.SchemaFile = filepath.Join(basepath, "schema.sql")
	db.AutoDumpSchema = true

	return &Migrate{db}, nil
}

func (m *Migrate) NewMigrate(name string) error {
	// Save original settings
	originalMigrationsDir := m.MigrationsDir
	originalFS := m.FS

	// Temporarily set to absolute path and disable embedded FS for file creation
	m.MigrationsDir = []string{filepath.Join(basepath, "migrations")}
	m.FS = nil

	// Create migration
	err := m.NewMigration(name)

	// Restore original settings
	m.MigrationsDir = originalMigrationsDir
	m.FS = originalFS

	if err != nil {
		return err
	}

	return nil
}

func (m *Migrate) MigrateUp() error {
	if err := m.CreateAndMigrate(); err != nil {
		return err
	}

	return nil
}

func (m *Migrate) MigrateDown() error {
	if err := m.Rollback(); err != nil {
		return err
	}

	return nil
}

func (m *Migrate) MigrateDownAll() error {
	for {
		if err := m.Rollback(); err != nil {
			if errors.Is(err, dbmate.ErrNoRollback) {
				break
			}
			return fmt.Errorf("cannot rollback migration: %w", err)
		}
	}

	return nil
}
