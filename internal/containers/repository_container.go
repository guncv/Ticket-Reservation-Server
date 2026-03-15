package containers

import (
	eventRepo "github.com/guncv/ticket-reservation-server/internal/service/event/repo"
	userRepo "github.com/guncv/ticket-reservation-server/internal/service/user/repo"
)

func (c *Container) RepositoryProvider() {
	if err := c.Container.Provide(userRepo.NewUserRepository); err != nil {
		c.Error = err
	}

	if err := c.Container.Provide(eventRepo.NewEventRepository); err != nil {
		c.Error = err
	}
}
