package app

import (
	v1handler "gocommerce/internal/handler/v1"
	"gocommerce/internal/repository"
	"gocommerce/internal/routes"
	v1routes "gocommerce/internal/routes/v1"
	v1service "gocommerce/internal/service/v1"
)

type UserModule struct {
	routes routes.Route
}

func NewUserModule(ctx *ModuleContext) *UserModule {
	userRepo := repository.NewSQLUserRepository(ctx.DB)
	userService := v1service.NewUserService(userRepo)
	userHandler := v1handler.NewUserHandler(userService)
	userRoutes := v1routes.NewUserRouter(userHandler, ctx.APIKey, ctx.Auth)

	return &UserModule{routes: userRoutes}
}
func (m *UserModule) Routes() routes.Route {
	return m.routes
}
