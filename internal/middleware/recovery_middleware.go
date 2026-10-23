package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"huydq2801.com/golearn/internal/utils"
)

func RecoveryMiddleware(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error().Str("request_id", c.GetString("request_id")).Str("route", c.FullPath()).Msg("request_panic")
				c.Abort()
				if !c.Writer.Written() {
					utils.ResponseError(c, utils.NewError("Internal server error", utils.ErrCodeInternal))
				}
			}
		}()
		c.Next()
	}
}
