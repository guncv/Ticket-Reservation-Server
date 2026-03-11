package containers

import (
	userRepo "github.com/guncv/ticket-reservation-server/internal/domain/user/repo"
)

func (c *Container) RepositoryProvider() {
	if err := c.Container.Provide(userRepo.NewUserRepository); err != nil {
		c.Error = err
	}
}
