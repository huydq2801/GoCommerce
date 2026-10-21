package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func RecoveryMiddleware(logger *zerolog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				logger.Error().
					Str("path", ctx.Request.URL.Path).
					Str("method", ctx.Request.Method).
					Str("client_ip", ctx.ClientIP()).
					Str("panic", fmt.Sprintf("%v", err)).
					Str("stack", ExtractFirstAppStackLine(stack)).
					Str("full_stack", string(stack)).
					Msg("Panic Recovery")
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Something went wrong",
				})
			}
		}()
		ctx.Next()
	}
}

var statckLineRegex = regexp.MustCompile(`(.+\.go:\d+)`)

func ExtractFirstAppStackLine(stack []byte) string {
	lines := bytes.Split(stack, []byte{'\n'})
	for _, line := range lines {
		if bytes.Contains(line, []byte(".go")) &&
			!bytes.Contains(line, []byte("debug")) &&
			!bytes.Contains(line, []byte("recovery_middleware")) &&
			!bytes.Contains(line, []byte("runtime")) {
			cleanLine := strings.TrimSpace(string(line))
			match := statckLineRegex.FindStringSubmatch(cleanLine)
			if len(match) > 1 {
				return match[1]
			}
			return cleanLine
		}
	}
	return ""
}
