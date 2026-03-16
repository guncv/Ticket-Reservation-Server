package dto

import (
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateEventReq struct {
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Price        decimal.Decimal `json:"price"`
	TotalTickets int             `json:"total_tickets"`
}

type CreateEventRes struct {
	ID uuid.UUID `json:"id"`
}

type UpdateEventReq struct {
	ID           uuid.UUID       `json:"id"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Price        decimal.Decimal `json:"price"`
	TotalTickets int             `json:"total_tickets"`
}

type Event struct {
	ID               uuid.UUID       `json:"id"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	Price            decimal.Decimal `json:"price"`
	TotalTickets     int             `json:"total_tickets"`
	AvailableTickets int             `json:"available_tickets"`
	CreatedAt        civil.DateTime  `json:"created_at"`
	UpdatedAt        civil.DateTime  `json:"updated_at"`
}
