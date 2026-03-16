package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
)

func (r *eventRepository) CreateEvent(ctx context.Context, event dto.CreateEventReq) (uuid.UUID, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	createEventQuery := `
		INSERT INTO events (title, description, price, total_tickets, available_tickets)
		VALUES ($1, $2, $3, $4, $4)
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
		return uuid.Nil, fmt.Errorf("failed to create event: %w", err)
	}

	if event.TotalTickets > 0 {
		if err := r.CreateTicketsForEvent(ctx, eventID, event.TotalTickets); err != nil {
			return uuid.Nil, err
		}
	}

	return eventID, nil
}

func (r *eventRepository) UpdateEvent(ctx context.Context, event dto.UpdateEventReq, previousTotalTickets int) error {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	additionalTickets := 0
	if event.TotalTickets > previousTotalTickets {
		additionalTickets = event.TotalTickets - previousTotalTickets
	}

	query := `
		UPDATE events
		SET title = $1,
			description = $2,
			price = $3,
			total_tickets = $4,
			available_tickets = available_tickets + $5,
			updated_at = NOW()
		WHERE id = $6
	`
	_, err = conn.Exec(ctx,
		query,
		event.Title,
		event.Description,
		event.Price,
		event.TotalTickets,
		additionalTickets,
		event.ID,
	)
	if err != nil {
		r.log.Error(ctx, "Failed to update event", err)
		return fmt.Errorf("failed to update event: %w", err)
	}

	if additionalTickets > 0 {
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
			id,
			title,
			description,
			price,
			total_tickets,
			available_tickets,
			created_at,
			updated_at
		FROM events
		WHERE id = $1
	`

	var event dto.Event
	err = conn.QueryRow(ctx, query, id).Scan(
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
			id,
			title,
			description,
			price,
			total_tickets,
			available_tickets,
			created_at,
			updated_at
		FROM events
		ORDER BY updated_at DESC
	`

	rows, err := conn.Query(ctx, query)
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

func (r *eventRepository) SyncAvailableTickets(ctx context.Context) error {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()

	query := `
		UPDATE events e
		SET available_tickets = (
			SELECT COUNT(*) FROM tickets t
			WHERE t.event_id = e.id AND t.status = $1
		)
	`

	_, err = conn.Exec(ctx, query, dto.TicketStatusAvailable)
	if err != nil {
		r.log.Error(ctx, "Failed to sync available tickets", err)
		return fmt.Errorf("failed to sync available tickets: %w", err)
	}

	return nil
}
