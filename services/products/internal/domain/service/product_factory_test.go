package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

func TestProductFactoryCreate(t *testing.T) {
	ctx := context.Background()
	catID := uuid.New()
	validPhoto := "products/123e4567-e89b-12d3-a456-426614174000.jpg"

	setup := func() *ProductFactory {
		products := newFakeProductRepo()
		categories := newFakeCategoryRepo()
		categories.byID[catID] = activeCategory(catID)
		return NewProductFactory(products, categories)
	}

	t.Run("success", func(t *testing.T) {
		f := setup()
		p, err := f.Create(ctx, CreateProductInput{
			Name: "Shirt", Description: "desc", PricePoints: 100,
			CategoryID: catID, PhotoKeys: []string{validPhoto, ""},
		})
		require.NoError(t, err)
		assert.Equal(t, "Shirt", p.Name)
		assert.Equal(t, catID, p.Category.ID)
		assert.Equal(t, []string{validPhoto}, p.PhotoKeyStrings())
	})

	t.Run("invalid price", func(t *testing.T) {
		_, err := setup().Create(ctx, CreateProductInput{Name: "Shirt", Description: "desc", PricePoints: 0, CategoryID: catID})
		assert.ErrorIs(t, err, domain.ErrInvalidPrice)
	})

	t.Run("invalid photo key", func(t *testing.T) {
		_, err := setup().Create(ctx, CreateProductInput{
			Name: "Shirt", Description: "desc", PricePoints: 100,
			CategoryID: catID, PhotoKeys: []string{"bad-key"},
		})
		assert.ErrorIs(t, err, domain.ErrInvalidPhotoKey)
	})

	t.Run("category not found", func(t *testing.T) {
		_, err := setup().Create(ctx, CreateProductInput{Name: "Shirt", Description: "desc", PricePoints: 100, CategoryID: uuid.New()})
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("inactive category rejected by model", func(t *testing.T) {
		products := newFakeProductRepo()
		categories := newFakeCategoryRepo()
		cat := activeCategory(catID)
		cat.Active = false
		categories.byID[catID] = cat
		f := NewProductFactory(products, categories)
		_, err := f.Create(ctx, CreateProductInput{Name: "Shirt", Description: "desc", PricePoints: 100, CategoryID: catID})
		assert.ErrorIs(t, err, domain.ErrCategoryNotActive)
	})
}

func TestProductFactoryUpdate(t *testing.T) {
	ctx := context.Background()
	catID := uuid.New()
	prodID := uuid.New()

	setup := func(t *testing.T) (*fakeProductRepo, *ProductFactory) {
		products := newFakeProductRepo()
		categories := newFakeCategoryRepo()
		categories.byID[catID] = activeCategory(catID)
		existing, err := model.NewProduct("Old", "old desc", vo.NewPricePointsFromStored(100), *activeCategory(catID), nil)
		require.NoError(t, err)
		existing.ID = prodID
		products.byID[prodID] = existing
		return products, NewProductFactory(products, categories)
	}

	t.Run("success", func(t *testing.T) {
		_, f := setup(t)
		p, err := f.Update(ctx, UpdateProductInput{
			ProductID: prodID, Name: "Renamed", Description: "new", PricePoints: 200,
			CategoryID: catID, Active: true, Version: 1,
		})
		require.NoError(t, err)
		assert.Equal(t, "Renamed", p.Name)
		assert.Equal(t, int64(200), p.Price.Int64())
	})

	t.Run("product not found", func(t *testing.T) {
		_, f := setup(t)
		_, err := f.Update(ctx, UpdateProductInput{ProductID: uuid.New(), Name: "x", Description: "y", PricePoints: 1, CategoryID: catID, Version: 1})
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})

	t.Run("version conflict", func(t *testing.T) {
		_, f := setup(t)
		_, err := f.Update(ctx, UpdateProductInput{ProductID: prodID, Name: "x", Description: "y", PricePoints: 1, CategoryID: catID, Version: 99})
		assert.ErrorIs(t, err, domain.ErrVersionConflict)
	})

	t.Run("invalid price", func(t *testing.T) {
		_, f := setup(t)
		_, err := f.Update(ctx, UpdateProductInput{ProductID: prodID, Name: "x", Description: "y", PricePoints: 0, CategoryID: catID, Version: 1})
		assert.ErrorIs(t, err, domain.ErrInvalidPrice)
	})

	t.Run("category not found", func(t *testing.T) {
		_, f := setup(t)
		_, err := f.Update(ctx, UpdateProductInput{ProductID: prodID, Name: "x", Description: "y", PricePoints: 1, CategoryID: uuid.New(), Version: 1})
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})
}
