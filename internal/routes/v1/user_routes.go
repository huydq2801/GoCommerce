package v1routes

import (
	"github.com/gin-gonic/gin"

	v1handler "gocommerce/internal/handler/v1"
	"gocommerce/internal/middleware"
)

type UserRouters struct {
	handler *v1handler.UserHandler
	apiKey  string
	auth    *middleware.Authenticator
}

func NewUserRouter(handler *v1handler.UserHandler, key string, auth *middleware.Authenticator) *UserRouters {
	return &UserRouters{handler: handler, apiKey: key, auth: auth}
}
func (r *UserRouters) RegisterRoutes(v1 *gin.RouterGroup) {
	users := v1.Group("/users")
	users.POST("", r.handler.CreateUser)
	users.Use(r.auth.RequireAdmin(r.apiKey))
	users.GET("", r.handler.GetAllUsers)
	users.GET("/:uuid", r.handler.GetUserByUUID)
	users.PATCH("/:uuid", r.handler.UpdateUser)
	users.DELETE("/:uuid", r.handler.DeleteUser)
}
