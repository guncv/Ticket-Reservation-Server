package containers_test

import (
	"context"
	"os"
	"testing"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/containers"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/guncv/ticket-reservation-server/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainer_TestEnv(t *testing.T) {
	os.Setenv("APP_ENV", shared.AppEnvTest)
	defer os.Unsetenv("APP_ENV")

	cfg, err := config.LoadConfig(nil)
	require.NoError(t, err)
	require.Equal(t, shared.AppEnvTest, cfg.AppConfig.AppEnv)

	container := containers.NewContainer(cfg)
	require.NoError(t, container.Error)

	var pool *db.PgPool
	err = container.Container.Invoke(func(p *db.PgPool) {
		pool = p
	})
	require.NoError(t, err)
	require.NotNil(t, pool)

	ctx := context.Background()

	err = pool.Pool.Ping(ctx)
	require.NoError(t, err)

	var result int
	err = pool.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)

	var dbName string
	err = pool.Pool.QueryRow(ctx, "SELECT current_database()").Scan(&dbName)
	require.NoError(t, err)
	assert.Equal(t, shared.TestDatabaseName, dbName)
}

func TestContainer_DevEnv(t *testing.T) {
	cleanup := setupTestDevEnv()
	defer cleanup()

	cfg, err := config.LoadConfig(nil)
	require.NoError(t, err)
	require.Equal(t, shared.AppEnvDev, cfg.AppConfig.AppEnv)

	container := containers.NewContainer(cfg)
	require.NoError(t, container.Error)

	var pool *db.PgPool
	err = container.Container.Invoke(func(p *db.PgPool) {
		pool = p
	})
	require.NoError(t, err)
	require.NotNil(t, pool)
	ctx := context.Background()

	var dbName string
	err = pool.Pool.QueryRow(ctx, "SELECT current_database()").Scan(&dbName)
	require.NoError(t, err)
	assert.Equal(t, config.DefaultPostgresApplicationName, dbName)
}

func setupTestDevEnv() func() {
	os.Setenv("APP_ENV", shared.AppEnvDev)
	os.Setenv("TOKEN_SECRET_KEY", shared.TestTokenSecretKey)

	return func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("TOKEN_SECRET_KEY")
	}
}
