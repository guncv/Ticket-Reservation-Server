package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/user/token"
	"github.com/guncv/ticket-reservation-server/internal/shared"
	"github.com/guncv/ticket-reservation-server/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("APP_ENV", shared.AppEnvTest)
	t.Setenv("REFRESH_TOKEN_DURATION", "24h")

	cfg, err := config.LoadConfig(nil)
	require.NoError(t, err)

	logger := log.NewLogger(cfg)

	testCases := []struct {
		name   string
		setup  func(t *testing.T, mockSvc *mocks.MockUserService)
		req    func(t *testing.T) *http.Request
		verify func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success_PopulatesContextAndCallsNext",
			setup: func(t *testing.T, mockSvc *mocks.MockUserService) {
				expectedUserID := "user-123"
				expectedRefreshToken := "valid-refresh-token"
				expectedAccessToken := "valid-access-token"

				mockSvc.On("VerifyAndRenewToken", mock.Anything, mock.MatchedBy(func(r dto.SessionReq) bool {
					return r.RefreshToken == expectedRefreshToken
				})).Return(dto.SessionResp{
					AccessToken:  expectedAccessToken,
					RefreshToken: expectedRefreshToken,
				}, token.TokenPayload{
					UserID:    expectedUserID,
					IssuedAt:  time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil)
			},
			req: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set(shared.AuthorizationHeaderKey, "Bearer valid-access-token")
				req.AddCookie(&http.Cookie{Name: string(shared.RefreshTokenCookieKey), Value: "valid-refresh-token"})
				return req
			},
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, w.Code)
				require.Equal(t, "user-123", strings.TrimSpace(w.Body.String()))
			},
		},
		{
			name: "Success_SetsXAccessTokenHeaderWhenTokenRenewed",
			setup: func(t *testing.T, mockSvc *mocks.MockUserService) {
				newAccessToken := "new-access-token-after-renewal"
				refreshToken := "refresh-token"

				mockSvc.On("VerifyAndRenewToken", mock.Anything, mock.Anything).Return(dto.SessionResp{
					AccessToken:  newAccessToken,
					RefreshToken: refreshToken,
				}, token.TokenPayload{
					UserID:    "user-456",
					IssuedAt:  time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil)
			},
			req: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set(shared.AuthorizationHeaderKey, "Bearer expired-access-token")
				req.AddCookie(&http.Cookie{Name: string(shared.RefreshTokenCookieKey), Value: "refresh-token"})
				return req
			},
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, w.Code)
				require.Equal(t, "new-access-token-after-renewal", w.Header().Get(shared.XAccessTokenHeaderKey))
			},
		},
		{
			name: "Error_NoAuthorizationHeader",
			setup: func(t *testing.T, mockSvc *mocks.MockUserService) {
				// no setup - VerifyAndRenewToken should not be called
			},
			req: func(t *testing.T) *http.Request {
				return httptest.NewRequest(http.MethodGet, "/protected", nil)
			},
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name:  "Error_InvalidAuthorizationFormat",
			setup: func(t *testing.T, mockSvc *mocks.MockUserService) {},
			req: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set(shared.AuthorizationHeaderKey, "invalid-no-bearer")
				return req
			},
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name:  "Error_NotBearerType",
			setup: func(t *testing.T, mockSvc *mocks.MockUserService) {},
			req: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set(shared.AuthorizationHeaderKey, "Basic some-token")
				return req
			},
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name:  "Error_NoRefreshTokenCookie",
			setup: func(t *testing.T, mockSvc *mocks.MockUserService) {},
			req: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set(shared.AuthorizationHeaderKey, "Bearer some-access-token")
				return req
			},
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name: "Error_VerifyAndRenewTokenFails",
			setup: func(t *testing.T, mockSvc *mocks.MockUserService) {
				mockSvc.On("VerifyAndRenewToken", mock.Anything, mock.MatchedBy(func(r dto.SessionReq) bool {
					return r.AccessToken == "access-token" && r.RefreshToken == "refresh-token"
				})).Return(dto.SessionResp{}, token.TokenPayload{}, errors.New("invalid access token"))
			},
			req: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set(shared.AuthorizationHeaderKey, "Bearer access-token")
				req.AddCookie(&http.Cookie{Name: string(shared.RefreshTokenCookieKey), Value: "refresh-token"})
				return req
			},
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mocks.NewMockUserService(t)
			tc.setup(t, mockSvc)

			authMw := NewAuthMiddleware(logger, cfg, mockSvc)

			router := gin.New()
			router.GET("/protected", authMw.AuthMiddleware(), func(c *gin.Context) {
				userID, _ := shared.GetUserIDFromContext(c.Request.Context())
				c.String(http.StatusOK, userID)
			})

			w := httptest.NewRecorder()
			req := tc.req(t)
			router.ServeHTTP(w, req)

			tc.verify(t, w)
		})
	}
}
