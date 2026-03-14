package user

import (
	"crypto/sha256"
	"encoding/hex"
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

func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
