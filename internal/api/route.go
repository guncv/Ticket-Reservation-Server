package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/api/handlers"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"go.uber.org/dig"
)

func RegisterRoutes(e *gin.Engine, c *dig.Container, cfg *config.Config) {

	corsOrigins := cfg.AppConfig.CORSOrigins
	if len(corsOrigins) == 0 {
		corsOrigins = []string{"http://localhost:3000"}
	}

	e.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Active-Role"},
		ExposeHeaders:    []string{"Content-Length", "Authorization", "Access-Control-Expose-Headers", "X-Access-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	e.Use(InjectRequestMetadata())

	if err := c.Invoke(func(
		userHandler *handlers.UserHandler,
		authMiddleware AuthMiddleware,
	) {
		api_v1 := e.Group("/api/v1")
		userRoutes(api_v1, userHandler, authMiddleware)
	}); err != nil {
		panic(err)
	}
}

func userRoutes(api_v1 *gin.RouterGroup, userHandler *handlers.UserHandler, authMiddleware AuthMiddleware) {
	userRoutes := api_v1.Group("/user")

	userRoutes.GET("/health", userHandler.HealthCheck)
	userRoutes.POST("/register", userHandler.CreateUser)
}
