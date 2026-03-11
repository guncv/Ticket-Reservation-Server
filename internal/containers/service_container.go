package containers

import (
	"github.com/guncv/ticket-reservation-server/internal/domain/user"
)

func (c *Container) ServiceProvider() {
	if err := c.Container.Provide(user.NewUserService); err != nil {
		c.Error = err
	}
}
