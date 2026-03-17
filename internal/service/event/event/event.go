package event

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/guncv/ticket-reservation-server/internal/service/event/dto"
	"github.com/shopspring/decimal"
)

const (
	MinEventTitleLength       = 5
	MaxEventTitleLength       = 100
	MinEventDescriptionLength = 10
	MaxEventDescriptionLength = 1_000
	MinEventPrice             = 0
	MaxEventPrice             = 1_000_000
	MinEventTotalTickets      = 1
	MaxEventTotalTickets      = 1_000_000
)

func validateCommonEventFields(req dto.CreateEventReq, checkTitleExists bool) error {
	if checkTitleExists {
		return errors.New("title already exists")
	}

	if utf8.RuneCountInString(req.Title) < MinEventTitleLength {
		return fmt.Errorf("title must be at least %d characters", MinEventTitleLength)
	}

	if utf8.RuneCountInString(req.Title) > MaxEventTitleLength {
		return fmt.Errorf("title must be at most %d characters", MaxEventTitleLength)
	}

	if utf8.RuneCountInString(req.Description) < MinEventDescriptionLength {
		return fmt.Errorf("description must be at least %d characters", MinEventDescriptionLength)
	}

	if utf8.RuneCountInString(req.Description) > MaxEventDescriptionLength {
		return fmt.Errorf("description must be at most %d characters", MaxEventDescriptionLength)
	}

	if req.Price.LessThanOrEqual(decimal.Zero) {
		return errors.New("price must be greater than 0")
	}

	if req.TotalTickets <= MinEventTotalTickets {
		return fmt.Errorf("total tickets must be at least %d", MinEventTotalTickets)
	}

	if req.TotalTickets > MaxEventTotalTickets {
		return fmt.Errorf("total tickets must be at most %d", MaxEventTotalTickets)
	}

	return nil
}

func ValidateCreateEvent(req dto.CreateEventReq, checkTitleExists bool) error {
	if err := validateCommonEventFields(req, checkTitleExists); err != nil {
		return err
	}

	return nil
}

func ValidateUpdateEvent(req dto.UpdateEventReq, prevEvent dto.Event, checkTitleExists bool) error {
	if req.ID != prevEvent.ID {
		return errors.New("event id cannot be changed")
	}

	createEventReq := dto.CreateEventReq{
		Title:        req.Title,
		Description:  req.Description,
		Price:        req.Price,
		TotalTickets: req.TotalTickets,
	}

	if err := validateCommonEventFields(createEventReq, checkTitleExists); err != nil {
		return err
	}

	if req.TotalTickets < prevEvent.TotalTickets {
		return fmt.Errorf("total tickets cannot be less than the previous total tickets")
	}

	return nil
}
