package user

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/domain/user/dto"
)

func (s *userService) HealthCheck(ctx context.Context) (*dto.HealthCheckResp, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result, err := s.userRepo.HealthCheck(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.HealthCheckResp{Status: result}, nil
}
