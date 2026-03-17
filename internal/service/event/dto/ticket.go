package dto

import (
	"github.com/google/uuid"
)

type ReserveEventTicketReq struct {
	EventID  uuid.UUID `json:"event_id"`
	Quantity int       `json:"quantity"`
}

type ReserveEventTicketRes struct {
	ReservationID uuid.UUID `json:"reservation_id"`
	Quantity      int       `json:"quantity"`
}
