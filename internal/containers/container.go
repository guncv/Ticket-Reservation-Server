package containers

import (
	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/api"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"go.uber.org/dig"
)

type Container struct {
	cfg       *config.Config
	Container *dig.Container
	Error     error
}

func (c *Container) configure() {
	c.Container = dig.New()

	c.Container.Provide(func() *config.Config {
		return c.cfg
	})
	c.RepositoryProvider()
	c.InfrastructureProvider()
	c.ServiceProvider()
	c.HandlerProvider()
}

func (c *Container) Run() *Container {
	router := gin.Default()

	api.RegisterRoutes(router, c.Container, c.cfg)

	port := c.cfg.AppConfig.AppPort
	if port == "" {
		panic("port is empty")
	}

	if err := router.Run(":" + port); err != nil {
		c.Error = err
	}

	return c
}

func NewContainer(cfg *config.Config) *Container {
	c := &Container{
		cfg: cfg,
	}
	c.configure()

	return c
}
