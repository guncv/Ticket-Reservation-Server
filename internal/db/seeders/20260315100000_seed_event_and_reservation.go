package seeders

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/containers"
	"github.com/guncv/ticket-reservation-server/internal/service/event"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
	"github.com/guncv/ticket-reservation-server/internal/shared"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var seededEventTitle = "Concert Night 2025"

func init() {
	if err := registerSeed(seedEventAndReservationUp, seedEventAndReservationDown); err != nil {
		panic(err)
	}
}

func seedEventAndReservationUp(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	c := containers.NewContainer(cfg)

	var eventService event.EventService
	if err := c.Container.Invoke(func(es event.EventService) {
		eventService = es
	}); err != nil {
		return err
	}

	eventID, err := eventService.CreateEvent(ctx, dto.CreateEventReq{
		Title:        seededEventTitle,
		Description:  "An amazing live concert experience",
		Price:        decimal.NewFromFloat(49.99),
		TotalTickets: 100,
	})
	if err != nil {
		return err
	}

	ctxWithUser := context.WithValue(ctx, shared.UserIDKey, seededAdminUserID)
	_, err = eventService.ReserveEventTicket(ctxWithUser, dto.ReserveEventTicketReq{
		EventID:  eventID,
		Quantity: 2,
	})
	if err != nil {
		return err
	}

	return nil
}

func seedEventAndReservationDown(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	return nil
}
