package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
)

func (r *eventRepository) CreateEvent(ctx context.Context, event dto.CreateEventReq) error {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	createEventQuery := `
		INSERT INTO events (title, description, price, total_tickets)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var eventID uuid.UUID
	err = conn.QueryRow(ctx,
		createEventQuery,
		event.Title,
		event.Description,
		event.Price,
		event.TotalTickets,
	).Scan(&eventID)
	if err != nil {
		r.log.Error(ctx, "Failed to create event", err)
		return fmt.Errorf("failed to create event: %w", err)
	}

	if event.TotalTickets > 0 {
		if err := r.CreateTicketsForEvent(ctx, eventID, event.TotalTickets); err != nil {
			return err
		}
	}

	return nil
}

func (r *eventRepository) UpdateEvent(ctx context.Context, event dto.UpdateEventReq, previousTotalTickets int) error {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	query := `
		UPDATE events
		SET title = $1,
			description = $2,
			price = $3,
			total_tickets = $4,
			updated_at = NOW()
		WHERE id = $5
	`
	_, err = conn.Exec(ctx,
		query,
		event.Title,
		event.Description,
		event.Price,
		event.TotalTickets,
		event.ID,
	)
	if err != nil {
		r.log.Error(ctx, "Failed to update event", err)
		return fmt.Errorf("failed to update event: %w", err)
	}

	if event.TotalTickets > previousTotalTickets {
		additionalTickets := event.TotalTickets - previousTotalTickets
		if err := r.CreateTicketsForEvent(ctx, event.ID, additionalTickets); err != nil {
			return fmt.Errorf("failed to create additional tickets: %w", err)
		}
	}

	return nil
}

func (r *eventRepository) CheckEventTitleExists(ctx context.Context, title string, excludeID uuid.UUID) (bool, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	query := `
		SELECT EXISTS(SELECT 1 FROM events WHERE title = $1 AND
			($2 = '00000000-0000-0000-0000-000000000000'::uuid OR id != $2))
	`

	var exists bool
	err = conn.QueryRow(ctx, query, title, excludeID).Scan(&exists)
	if err != nil {
		r.log.Error(ctx, "Failed to check event title exists", err)
		return false, fmt.Errorf("failed to check event title exists: %w", err)
	}

	return exists, nil
}

func (r *eventRepository) GetEventByID(ctx context.Context, id uuid.UUID) (dto.Event, error) {

	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return dto.Event{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	query := `
		SELECT
			e.id,
			e.title,
			e.description,
			e.price,
			e.total_tickets,
			COUNT(t.id) FILTER (WHERE t.status = $2) AS available_tickets,
			e.created_at,
			e.updated_at
		FROM events e
		LEFT JOIN tickets t ON t.event_id = e.id
		WHERE e.id = $1
		GROUP BY e.id
	`

	var event dto.Event
	err = conn.QueryRow(ctx, query, id, dto.TicketStatusAvailable).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.Price,
		&event.TotalTickets,
		&event.AvailableTickets,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		r.log.Error(ctx, "Failed to get event by id", err)
		return dto.Event{}, fmt.Errorf("failed to get event by id: %w", err)
	}

	return event, nil
}

func (r *eventRepository) GetAllEvents(ctx context.Context) ([]dto.Event, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	query := `
		SELECT
			e.id,
			e.title,
			e.description,
			e.price,
			e.total_tickets,
			COUNT(t.id) FILTER (WHERE t.status = $1) AS available_tickets,
			e.created_at,
			e.updated_at
		FROM events e
		LEFT JOIN tickets t ON t.event_id = e.id
		GROUP BY e.id
		ORDER BY e.updated_at DESC
	`

	rows, err := conn.Query(ctx, query, dto.TicketStatusAvailable)
	if err != nil {
		r.log.Error(ctx, "Failed to get all events", err)
		return nil, fmt.Errorf("failed to get all events: %w", err)
	}
	defer rows.Close()

	var events []dto.Event
	for rows.Next() {
		var event dto.Event
		err = rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.Price,
			&event.TotalTickets,
			&event.AvailableTickets,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			r.log.Error(ctx, "Failed to scan event", err)
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
