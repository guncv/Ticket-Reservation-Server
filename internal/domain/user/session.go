package user

import (
	"context"

	"cloud.google.com/go/civil"
	"github.com/guncv/ticket-reservation-server/internal/domain/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/domain/user/repo"
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
		s.logger.Error(ctx, "Failed to generate access token", err)
		return dto.CreateSessionResp{}, err
	}

	refreshToken, expiresAt, err := s.token.GenerateRefreshToken(userID)
	if err != nil {
		s.logger.Error(ctx, "Failed to generate refresh token", err)
		return dto.CreateSessionResp{}, err
	}

	userAgent := ctx.Value(shared.UserAgentKey).(string)
	ipAddress := ctx.Value(shared.ClientIPKey).(string)

	hashedRefreshToken := shared.HashRefreshToken(refreshToken)

	session := repo.CreateSessionParams{
		UserID:             userID,
		HashedRefreshToken: hashedRefreshToken,
		UserAgent:          userAgent,
		IPAddress:          ipAddress,
		ExpiresAt:          civil.TimeOf(expiresAt),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		s.logger.Error(ctx, "Failed to create session", err)
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

// func (s *userService) VerifyAndRenewToken(ctx context.Context, req *dto.SessionReq) (*dto.SessionResult, error) {
// 	hashedRefreshToken := shared.HashRefreshToken(req.RefreshToken)
// 	session, err := s.userRepo.GetSession(ctx, hashedRefreshToken)
// 	if err != nil {
// 		s.logger.Error(ctx, "Failed to get session", err)
// 		return nil, err
// 	}

// 	if session.IsRevoked {
// 		s.logger.Error(ctx, "Session is revoked")
// 		return nil, errors.New("session is revoked")
// 	}

// 	_, err = s.token.VerifyToken(req.AccessToken, session.UserID)
// 	if err == nil {
// 		return &dto.SessionResult{
// 			UserID:       session.UserID,
// 			AccessToken:  req.AccessToken,
// 			RefreshToken: req.RefreshToken,
// 		}, nil
// 	}

// 	if !errors.Is(err, token.ErrTokenExpired) {
// 		s.logger.Error(ctx, "Failed to verify access token", err)
// 		return nil, errors.New("invalid access token")
// 	}

// 	refreshTokenPayload, err := s.token.VerifyToken(req.RefreshToken, session.UserID)
// 	if err != nil {
// 		s.logger.Error(ctx, "Failed to verify refresh token", err)
// 		return nil, errors.New("invalid refresh token")
// 	}

// 	newAccessToken, err := s.token.GenerateAccessToken(refreshTokenPayload.UserID)
// 	if err != nil {
// 		s.logger.Error(ctx, "Failed to generate new access token", err)
// 		return nil, err
// 	}

// 	return &dto.SessionResult{
// 		UserID:       refreshTokenPayload.UserID,
// 		AccessToken:  newAccessToken,
// 		RefreshToken: req.RefreshToken,
// 	}, nil
// }

// func (s *userService) RevokeSession(ctx context.Context, refreshToken string) error {
// 	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback(ctx)

// 	hashedRefreshToken := shared.HashRefreshToken(refreshToken)
// 	if err := s.userRepo.RevokeSession(ctx, hashedRefreshToken); err != nil {
// 		return err
// 	}

// 	if err := tx.Commit(ctx); err != nil {
// 		return err
// 	}

// 	return nil
// }
