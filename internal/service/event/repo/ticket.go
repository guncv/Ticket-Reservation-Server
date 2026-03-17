package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
)

var ErrNoAvailableTickets = errors.New("no available tickets")

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

func (r *eventRepository) ReserveTickets(ctx context.Context, eventID, userID uuid.UUID, quantity int) (dto.ReserveEventTicketRes, error) {
	if quantity <= 0 {
		return dto.ReserveEventTicketRes{}, nil
	}

	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return dto.ReserveEventTicketRes{}, fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()

	// Create reservation
	var reservationID uuid.UUID
	createReservationQuery := `
		INSERT INTO reservations (event_id, user_id)
		VALUES ($1, $2)
		RETURNING id
	`
	err = conn.QueryRow(ctx, createReservationQuery, eventID, userID).Scan(&reservationID)
	if err != nil {
		r.log.Error(ctx, "Failed to create reservation", err)
		return dto.ReserveEventTicketRes{}, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Reserve tickets
	reserveTicketsQuery := `
		UPDATE tickets
		SET reservation_id = $1,
			status = $2,
			updated_at = NOW()
		WHERE id IN (
			SELECT id FROM tickets
			WHERE event_id = $3 AND status = $4
			LIMIT $5
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id
	`

	rows, err := conn.Query(
		ctx,
		reserveTicketsQuery,
		reservationID,
		dto.TicketStatusSold,
		eventID,
		dto.TicketStatusAvailable,
		quantity,
	)
	if err != nil {
		r.log.Error(ctx, "Failed to reserve tickets", err)
		return dto.ReserveEventTicketRes{}, fmt.Errorf("failed to reserve tickets: %w", err)
	}
	defer rows.Close()

	var ticketIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return dto.ReserveEventTicketRes{}, fmt.Errorf("failed to scan ticket id: %w", err)
		}
		ticketIDs = append(ticketIDs, id)
	}
	if err := rows.Err(); err != nil {
		return dto.ReserveEventTicketRes{}, fmt.Errorf("failed to iterate ticket rows: %w", err)
	}

	if len(ticketIDs) < quantity {
		return dto.ReserveEventTicketRes{}, ErrNoAvailableTickets
	}

	return dto.ReserveEventTicketRes{
		ReservationID: reservationID,
		TicketIDs:     ticketIDs,
	}, nil
}
