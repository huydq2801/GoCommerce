package v1dto

import (
	"encoding/json"

	"github.com/google/uuid"

	"huydq2801.com/golearn/internal/db/sqlc"
	"huydq2801.com/golearn/internal/utils"
)

type CategoryInput struct {
	Name string `json:"name" binding:"required,max=100"`
	Slug string `json:"slug" binding:"required,max=100"`
}
type ProductInput struct {
	CategoryID     string            `json:"category_id" binding:"required,uuid"`
	SKU            string            `json:"sku" binding:"required,max=64"`
	Name           string            `json:"name" binding:"required,max=200"`
	Slug           string            `json:"slug" binding:"required,max=200"`
	Brand          string            `json:"brand" binding:"required,max=100"`
	Model          string            `json:"model" binding:"required,max=100"`
	Description    string            `json:"description" binding:"max=10000"`
	PriceVND       int64             `json:"price_vnd" binding:"min=1,max=1000000000000"`
	StockQuantity  int32             `json:"stock_quantity" binding:"min=0,max=1000000"`
	Specifications map[string]string `json:"specifications" binding:"max=30"`
}
type UpdateProductInput struct {
	ProductInput
	Version  int64 `json:"version" binding:"min=1"`
	IsActive *bool `json:"is_active" binding:"required"`
}

func (i ProductInput) Params() sqlc.CreateProductParams {
	id, _ := uuid.Parse(i.CategoryID)
	specs := i.Specifications
	if specs == nil {
		specs = map[string]string{}
	}
	raw, _ := json.Marshal(specs)
	return sqlc.CreateProductParams{CategoryID: id, Sku: i.SKU, Name: i.Name, Slug: i.Slug, Brand: i.Brand, Model: i.Model, Description: i.Description, PriceVnd: i.PriceVND, StockQuantity: i.StockQuantity, Specifications: raw}
}

type ProductFilter struct {
	utils.Page
	CategoryID string `form:"category_id" binding:"omitempty,uuid"`
	Brand      string `form:"brand" binding:"max=100"`
	MinPrice   *int64 `form:"min_price" binding:"omitempty,min=0,max=1000000000000"`
	MaxPrice   *int64 `form:"max_price" binding:"omitempty,min=0,max=1000000000000"`
	Sort       string `form:"sort" binding:"omitempty,oneof=newest price_asc price_desc"`
}

func (f ProductFilter) Params() sqlc.ListProductsParams {
	id, _ := uuid.Parse(f.CategoryID)
	min, max := int64(-1), int64(-1)
	if f.MinPrice != nil {
		min = *f.MinPrice
	}
	if f.MaxPrice != nil {
		max = *f.MaxPrice
	}
	return sqlc.ListProductsParams{Search: f.Search, CategoryID: id, Brand: f.Brand, MinPrice: min, MaxPrice: max, SortBy: f.Sort, PageLimit: int32(f.Limit), PageOffset: f.Offset()}
}
