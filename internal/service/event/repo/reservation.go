package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
)

func (r *eventRepository) GetAllReservations(ctx context.Context) ([]dto.Reservation, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()

	query := `
		SELECT
			r.id,
			r.event_id,
			e.title,
			e.description,
			e.price,
			e.total_tickets,
			e.available_tickets,
			r.quantity,
			r.user_id,
			u.user_name,
			r.created_at,
			r.updated_at
		FROM reservations r
		JOIN events e ON e.id = r.event_id
		JOIN users u ON u.id = r.user_id
		ORDER BY r.created_at DESC
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		r.log.Error(ctx, "Failed to get all reservations", err)
		return nil, fmt.Errorf("failed to get all reservations: %w", err)
	}
	defer rows.Close()

	var reservations []dto.Reservation
	for rows.Next() {
		var res dto.Reservation
		err = rows.Scan(
			&res.ID,
			&res.EventID,
			&res.EventTitle,
			&res.EventDescription,
			&res.EventPrice,
			&res.TotalTickets,
			&res.AvailableTickets,
			&res.Quantity,
			&res.UserID,
			&res.UserName,
			&res.CreatedAt,
			&res.UpdatedAt,
		)
		if err != nil {
			r.log.Error(ctx, "Failed to scan reservation", err)
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, res)
	}
	return reservations, rows.Err()
}

func (r *eventRepository) GetReservationByID(ctx context.Context, id uuid.UUID) (dto.Reservation, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return dto.Reservation{}, fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()

	query := `
		SELECT
			r.id,
			r.event_id,
			e.title,
			e.description,
			e.price,
			e.total_tickets,
			e.available_tickets,
			r.quantity,
			r.user_id,
			u.user_name,
			r.created_at,
			r.updated_at
		FROM reservations r
		JOIN events e ON e.id = r.event_id
		JOIN users u ON u.id = r.user_id
		WHERE r.id = $1
	`

	var res dto.Reservation
	err = conn.QueryRow(ctx, query, id).Scan(
		&res.ID,
		&res.EventID,
		&res.EventTitle,
		&res.EventDescription,
		&res.EventPrice,
		&res.TotalTickets,
		&res.AvailableTickets,
		&res.Quantity,
		&res.UserID,
		&res.UserName,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		r.log.Error(ctx, "Failed to get reservation by id", err)
		return dto.Reservation{}, fmt.Errorf("failed to get reservation by id: %w", err)
	}
	return res, nil
}

func (r *eventRepository) GetReservationByEventID(ctx context.Context, eventID uuid.UUID) ([]dto.Reservation, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()

	query := `
		SELECT
			r.id,
			r.event_id,
			e.title,
			e.description,
			e.price,
			e.total_tickets,
			e.available_tickets,
			r.quantity,
			r.user_id,
			u.user_name,
			r.created_at,
			r.updated_at
		FROM reservations r
		JOIN events e ON e.id = r.event_id
		JOIN users u ON u.id = r.user_id
		WHERE r.event_id = $1
		ORDER BY r.created_at DESC
	`

	rows, err := conn.Query(ctx, query, eventID)
	if err != nil {
		r.log.Error(ctx, "Failed to get reservations by event id", err)
		return nil, fmt.Errorf("failed to get reservations by event id: %w", err)
	}
	defer rows.Close()

	var reservations []dto.Reservation
	for rows.Next() {
		var res dto.Reservation
		err = rows.Scan(
			&res.ID,
			&res.EventID,
			&res.EventTitle,
			&res.EventDescription,
			&res.EventPrice,
			&res.TotalTickets,
			&res.AvailableTickets,
			&res.Quantity,
			&res.UserID,
			&res.UserName,
			&res.CreatedAt,
			&res.UpdatedAt,
		)
		if err != nil {
			r.log.Error(ctx, "Failed to scan reservation", err)
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, res)
	}
	return reservations, rows.Err()
}

func (r *eventRepository) GetAllReservationByUserID(ctx context.Context, userID uuid.UUID) ([]dto.Reservation, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()

	query := `
		SELECT
			r.id,
			r.event_id,
			e.title,
			e.description,
			e.price,
			e.total_tickets,
			e.available_tickets,
			r.quantity,
			r.user_id,
			u.user_name,
			r.created_at,
			r.updated_at
		FROM reservations r
		JOIN events e ON e.id = r.event_id
		JOIN users u ON u.id = r.user_id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
	`

	rows, err := conn.Query(ctx, query, userID)
	if err != nil {
		r.log.Error(ctx, "Failed to get reservations by user id", err)
		return nil, fmt.Errorf("failed to get reservations by user id: %w", err)
	}
	defer rows.Close()

	var reservations []dto.Reservation
	for rows.Next() {
		var res dto.Reservation
		err = rows.Scan(
			&res.ID,
			&res.EventID,
			&res.EventTitle,
			&res.EventDescription,
			&res.EventPrice,
			&res.TotalTickets,
			&res.AvailableTickets,
			&res.Quantity,
			&res.UserID,
			&res.UserName,
			&res.CreatedAt,
			&res.UpdatedAt,
		)
		if err != nil {
			r.log.Error(ctx, "Failed to scan reservation", err)
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, res)
	}
	return reservations, rows.Err()
}
