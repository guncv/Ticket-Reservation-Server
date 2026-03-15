package seeders

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/containers"
	"github.com/guncv/ticket-reservation-server/internal/service/user"
	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
	"github.com/jackc/pgx/v5/pgxpool"
)

var seededAdminUserID string

func init() {
	if err := registerSeed(seedUserAdminUp, seedUserAdminDown); err != nil {
		panic(err)
	}
}

func seedUserAdminUp(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	c := containers.NewContainer(cfg)

	var userService user.UserService
	if err := c.Container.Invoke(func(us user.UserService) {
		userService = us
	}); err != nil {
		return err
	}

	resp, err := userService.CreateAdminUser(ctx, dto.CreateUserReq{
		UserName: "admin",
		Password: "password",
	})
	if err != nil {
		return err
	}

	seededAdminUserID = resp.UserID

	return nil
}

func seedUserAdminDown(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	c := containers.NewContainer(cfg)

	if err := c.Container.Invoke(func() {
		// TODO: Implement rollback logic here
	}); err != nil {
		return err
	}

	return nil
}
