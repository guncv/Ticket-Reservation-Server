package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	t.Setenv("ACCESS_TOKEN_DURATION", "1h")
	t.Setenv("REFRESH_TOKEN_DURATION", "24h")

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
		req    func(t *testing.T) dto.SessionReq
		setup  func(t *testing.T, userService user.UserService, pool *db.PgPool)
		verify func(t *testing.T, actual dto.SessionResp, payload token.TokenPayload, actualErr error)
	}{
		{
			name: "Success_AccessTokenValid",
			req: func(t *testing.T) dto.SessionReq {
				createResp, _ := userService.CreateUser(ctx, dto.CreateUserReq{
					UserName: "renew_valid_user",
					Password: "renew_valid_pass",
				})
				return dto.SessionReq{
					AccessToken:  createResp.AccessToken,
					RefreshToken: createResp.RefreshToken,
				}
			},
			setup: nil,
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
			req: func(t *testing.T) dto.SessionReq {
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
			},
		},
		{
			name: "Error_SessionNotFound",
			req: func(t *testing.T) dto.SessionReq {
				userService.CreateUser(ctx, dto.CreateUserReq{
					UserName: "renew_notfound_user",
					Password: "renew_notfound_pass",
				})
				return dto.SessionReq{
					AccessToken:  mustCreateExpiredAccessToken(t),
					RefreshToken: "fake_refresh_token_that_matches_no_session",
				}
			},
			setup: nil,
			verify: func(t *testing.T, actual dto.SessionResp, payload token.TokenPayload, actualErr error) {
				require.Error(t, actualErr)
			},
		},
		{
			name: "Error_AccessAndRefreshTokenDifferentUserID",
			req: func(t *testing.T) dto.SessionReq {
				createRespA, _ := userService.CreateUser(ctx, dto.CreateUserReq{
					UserName: "renew_user_a",
					Password: "renew_pass_a",
				})
				createRespB, _ := userService.CreateUser(ctx, dto.CreateUserReq{
					UserName: "renew_user_b",
					Password: "renew_pass_b",
				})
				userAID := mustExtractUserIDFromToken(t, createRespA.AccessToken)
				return dto.SessionReq{
					AccessToken:  mustCreateExpiredAccessTokenForUser(t, userAID),
					RefreshToken: createRespB.RefreshToken,
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

			req := tc.req(t)
			actual, payload, actualErr := userService.VerifyAndRenewToken(ctx, req)
			tc.verify(t, actual, payload, actualErr)

			err = test.TruncateAllTables()
			require.NoError(t, err)
		})
	}
}

func mustCreateExpiredAccessToken(t *testing.T) string {
	t.Helper()
	return mustCreateExpiredAccessTokenForUser(t, "")
}

func mustCreateExpiredAccessTokenForUser(t *testing.T, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(shared.TestTokenSecretKey))
	require.NoError(t, err)
	return signed
}

func mustExtractUserIDFromToken(t *testing.T, tokenString string) string {
	t.Helper()
	tok, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	require.NoError(t, err)
	claims, ok := tok.Claims.(jwt.MapClaims)
	require.True(t, ok)
	userID, ok := claims["user_id"].(string)
	require.True(t, ok)
	return userID
}
