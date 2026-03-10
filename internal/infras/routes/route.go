package routes

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/handlers"
	"go.uber.org/dig"
)

func RegisterRoutes(e *gin.Engine, c *dig.Container, cfg *config.Config) {

	e.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Active-Role"},
		ExposeHeaders:    []string{"Content-Length", "Authorization", "Access-Control-Expose-Headers", "X-Access-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	e.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	e.RedirectTrailingSlash = false

	e.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "interview-backend-server",
		})
	})

	if err := c.Invoke(func(
		userHandler *handlers.UserHandler,
	) {
		api_v1 := e.Group("/api/v1")
		userRoutes(api_v1, userHandler)
	}); err != nil {
		panic(err)
	}
}

func userRoutes(eg *gin.RouterGroup, userHandler *handlers.UserHandler) {
	userRoutes := eg.Group("/auth")

	{
		userRoutes.GET("/health", userHandler.HealthCheck)
	}
}
