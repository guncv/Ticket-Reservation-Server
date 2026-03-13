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

func (s *userService) CreateSession(ctx context.Context, userID string) (dto.CreateSessionResp, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return dto.CreateSessionResp{}, err
	}
	defer tx.Rollback(ctx)

	accessToken, err := s.token.GenerateAccessToken(userID)
	if err != nil {
		s.log.Error(ctx, "Failed to generate access token", err)
		return dto.CreateSessionResp{}, err
	}

	refreshToken, expiresAt, err := s.token.GenerateRefreshToken()
	if err != nil {
		s.log.Error(ctx, "Failed to generate refresh token", err)
		return dto.CreateSessionResp{}, err
	}

	userAgent, _ := ctx.Value(shared.UserAgentKey).(string)
	ipAddress, _ := ctx.Value(shared.ClientIPKey).(string)

	hashedRefreshToken := shared.HashRefreshToken(refreshToken)

	session := repo.CreateSessionParams{
		UserID:             userID,
		HashedRefreshToken: hashedRefreshToken,
		UserAgent:          userAgent,
		IPAddress:          ipAddress,
		ExpiresAt:          expiresAt,
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		s.log.Error(ctx, "Failed to create session", err)
		return dto.CreateSessionResp{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.CreateSessionResp{}, err
	}

	return dto.CreateSessionResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *userService) VerifyAndRenewToken(ctx context.Context, req dto.RenewTokenReq) (dto.RenewTokenResp, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return dto.RenewTokenResp{}, err
	}
	defer tx.Rollback(ctx)

	_, err = s.token.VerifyAccessToken(req.AccessToken)
	if err == nil {
		return dto.RenewTokenResp{
			AccessToken:  req.AccessToken,
			RefreshToken: req.RefreshToken,
		}, nil
	}

	if !errors.Is(err, token.ErrTokenExpired) {
		s.log.Error(ctx, "Invalid access token", err)
		return dto.RenewTokenResp{}, errors.New("invalid access token")
	}

	hashedRefreshToken := shared.HashRefreshToken(req.RefreshToken)
	session, err := s.userRepo.GetSessionByRefreshToken(ctx, hashedRefreshToken)
	if err != nil {
		s.log.Error(ctx, "Failed to get session", err)
		return dto.RenewTokenResp{}, err
	}

	err = user.VerifySession(session)
	if err != nil {
		s.log.Error(ctx, "Failed to verify session", err)
		return dto.RenewTokenResp{}, err
	}

	newAccessToken, err := s.token.GenerateAccessToken(session.UserID)
	if err != nil {
		s.log.Error(ctx, "Failed to generate new access token", err)
		return dto.RenewTokenResp{}, err
	}

	return dto.RenewTokenResp{
		AccessToken:  newAccessToken,
		RefreshToken: req.RefreshToken,
	}, nil
}
