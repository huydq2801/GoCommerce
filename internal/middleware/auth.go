package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"gocommerce/internal/db/sqlc"
	"gocommerce/internal/utils"
)

type UserLookup interface {
	GetUser(context.Context, uuid.UUID) (sqlc.User, error)
}
type Authenticator struct {
	secret []byte
	users  UserLookup
}

func NewAuthenticator(secret string, users UserLookup) *Authenticator {
	return &Authenticator{secret: []byte(secret), users: users}
}
func (a *Authenticator) authenticate(c *gin.Context) bool {
	header := strings.Fields(c.GetHeader("Authorization"))
	if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
		return a.reject(c)
	}
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(header[1], claims, func(t *jwt.Token) (any, error) { return a.secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuer("gocommerce"), jwt.WithAudience("gocommerce-api"))
	if err != nil || !token.Valid {
		return a.reject(c)
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return a.reject(c)
	}
	user, err := a.users.GetUser(c.Request.Context(), id)
	if err != nil || user.UserStatus != 1 {
		return a.reject(c)
	}
	c.Set("current_user", user)
	return true
}
func (a *Authenticator) reject(c *gin.Context) bool {
	utils.ResponseError(c, utils.NewError("Authentication required", utils.ErrCodeUnauthorized))
	c.Abort()
	return false
}
func (a *Authenticator) RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if a.authenticate(c) {
			c.Next()
		}
	}
}
func (a *Authenticator) RequireAdmin(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// The bootstrap key is a server-side admin credential, never embed it in a frontend.
		if c.GetHeader("X-API-Key") != "" {
			ApiKeyMiddleware(apiKey)(c)
			return
		}
		if !a.authenticate(c) {
			return
		}
		if CurrentUser(c).UserLevel != 1 {
			c.AbortWithStatusJSON(403, gin.H{"error": gin.H{"code": "FORBIDDEN", "message": "Administrator role required"}})
			return
		}
		c.Next()
	}
}
func CurrentUser(c *gin.Context) sqlc.User {
	value, _ := c.Get("current_user")
	user, _ := value.(sqlc.User)
	return user
}
