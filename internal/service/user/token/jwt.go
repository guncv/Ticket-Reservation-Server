package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
)

type JWTToken struct {
	cfg *config.Config
	log log.Logger
}

func NewJWTToken(cfg *config.Config, log log.Logger) (*JWTToken, error) {
	if cfg.TokenConfig.SecretKey == "" {
		return nil, errors.New("token secret key is required")
	}

	return &JWTToken{
		cfg: cfg,
		log: log,
	}, nil
}

type JWTClaims struct {
	UserID string    `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

func (j *JWTToken) GenerateAccessToken(userID string) (string, error) {
	token, _, err := j.generateToken(userID, AccessToken, j.cfg.AuthConfig.AccessTokenDuration)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (j *JWTToken) GenerateRefreshToken() (string, time.Time, error) {
	token := uuid.NewString()
	expiresAt := time.Now().Add(j.cfg.AuthConfig.RefreshTokenDuration)
	return token, expiresAt, nil
}

func (j *JWTToken) VerifyAccessToken(tokenString string) (*TokenPayload, error) {
	payload, err := j.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if payload.Type != AccessToken {
		return nil, ErrTokenInvalid
	}

	return payload, nil
}


func (j *JWTToken) generateToken(
	userID string, tokenType TokenType, duration time.Duration,
) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(duration)

	claims := JWTClaims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.cfg.TokenConfig.TokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.cfg.TokenConfig.SecretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

func (j *JWTToken) parseToken(tokenString string) (*TokenPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return []byte(j.cfg.TokenConfig.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return &TokenPayload{
		UserID:    claims.UserID,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
		Type:      claims.Type,
	}, nil
}
