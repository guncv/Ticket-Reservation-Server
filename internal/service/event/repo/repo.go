package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event dto.CreateEventReq) (uuid.UUID, error)
	UpdateEvent(ctx context.Context, event dto.UpdateEventReq, previousTotalTickets int) error
	CheckEventTitleExists(ctx context.Context, title string, excludeID uuid.UUID) (bool, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (dto.Event, error)
	GetAllEvents(ctx context.Context) ([]dto.Event, error)

	CreateTicketsForEvent(ctx context.Context, eventID uuid.UUID, count int) error
	ReserveTickets(ctx context.Context, eventID, userID uuid.UUID, quantity int) (dto.ReserveEventTicketRes, error)
	GetAllReservations(ctx context.Context) ([]dto.Reservation, error)
}

type eventRepository struct {
	db  *db.PgPool
	log log.Logger
}

func NewEventRepository(
	db *db.PgPool,
	log log.Logger,
) EventRepository {
	return &eventRepository{
		db:  db,
		log: log,
	}
}
