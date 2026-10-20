package routes

import (
	"github.com/gin-gonic/gin"
	"huydq2801.com/golearn/internal/middleware"
)

type Route interface {
	RegisterRoutes(r *gin.RouterGroup)
}

func RegisterRoutes(r *gin.Engine, routes ...Route) {
	r.Use(
		middleware.LoggerMiddleware(),
		middleware.ApiKeyMiddleware(),
		middleware.AuthMiddleware(),
		middleware.RateLimiterMiddleware(),
	)
	api := r.Group("/api/v1")
	for _, route := range routes {
		route.RegisterRoutes(api)
	}
}
