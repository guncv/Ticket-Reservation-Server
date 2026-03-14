package user_test

import (
	"context"
	"testing"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/containers"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/guncv/ticket-reservation-server/internal/infra/test"
	"github.com/guncv/ticket-reservation-server/internal/service/user"
	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/user/token"
	"github.com/guncv/ticket-reservation-server/internal/shared"
	"github.com/stretchr/testify/require"
)

func TestVerifyAndRenewToken(t *testing.T) {
	t.Setenv("APP_ENV", shared.AppEnvTest)

	cfg, err := config.LoadConfig(nil)
	require.NoError(t, err)

	container := containers.NewContainer(cfg)
	require.NoError(t, container.Error)

	var userService user.UserService
	var pool *db.PgPool

	err = container.Container.Invoke(func(us user.UserService, p *db.PgPool) {
		userService = us
		pool = p
	})
	require.NoError(t, err)

	ctx := context.Background()

	testCases := []struct {
		name   string
		req    func() dto.SessionReq
		setup  func(t *testing.T, userService user.UserService, pool *db.PgPool)
		verify func(t *testing.T, actual dto.SessionResp, payload token.TokenPayload, actualErr error)
	}{
		{
			name: "Success_AccessTokenValid",
			req: func() dto.SessionReq {
				createResp, _ := userService.CreateUser(ctx, dto.CreateUserReq{
					UserName: "renew_valid_user",
					Password: "renew_valid_pass",
				})
				return dto.SessionReq{
					AccessToken:  createResp.AccessToken,
					RefreshToken: createResp.RefreshToken,
				}
			},
			setup: nil, // req creates user and returns tokens
			verify: func(t *testing.T, actual dto.SessionResp, payload token.TokenPayload, actualErr error) {
				require.NoError(t, actualErr)
				require.NotEmpty(t, actual.AccessToken)
				require.NotEmpty(t, actual.RefreshToken)
				require.NotEmpty(t, payload.UserID)
				require.NotEmpty(t, payload.IssuedAt)
				require.NotEmpty(t, payload.ExpiresAt)
			},
		},
		{
			name: "Error_InvalidAccessToken",
			req: func() dto.SessionReq {
				createResp, _ := userService.CreateUser(ctx, dto.CreateUserReq{
					UserName: "renew_invalid_at_user",
					Password: "renew_invalid_at_pass",
				})
				return dto.SessionReq{
					AccessToken:  "invalid.jwt.token",
					RefreshToken: createResp.RefreshToken,
				}
			},
			setup: nil,
			verify: func(t *testing.T, actual dto.SessionResp, payload token.TokenPayload, actualErr error) {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), "invalid access token")
			},
		},
		{
			name: "Error_SessionNotFound",
			req: func() dto.SessionReq {
				createResp, _ := userService.CreateUser(ctx, dto.CreateUserReq{
					UserName: "renew_notfound_user",
					Password: "renew_notfound_pass",
				})
				return dto.SessionReq{
					AccessToken:  createResp.AccessToken,
					RefreshToken: "fake_refresh_token_that_matches_no_session",
				}
			},
			setup: nil,
			verify: func(t *testing.T, actual dto.SessionResp, payload token.TokenPayload, actualErr error) {
				require.Error(t, actualErr)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(t, userService, pool)
			}

			req := tc.req()
			actual, payload, actualErr := userService.VerifyAndRenewToken(ctx, req)
			tc.verify(t, actual, payload, actualErr)

			err = test.TruncateAllTables()
			require.NoError(t, err)
		})
	}
}
