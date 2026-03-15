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
		eventHandler *handlers.EventHandler,
		authMiddleware AuthMiddleware,
	) {
		api_v1 := e.Group("/api/v1")
		userRoutes(api_v1, userHandler, authMiddleware)
		eventRoutes(api_v1, eventHandler, authMiddleware)
	}); err != nil {
		panic(err)
	}
}

func userRoutes(api_v1 *gin.RouterGroup, userHandler *handlers.UserHandler, authMiddleware AuthMiddleware) {
	userRoutes := api_v1.Group("/user")

	userRoutes.POST("/register", userHandler.CreateUser)
	userRoutes.POST("/login", userHandler.LoginUser)

	userMiddleware := authMiddleware.AuthMiddleware()
	userRoutes.POST("/logout", userMiddleware, userHandler.LogoutUser)
}

func eventRoutes(api_v1 *gin.RouterGroup, eventHandler *handlers.EventHandler, authMiddleware AuthMiddleware) {
	eventRoutes := api_v1.Group("/event")

	eventMiddleware := authMiddleware.AuthMiddleware()
	eventRoutes.POST("/", eventMiddleware, eventHandler.CreateEvent)
	eventRoutes.PUT("/:id", eventMiddleware, eventHandler.UpdateEvent)
	eventRoutes.GET("/", eventMiddleware, eventHandler.GetAllEvents)
	eventRoutes.GET("/reservations", eventMiddleware, eventHandler.GetAllReservations)
	eventRoutes.POST("/ticket", eventMiddleware, eventHandler.ReserveEventTicket)
}
