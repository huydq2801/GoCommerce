package app

import (
	v1handler "gocommerce/internal/handler/v1"
	"gocommerce/internal/repository"
	"gocommerce/internal/routes"
	v1routes "gocommerce/internal/routes/v1"
	v1service "gocommerce/internal/service/v1"
)

type CatalogModule struct{ routes routes.Route }

func NewCatalogModule(ctx *ModuleContext) *CatalogModule {
	repo := repository.NewSQLCatalogRepository(ctx.DB)
	service := v1service.NewCatalogService(repo)
	handler := v1handler.NewCatalogHandler(service)
	return &CatalogModule{routes: v1routes.NewCatalogRoutes(handler, ctx.APIKey, ctx.Auth)}
}
func (m *CatalogModule) Routes() routes.Route { return m.routes }
