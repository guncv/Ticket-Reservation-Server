package event

import (
	"errors"

	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
)

const MaxReserveQuantity = 10

func ValidateReserveEventTicket(req dto.ReserveEventTicketReq) error {
	if req.EventID == uuid.Nil {
		return errors.New("event_id is required")
	}

	if req.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	if req.Quantity > MaxReserveQuantity {
		return errors.New("quantity must be less than or equal to 10")
	}

	return nil
}
