package v1handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	v1dto "gocommerce/internal/dto/v1"
	"gocommerce/internal/middleware"
	v1service "gocommerce/internal/service/v1"
	"gocommerce/internal/utils"
)

type ShoppingHandler struct{ service *v1service.ShoppingService }

func NewShoppingHandler(s *v1service.ShoppingService) *ShoppingHandler {
	return &ShoppingHandler{service: s}
}
func (h *ShoppingHandler) GetCart(c *gin.Context) {
	cart, err := h.service.GetCart(c.Request.Context(), middleware.CurrentUser(c).UserID)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, cart)
}
func (h *ShoppingHandler) SetCartItem(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	var input struct {
		Quantity int32 `json:"quantity" binding:"min=1,max=1000"`
	}
	if !utils.DecodeJSON(c, &input) {
		return
	}
	if err := h.service.SetCartItem(c.Request.Context(), middleware.CurrentUser(c).UserID, id, input.Quantity); err != nil {
		utils.ResponseError(c, err)
		return
	}
	c.Status(204)
}
func (h *ShoppingHandler) DeleteCartItem(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	h.removeCart(c, id)
}
func (h *ShoppingHandler) ClearCart(c *gin.Context) { h.removeCart(c, uuid.Nil) }
func (h *ShoppingHandler) removeCart(c *gin.Context, id uuid.UUID) {
	if err := h.service.RemoveCartItem(c.Request.Context(), middleware.CurrentUser(c).UserID, id); err != nil {
		utils.ResponseError(c, err)
		return
	}
	c.Status(204)
}
func (h *ShoppingHandler) Checkout(c *gin.Context) {
	key, err := uuid.Parse(c.GetHeader("Idempotency-Key"))
	if err != nil {
		utils.ResponseError(c, utils.NewError("Idempotency-Key header must be a UUID", utils.ErrCodeBadRequest))
		return
	}
	var input v1dto.CheckoutInput
	if !utils.DecodeJSON(c, &input) {
		return
	}
	order, replay, err := h.service.Checkout(c.Request.Context(), middleware.CurrentUser(c).UserID, key, input)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	status := 201
	if replay {
		status = 200
		c.Header("Idempotency-Replayed", "true")
	}
	c.Header("Location", "/api/v1/orders/"+order.ID.String())
	utils.ResponseSuccess(c, status, order)
}
func (h *ShoppingHandler) GetOrder(c *gin.Context)      { h.getOrder(c, false) }
func (h *ShoppingHandler) GetAdminOrder(c *gin.Context) { h.getOrder(c, true) }
func (h *ShoppingHandler) getOrder(c *gin.Context, admin bool) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	order, err := h.service.GetOrder(c.Request.Context(), middleware.CurrentUser(c).UserID, id, admin)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, order)
}
func (h *ShoppingHandler) ListOrders(c *gin.Context) {
	h.listOrders(c, middleware.CurrentUser(c).UserID)
}
func (h *ShoppingHandler) ListAdminOrders(c *gin.Context) { h.listOrders(c, 0) }
func (h *ShoppingHandler) listOrders(c *gin.Context, userID int32) {
	var p utils.Page
	if err := c.ShouldBindQuery(&p); err != nil {
		utils.ResponseError(c, utils.NewError("Invalid pagination", utils.ErrCodeBadRequest))
		return
	}
	p.Defaults()
	items, total, err := h.service.ListOrders(c.Request.Context(), userID, p)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponsePage(c, items, p, total)
}
func (h *ShoppingHandler) CancelOrder(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	order, err := h.service.CancelOrder(c.Request.Context(), middleware.CurrentUser(c).UserID, id)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, order)
}
func (h *ShoppingHandler) AdvanceOrder(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	var input struct {
		Status string `json:"status" binding:"required,oneof=processing shipped completed"`
	}
	if !utils.DecodeJSON(c, &input) {
		return
	}
	order, err := h.service.AdvanceOrder(c.Request.Context(), id, input.Status)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, order)
}
