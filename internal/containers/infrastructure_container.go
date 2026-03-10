package containers

import (
	"database/sql"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/infras/database"
	"github.com/guncv/ticket-reservation-server/internal/infras/log"
	"github.com/guncv/ticket-reservation-server/internal/infras/server"
	"gorm.io/gorm"
)

func (c *Container) InfrastructureProvider() {

	c.Container.Provide(func(cfg *config.Config) *log.Logger {
		return log.Initialize(cfg.AppConfig.AppEnv)
	})

	if err := c.Container.Provide(func(cfg *config.Config) *server.GinServer {
		return server.NewGinServer(cfg, c.Container)
	}); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(database.ConnectPostgres); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(func(conn *database.DBConnections) *gorm.DB {
		return conn.GormDB
	}); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(func(conn *database.DBConnections) *sql.DB {
		return conn.SqlDB
	}); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(database.NewRedisClient); err != nil {
		c.Error = err
	}

}
