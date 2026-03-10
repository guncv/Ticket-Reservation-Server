package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/infras/log"
	"github.com/guncv/ticket-reservation-server/internal/services/user"
)

type UserHandler struct {
	userService user.UserService
	log         *log.Logger
	config      *config.Config
}

func NewUserHandler(
	l *log.Logger,
	userService user.UserService,
	config *config.Config,
) *UserHandler {
	return &UserHandler{
		log:         l,
		userService: userService,
		config:      config,
	}
}

func (h *UserHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	res, err := h.userService.HealthCheck(ctx)
	if err != nil {
		h.log.ErrorWithID(ctx, "[Handler: HealthCheck] Error checking health", err)
		return
	}

	c.JSON(http.StatusOK, res)
}
