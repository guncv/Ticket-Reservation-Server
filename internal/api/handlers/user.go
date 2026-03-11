package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/domain/user"
	cookies "github.com/guncv/ticket-reservation-server/internal/infra/http"
)

type UserHandler struct {
	config  *config.Config
	userSrv user.UserService
	cookies cookies.Cookies
}

func NewUserHandler(
	config *config.Config,
	userSrv user.UserService,
	cookies cookies.Cookies,
) *UserHandler {
	return &UserHandler{
		config:  config,
		userSrv: userSrv,
		cookies: cookies,
	}
}

func (h *UserHandler) HealthCheck(c *gin.Context) {
	result, err := h.userSrv.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
