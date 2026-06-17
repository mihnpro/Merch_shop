package query

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/products/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/repository"
)

func TestGetProduct(t *testing.T) {
	ctx := context.Background()
	prodID := uuid.New()
	catID := uuid.New()

	t.Run("success", func(t *testing.T) {
		products := newFakeProductRepo()
		products.byID[prodID] = sampleProduct(prodID, catID)
		svc := NewCatalogReadService(products, newFakeCategoryRepo())
		view, err := svc.GetProduct(ctx, prodID.String())
		require.NoError(t, err)
		assert.Equal(t, prodID.String(), view.ID)
	})

	t.Run("malformed id", func(t *testing.T) {
		svc := NewCatalogReadService(newFakeProductRepo(), newFakeCategoryRepo())
		_, err := svc.GetProduct(ctx, "bad")
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})

	t.Run("not found", func(t *testing.T) {
		svc := NewCatalogReadService(newFakeProductRepo(), newFakeCategoryRepo())
		_, err := svc.GetProduct(ctx, uuid.New().String())
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})
}

func TestListProducts(t *testing.T) {
	ctx := context.Background()
	catID := uuid.New()

	t.Run("maps products and clamps page size", func(t *testing.T) {
		products := newFakeProductRepo()
		products.page = repository.ProductPage{Products: []*model.Product{sampleProduct(uuid.New(), catID)}}
		svc := NewCatalogReadService(products, newFakeCategoryRepo())
		res, err := svc.ListProducts(ctx, dto.ListProductsInput{PageSize: 0})
		require.NoError(t, err)
		assert.Len(t, res.Products, 1)
		assert.Equal(t, defaultPageSize, products.lastFilter.Limit)
	})

	t.Run("page size over max is clamped", func(t *testing.T) {
		products := newFakeProductRepo()
		svc := NewCatalogReadService(products, newFakeCategoryRepo())
		_, err := svc.ListProducts(ctx, dto.ListProductsInput{PageSize: 1000})
		require.NoError(t, err)
		assert.Equal(t, maxPageSize, products.lastFilter.Limit)
	})

	t.Run("category filter set", func(t *testing.T) {
		products := newFakeProductRepo()
		svc := NewCatalogReadService(products, newFakeCategoryRepo())
		_, err := svc.ListProducts(ctx, dto.ListProductsInput{CategoryID: catID.String()})
		require.NoError(t, err)
		require.NotNil(t, products.lastFilter.CategoryID)
		assert.Equal(t, catID, *products.lastFilter.CategoryID)
	})

	t.Run("invalid category id", func(t *testing.T) {
		svc := NewCatalogReadService(newFakeProductRepo(), newFakeCategoryRepo())
		_, err := svc.ListProducts(ctx, dto.ListProductsInput{CategoryID: "bad"})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("invalid page token", func(t *testing.T) {
		svc := NewCatalogReadService(newFakeProductRepo(), newFakeCategoryRepo())
		_, err := svc.ListProducts(ctx, dto.ListProductsInput{PageToken: "!!!not-base64!!!"})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("page token round trip", func(t *testing.T) {
		cursorID := uuid.New()
		token := encodeCursor(&repository.ProductCursor{CreatedAt: time.Unix(0, 1700000000000000000).UTC(), ID: cursorID})
		require.NotEmpty(t, token)

		products := newFakeProductRepo()
		svc := NewCatalogReadService(products, newFakeCategoryRepo())
		_, err := svc.ListProducts(ctx, dto.ListProductsInput{PageToken: token})
		require.NoError(t, err)
		require.NotNil(t, products.lastFilter.Cursor)
		assert.Equal(t, cursorID, products.lastFilter.Cursor.ID)
	})

	t.Run("next page token encoded from result", func(t *testing.T) {
		next := &repository.ProductCursor{CreatedAt: time.Unix(0, 1).UTC(), ID: uuid.New()}
		products := newFakeProductRepo()
		products.page = repository.ProductPage{Next: next}
		svc := NewCatalogReadService(products, newFakeCategoryRepo())
		res, err := svc.ListProducts(ctx, dto.ListProductsInput{})
		require.NoError(t, err)
		assert.NotEmpty(t, res.NextPageToken)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		products := newFakeProductRepo()
		products.listErr = errBoom
		svc := NewCatalogReadService(products, newFakeCategoryRepo())
		_, err := svc.ListProducts(ctx, dto.ListProductsInput{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestListCategories(t *testing.T) {
	ctx := context.Background()

	t.Run("maps categories and forwards active flag", func(t *testing.T) {
		categories := newFakeCategoryRepo()
		categories.list = []*model.Category{activeCategory(uuid.New())}
		svc := NewCatalogReadService(newFakeProductRepo(), categories)
		views, err := svc.ListCategories(ctx, true)
		require.NoError(t, err)
		assert.Len(t, views, 1)
		assert.True(t, categories.lastActiveOnly)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		categories := newFakeCategoryRepo()
		categories.listErr = errBoom
		svc := NewCatalogReadService(newFakeProductRepo(), categories)
		_, err := svc.ListCategories(ctx, false)
		assert.ErrorIs(t, err, errBoom)
	})
}
