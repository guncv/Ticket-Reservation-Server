package event_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/event/event"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func validCreateReq() dto.CreateEventReq {
	return dto.CreateEventReq{
		Title:        "Valid Event Title",
		Description:  "A valid description that is long enough.",
		Price:        decimal.NewFromFloat(99.99),
		TotalTickets: 100,
	}
}

func TestValidateCreateEvent(t *testing.T) {
	testCases := []struct {
		name             string
		req              func() dto.CreateEventReq
		checkTitleExists bool
		verify           func(t *testing.T, actualErr error)
	}{
		{
			name:             "Success",
			req:              validCreateReq,
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.NoError(t, actualErr)
			},
		},
		{
			name:             "Error_TitleAlreadyExists",
			req:              validCreateReq,
			checkTitleExists: true,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "title already exists")
			},
		},
		{
			name: "Error_TitleTooShort",
			req: func() dto.CreateEventReq {
				req := validCreateReq()
				req.Title = "Hi"
				return req
			},
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "title must be at least")
			},
		},
		{
			name: "Error_TitleTooLong",
			req: func() dto.CreateEventReq {
				req := validCreateReq()
				req.Title = strings.Repeat("a", event.MaxEventTitleLength+1)
				return req
			},
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "title must be at most")
			},
		},
		{
			name: "Error_DescriptionTooShort",
			req: func() dto.CreateEventReq {
				req := validCreateReq()
				req.Description = "Short"
				return req
			},
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "description must be at least")
			},
		},
		{
			name: "Error_DescriptionTooLong",
			req: func() dto.CreateEventReq {
				req := validCreateReq()
				req.Description = strings.Repeat("a", event.MaxEventDescriptionLength+1)
				return req
			},
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "description must be at most")
			},
		},
		{
			name: "Error_PriceZero",
			req: func() dto.CreateEventReq {
				req := validCreateReq()
				req.Price = decimal.Zero
				return req
			},
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "price must be greater than 0")
			},
		},
		{
			name: "Error_PriceNegative",
			req: func() dto.CreateEventReq {
				req := validCreateReq()
				req.Price = decimal.NewFromFloat(-1)
				return req
			},
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "price must be greater than 0")
			},
		},
		{
			name: "Error_TotalTicketsTooLow",
			req: func() dto.CreateEventReq {
				req := validCreateReq()
				req.TotalTickets = event.MinEventTotalTickets
				return req
			},
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "total tickets must be at least")
			},
		},
		{
			name: "Error_TotalTicketsTooHigh",
			req: func() dto.CreateEventReq {
				req := validCreateReq()
				req.TotalTickets = event.MaxEventTotalTickets + 1
				return req
			},
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "total tickets must be at most")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := event.ValidateCreateEvent(tc.req(), tc.checkTitleExists)
			tc.verify(t, actualErr)
		})
	}
}

func TestValidateUpdateEvent(t *testing.T) {
	fixedID := uuid.New()

	validPrevEvent := dto.Event{
		ID:           fixedID,
		Title:        "Original Title",
		Description:  "Original description long enough.",
		Price:        decimal.NewFromFloat(99.99),
		TotalTickets: 100,
	}

	validUpdateReq := func() dto.UpdateEventReq {
		return dto.UpdateEventReq{
			ID:           fixedID,
			Title:        "Updated Event Title",
			Description:  "An updated description that is long enough.",
			Price:        decimal.NewFromFloat(149.99),
			TotalTickets: 200,
		}
	}

	testCases := []struct {
		name             string
		req              func() dto.UpdateEventReq
		prevEvent        dto.Event
		checkTitleExists bool
		verify           func(t *testing.T, actualErr error)
	}{
		{
			name:             "Success",
			req:              validUpdateReq,
			prevEvent:        validPrevEvent,
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.NoError(t, actualErr)
			},
		},
		{
			name: "Error_IDMismatch",
			req: func() dto.UpdateEventReq {
				req := validUpdateReq()
				req.ID = uuid.New()
				return req
			},
			prevEvent:        validPrevEvent,
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "event id cannot be changed")
			},
		},
		{
			name:             "Error_TitleAlreadyExistsOnAnotherEvent",
			req:              validUpdateReq,
			prevEvent:        validPrevEvent,
			checkTitleExists: true,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "title already exists")
			},
		},
		{
			name: "Error_TitleTooShort",
			req: func() dto.UpdateEventReq {
				req := validUpdateReq()
				req.Title = "Hi"
				return req
			},
			prevEvent:        validPrevEvent,
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "title must be at least")
			},
		},
		{
			name: "Error_TitleTooLong",
			req: func() dto.UpdateEventReq {
				req := validUpdateReq()
				req.Title = strings.Repeat("a", event.MaxEventTitleLength+1)
				return req
			},
			prevEvent:        validPrevEvent,
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "title must be at most")
			},
		},
		{
			name: "Error_DescriptionTooShort",
			req: func() dto.UpdateEventReq {
				req := validUpdateReq()
				req.Description = "Short"
				return req
			},
			prevEvent:        validPrevEvent,
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "description must be at least")
			},
		},
		{
			name: "Error_PriceZero",
			req: func() dto.UpdateEventReq {
				req := validUpdateReq()
				req.Price = decimal.Zero
				return req
			},
			prevEvent:        validPrevEvent,
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "price must be greater than 0")
			},
		},
		{
			name: "Error_ReduceTotalTickets",
			req: func() dto.UpdateEventReq {
				req := validUpdateReq()
				req.TotalTickets = 50
				return req
			},
			prevEvent:        validPrevEvent,
			checkTitleExists: false,
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "total tickets cannot be less")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := event.ValidateUpdateEvent(tc.req(), tc.prevEvent, tc.checkTitleExists)
			tc.verify(t, actualErr)
		})
	}
}
