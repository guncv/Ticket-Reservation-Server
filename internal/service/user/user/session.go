package user

import (
	"errors"
	"time"

	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
)

func VerifySession(session dto.Session) error {
	if session.IsRevoked {
		return errors.New("session is revoked")
	}

	if time.Now().After(session.ExpiresAt) {
		return errors.New("session expired")
	}

	return nil
}
