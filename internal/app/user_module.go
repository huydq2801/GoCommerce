package app

import (
	v1handler "huydq2801.com/golearn/internal/handler/v1"
	"huydq2801.com/golearn/internal/repository"
	"huydq2801.com/golearn/internal/routes"
	v1routes "huydq2801.com/golearn/internal/routes/v1"
	v1service "huydq2801.com/golearn/internal/service/v1"
)

type UserModule struct {
	routes routes.Route
}

func NewUserModule() *UserModule {
	userRepo := repository.NewInMemoryUserRepository()
	userService := v1service.NewUserService(userRepo)
	userHandler := v1handler.NewUserHandler(userService)
	userRoutes := v1routes.NewUserRouter(userHandler)

	return &UserModule{routes: userRoutes}
}
func (m *UserModule) Routes() routes.Route {
	return m.routes
}
