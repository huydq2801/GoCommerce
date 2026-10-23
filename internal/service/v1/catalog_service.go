package v1service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"gocommerce/internal/db/sqlc"
	"gocommerce/internal/repository"
	"gocommerce/internal/utils"
)

type CatalogService struct{ repo repository.CatalogRepository }

func NewCatalogService(repo repository.CatalogRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

var catalogSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func categoryValues(name, slug string) (string, string, error) {
	name = strings.TrimSpace(name)
	slug = strings.ToLower(strings.TrimSpace(slug))
	if name == "" || !catalogSlug.MatchString(slug) {
		return "", "", utils.NewError("Name and a valid lowercase slug are required", utils.ErrCodeBadRequest)
	}
	return name, slug, nil
}
func (s *CatalogService) CreateCategory(ctx context.Context, p sqlc.CreateCategoryParams) (sqlc.Category, error) {
	name, slug, err := categoryValues(p.Name, p.Slug)
	if err != nil {
		return sqlc.Category{}, err
	}
	p.Name = name
	p.Slug = slug
	c, err := s.repo.CreateCategory(ctx, p)
	return c, utils.DatabaseError(err)
}
func (s *CatalogService) ListCategories(ctx context.Context) ([]sqlc.Category, error) {
	items, err := s.repo.ListCategories(ctx)
	return items, utils.DatabaseError(err)
}
func (s *CatalogService) UpdateCategory(ctx context.Context, p sqlc.UpdateCategoryParams) (sqlc.Category, error) {
	name, slug, err := categoryValues(p.Name, p.Slug)
	if err != nil {
		return sqlc.Category{}, err
	}
	p.Name = name
	p.Slug = slug
	c, err := s.repo.UpdateCategory(ctx, p)
	return c, utils.DatabaseError(err)
}
func (s *CatalogService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	rows, err := s.repo.DeleteCategory(ctx, id)
	if err != nil {
		return utils.DatabaseError(err)
	}
	if rows == 0 {
		return utils.NewError("Category not found", utils.ErrCodeNotFound)
	}
	return nil
}
func normalizeProduct(p *sqlc.CreateProductParams) error {
	p.Sku = strings.ToUpper(strings.TrimSpace(p.Sku))
	p.Name = strings.TrimSpace(p.Name)
	p.Slug = strings.ToLower(strings.TrimSpace(p.Slug))
	p.Brand = strings.TrimSpace(p.Brand)
	p.Model = strings.TrimSpace(p.Model)
	if p.CategoryID == uuid.Nil || p.Sku == "" || p.Name == "" || p.Brand == "" || p.Model == "" || !catalogSlug.MatchString(p.Slug) {
		return utils.NewError("Invalid product identity or slug", utils.ErrCodeBadRequest)
	}
	if p.PriceVnd < 1 || p.PriceVnd > 1000000000000 || p.StockQuantity < 0 || p.StockQuantity > 1000000 {
		return utils.NewError("Price or stock is out of range", utils.ErrCodeBadRequest)
	}
	var specs map[string]string
	if err := json.Unmarshal(p.Specifications, &specs); err != nil || specs == nil || len(specs) > 30 {
		return utils.NewError("Specifications must be an object of at most 30 string fields", utils.ErrCodeBadRequest)
	}
	for key, value := range specs {
		if strings.TrimSpace(key) == "" || len(key) > 64 || len(value) > 256 {
			return utils.NewError("Invalid specification key or value", utils.ErrCodeBadRequest)
		}
	}
	return nil
}
func (s *CatalogService) CreateProduct(ctx context.Context, p sqlc.CreateProductParams) (sqlc.Product, error) {
	if err := normalizeProduct(&p); err != nil {
		return sqlc.Product{}, err
	}
	item, err := s.repo.CreateProduct(ctx, p)
	return item, utils.DatabaseError(err)
}
func (s *CatalogService) GetProduct(ctx context.Context, id uuid.UUID, admin bool) (sqlc.Product, error) {
	var item sqlc.Product
	var err error
	if admin {
		item, err = s.repo.GetAdminProduct(ctx, id)
	} else {
		item, err = s.repo.GetProduct(ctx, id)
	}
	return item, utils.DatabaseError(err)
}
func (s *CatalogService) ListProducts(ctx context.Context, p sqlc.ListProductsParams) ([]sqlc.Product, int64, error) {
	if p.MinPrice >= 0 && p.MaxPrice >= 0 && p.MinPrice > p.MaxPrice {
		return nil, 0, utils.NewError("min_price must not exceed max_price", utils.ErrCodeBadRequest)
	}
	p.Search = strings.TrimSpace(p.Search)
	p.Brand = strings.TrimSpace(p.Brand)
	items, err := s.repo.ListProducts(ctx, p)
	if err != nil {
		return nil, 0, utils.DatabaseError(err)
	}
	total, err := s.repo.CountProducts(ctx, sqlc.CountProductsParams{Search: p.Search, CategoryID: p.CategoryID, Brand: p.Brand, MinPrice: p.MinPrice, MaxPrice: p.MaxPrice})
	return items, total, utils.DatabaseError(err)
}
func (s *CatalogService) UpdateProduct(ctx context.Context, p sqlc.UpdateProductParams) (sqlc.Product, error) {
	values := sqlc.CreateProductParams{CategoryID: p.CategoryID, Sku: p.Sku, Name: p.Name, Slug: p.Slug, Brand: p.Brand, Model: p.Model, Description: p.Description, PriceVnd: p.PriceVnd, StockQuantity: p.StockQuantity, Specifications: p.Specifications}
	if err := normalizeProduct(&values); err != nil {
		return sqlc.Product{}, err
	}
	p.Sku = values.Sku
	p.Name = values.Name
	p.Slug = values.Slug
	p.Brand = values.Brand
	p.Model = values.Model
	item, err := s.repo.UpdateProduct(ctx, p)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, getErr := s.repo.GetAdminProduct(ctx, p.ID); getErr != nil {
			return sqlc.Product{}, utils.DatabaseError(getErr)
		}
		return sqlc.Product{}, utils.NewError("Product changed; reload it and retry with the latest version", utils.ErrCodeConflict)
	}
	return item, utils.DatabaseError(err)
}
func (s *CatalogService) DeactivateProduct(ctx context.Context, id uuid.UUID) error {
	rows, err := s.repo.DeactivateProduct(ctx, id)
	if err != nil {
		return utils.DatabaseError(err)
	}
	if rows == 0 {
		_, err = s.repo.GetAdminProduct(ctx, id)
		return utils.DatabaseError(err)
	}
	return nil
}
