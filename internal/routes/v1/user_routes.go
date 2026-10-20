package v1routes

import (
	"github.com/gin-gonic/gin"
	v1handler "huydq2801.com/golearn/internal/handler/v1"
)

type UserRouters struct {
	handler *v1handler.UserHandler
}

func NewUserRouter(handler *v1handler.UserHandler) *UserRouters {
	return &UserRouters{
		handler: handler,
	}
}

func (u *UserRouters) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.GET("/", u.handler.GetAllUsers)
		users.POST("/", u.handler.CreateUser)
		users.GET("/:uuid", u.handler.GetUserByUUID)
		users.PUT("/:uuid", u.handler.UpdateUser)
		users.DELETE("/:uuid", u.handler.DeleteUser)
	}
}
