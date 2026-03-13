package repo

import (
	"context"
	"fmt"

	"cloud.google.com/go/civil"
	"github.com/guncv/ticket-reservation-server/internal/domain/user/dto"
)

type CreateSessionParams struct {
	UserID             string
	HashedRefreshToken string
	UserAgent          string
	IPAddress          string
	ExpiresAt          civil.Time
}

func (r *userRepository) CreateSession(ctx context.Context, params CreateSessionParams) error {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	query := `
		INSERT INTO sessions (
			user_id,
			hashed_refresh_token,
			user_agent,
			ip_address,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = conn.Exec(ctx,
		query,
		params.UserID,
		params.HashedRefreshToken,
		params.UserAgent,
		params.IPAddress,
		params.ExpiresAt,
	)
	if err != nil {
		r.log.Error(ctx, "Failed to create session", err)
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r *userRepository) GetSessionByRefreshToken(ctx context.Context, hashedRefreshToken string) (dto.Session, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return dto.Session{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	query := `
		SELECT id,
			user_id,
			hashed_refresh_token,
			is_revoked,
			user_agent,
			ip_address,
			revoked_at,
			expires_at,
			created_at
		FROM sessions
		WHERE hashed_refresh_token = $1
	`

	var session dto.Session
	err = conn.QueryRow(
		ctx,
		query,
		hashedRefreshToken,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.IsRevoked,
		&session.UserAgent,
		&session.IPAddress,
		&session.RevokedAt,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		r.log.Error(ctx, "Failed to get session by refresh token", err)
		return dto.Session{}, fmt.Errorf("failed to get session by refresh token: %w", err)
	}

	return session, nil
}

func (r *userRepository) RevokeSession(ctx context.Context, sessionID string, revokedAt civil.Time) error {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	query := `
		UPDATE sessions SET is_revoked = TRUE, revoked_at = $2 WHERE id = $1
	`

	_, err = conn.Exec(
		ctx,
		query,
		sessionID,
		revokedAt,
	)
	if err != nil {
		r.log.Error(ctx, "Failed to revoke session", err)
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}
