package v1routes

import (
	"github.com/gin-gonic/gin"

	v1handler "huydq2801.com/golearn/internal/handler/v1"
	"huydq2801.com/golearn/internal/middleware"
)

type ShoppingRoutes struct {
	handler *v1handler.ShoppingHandler
	auth    *middleware.Authenticator
	apiKey  string
}

func NewShoppingRoutes(h *v1handler.ShoppingHandler, auth *middleware.Authenticator, key string) *ShoppingRoutes {
	return &ShoppingRoutes{handler: h, auth: auth, apiKey: key}
}
func (r *ShoppingRoutes) RegisterRoutes(v1 *gin.RouterGroup) {
	user := v1.Group("", r.auth.RequireUser())
	user.GET("/cart", r.handler.GetCart)
	user.PUT("/cart/items/:uuid", r.handler.SetCartItem)
	user.DELETE("/cart/items/:uuid", r.handler.DeleteCartItem)
	user.DELETE("/cart/items", r.handler.ClearCart)
	user.POST("/orders", r.handler.Checkout)
	user.GET("/orders", r.handler.ListOrders)
	user.GET("/orders/:uuid", r.handler.GetOrder)
	user.POST("/orders/:uuid/cancel", r.handler.CancelOrder)
	admin := v1.Group("/admin", r.auth.RequireAdmin(r.apiKey))
	admin.GET("/orders", r.handler.ListAdminOrders)
	admin.GET("/orders/:uuid", r.handler.GetAdminOrder)
	admin.PATCH("/orders/:uuid/status", r.handler.AdvanceOrder)
}
