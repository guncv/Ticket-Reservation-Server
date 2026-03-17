package event_test

import (
	"context"
	"sync"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/infra/test"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/event/repo"
	"github.com/guncv/ticket-reservation-server/internal/service/user"
	userdto "github.com/guncv/ticket-reservation-server/internal/service/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/shared"
	"github.com/stretchr/testify/require"
)

func ctxWithUserID(userID string) context.Context {
	return context.WithValue(context.Background(), shared.UserIDKey, userID)
}

func mustExtractUserIDFromToken(t *testing.T, tokenString string) string {
	t.Helper()
	tok, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	require.NoError(t, err)
	claims, ok := tok.Claims.(jwt.MapClaims)
	require.True(t, ok)
	userID, ok := claims["user_id"].(string)
	require.True(t, ok)
	return userID
}

func createUserAndGetID(t *testing.T, userService user.UserService) string {
	t.Helper()
	userName := "u_" + uuid.New().String()[:8]
	resp, err := userService.CreateUser(context.Background(), userdto.CreateUserReq{
		UserName: userName,
		Password: "reservepass123",
	})
	require.NoError(t, err)
	return mustExtractUserIDFromToken(t, resp.AccessToken)
}

func TestReserveEventTicket(t *testing.T) {
	eventService, userService := setupEventAndUserServices(t)

	createEventAndGetID := func(t *testing.T) uuid.UUID {
		t.Helper()
		eventRes, err := eventService.CreateEvent(context.Background(), validCreateReqWithTitle("reserve_"+uuid.New().String()))
		require.NoError(t, err)
		return eventRes.ID
	}

	testCases := []struct {
		name   string
		ctx    func(t *testing.T) context.Context
		req    func(t *testing.T) dto.ReserveEventTicketReq
		verify func(t *testing.T, res dto.ReserveEventTicketRes, actualErr error)
	}{
		{
			name: "Success",
			ctx: func(t *testing.T) context.Context {
				userID := createUserAndGetID(t, userService)
				return ctxWithUserID(userID)
			},
			req: func(t *testing.T) dto.ReserveEventTicketReq {
				eventID := createEventAndGetID(t)
				return dto.ReserveEventTicketReq{EventID: eventID, Quantity: 1}
			},
			verify: func(t *testing.T, res dto.ReserveEventTicketRes, actualErr error) {
				require.NoError(t, actualErr)
				require.NotEqual(t, uuid.Nil, res.ReservationID)
				require.Equal(t, 1, res.Quantity)
			},
		},
		{
			name: "Success_MultipleReservations",
			ctx: func(t *testing.T) context.Context {
				userID := createUserAndGetID(t, userService)
				return ctxWithUserID(userID)
			},
			req: func(t *testing.T) dto.ReserveEventTicketReq {
				eventID := createEventAndGetID(t)

				userAID := createUserAndGetID(t, userService)
				res, err := eventService.ReserveEventTicket(ctxWithUserID(userAID), dto.ReserveEventTicketReq{EventID: eventID, Quantity: 1})
				require.NoError(t, err)
				require.NotEqual(t, uuid.Nil, res.ReservationID)
				require.Equal(t, 1, res.Quantity)

				return dto.ReserveEventTicketReq{EventID: eventID, Quantity: 1}
			},
			verify: func(t *testing.T, res dto.ReserveEventTicketRes, actualErr error) {
				require.NoError(t, actualErr)
				require.NotEqual(t, uuid.Nil, res.ReservationID)
				require.Equal(t, 1, res.Quantity)
			},
		},
		{
			name: "Error_ReserveQuantityExceeded",
			ctx: func(t *testing.T) context.Context {
				userID := createUserAndGetID(t, userService)
				return ctxWithUserID(userID)
			},
			req: func(t *testing.T) dto.ReserveEventTicketReq {
				eventID := createEventAndGetID(t)
				return dto.ReserveEventTicketReq{EventID: eventID, Quantity: 11}
			},
			verify: func(t *testing.T, res dto.ReserveEventTicketRes, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "quantity")
			},
		},
		{
			name: "Success_MultipleTicketsInOneRequest",
			ctx: func(t *testing.T) context.Context {
				userID := createUserAndGetID(t, userService)
				return ctxWithUserID(userID)
			},
			req: func(t *testing.T) dto.ReserveEventTicketReq {
				eventID := createEventAndGetID(t)
				return dto.ReserveEventTicketReq{EventID: eventID, Quantity: 3}
			},
			verify: func(t *testing.T, res dto.ReserveEventTicketRes, actualErr error) {
				require.NoError(t, actualErr)
				require.NotEqual(t, uuid.Nil, res.ReservationID)
				require.Equal(t, 3, res.Quantity)
			},
		},
		{
			name: "Error_UserIDNotInContext",
			ctx: func(t *testing.T) context.Context {
				return context.Background()
			},
			req: func(t *testing.T) dto.ReserveEventTicketReq {
				eventID := createEventAndGetID(t)
				return dto.ReserveEventTicketReq{EventID: eventID, Quantity: 1}
			},
			verify: func(t *testing.T, res dto.ReserveEventTicketRes, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "user ID not found in context")
			},
		},
		{
			name: "Error_NoAvailableTickets",
			ctx: func(t *testing.T) context.Context {
				userID := createUserAndGetID(t, userService)
				return ctxWithUserID(userID)
			},
			req: func(t *testing.T) dto.ReserveEventTicketReq {
				eventID := createEventAndGetID(t)

				for range 100 {
					uid := createUserAndGetID(t, userService)
					_, err := eventService.ReserveEventTicket(ctxWithUserID(uid), dto.ReserveEventTicketReq{EventID: eventID, Quantity: 1})
					require.NoError(t, err)
				}

				return dto.ReserveEventTicketReq{EventID: eventID, Quantity: 1}
			},
			verify: func(t *testing.T, res dto.ReserveEventTicketRes, actualErr error) {
				require.Error(t, actualErr)
				require.ErrorIs(t, actualErr, repo.ErrNoAvailableTickets)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() { require.NoError(t, test.TruncateAllTables()) }()

			ctx := tc.ctx(t)
			req := tc.req(t)
			res, actualErr := eventService.ReserveEventTicket(ctx, req)
			tc.verify(t, res, actualErr)
		})
	}
}

