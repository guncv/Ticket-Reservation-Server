package containers

import (
	"github.com/guncv/ticket-reservation-server/internal/services/user/repo"
)

func (c *Container) RepositoryProvider() {
	if err := c.Container.Provide(repo.NewUserRepository); err != nil {
		c.Error = err
	}
}
