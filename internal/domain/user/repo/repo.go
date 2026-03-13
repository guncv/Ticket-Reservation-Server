package repo

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/domain/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"

	"github.com/guncv/ticket-reservation-server/internal/db"
)

type UserRepository interface {
	HealthCheck(ctx context.Context) (string, error)
	CreateUser(ctx context.Context, params CreateUserParams) (string, error)
	CheckUserNameExists(ctx context.Context, userName string) (bool, error)
	CreateSession(ctx context.Context, params CreateSessionParams) error
	GetSessionByRefreshToken(ctx context.Context, hashedRefreshToken string) (dto.Session, error)
}

type userRepository struct {
	db  *db.PgPool
	log log.Logger
}

func NewUserRepository(
	db *db.PgPool,
	log log.Logger,
) UserRepository {
	return &userRepository{
		db:  db,
		log: log,
	}
}