func TestReserveEventTicket_Concurrent(t *testing.T) {
	defer func() { require.NoError(t, test.TruncateAllTables()) }()

	eventService, userService := setupEventAndUserServices(t)

	numTickets := 50
	eventRes, err := eventService.CreateEvent(context.Background(), validCreateReqWithTitleAndTickets("concurrent_"+uuid.New().String(), numTickets))
	require.NoError(t, err)

	// 60 goroutines try to reserve; exactly 50 should succeed, 10 should fail
	const numGoroutines = 60
	type result struct {
		reservationID uuid.UUID
		err           error
	}
	results := make([]result, numGoroutines)
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			userID := createUserAndGetID(t, userService)
			ctx := ctxWithUserID(userID)
			res, err := eventService.ReserveEventTicket(ctx, dto.ReserveEventTicketReq{EventID: eventRes.ID, Quantity: 1})
			var reservationID uuid.UUID
			if err == nil {
				reservationID = res.ReservationID
			}
			results[idx] = result{reservationID: reservationID, err: err}
		}(i)
	}
	wg.Wait()

	successes := 0
	reservationIDs := make(map[uuid.UUID]bool)
	failures := 0

	for _, r := range results {
		if r.err == nil {
			successes++
			require.NotEqual(t, uuid.Nil, r.reservationID, "successful reservation must return non-nil reservation ID")
			require.False(t, reservationIDs[r.reservationID], "each reservation must be unique (no double-booking)")
			reservationIDs[r.reservationID] = true
		} else {
			failures++
			require.ErrorIs(t, r.err, repo.ErrNoAvailableTickets)
		}
	}

	require.Equal(t, numTickets, successes, "exactly %d reservations should succeed", numTickets)
	require.Equal(t, numGoroutines-numTickets, failures)
	require.Len(t, reservationIDs, numTickets, "all %d reservation IDs must be unique", numTickets)
}
