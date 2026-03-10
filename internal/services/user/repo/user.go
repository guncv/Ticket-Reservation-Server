package repo

import (
	"context"
)

func (r *userRepository) HealthCheck(ctx context.Context) (string, error) {
	return "Status OK", nil
}
