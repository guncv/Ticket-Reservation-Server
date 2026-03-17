package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
)

var ErrNoAvailableTickets = errors.New("no available tickets")

func (r *eventRepository) ReserveTickets(ctx context.Context, eventID, userID uuid.UUID, quantity int) (dto.ReserveEventTicketRes, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return dto.ReserveEventTicketRes{}, fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()

	// Atomically decrement available_tickets
	updateQuery := `
		UPDATE events
		SET available_tickets = available_tickets - $1,
			updated_at = NOW()
		WHERE id = $2
	`
	_, err = conn.Exec(ctx, updateQuery, quantity, eventID)
	if err != nil {
		r.log.Error(ctx, "Failed to decrement available tickets", err)
		return dto.ReserveEventTicketRes{}, fmt.Errorf("failed to update available tickets: %w", err)
	}

	// Create the reservation
	var reservationID uuid.UUID
	insertQuery := `
		INSERT INTO reservations (event_id, user_id, quantity)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err = conn.QueryRow(ctx, insertQuery, eventID, userID, quantity).Scan(&reservationID)
	if err != nil {
		r.log.Error(ctx, "Failed to create reservation", err)
		return dto.ReserveEventTicketRes{}, fmt.Errorf("failed to create reservation: %w", err)
	}

	return dto.ReserveEventTicketRes{
		ReservationID: reservationID,
		Quantity:      quantity,
	}, nil
}
