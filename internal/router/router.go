package router

import (
	"github.com/gin-gonic/gin"
	"github.com/intern/inventory-api/internal/config"
	"github.com/intern/inventory-api/internal/handler"
	"github.com/intern/inventory-api/internal/middleware"
	"github.com/intern/inventory-api/internal/pkg/token"
)

type Handlers struct {
	Auth   *handler.AuthHandler
	Item   *handler.ItemHandler
	Health *handler.HealthHandler
}

func SetupRouter(cfg *config.Config, tokenManager *token.Manager, h *Handlers) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	r.GET("/health", h.Health.Check)

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", h.Auth.Register)
		authGroup.POST("/login", h.Auth.Login)
	}

	itemsGroup := r.Group("/items")
	itemsGroup.Use(middleware.Auth(tokenManager))
	{
		itemsGroup.POST("", h.Item.Create)
		itemsGroup.GET("", h.Item.List)
		itemsGroup.GET("/:id", h.Item.Get)
		itemsGroup.PUT("/:id", h.Item.Update)
		itemsGroup.DELETE("/:id", h.Item.Delete)
		itemsGroup.PATCH("/:id/stock", h.Item.AdjustStock)
	}

	return r
}

