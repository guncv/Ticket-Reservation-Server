package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/service/user"
	"github.com/guncv/ticket-reservation-server/internal/shared"
)

type AuthMiddleware interface {
	AuthMiddleware() gin.HandlerFunc
}

type authMiddleware struct {
	log            log.Logger
	cfg            *config.Config
	sessionService user.UserService
}

func NewAuthMiddleware(
	log log.Logger,
	cfg *config.Config,
	sessionService user.UserService,
) AuthMiddleware {
	return &authMiddleware{
		log:            log,
		cfg:            cfg,
		sessionService: sessionService,
	}
}

func (m *authMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(shared.AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errors.New("authorization header is not provided"))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errors.New("invalid authorization header format"))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != shared.AuthorizationTypeBearer {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errors.New("authorization header must start with "+shared.AuthorizationTypeBearer))
			return
		}

		// cookie, err := ctx.Request.Cookie(string(shared.RefreshTokenCookieKey))
		// if err != nil {
		// 	ctx.AbortWithStatusJSON(http.StatusUnauthorized, errors.New("refresh token cookie not found"))
		// 	return
		// }

		// sessionReq := &dto.SessionReq{
		// 	AccessToken:  fields[1],
		// 	RefreshToken: cookie.Value,
		// }

		// sessionResult, err := m.sessionService.VerifyAndRenewToken(ctx, sessionReq)
		// if err != nil {
		// 	ctx.AbortWithStatusJSON(http.StatusUnauthorized, err)
		// 	return
		// }

		// if sessionResult.AccessToken != sessionReq.AccessToken {
		// 	ctx.Header(shared.XAccessTokenHeaderKey, sessionResult.AccessToken)
		// }

		// reqCtx := context.WithValue(ctx.Request.Context(), shared.UserIDKey, sessionResult.UserID)
		// ctx.Request = ctx.Request.WithContext(reqCtx)

		ctx.Next()
	}
}

func InjectRequestMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), shared.UserAgentKey, c.Request.UserAgent())
		ctx = context.WithValue(ctx, shared.ClientIPKey, c.ClientIP())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
