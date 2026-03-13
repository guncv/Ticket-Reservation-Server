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
