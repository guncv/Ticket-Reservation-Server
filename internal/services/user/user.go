package user

import (
	"context"
	"errors"

	"github.com/guncv/ticket-reservation-server/internal/services/user/dto"
)

func (s *userService) HealthCheck(ctx context.Context) (dto.HealthCheckRes, error) {

	res, err := s.userRepo.HealthCheck(ctx)
	if err != nil {
		return dto.HealthCheckRes{}, errors.New(err.Error())
	}

	result := dto.HealthCheckRes{
		Status: res,
	}

	return result, nil
}
