package user

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/guncv/ticket-reservation-server/internal/domain/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/domain/user/repo"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/infra/redis"
)

type UserService interface {
	HealthCheck(ctx context.Context) (*dto.HealthCheckResp, error)
}

type userService struct {
	userRepo repo.UserRepository
	db       *db.PgPool
	config   *config.Config
	logger   log.Logger
	redis    redis.RedisClient
}

func NewUserService(
	userRepo repo.UserRepository,
	db *db.PgPool,
	config *config.Config,
	logger log.Logger,
	redis redis.RedisClient,
) UserService {
	return &userService{
		userRepo: userRepo,
		db:       db,
		config:   config,
		logger:   logger,
		redis:    redis,
	}
}
