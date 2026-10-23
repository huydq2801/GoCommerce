package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.NewString()
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	}
}
func LoggerMiddleware(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		event := logger.Info()
		if c.Writer.Status() >= 500 {
			event = logger.Error()
		} else if c.Writer.Status() >= 400 {
			event = logger.Warn()
		}
		event.Str("request_id", c.GetString("request_id")).Str("method", c.Request.Method).Str("route", c.FullPath()).Int("status", c.Writer.Status()).Int64("duration_ms", time.Since(start).Milliseconds()).Msg("http_request")
	}
}
