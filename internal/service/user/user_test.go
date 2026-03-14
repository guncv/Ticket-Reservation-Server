package user_test

import (
	"context"
	"testing"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/containers"
	"github.com/guncv/ticket-reservation-server/internal/infra/test"
	"github.com/guncv/ticket-reservation-server/internal/service/user"
	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/shared"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	t.Setenv("APP_ENV", shared.AppEnvTest)

	cfg, err := config.LoadConfig(nil)
	require.NoError(t, err)

	container := containers.NewContainer(cfg)
	require.NoError(t, container.Error)

	var userService user.UserService

	err = container.Container.Invoke(func(
		us user.UserService,
	) {
		userService = us
	})
	require.NoError(t, err)

	testCases := []struct {
		name   string
		req    func() dto.CreateUserReq
		verify func(t *testing.T, actual dto.CreateUserResp, actualErr error)
	}{
		{
			name: "Success",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: "test_user",
					Password: "test_password",
				}
			},
			verify: func(t *testing.T, actual dto.CreateUserResp, actualErr error) {
				require.NoError(t, actualErr)
				require.NotEmpty(t, actual.AccessToken)
				require.NotEmpty(t, actual.RefreshToken)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, actualErr := userService.CreateUser(context.Background(), tc.req())
			tc.verify(t, actual, actualErr)

			err = test.TruncateAllTables()
			require.NoError(t, err)
		})
	}
}

func TestLoginUser(t *testing.T) {
	t.Setenv("APP_ENV", shared.AppEnvTest)
	t.Setenv("REFRESH_TOKEN_DURATION", "24h")

	cfg, err := config.LoadConfig(nil)
	require.NoError(t, err)

	container := containers.NewContainer(cfg)
	require.NoError(t, container.Error)

	var userService user.UserService

	err = container.Container.Invoke(func(us user.UserService) {
		userService = us
	})
	require.NoError(t, err)

	testCases := []struct {
		name   string
		req    func() dto.LoginUserReq
		setup  func(t *testing.T, userService user.UserService)
		verify func(t *testing.T, actual dto.LoginUserResp, actualErr error)
	}{
		{
			name: "Success",
			req: func() dto.LoginUserReq {
				return dto.LoginUserReq{
					UserName: "login_test_user",
					Password: "login_test_pass",
				}
			},
			setup: func(t *testing.T, userService user.UserService) {
				_, err := userService.CreateUser(context.Background(), dto.CreateUserReq{
					UserName: "login_test_user",
					Password: "login_test_pass",
				})
				require.NoError(t, err)
			},
			verify: func(t *testing.T, actual dto.LoginUserResp, actualErr error) {
				require.NoError(t, actualErr)
				require.NotEmpty(t, actual.AccessToken)
				require.NotEmpty(t, actual.RefreshToken)
			},
		},
		{
			name: "Error_WrongPassword",
			req: func() dto.LoginUserReq {
				return dto.LoginUserReq{
					UserName: "wrong_pass_user",
					Password: "wrong_password",
				}
			},
			setup: func(t *testing.T, userService user.UserService) {
				_, err := userService.CreateUser(context.Background(), dto.CreateUserReq{
					UserName: "wrong_pass_user",
					Password: "correct_password",
				})
				require.NoError(t, err)
			},
			verify: func(t *testing.T, actual dto.LoginUserResp, actualErr error) {
				require.Error(t, actualErr)
				require.Empty(t, actual.AccessToken)
				require.Empty(t, actual.RefreshToken)
			},
		},
		{
			name: "Error_UserNotFound",
			req: func() dto.LoginUserReq {
				return dto.LoginUserReq{
					UserName: "nonexistent_user",
					Password: "any_password",
				}
			},
			setup: func(t *testing.T, userService user.UserService) {
				// No setup - user does not exist
			},
			verify: func(t *testing.T, actual dto.LoginUserResp, actualErr error) {
				require.Error(t, actualErr)
				require.Empty(t, actual.AccessToken)
				require.Empty(t, actual.RefreshToken)
			},
		},
		{
			name: "Success_LoginAfterSessionRevoked",
			req: func() dto.LoginUserReq {
				return dto.LoginUserReq{
					UserName: "revoke_then_login_user",
					Password: "revoke_then_login_pass",
				}
			},
			setup: func(t *testing.T, userService user.UserService) {
				createResp, err := userService.CreateUser(context.Background(), dto.CreateUserReq{
					UserName: "revoke_then_login_user",
					Password: "revoke_then_login_pass",
				})
				require.NoError(t, err)
				ctx := context.WithValue(context.Background(), shared.RefreshTokenKey, createResp.RefreshToken)
				err = userService.LogoutUser(ctx)
				require.NoError(t, err)
			},
			verify: func(t *testing.T, actual dto.LoginUserResp, actualErr error) {
				require.NoError(t, actualErr)
				require.NotEmpty(t, actual.AccessToken)
				require.NotEmpty(t, actual.RefreshToken)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(t, userService)
			}

			actual, actualErr := userService.LoginUser(context.Background(), tc.req())
			tc.verify(t, actual, actualErr)

			err = test.TruncateAllTables()
			require.NoError(t, err)
		})
	}
}

func TestLogoutUser(t *testing.T) {
	t.Setenv("APP_ENV", shared.AppEnvTest)
	t.Setenv("REFRESH_TOKEN_DURATION", "24h")

	cfg, err := config.LoadConfig(nil)
	require.NoError(t, err)

	container := containers.NewContainer(cfg)
	require.NoError(t, container.Error)

	var userService user.UserService

	err = container.Container.Invoke(func(us user.UserService) {
		userService = us
	})
	require.NoError(t, err)

	testCases := []struct {
		name   string
		setup  func(t *testing.T, userService user.UserService) context.Context
		verify func(t *testing.T, actualErr error)
	}{
		{
			name: "Success",
			setup: func(t *testing.T, userService user.UserService) context.Context {
				resp, err := userService.CreateUser(context.Background(), dto.CreateUserReq{
					UserName: "logout_test_user",
					Password: "logout_test_pass",
				})
				require.NoError(t, err)
				return context.WithValue(context.Background(), shared.RefreshTokenKey, resp.RefreshToken)
			},
			verify: func(t *testing.T, actualErr error) {
				require.NoError(t, actualErr)
			},
		},
		{
			name: "Error_RefreshTokenNotFoundInContext",
			setup: func(t *testing.T, userService user.UserService) context.Context {
				return context.Background()
			},
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "refresh token not found")
			},
		},
		{
			name: "Error_SessionAlreadyRevoked",
			setup: func(t *testing.T, userService user.UserService) context.Context {
				resp, err := userService.CreateUser(context.Background(), dto.CreateUserReq{
					UserName: "revoked_twice_user",
					Password: "revoked_twice_pass",
				})
				require.NoError(t, err)
				ctx := context.WithValue(context.Background(), shared.RefreshTokenKey, resp.RefreshToken)
				require.NoError(t, userService.LogoutUser(ctx))
				return ctx
			},
			verify: func(t *testing.T, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "session is revoked")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setup(t, userService)
			actualErr := userService.LogoutUser(ctx)
			tc.verify(t, actualErr)

			err = test.TruncateAllTables()
			require.NoError(t, err)
		})
	}
}
