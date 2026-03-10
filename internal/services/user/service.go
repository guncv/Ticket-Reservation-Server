package user

import (
	"context"
	"log"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/services/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/services/user/repo"
)

type UserService interface {
	HealthCheck(ctx context.Context) (dto.HealthCheckRes, error)
}

type userService struct {
	log      *log.Logger
	userRepo repo.UserRepository
	config   *config.Config
}

func NewUserService(l *log.Logger,
	r repo.UserRepository,
	config *config.Config,
) UserService {
	return &userService{
		log:      l,
		userRepo: r,
		config:   config,
	}
}
