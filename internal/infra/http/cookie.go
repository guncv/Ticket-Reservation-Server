package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/shared"
)

type Cookies interface {
	SetRefreshTokenCookie(c *gin.Context, refreshToken string)
	ClearRefreshTokenCookie(c *gin.Context)
}

type cookies struct {
	config *config.Config
}

func NewCookies(config *config.Config) Cookies {
	return &cookies{config: config}
}

func (c *cookies) SetRefreshTokenCookie(ctx *gin.Context, refreshToken string) {
	duration := c.config.AuthConfig.RefreshTokenDuration
	domain := c.config.AuthConfig.CookieDomain
	expiration := time.Now().Add(duration)
	secure := c.config.AppConfig.AppEnv != shared.AppEnvDev
	c.setCookie(ctx, refreshToken, expiration, domain, secure)
}

func (c *cookies) ClearRefreshTokenCookie(ctx *gin.Context) {
	expiration := time.Now().Add(-24 * time.Hour)
	domain := c.config.AuthConfig.CookieDomain
	secure := c.config.AppConfig.AppEnv != shared.AppEnvDev
	c.setCookie(ctx, "", expiration, domain, secure)
}

func (c *cookies) setCookie(ctx *gin.Context, value string, expires time.Time, domain string, secure bool) {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     string(shared.RefreshTokenCookieKey),
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Domain:   domain,
		Expires:  expires,
	})
}
