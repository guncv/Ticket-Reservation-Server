package repo

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"

	"github.com/guncv/ticket-reservation-server/internal/db"
)

type User struct {
	ID             string
	UserName       string
	HashedPassword string
	Role           string
}

type UserRepository interface {
	CreateUser(ctx context.Context, params CreateUserParams) (string, error)
	CheckUserNameExists(ctx context.Context, userName string) (bool, error)
	GetUserByUserName(ctx context.Context, userName string) (User, error)
	CreateSession(ctx context.Context, params CreateSessionParams) error
	GetSessionByRefreshToken(ctx context.Context, hashedRefreshToken string) (dto.Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
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
