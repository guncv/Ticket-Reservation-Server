package shared

import (
	"context"
	"errors"
)

func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user ID not found in context")
	}

	return userID, nil
}

func GetRefreshTokenFromContext(ctx context.Context) (string, error) {
	refreshToken, ok := ctx.Value(RefreshTokenKey).(string)
	if !ok || refreshToken == "" {
		return "", errors.New("refresh token not found in context")
	}

	return refreshToken, nil
}
