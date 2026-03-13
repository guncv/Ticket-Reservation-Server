package repo

import (
	"context"
	"fmt"
)

func (r *userRepository) HealthCheck(ctx context.Context) (string, error) {
	return "ok", nil
}

type CreateUserParams struct {
	UserName       string
	HashedPassword string
	Role           string
}

func (r *userRepository) CreateUser(ctx context.Context, params CreateUserParams) (string, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	query := `
		INSERT INTO users (user_name, hashed_password, role)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var userID string
	err = conn.QueryRow(ctx, query, params.UserName, params.HashedPassword, params.Role).Scan(&userID)
	if err != nil {
		r.log.Error(ctx, "Failed to create user", err)
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (r *userRepository) CheckUserNameExists(ctx context.Context, userName string) (bool, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer conn.Release()

	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE user_name = $1)
	`

	var exists bool
	err = conn.QueryRow(ctx, query, userName).Scan(&exists)
	if err != nil {
		r.log.Error(ctx, "Failed to check user name exists", err)
		return false, fmt.Errorf("failed to check user name exists: %w", err)
	}

	return exists, nil
}

func (r *userRepository) GetUserByUserName(ctx context.Context, userName string) (User, error) {
	ctx, conn, err := r.db.EnsureConnFromCtx(ctx)
	if err != nil {
		return User{}, fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()

	query := `
		SELECT id, user_name, hashed_password, role
		FROM users
		WHERE user_name = $1
	`

	var u User
	err = conn.QueryRow(ctx, query, userName).Scan(&u.ID, &u.UserName, &u.HashedPassword, &u.Role)
	if err != nil {
		r.log.Error(ctx, "Failed to get user by user name", err)
		return User{}, fmt.Errorf("failed to get user by user name: %w", err)
	}

	return u, nil
}
