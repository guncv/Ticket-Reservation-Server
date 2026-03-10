package repo

import (
	"context"

	db "github.com/guncv/ticket-reservation-server/internal/db/sqlc"
	"github.com/guncv/ticket-reservation-server/internal/infras/log"
)

type UserRepository interface {
	HealthCheck(ctx context.Context) (string, error)
}

type userRepository struct {
	log *log.Logger
	db  db.Store
}

func NewUserRepository(l *log.Logger, db db.Store) UserRepository {
	return &userRepository{
		log: l,
		db:  db,
	}
}
