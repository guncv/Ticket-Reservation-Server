package token

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"time"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/o1egl/paseto"
)

type PasetoToken struct {
	paseto       *paseto.V2
	symmetricKey []byte
	cfg          *config.Config
	log          log.Logger
}

func NewPasetoToken(cfg *config.Config, log log.Logger) (*PasetoToken, error) {
	symmetricKey, err := hex.DecodeString(cfg.TokenConfig.SecretKey)
	if err != nil {
		log.Error(context.Background(), "failed to decode secret key", "error", err)
		return nil, err
	}

	if len(symmetricKey) != ed25519.SeedSize {
		return nil, errors.New("invalid key length: must be 64 hex characters (32 bytes)")
	}

	return &PasetoToken{
		paseto:       paseto.NewV2(),
		symmetricKey: symmetricKey,
		cfg:          cfg,
		log:          log,
	}, nil
}

type PasetoClaims struct {
	UserID    string    `json:"user_id"`
	Type      TokenType `json:"type"`
	Issuer    string    `json:"iss"`
	IssuedAt  time.Time `json:"iat"`
	ExpiresAt time.Time `json:"exp"`
}

func (p *PasetoToken) GenerateAccessToken(userID string) (string, error) {
	token, _, err := p.generateToken(userID, AccessToken, p.cfg.AuthConfig.AccessTokenDuration)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (p *PasetoToken) GenerateRefreshToken(userID string) (string, time.Time, error) {
	return p.generateToken(userID, RefreshToken, p.cfg.AuthConfig.RefreshTokenDuration)
}

func (p *PasetoToken) VerifyToken(tokenString string) (TokenPayload, error) {
	payload, err := p.decryptToken(tokenString)
	if err != nil {
		return TokenPayload{}, err
	}

	return payload, nil
}

func (p *PasetoToken) generateToken(
	userID string, tokenType TokenType, duration time.Duration,
) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(duration)

	claims := PasetoClaims{
		UserID:    userID,
		Type:      tokenType,
		Issuer:    p.cfg.TokenConfig.TokenIssuer,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
	}

	token, err := p.paseto.Encrypt(p.symmetricKey, claims, nil)
	if err != nil {
		p.log.Error(context.Background(), "failed to encrypt token", "error", err)
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (p *PasetoToken) decryptToken(tokenString string) (TokenPayload, error) {
	var claims PasetoClaims
	err := p.paseto.Decrypt(tokenString, p.symmetricKey, &claims, nil)
	if err != nil {
		p.log.Error(context.Background(), "failed to decrypt token", "error", err)
		return TokenPayload{}, ErrTokenInvalid
	}

	if time.Now().After(claims.ExpiresAt) {
		return TokenPayload{}, ErrTokenExpired
	}

	return TokenPayload{
		UserID:    claims.UserID,
		IssuedAt:  claims.IssuedAt,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}
