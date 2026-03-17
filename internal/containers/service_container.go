package containers

import (
	"github.com/guncv/ticket-reservation-server/internal/service/event"
	"github.com/guncv/ticket-reservation-server/internal/service/user"
)

func (c *Container) ServiceProvider() {
	if err := c.Container.Provide(user.NewUserService); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(event.NewEventService); err != nil {
		c.Error = err
	}
}
