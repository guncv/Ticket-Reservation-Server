package event

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
)

func (s *eventService) GetAllReservations(ctx context.Context) ([]dto.Reservation, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	reservations, err := s.eventRepo.GetAllReservations(ctx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return reservations, nil
}
