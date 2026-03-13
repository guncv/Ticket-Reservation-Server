package test

import (
	"context"
	"fmt"
	"sync"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/guncv/ticket-reservation-server/internal/shared"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	globalContainer *postgres.PostgresContainer
	globalPool      *db.PgPool
	globalConfig    *config.Config
	globalCtx       context.Context
	containerOnce   sync.Once
	containerErr    error
)

func NewTestPgPool(cfg *config.Config) (*db.PgPool, error) {
	containerOnce.Do(func() {
		globalCtx = context.Background()

		pgContainer, err := postgres.Run(globalCtx,
			shared.TestPostgresImage,
			postgres.WithDatabase(shared.TestDatabaseName),
			postgres.WithUsername(shared.TestDatabaseUser),
			postgres.WithPassword(shared.TestDatabasePass),
			testcontainers.WithWaitStrategy(
				wait.ForLog(shared.TestContainerWaitLogMessage).
					WithOccurrence(shared.TestContainerWaitOccurrence).
					WithStartupTimeout(shared.TestContainerStartupTimeout),
			),
		)
		if err != nil {
			containerErr = fmt.Errorf("failed to start test container: %w", err)
			return
		}
		globalContainer = pgContainer

		connStr, err := pgContainer.ConnectionString(globalCtx, "sslmode="+shared.TestSSLMode)
		if err != nil {
			containerErr = fmt.Errorf("failed to get connection string: %w", err)
			return
		}

		host, err := pgContainer.Host(globalCtx)
		if err != nil {
			containerErr = fmt.Errorf("failed to get container host: %w", err)
			return
		}
		mappedPort, err := pgContainer.MappedPort(globalCtx, "5432")
		if err != nil {
			containerErr = fmt.Errorf("failed to get mapped port: %w", err)
			return
		}

		globalConfig = &config.Config{
			AppConfig: cfg.AppConfig,
			DatabaseConfig: config.DatabaseConfig{
				Host:                  host,
				Port:                  mappedPort.Port(),
				User:                  shared.TestDatabaseUser,
				Password:              shared.TestDatabasePass,
				DbName:                shared.TestDatabaseName,
				ApplicationName:       shared.TestAppName,
				SSLMode:               shared.TestSSLMode,
				ConnectTimeout:        shared.TestConnectTimeout,
				MaxOpenConns:          shared.TestMaxOpenConns,
				MaxIdleConns:          shared.TestMaxIdleConns,
				ConnMaxLifetime:       shared.TestConnMaxLifetime,
				ConnMaxLifetimeJitter: shared.TestConnMaxLifetimeJitter,
				ConnMaxIdleTime:       shared.TestConnMaxIdleTime,
				HealthCheckPeriod:     shared.TestHealthCheckPeriod,
				EventTimeout:          shared.TestEventTimeout,
			},
		}

		poolConfig, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			containerErr = fmt.Errorf("failed to parse connection string: %w", err)
			return
		}

		poolConfig.MaxConns = int32(globalConfig.DatabaseConfig.MaxOpenConns)
		poolConfig.MinConns = int32(globalConfig.DatabaseConfig.MaxIdleConns)
		poolConfig.MaxConnLifetime = globalConfig.DatabaseConfig.ConnMaxLifetime
		poolConfig.MaxConnIdleTime = globalConfig.DatabaseConfig.ConnMaxIdleTime
		poolConfig.HealthCheckPeriod = globalConfig.DatabaseConfig.HealthCheckPeriod

		pool, err := pgxpool.NewWithConfig(globalCtx, poolConfig)
		if err != nil {
			containerErr = fmt.Errorf("failed to create connection pool: %w", err)
			return
		}

		if err := pool.Ping(globalCtx); err != nil {
			containerErr = fmt.Errorf("failed to ping database: %w", err)
			return
		}

		globalPool = &db.PgPool{Pool: pool}

		if err := runMigrations(globalConfig); err != nil {
			containerErr = fmt.Errorf("failed to run migrations: %w", err)
			return
		}
	})

	if containerErr != nil {
		return nil, containerErr
	}

	return globalPool, nil
}

var (
	cleanupOnce sync.Once
)

func CleanupTestContainer() error {
	var cleanupErr error
	cleanupOnce.Do(func() {
		if globalPool != nil && globalPool.Pool != nil {
			globalPool.Pool.Close()
		}

		if globalContainer != nil {
			cleanupErr = globalContainer.Terminate(globalCtx)
		}
	})
	return cleanupErr
}

func TruncateAllTables() error {
	if globalPool == nil {
		return fmt.Errorf("test database not initialized")
	}

	_, _ = globalPool.Pool.Exec(globalCtx, `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = current_database()
			AND pid <> pg_backend_pid()
			AND state = 'idle in transaction';
	`)

	sql := `TRUNCATE TABLE users, sessions RESTART IDENTITY CASCADE;`

	_, err := globalPool.Pool.Exec(globalCtx, sql)
	return err
}

func runMigrations(testConfig *config.Config) error {
	migrate, err := db.LoadMigrate(testConfig)
	if err != nil {
		return fmt.Errorf("failed to load migrate: %w", err)
	}

	if err := migrate.MigrateUp(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
