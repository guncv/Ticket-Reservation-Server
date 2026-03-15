package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	cookies "github.com/guncv/ticket-reservation-server/internal/infra/http"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/service/user"
	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
)

type UserHandler struct {
	userSrv user.UserService
	cookies cookies.Cookies
	log     log.Logger
}

func NewUserHandler(
	userSrv user.UserService,
	cookies cookies.Cookies,
	log log.Logger,
) *UserHandler {
	return &UserHandler{
		userSrv: userSrv,
		cookies: cookies,
		log:     log,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userResp, err := h.userSrv.CreateUser(c.Request.Context(), req)
	if err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cookies.SetRefreshTokenCookie(c, userResp.RefreshToken)

	c.JSON(http.StatusOK, userResp)
}

func (h *UserHandler) LoginUser(c *gin.Context) {
	var req dto.LoginUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userResp, err := h.userSrv.LoginUser(c.Request.Context(), req)
	if err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cookies.SetRefreshTokenCookie(c, userResp.RefreshToken)

	c.JSON(http.StatusOK, userResp)
}

func (h *UserHandler) LogoutUser(c *gin.Context) {
	err := h.userSrv.LogoutUser(c.Request.Context())
	if err != nil {
		h.log.Error(c.Request.Context(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cookies.ClearRefreshTokenCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
