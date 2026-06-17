package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/products/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

func TestCatalogCreateProduct(t *testing.T) {
	ctx := context.Background()
	catID := uuid.New()

	setup := func() (*fakeProductRepo, CatalogService) {
		products := newFakeProductRepo()
		categories := newFakeCategoryRepo()
		categories.byID[catID] = activeCategory(catID)
		return products, NewCatalogService(products, categories)
	}

	t.Run("success", func(t *testing.T) {
		products, svc := setup()
		view, err := svc.CreateProduct(ctx, dto.CreateProductInput{
			Name: "Shirt", Description: "desc", PricePoints: 100, CategoryID: catID.String(),
		})
		require.NoError(t, err)
		assert.Equal(t, "Shirt", view.Name)
		assert.NotEmpty(t, view.ID)
		assert.NotNil(t, products.created)
	})

	t.Run("malformed category id maps to category not found", func(t *testing.T) {
		_, svc := setup()
		_, err := svc.CreateProduct(ctx, dto.CreateProductInput{Name: "Shirt", Description: "desc", PricePoints: 100, CategoryID: "not-a-uuid"})
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("factory error propagates", func(t *testing.T) {
		_, svc := setup()
		_, err := svc.CreateProduct(ctx, dto.CreateProductInput{Name: "Shirt", Description: "desc", PricePoints: 0, CategoryID: catID.String()})
		assert.ErrorIs(t, err, domain.ErrInvalidPrice)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		products, svc := setup()
		products.createErr = errBoom
		_, err := svc.CreateProduct(ctx, dto.CreateProductInput{Name: "Shirt", Description: "desc", PricePoints: 100, CategoryID: catID.String()})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestCatalogUpdateProduct(t *testing.T) {
	ctx := context.Background()
	catID := uuid.New()
	prodID := uuid.New()

	setup := func(t *testing.T) (*fakeProductRepo, CatalogService) {
		products := newFakeProductRepo()
		categories := newFakeCategoryRepo()
		categories.byID[catID] = activeCategory(catID)
		existing, err := model.NewProduct("Old", "old desc", vo.NewPricePointsFromStored(100), *activeCategory(catID), nil)
		require.NoError(t, err)
		existing.ID = prodID
		products.byID[prodID] = existing
		return products, NewCatalogService(products, categories)
	}

	t.Run("success passes expected version to repo", func(t *testing.T) {
		products, svc := setup(t)
		view, err := svc.UpdateProduct(ctx, dto.UpdateProductInput{
			ProductID: prodID.String(), Name: "New", Description: "new", PricePoints: 200,
			CategoryID: catID.String(), Active: true, Version: 1,
		})
		require.NoError(t, err)
		assert.Equal(t, "New", view.Name)
		assert.Equal(t, 1, products.lastVersion)
	})

	t.Run("malformed product id", func(t *testing.T) {
		_, svc := setup(t)
		_, err := svc.UpdateProduct(ctx, dto.UpdateProductInput{ProductID: "bad", CategoryID: catID.String(), Version: 1})
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})

	t.Run("malformed category id", func(t *testing.T) {
		_, svc := setup(t)
		_, err := svc.UpdateProduct(ctx, dto.UpdateProductInput{ProductID: prodID.String(), CategoryID: "bad", Version: 1})
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("version conflict", func(t *testing.T) {
		_, svc := setup(t)
		_, err := svc.UpdateProduct(ctx, dto.UpdateProductInput{
			ProductID: prodID.String(), Name: "New", Description: "new", PricePoints: 200, CategoryID: catID.String(), Version: 5,
		})
		assert.ErrorIs(t, err, domain.ErrVersionConflict)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		products, svc := setup(t)
		products.updateErr = errBoom
		_, err := svc.UpdateProduct(ctx, dto.UpdateProductInput{
			ProductID: prodID.String(), Name: "New", Description: "new", PricePoints: 200, CategoryID: catID.String(), Version: 1,
		})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestCatalogDeactivateProduct(t *testing.T) {
	ctx := context.Background()
	prodID := uuid.New()

	t.Run("success", func(t *testing.T) {
		products := newFakeProductRepo()
		svc := NewCatalogService(products, newFakeCategoryRepo())
		require.NoError(t, svc.DeactivateProduct(ctx, prodID.String()))
		assert.Equal(t, []uuid.UUID{prodID}, products.deactivated)
	})

	t.Run("malformed id", func(t *testing.T) {
		svc := NewCatalogService(newFakeProductRepo(), newFakeCategoryRepo())
		assert.ErrorIs(t, svc.DeactivateProduct(ctx, "bad"), domain.ErrProductNotFound)
	})

	t.Run("repository error", func(t *testing.T) {
		products := newFakeProductRepo()
		products.deactErr = errBoom
		svc := NewCatalogService(products, newFakeCategoryRepo())
		assert.ErrorIs(t, svc.DeactivateProduct(ctx, prodID.String()), errBoom)
	})
}

func TestCatalogCreateCategory(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		categories := newFakeCategoryRepo()
		svc := NewCatalogService(newFakeProductRepo(), categories)
		view, err := svc.CreateCategory(ctx, dto.CreateCategoryInput{Code: "clothing", Name: "Clothes"})
		require.NoError(t, err)
		assert.Equal(t, "clothing", view.Code)
		assert.NotNil(t, categories.created)
	})

	t.Run("invalid code", func(t *testing.T) {
		svc := NewCatalogService(newFakeProductRepo(), newFakeCategoryRepo())
		_, err := svc.CreateCategory(ctx, dto.CreateCategoryInput{Code: "BAD", Name: "x"})
		assert.ErrorIs(t, err, domain.ErrInvalidCategoryCode)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		categories := newFakeCategoryRepo()
		categories.createErr = errBoom
		svc := NewCatalogService(newFakeProductRepo(), categories)
		_, err := svc.CreateCategory(ctx, dto.CreateCategoryInput{Code: "clothing", Name: "Clothes"})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestCatalogUpdateCategory(t *testing.T) {
	ctx := context.Background()
	catID := uuid.New()

	setup := func() (*fakeCategoryRepo, CatalogService) {
		categories := newFakeCategoryRepo()
		categories.byID[catID] = activeCategory(catID)
		return categories, NewCatalogService(newFakeProductRepo(), categories)
	}

	t.Run("success", func(t *testing.T) {
		_, svc := setup()
		view, err := svc.UpdateCategory(ctx, dto.UpdateCategoryInput{ID: catID.String(), Name: "Apparel", Active: false})
		require.NoError(t, err)
		assert.Equal(t, "Apparel", view.Name)
		assert.False(t, view.Active)
	})

	t.Run("malformed id", func(t *testing.T) {
		_, svc := setup()
		_, err := svc.UpdateCategory(ctx, dto.UpdateCategoryInput{ID: "bad", Name: "x"})
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		categories, svc := setup()
		categories.updateErr = errBoom
		_, err := svc.UpdateCategory(ctx, dto.UpdateCategoryInput{ID: catID.String(), Name: "Apparel", Active: true})
		assert.ErrorIs(t, err, errBoom)
	})
}
