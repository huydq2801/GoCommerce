package v1routes

import (
	"github.com/gin-gonic/gin"

	v1handler "gocommerce/internal/handler/v1"
	"gocommerce/internal/middleware"
)

type CatalogRoutes struct {
	handler *v1handler.CatalogHandler
	apiKey  string
	auth    *middleware.Authenticator
}

func NewCatalogRoutes(h *v1handler.CatalogHandler, key string, auth *middleware.Authenticator) *CatalogRoutes {
	return &CatalogRoutes{handler: h, apiKey: key, auth: auth}
}
func (r *CatalogRoutes) RegisterRoutes(v1 *gin.RouterGroup) {
	v1.GET("/categories", r.handler.ListCategories)
	v1.GET("/products", r.handler.ListProducts)
	v1.GET("/products/:uuid", r.handler.GetProduct)
	admin := v1.Group("", r.auth.RequireAdmin(r.apiKey))
	admin.POST("/categories", r.handler.CreateCategory)
	admin.PUT("/categories/:uuid", r.handler.UpdateCategory)
	admin.DELETE("/categories/:uuid", r.handler.DeleteCategory)
	admin.POST("/products", r.handler.CreateProduct)
	admin.PUT("/products/:uuid", r.handler.UpdateProduct)
	admin.DELETE("/products/:uuid", r.handler.DeleteProduct)
	admin.GET("/admin/products/:uuid", r.handler.GetAdminProduct)
}
