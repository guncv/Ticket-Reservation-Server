package event_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/containers"
	"github.com/guncv/ticket-reservation-server/internal/infra/test"
	"github.com/guncv/ticket-reservation-server/internal/service/event"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/user"
	"github.com/guncv/ticket-reservation-server/internal/shared"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func validCreateReq() dto.CreateEventReq {
	return validCreateReqWithTitle("Valid Event Title")
}

func validCreateReqWithTitle(title string) dto.CreateEventReq {
	return validCreateReqWithTitleAndTickets(title, 100)
}

func validCreateReqWithTitleAndTickets(title string, tickets int) dto.CreateEventReq {
	return dto.CreateEventReq{
		Title:        title,
		Description:  "A valid description that is long enough.",
		Price:        decimal.NewFromFloat(99.99),
		TotalTickets: tickets,
	}
}

func setupEventService(t *testing.T) event.EventService {
	t.Helper()
	eventService, _ := setupEventAndUserServices(t)
	return eventService
}

func setupEventAndUserServices(t *testing.T) (event.EventService, user.UserService) {
	t.Helper()
	t.Setenv("APP_ENV", shared.AppEnvTest)

	cfg, err := config.LoadConfig(nil)
	require.NoError(t, err)

	container := containers.NewContainer(cfg)
	require.NoError(t, container.Error)

	var eventService event.EventService
	var userService user.UserService
	err = container.Container.Invoke(func(es event.EventService, us user.UserService) {
		eventService = es
		userService = us
	})
	require.NoError(t, err)
	return eventService, userService
}

func TestCreateEvent(t *testing.T) {
	eventService := setupEventService(t)

	testCases := []struct {
		name   string
		req    func() dto.CreateEventReq
		verify func(t *testing.T, actualErr error)
	}{
		{
			name: "Success",
			req:  validCreateReq,
			verify: func(t *testing.T, actualErr error) {
				require.NoError(t, actualErr)

				events, err := eventService.GetAllEvents(context.Background())
				require.NoError(t, err)
				require.Len(t, events, 1)

				got := events[0]
				req := validCreateReq()
				require.Equal(t, req.Title, got.Title)
				require.Equal(t, req.Description, got.Description)
				require.True(t, req.Price.Equal(got.Price))
				require.Equal(t, req.TotalTickets, got.TotalTickets)
				require.Equal(t, req.TotalTickets, got.AvailableTickets)
				require.NotEqual(t, uuid.Nil, got.ID)
			},
		},
		{
			name: "Error_DuplicateTitle",
			req:  validCreateReq,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "title already exists")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "Error_DuplicateTitle" {
				_, err := eventService.CreateEvent(context.Background(), validCreateReq())
				require.NoError(t, err)
			}

			_, err := eventService.CreateEvent(context.Background(), tc.req())
			tc.verify(t, err)

			require.NoError(t, test.TruncateAllTables())
		})
	}
}

func TestUpdateEvent(t *testing.T) {
	eventService := setupEventService(t)

	createAndGetID := func(t *testing.T) uuid.UUID {
		t.Helper()
		eventRes, err := eventService.CreateEvent(context.Background(), validCreateReq())
		require.NoError(t, err)
		return eventRes.ID
	}

	testCases := []struct {
		name   string
		req    func(t *testing.T) dto.UpdateEventReq
		verify func(t *testing.T, actualErr error)
	}{
		{
			name: "Success",
			req: func(t *testing.T) dto.UpdateEventReq {
				id := createAndGetID(t)
				return dto.UpdateEventReq{
					ID:           id,
					Title:        "Updated Event Title",
					Description:  "An updated description that is long enough.",
					Price:        decimal.NewFromFloat(149.99),
					TotalTickets: 200,
				}
			},
			verify: func(t *testing.T, actualErr error) {
				require.NoError(t, actualErr)

				events, err := eventService.GetAllEvents(context.Background())
				require.NoError(t, err)
				require.Len(t, events, 1)

				got := events[0]
				require.Equal(t, "Updated Event Title", got.Title)
				require.True(t, decimal.NewFromFloat(149.99).Equal(got.Price))
				require.Equal(t, 200, got.TotalTickets)
				require.Equal(t, 200, got.AvailableTickets)
			},
		},
		{
			name: "Success_SameTitleAsOwnEvent",
			req: func(t *testing.T) dto.UpdateEventReq {
				id := createAndGetID(t)
				req := validCreateReq()
				return dto.UpdateEventReq{
					ID:           id,
					Title:        req.Title,
					Description:  req.Description,
					Price:        req.Price,
					TotalTickets: req.TotalTickets,
				}
			},
			verify: func(t *testing.T, actualErr error) {
				require.NoError(t, actualErr)

				events, err := eventService.GetAllEvents(context.Background())
				require.NoError(t, err)
				require.Len(t, events, 1)
				require.Equal(t, validCreateReq().Title, events[0].Title)
			},
		},
		{
			name: "Error_EventNotFound",
			req: func(t *testing.T) dto.UpdateEventReq {
				return dto.UpdateEventReq{
					ID:           uuid.New(),
					Title:        "Updated Event Title",
					Description:  "A valid description that is long enough.",
					Price:        decimal.NewFromFloat(50.00),
					TotalTickets: 10,
				}
			},
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
			},
		},
		{
			name: "Error_DuplicateTitleFromAnotherEvent",
			req: func(t *testing.T) dto.UpdateEventReq {
				id := createAndGetID(t)

				otherReq := validCreateReq()
				otherReq.Title = "Another Event Title"
				_, err := eventService.CreateEvent(context.Background(), otherReq)
				require.NoError(t, err)

				return dto.UpdateEventReq{
					ID:           id,
					Title:        "Another Event Title",
					Description:  "A valid description that is long enough.",
					Price:        decimal.NewFromFloat(99.99),
					TotalTickets: 100,
				}
			},
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "title already exists")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.req(t)
			actualErr := eventService.UpdateEvent(context.Background(), req)
			tc.verify(t, actualErr)

			require.NoError(t, test.TruncateAllTables())
		})
	}
}
