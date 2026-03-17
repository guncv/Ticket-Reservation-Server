package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("invalid token")
)

type TokenPayload struct {
	UserID    string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Token interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, time.Time, error)
	VerifyToken(token string) (TokenPayload, error)
}

func NewToken(
	cfg *config.Config,
	log log.Logger,
) (Token, error) {
	providerFactory, exists := providers[cfg.TokenConfig.TokenType]
	if !exists {
		return nil, fmt.Errorf("unsupported token type: %s", cfg.TokenConfig.TokenType)
	}
	return providerFactory(cfg, log)
}

type TokenProviderFactory func(cfg *config.Config, log log.Logger) (Token, error)

var providers = map[string]TokenProviderFactory{
	config.TokenTypeJWT: func(cfg *config.Config, log log.Logger) (Token, error) {
		return NewJWTToken(cfg, log)
	},

	config.TokenTypePASETO: func(cfg *config.Config, log log.Logger) (Token, error) {
		return NewPasetoToken(cfg, log)
	},
}
