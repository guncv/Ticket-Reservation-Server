package user

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/infra/redis"
	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/user/repo"
	"github.com/guncv/ticket-reservation-server/internal/service/user/token"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.CreateUserReq) (dto.CreateUserResp, error)
	CreateAdminUser(ctx context.Context, req dto.CreateUserReq) (dto.CreateUserResp, error)
	VerifyAndRenewToken(ctx context.Context, req dto.SessionReq) (dto.SessionResp, token.TokenPayload, error)
	LoginUser(ctx context.Context, req dto.LoginUserReq) (dto.LoginUserResp, error)
	LogoutUser(ctx context.Context) error
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
