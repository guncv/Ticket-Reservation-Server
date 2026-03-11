package test

import (
	"context"
	"fmt"
	"sync"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	redisClient "github.com/guncv/ticket-reservation-server/internal/infra/redis"
	"github.com/guncv/ticket-reservation-server/internal/shared"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	globalRedisContainer *redis.RedisContainer
	globalRedisClient    redisClient.RedisClient
	globalRedisConfig    *config.Config
	globalRedisCtx       context.Context
	redisContainerOnce   sync.Once
	redisContainerErr    error
)

func NewTestRedisClient(cfg *config.Config) (redisClient.RedisClient, error) {
	redisContainerOnce.Do(func() {
		globalRedisCtx = context.Background()

		redisContainer, err := redis.Run(globalRedisCtx,
			shared.TestRedisImage,
			testcontainers.WithWaitStrategy(
				wait.ForLog(shared.TestRedisWaitLogMessage).
					WithStartupTimeout(shared.TestContainerStartupTimeout),
			),
		)
		if err != nil {
			redisContainerErr = fmt.Errorf("failed to start redis test container: %w", err)
			return
		}
		globalRedisContainer = redisContainer

		host, err := redisContainer.Host(globalRedisCtx)
		if err != nil {
			redisContainerErr = fmt.Errorf("failed to get redis container host: %w", err)
			return
		}

		mappedPort, err := redisContainer.MappedPort(globalRedisCtx, "6379")
		if err != nil {
			redisContainerErr = fmt.Errorf("failed to get redis mapped port: %w", err)
			return
		}

		globalRedisConfig = &config.Config{
			AppConfig: cfg.AppConfig,
			RedisConfig: config.RedisConfig{
				Host:     host,
				Port:     mappedPort.Port(),
				Password: "",
				DB:       0,
			},
			DatabaseConfig: cfg.DatabaseConfig,
			PasswordConfig: cfg.PasswordConfig,
			TokenConfig:    cfg.TokenConfig,
			AuthConfig:     cfg.AuthConfig,
			OAuthConfig:    cfg.OAuthConfig,
		}

		// Create logger for test redis
		logger := log.NewLogger(globalRedisConfig)

		// Create redis client with test config
		globalRedisClient = redisClient.NewRedisClient(globalRedisConfig, logger)
	})

	if redisContainerErr != nil {
		return nil, redisContainerErr
	}

	return globalRedisClient, nil
}

var (
	redisCleanupOnce sync.Once
)

func CleanupTestRedisContainer() error {
	var cleanupErr error
	redisCleanupOnce.Do(func() {
		if globalRedisContainer != nil {
			cleanupErr = globalRedisContainer.Terminate(globalRedisCtx)
		}
	})
	return cleanupErr
}

func FlushRedis() error {
	if globalRedisClient == nil {
		return fmt.Errorf("test redis not initialized")
	}

	return globalRedisClient.FlushAll(globalRedisCtx)
}
