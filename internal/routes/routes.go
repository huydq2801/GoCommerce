package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"huydq2801.com/golearn/internal/middleware"
	"huydq2801.com/golearn/internal/utils"
	"huydq2801.com/golearn/pkg/logger"
)

type Route interface {
	RegisterRoutes(r *gin.RouterGroup)
}

func RegisterRoutes(r *gin.Engine, routes ...Route) {

	httplogger := newLoggerWithPath("../../internal/logs/http.log", "info")
	recoverylogger := newLoggerWithPath("../../internal/logs/recovery.log", "warning")
	rateLimiterLogger := newLoggerWithPath("../../internal/logs/recovery.log", "warning")
	r.Use(
		middleware.RateLimiterMiddleware(rateLimiterLogger),
		middleware.LoggerMiddleware(httplogger),
		middleware.RecoveryMiddleware(recoverylogger),
		middleware.ApiKeyMiddleware(),
		middleware.AuthMiddleware(),
	)
	v1api := r.Group("/api/v1")
	for _, route := range routes {
		route.RegisterRoutes(v1api)
	}
}

func newLoggerWithPath(path string, level string) *zerolog.Logger {
	config := logger.LoggerConfig{
		Level:      level,
		Filename:   path,
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     5,    //days
		Compress:   true, // disabled by default
		LocalTime:  true,
		IsDev:      utils.GetEnv("APP_ENV", "development"),
	}
	return logger.NewLoggerConfig(config)
}
