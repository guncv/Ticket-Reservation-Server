package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/event/event"
)

func (s *eventService) CreateEvent(ctx context.Context, req dto.CreateEventReq) (dto.CreateEventRes, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return dto.CreateEventRes{}, err
	}
	defer tx.Rollback(ctx)

	exists, err := s.eventRepo.CheckEventTitleExists(ctx, req.Title, uuid.Nil)
	if err != nil {
		return dto.CreateEventRes{}, err
	}

	if err := event.ValidateCreateEvent(req, exists); err != nil {
		return dto.CreateEventRes{}, err
	}

	eventID, err := s.eventRepo.CreateEvent(ctx, req)
	if err != nil {
		return dto.CreateEventRes{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.CreateEventRes{}, err
	}

	return dto.CreateEventRes{ID: eventID}, nil
}

func (s *eventService) UpdateEvent(ctx context.Context, req dto.UpdateEventReq) error {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	exists, err := s.eventRepo.CheckEventTitleExists(ctx, req.Title, req.ID)
	if err != nil {
		return err
	}

	prevEvent, err := s.eventRepo.GetEventByID(ctx, req.ID)
	if err != nil {
		return err
	}

	if err := event.ValidateUpdateEvent(req, prevEvent, exists); err != nil {
		return err
	}

	if err := s.eventRepo.UpdateEvent(ctx, req, prevEvent.TotalTickets); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *eventService) GetAllEvents(ctx context.Context) ([]dto.Event, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	events, err := s.eventRepo.GetAllEvents(ctx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return events, nil
}
