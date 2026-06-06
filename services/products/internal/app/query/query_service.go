package query

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/products/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/repository"
)

const (
	defaultPageSize = 24
	maxPageSize     = 100
)

type CatalogReadService interface {
	GetProduct(ctx context.Context, productID string) (dto.ProductView, error)
	ListProducts(ctx context.Context, in dto.ListProductsInput) (dto.ListProductsResult, error)
	ListCategories(ctx context.Context, activeOnly bool) ([]dto.CategoryView, error)
}

type catalogReadService struct {
	products   repository.ProductRepository
	categories repository.CategoryRepository
}

func NewCatalogReadService(products repository.ProductRepository, categories repository.CategoryRepository) CatalogReadService {
	return &catalogReadService{products: products, categories: categories}
}

func (s *catalogReadService) GetProduct(ctx context.Context, productID string) (dto.ProductView, error) {
	id, err := uuid.Parse(productID)
	if err != nil {
		return dto.ProductView{}, domain.ErrProductNotFound
	}
	product, err := s.products.GetByID(ctx, id)
	if err != nil {
		return dto.ProductView{}, err
	}
	return dto.ToProductView(product), nil
}

func (s *catalogReadService) ListProducts(ctx context.Context, in dto.ListProductsInput) (dto.ListProductsResult, error) {
	filter := repository.ProductFilter{
		Search:     in.Search,
		ActiveOnly: in.ActiveOnly,
		Limit:      clampPageSize(in.PageSize),
	}

	if in.CategoryID != "" {
		categoryID, err := uuid.Parse(in.CategoryID)
		if err != nil {
			return dto.ListProductsResult{}, domain.ErrInvalidInput
		}
		filter.CategoryID = &categoryID
	}

	cursor, err := decodeCursor(in.PageToken)
	if err != nil {
		return dto.ListProductsResult{}, err
	}
	filter.Cursor = cursor

	page, err := s.products.List(ctx, filter)
	if err != nil {
		return dto.ListProductsResult{}, err
	}

	views := make([]dto.ProductView, 0, len(page.Products))
	for _, p := range page.Products {
		views = append(views, dto.ToProductView(p))
	}
	return dto.ListProductsResult{
		Products:      views,
		NextPageToken: encodeCursor(page.Next),
	}, nil
}

func (s *catalogReadService) ListCategories(ctx context.Context, activeOnly bool) ([]dto.CategoryView, error) {
	categories, err := s.categories.List(ctx, activeOnly)
	if err != nil {
		return nil, err
	}
	views := make([]dto.CategoryView, 0, len(categories))
	for _, c := range categories {
		views = append(views, dto.ToCategoryView(c))
	}
	return views, nil
}

func clampPageSize(n int) int {
	if n <= 0 {
		return defaultPageSize
	}
	if n > maxPageSize {
		return maxPageSize
	}
	return n
}

func encodeCursor(c *repository.ProductCursor) string {
	if c == nil {
		return ""
	}
	raw := strconv.FormatInt(c.CreatedAt.UnixNano(), 10) + "|" + c.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(token string) (*repository.ProductCursor, error) {
	if token == "" {
		return nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	parts := strings.SplitN(string(data), "|", 2)
	if len(parts) != 2 {
		return nil, domain.ErrInvalidInput
	}
	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	return &repository.ProductCursor{CreatedAt: time.Unix(0, nanos).UTC(), ID: id}, nil
}
