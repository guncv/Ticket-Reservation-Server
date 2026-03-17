package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Reservation struct {
	ID               uuid.UUID       `json:"id"`
	EventID          uuid.UUID       `json:"event_id"`
	EventTitle       string          `json:"event_title"`
	EventDescription string          `json:"event_description"`
	EventPrice       decimal.Decimal `json:"event_price"`
	TotalTickets     int             `json:"total_tickets"`
	AvailableTickets int             `json:"available_tickets"`
	Quantity         int             `json:"quantity"`
	UserID           uuid.UUID       `json:"user_id"`
	UserName         string          `json:"user_name"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
