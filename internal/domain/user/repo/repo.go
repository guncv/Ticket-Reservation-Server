package repo

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/infra/log"

	"github.com/guncv/ticket-reservation-server/internal/db"
)

type UserRepository interface {
	HealthCheck(ctx context.Context) (string, error)
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
