package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
)

func TestCategoryFactoryCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		categories := newFakeCategoryRepo()
		c, err := NewCategoryFactory(categories).Create(ctx, CreateCategoryInput{Code: "clothing", Name: "Clothes"})
		require.NoError(t, err)
		assert.Equal(t, "clothing", c.Code.String())
		assert.True(t, c.Active)
	})

	t.Run("invalid code", func(t *testing.T) {
		_, err := NewCategoryFactory(newFakeCategoryRepo()).Create(ctx, CreateCategoryInput{Code: "BAD CODE", Name: "x"})
		assert.ErrorIs(t, err, domain.ErrInvalidCategoryCode)
	})

	t.Run("already exists", func(t *testing.T) {
		categories := newFakeCategoryRepo()
		categories.existsRet = true
		_, err := NewCategoryFactory(categories).Create(ctx, CreateCategoryInput{Code: "clothing", Name: "Clothes"})
		assert.ErrorIs(t, err, domain.ErrCategoryAlreadyExists)
	})

	t.Run("exists check fails", func(t *testing.T) {
		categories := newFakeCategoryRepo()
		categories.existsErr = errBoom
		_, err := NewCategoryFactory(categories).Create(ctx, CreateCategoryInput{Code: "clothing", Name: "Clothes"})
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := NewCategoryFactory(newFakeCategoryRepo()).Create(ctx, CreateCategoryInput{Code: "clothing", Name: "  "})
		assert.ErrorIs(t, err, domain.ErrEmptyCategoryName)
	})
}

func TestCategoryFactoryUpdate(t *testing.T) {
	ctx := context.Background()
	catID := uuid.New()

	setup := func() *fakeCategoryRepo {
		categories := newFakeCategoryRepo()
		categories.byID[catID] = activeCategory(catID)
		return categories
	}

	t.Run("success renames and sets active", func(t *testing.T) {
		categories := setup()
		c, err := NewCategoryFactory(categories).Update(ctx, UpdateCategoryInput{ID: catID, Name: "Apparel", Active: false})
		require.NoError(t, err)
		assert.Equal(t, "Apparel", c.Name)
		assert.False(t, c.Active)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := NewCategoryFactory(setup()).Update(ctx, UpdateCategoryInput{ID: uuid.New(), Name: "x", Active: true})
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("invalid name", func(t *testing.T) {
		_, err := NewCategoryFactory(setup()).Update(ctx, UpdateCategoryInput{ID: catID, Name: "", Active: true})
		assert.ErrorIs(t, err, domain.ErrEmptyCategoryName)
	})
}
