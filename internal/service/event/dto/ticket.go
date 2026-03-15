package dto

import (
	"github.com/google/uuid"
)

type TicketStatus = string

const (
	TicketStatusAvailable TicketStatus = "available"
	TicketStatusSold      TicketStatus = "sold"
)

type ReserveEventTicketReq struct {
	EventID  uuid.UUID `json:"event_id"`
	Quantity int       `json:"quantity"`
}

type ReserveEventTicketRes struct {
	ReservationID uuid.UUID   `json:"reservation_id"`
	TicketIDs     []uuid.UUID `json:"ticket_ids"`
}
