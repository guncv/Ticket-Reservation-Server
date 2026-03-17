package containers

import (
	"github.com/guncv/ticket-reservation-server/internal/api"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/guncv/ticket-reservation-server/internal/infra/http"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/infra/redis"
	"github.com/guncv/ticket-reservation-server/internal/infra/test"
	"github.com/guncv/ticket-reservation-server/internal/service/user/token"
	"github.com/guncv/ticket-reservation-server/internal/shared"
)

func (c *Container) InfrastructureProvider() {
	if err := c.Container.Provide(func(cfg *config.Config) (*db.PgPool, error) {
		if cfg.AppConfig.AppEnv == shared.AppEnvTest {
			return test.NewTestPgPool(cfg)
		}
		return db.NewPgPool(cfg)
	}); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(log.NewLogger); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(token.NewToken); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(func(cfg *config.Config, logger log.Logger) (redis.RedisClient, error) {
		if cfg.AppConfig.AppEnv == shared.AppEnvTest {
			return test.NewTestRedisClient(cfg)
		}
		return redis.NewRedisClient(cfg, logger), nil
	}); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(http.NewCookies); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(api.NewAuthMiddleware); err != nil {
		c.Error = err
	}
}
