package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/event/repo"
)

type EventService interface {
	CreateEvent(ctx context.Context, req dto.CreateEventReq) error
	UpdateEvent(ctx context.Context, req dto.UpdateEventReq) error
	GetAllEvents(ctx context.Context) ([]dto.Event, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (dto.Event, error)

	ReserveEventTicket(ctx context.Context, req dto.ReserveEventTicketReq) (dto.ReserveEventTicketRes, error)
}

type eventService struct {
	eventRepo repo.EventRepository
	log       log.Logger
	db        *db.PgPool
}

func NewEventService(
	eventRepo repo.EventRepository,
	log log.Logger,
	db *db.PgPool,
) EventService {
	return &eventService{
		eventRepo: eventRepo,
		log:       log,
		db:        db,
	}
}
