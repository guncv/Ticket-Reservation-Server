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
	EventID uuid.UUID `json:"event_id"`
}

type ReserveEventTicketRes struct {
	TicketID uuid.UUID `json:"ticket_id"`
}
