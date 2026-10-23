package repository

import (
	"context"

	"github.com/google/uuid"

	"huydq2801.com/golearn/internal/db/sqlc"
)

// sqlc already implements these persistence operations; no duplicate SQL wrapper is needed.
type CatalogRepository interface {
	CreateCategory(context.Context, sqlc.CreateCategoryParams) (sqlc.Category, error)
	ListCategories(context.Context) ([]sqlc.Category, error)
	UpdateCategory(context.Context, sqlc.UpdateCategoryParams) (sqlc.Category, error)
	DeleteCategory(context.Context, uuid.UUID) (int64, error)
	CreateProduct(context.Context, sqlc.CreateProductParams) (sqlc.Product, error)
	GetProduct(context.Context, uuid.UUID) (sqlc.Product, error)
	GetAdminProduct(context.Context, uuid.UUID) (sqlc.Product, error)
	ListProducts(context.Context, sqlc.ListProductsParams) ([]sqlc.Product, error)
	CountProducts(context.Context, sqlc.CountProductsParams) (int64, error)
	UpdateProduct(context.Context, sqlc.UpdateProductParams) (sqlc.Product, error)
	DeactivateProduct(context.Context, uuid.UUID) (int64, error)
}

func NewSQLCatalogRepository(queries sqlc.Querier) CatalogRepository { return queries }
