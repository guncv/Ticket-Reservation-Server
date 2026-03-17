package user

import (
	"context"
	"errors"

	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/user/repo"
	"github.com/guncv/ticket-reservation-server/internal/service/user/token"
	"github.com/guncv/ticket-reservation-server/internal/service/user/user"
	"github.com/guncv/ticket-reservation-server/internal/shared"
)

func (s *userService) CreateSession(ctx context.Context, userID string) (dto.SessionResp, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return dto.SessionResp{}, err
	}
	defer tx.Rollback(ctx)

	accessToken, err := s.token.GenerateAccessToken(userID)
	if err != nil {
		s.log.Error(ctx, "Failed to generate access token", err)
		return dto.SessionResp{}, err
	}

	refreshToken, expiresAt, err := s.token.GenerateRefreshToken(userID)
	if err != nil {
		s.log.Error(ctx, "Failed to generate refresh token", err)
		return dto.SessionResp{}, err
	}

	userAgent, _ := ctx.Value(shared.UserAgentKey).(string)
	ipAddress, _ := ctx.Value(shared.ClientIPKey).(string)

	hashedRefreshToken := user.HashRefreshToken(refreshToken)

	session := repo.CreateSessionParams{
		UserID:             userID,
		HashedRefreshToken: hashedRefreshToken,
		UserAgent:          userAgent,
		IPAddress:          ipAddress,
		ExpiresAt:          expiresAt,
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		s.log.Error(ctx, "Failed to create session", err)
		return dto.SessionResp{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.SessionResp{}, err
	}

	return dto.SessionResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *userService) VerifyAndRenewToken(ctx context.Context, req dto.SessionReq) (dto.SessionResp, token.TokenPayload, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return dto.SessionResp{}, token.TokenPayload{}, err
	}
	defer tx.Rollback(ctx)

	accessPayload, accessErr := s.token.VerifyToken(req.AccessToken)

	if accessErr != nil && !errors.Is(accessErr, token.ErrTokenExpired) {
		s.log.Error(ctx, "Invalid access token", accessErr)
		return dto.SessionResp{}, token.TokenPayload{}, errors.New("invalid access token")
	}

	refreshPayload, refreshErr := s.token.VerifyToken(req.RefreshToken)
	if refreshErr != nil {
		s.log.Error(ctx, "Invalid refresh token", err)
		return dto.SessionResp{}, refreshPayload, errors.New("invalid refresh token")
	}

	if refreshPayload.UserID != accessPayload.UserID {
		s.log.Error(ctx, "Invalid refresh token", err)
		return dto.SessionResp{}, refreshPayload, errors.New("invalid refresh token")
	}

	if accessErr == nil {
		return dto.SessionResp{
			AccessToken:  req.AccessToken,
			RefreshToken: req.RefreshToken,
		}, accessPayload, nil
	}

	session, err := s.GetSessionByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.log.Error(ctx, "Failed to get session", err)
		return dto.SessionResp{}, token.TokenPayload{}, err
	}

	newAccessToken, err := s.token.GenerateAccessToken(session.UserID)
	if err != nil {
		s.log.Error(ctx, "Failed to generate new access token", err)
		return dto.SessionResp{}, token.TokenPayload{}, err
	}

	return dto.SessionResp{
			AccessToken:  newAccessToken,
			RefreshToken: req.RefreshToken,
		}, token.TokenPayload{
			UserID:    session.UserID,
			IssuedAt:  session.CreatedAt,
			ExpiresAt: session.ExpiresAt,
		}, nil
}

func (s *userService) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (dto.Session, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return dto.Session{}, err
	}
	defer tx.Rollback(ctx)

	hashedRefreshToken := user.HashRefreshToken(refreshToken)
	session, err := s.userRepo.GetSessionByRefreshToken(ctx, hashedRefreshToken)
	if err != nil {
		s.log.Error(ctx, "Failed to get session", err)
		return dto.Session{}, err
	}

	err = user.VerifySession(session)
	if err != nil {
		s.log.Error(ctx, "Failed to verify session", err)
		return dto.Session{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.Session{}, err
	}

	return session, nil
}
