package user

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/guncv/ticket-reservation-server/internal/domain/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/domain/user/repo"
	"github.com/guncv/ticket-reservation-server/internal/domain/user/token"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/infra/redis"
)

type UserService interface {
	HealthCheck(ctx context.Context) (*dto.HealthCheckResp, error)
	CreateUser(ctx context.Context, req dto.CreateUserReq) (dto.CreateUserResp, error)
	VerifyAndRenewToken(ctx context.Context, req dto.RenewTokenReq) (dto.RenewTokenResp, error)
}

type userService struct {
	userRepo repo.UserRepository
	db       *db.PgPool
	config   *config.Config
	token    token.Token
	log      log.Logger
	redis    redis.RedisClient
}

func NewUserService(
	userRepo repo.UserRepository,
	db *db.PgPool,
	config *config.Config,
	token token.Token,
	logger log.Logger,
	redis redis.RedisClient,
) UserService {
	return &userService{
		userRepo: userRepo,
		db:       db,
		config:   config,
		token:    token,
		log:      logger,
		redis:    redis,
	}
}

func (s *userService) HealthCheck(ctx context.Context) (*dto.HealthCheckResp, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result, err := s.userRepo.HealthCheck(ctx)
	if err != nil {
		s.log.Error(ctx, "Failed to health check", err)
		return nil, err
	}

	return &dto.HealthCheckResp{Status: result}, nil
}
