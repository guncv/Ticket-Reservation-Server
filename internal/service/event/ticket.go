package event

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/event/event"
	"github.com/guncv/ticket-reservation-server/internal/shared"
)

func (s *eventService) ReserveEventTicket(ctx context.Context, req dto.ReserveEventTicketReq) (dto.ReserveEventTicketRes, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return dto.ReserveEventTicketRes{}, err
	}
	defer tx.Rollback(ctx)

	if err := event.ValidateReserveEventTicket(req); err != nil {
		return dto.ReserveEventTicketRes{}, err
	}

	userIDStr, err := shared.GetUserIDFromContext(ctx)
	if err != nil {
		return dto.ReserveEventTicketRes{}, err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return dto.ReserveEventTicketRes{}, errors.New("invalid user_id in context")
	}

	result, err := s.eventRepo.ReserveTickets(ctx, req.EventID, userID, req.Quantity)
	if err != nil {
		return dto.ReserveEventTicketRes{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.ReserveEventTicketRes{}, err
	}

	return result, nil
}
