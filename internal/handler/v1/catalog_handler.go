package v1handler

import (
	"github.com/gin-gonic/gin"

	"gocommerce/internal/db/sqlc"
	v1dto "gocommerce/internal/dto/v1"
	v1service "gocommerce/internal/service/v1"
	"gocommerce/internal/utils"
)

type CatalogHandler struct{ service *v1service.CatalogService }

func NewCatalogHandler(s *v1service.CatalogService) *CatalogHandler {
	return &CatalogHandler{service: s}
}
func (h *CatalogHandler) ListCategories(c *gin.Context) {
	items, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, items)
}
func (h *CatalogHandler) CreateCategory(c *gin.Context) {
	var input v1dto.CategoryInput
	if !utils.DecodeJSON(c, &input) {
		return
	}
	item, err := h.service.CreateCategory(c.Request.Context(), sqlc.CreateCategoryParams{Name: input.Name, Slug: input.Slug})
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	c.Header("Location", "/api/v1/categories/"+item.ID.String())
	utils.ResponseSuccess(c, 201, item)
}
func (h *CatalogHandler) UpdateCategory(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	var input v1dto.CategoryInput
	if !utils.DecodeJSON(c, &input) {
		return
	}
	item, err := h.service.UpdateCategory(c.Request.Context(), sqlc.UpdateCategoryParams{ID: id, Name: input.Name, Slug: input.Slug})
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, item)
}
func (h *CatalogHandler) DeleteCategory(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteCategory(c.Request.Context(), id); err != nil {
		utils.ResponseError(c, err)
		return
	}
	c.Status(204)
}
func (h *CatalogHandler) ListProducts(c *gin.Context) {
	var filter v1dto.ProductFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.ResponseError(c, utils.NewError("Invalid product filters", utils.ErrCodeBadRequest))
		return
	}
	filter.Defaults()
	items, total, err := h.service.ListProducts(c.Request.Context(), filter.Params())
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponsePage(c, items, filter.Page, total)
}
func (h *CatalogHandler) GetProduct(c *gin.Context)      { h.getProduct(c, false) }
func (h *CatalogHandler) GetAdminProduct(c *gin.Context) { h.getProduct(c, true) }
func (h *CatalogHandler) getProduct(c *gin.Context, admin bool) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	item, err := h.service.GetProduct(c.Request.Context(), id, admin)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, item)
}
func (h *CatalogHandler) CreateProduct(c *gin.Context) {
	var input v1dto.ProductInput
	if !utils.DecodeJSON(c, &input) {
		return
	}
	item, err := h.service.CreateProduct(c.Request.Context(), input.Params())
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	c.Header("Location", "/api/v1/products/"+item.ID.String())
	utils.ResponseSuccess(c, 201, item)
}
func (h *CatalogHandler) UpdateProduct(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	var input v1dto.UpdateProductInput
	if !utils.DecodeJSON(c, &input) {
		return
	}
	p := input.Params()
	item, err := h.service.UpdateProduct(c.Request.Context(), sqlc.UpdateProductParams{ID: id, CategoryID: p.CategoryID, Sku: p.Sku, Name: p.Name, Slug: p.Slug, Brand: p.Brand, Model: p.Model, Description: p.Description, PriceVnd: p.PriceVnd, StockQuantity: p.StockQuantity, Specifications: p.Specifications, ExpectedVersion: input.Version, IsActive: *input.IsActive})
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, item)
}
func (h *CatalogHandler) DeleteProduct(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	if err := h.service.DeactivateProduct(c.Request.Context(), id); err != nil {
		utils.ResponseError(c, err)
		return
	}
	c.Status(204)
}
