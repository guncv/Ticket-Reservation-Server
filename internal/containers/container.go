package containers

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/api"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/infra/job"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/service/event"
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
	// Start pprof server for profiling (CPU, memory, goroutines)
	go func() {
		http.ListenAndServe(":6060", nil)
	}()

	router := gin.Default()

	api.RegisterRoutes(router, c.Container, c.cfg)

	if err := c.Container.Invoke(func(eventService event.EventService, logger log.Logger) {
		ticketCounterJob := job.NewTicketCounterJob(eventService, logger, 3*time.Second)
		ticketCounterJob.Start(context.Background())
	}); err != nil {
		c.Error = err
		return c
	}

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
