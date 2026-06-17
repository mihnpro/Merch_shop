package model

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

func activeCategory() Category {
	return Category{ID: uuid.New(), Code: vo.NewCategoryCodeFromStored("clothing"), Name: "Clothes", Active: true}
}

func TestNewProduct(t *testing.T) {
	price := vo.NewPricePointsFromStored(100)

	t.Run("valid trims and sets defaults", func(t *testing.T) {
		p, err := NewProduct("  Shirt  ", "  Nice shirt  ", price, activeCategory(), nil)
		require.NoError(t, err)
		assert.Equal(t, "Shirt", p.Name)
		assert.Equal(t, "Nice shirt", p.Description)
		assert.True(t, p.Active)
		assert.Equal(t, 1, p.Version)
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := NewProduct("   ", "desc", price, activeCategory(), nil)
		assert.ErrorIs(t, err, domain.ErrEmptyName)
	})

	t.Run("name too long", func(t *testing.T) {
		_, err := NewProduct(strings.Repeat("a", 501), "desc", price, activeCategory(), nil)
		assert.ErrorIs(t, err, domain.ErrNameTooLong)
	})

	t.Run("empty description", func(t *testing.T) {
		_, err := NewProduct("Shirt", "  ", price, activeCategory(), nil)
		assert.ErrorIs(t, err, domain.ErrEmptyDescription)
	})

	t.Run("description too long", func(t *testing.T) {
		_, err := NewProduct("Shirt", strings.Repeat("a", 5001), price, activeCategory(), nil)
		assert.ErrorIs(t, err, domain.ErrDescriptionTooLong)
	})

	t.Run("invalid price", func(t *testing.T) {
		_, err := NewProduct("Shirt", "desc", vo.NewPricePointsFromStored(0), activeCategory(), nil)
		assert.ErrorIs(t, err, domain.ErrInvalidPrice)
	})

	t.Run("category without id", func(t *testing.T) {
		cat := activeCategory()
		cat.ID = uuid.Nil
		_, err := NewProduct("Shirt", "desc", price, cat, nil)
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("inactive category", func(t *testing.T) {
		cat := activeCategory()
		cat.Active = false
		_, err := NewProduct("Shirt", "desc", price, cat, nil)
		assert.ErrorIs(t, err, domain.ErrCategoryNotActive)
	})
}

func TestProductApplyUpdate(t *testing.T) {
	price := vo.NewPricePointsFromStored(100)
	newPrice := vo.NewPricePointsFromStored(250)

	newProduct := func(t *testing.T) *Product {
		p, err := NewProduct("Shirt", "desc", price, activeCategory(), nil)
		require.NoError(t, err)
		return p
	}

	t.Run("version conflict", func(t *testing.T) {
		p := newProduct(t)
		err := p.ApplyUpdate("New", "new desc", newPrice, activeCategory(), nil, true, 2)
		assert.ErrorIs(t, err, domain.ErrVersionConflict)
	})

	t.Run("success updates fields", func(t *testing.T) {
		p := newProduct(t)
		require.NoError(t, p.ApplyUpdate("  New  ", "  new desc  ", newPrice, activeCategory(), nil, false, 1))
		assert.Equal(t, "New", p.Name)
		assert.Equal(t, "new desc", p.Description)
		assert.Equal(t, int64(250), p.Price.Int64())
		assert.False(t, p.Active)
	})

	t.Run("validates after update", func(t *testing.T) {
		p := newProduct(t)
		err := p.ApplyUpdate("", "desc", newPrice, activeCategory(), nil, true, 1)
		assert.ErrorIs(t, err, domain.ErrEmptyName)
	})
}

func TestProductDeactivateAndPhotoKeys(t *testing.T) {
	keys := []vo.PhotoKey{
		vo.NewPhotoKeyFromStored("products/a.jpg"),
		vo.NewPhotoKeyFromStored("products/b.png"),
	}
	p, err := NewProduct("Shirt", "desc", vo.NewPricePointsFromStored(100), activeCategory(), keys)
	require.NoError(t, err)

	p.Deactivate()
	assert.False(t, p.Active)
	assert.Equal(t, []string{"products/a.jpg", "products/b.png"}, p.PhotoKeyStrings())
}
