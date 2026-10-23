package middleware

import (
	"crypto/sha256"
	"crypto/subtle"

	"github.com/gin-gonic/gin"

	"gocommerce/internal/utils"
)

// API keys protect administrative endpoints only until JWT/RBAC is implemented.
func ApiKeyMiddleware(expected string) gin.HandlerFunc {
	expectedHash := sha256.Sum256([]byte(expected))
	return func(c *gin.Context) {
		candidate := sha256.Sum256([]byte(c.GetHeader("X-API-Key")))
		if expected == "" || subtle.ConstantTimeCompare(candidate[:], expectedHash[:]) != 1 {
			utils.ResponseError(c, utils.NewError("Invalid administrator credential", utils.ErrCodeUnauthorized))
			c.Abort()
			return
		}
		c.Next()
	}
}
