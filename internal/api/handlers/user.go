package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/config"
	cookies "github.com/guncv/ticket-reservation-server/internal/infra/http"
	"github.com/guncv/ticket-reservation-server/internal/service/user"
	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
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

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userResp, err := h.userSrv.CreateUser(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cookies.SetRefreshTokenCookie(c, userResp.RefreshToken)

	c.JSON(http.StatusOK, userResp)
}

func (h *UserHandler) LoginUser(c *gin.Context) {
	var req dto.LoginUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userResp, err := h.userSrv.LoginUser(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cookies.SetRefreshTokenCookie(c, userResp.RefreshToken)

	c.JSON(http.StatusOK, userResp)
}

func (h *UserHandler) LogoutUser(c *gin.Context) {
	err := h.userSrv.LogoutUser(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cookies.ClearRefreshTokenCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
