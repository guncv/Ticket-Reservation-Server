package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
)

func (r *eventRepository) CreateTicketsForEvent(ctx context.Context, eventID uuid.UUID, count int) error {
	if count <= 0 {
		return nil
	}

	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()

	query := `
		INSERT INTO tickets (event_id, status)
		SELECT $1, $2
		FROM generate_series(1, $3)
	`
	_, err = conn.Exec(ctx, query, eventID, dto.TicketStatusAvailable, count)
	if err != nil {
		r.log.Error(ctx, "Failed to create tickets for event", err)
		return fmt.Errorf("failed to create tickets: %w", err)
	}

	return nil
}

var ErrNoAvailableTickets = errors.New("no available tickets")

func (r *eventRepository) ReserveTicket(ctx context.Context, eventID, userID uuid.UUID) (uuid.UUID, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()

	query := `
		UPDATE tickets
		SET user_id = $1,
			status = $2,
			updated_at = NOW()
		WHERE id = (
			SELECT id FROM tickets
			WHERE event_id = $3 AND status = $4
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id
	`

	var ticketID uuid.UUID
	err = conn.QueryRow(
		ctx,
		query,
		userID,
		dto.TicketStatusSold,
		eventID,
		dto.TicketStatusAvailable,
	).Scan(&ticketID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNoAvailableTickets
		}
		r.log.Error(ctx, "Failed to reserve ticket", err)
		return uuid.Nil, fmt.Errorf("failed to reserve ticket: %w", err)
	}

	return ticketID, nil
}
