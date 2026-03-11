package containers

import "github.com/guncv/ticket-reservation-server/internal/api/handlers"

func (c *Container) HandlerProvider() {
	if err := c.Container.Provide(handlers.NewUserHandler); err != nil {
		c.Error = err
	}
}
