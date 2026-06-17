package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

func TestNewCategory(t *testing.T) {
	code := vo.NewCategoryCodeFromStored("clothing")

	t.Run("valid trims name and is active", func(t *testing.T) {
		c, err := NewCategory(code, "  Clothes  ")
		require.NoError(t, err)
		assert.Equal(t, "Clothes", c.Name)
		assert.True(t, c.Active)
	})

	t.Run("empty code", func(t *testing.T) {
		_, err := NewCategory(vo.NewCategoryCodeFromStored(""), "Clothes")
		assert.ErrorIs(t, err, domain.ErrInvalidCategoryCode)
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := NewCategory(code, "   ")
		assert.ErrorIs(t, err, domain.ErrEmptyCategoryName)
	})

	t.Run("name too long", func(t *testing.T) {
		_, err := NewCategory(code, strings.Repeat("a", 101))
		assert.ErrorIs(t, err, domain.ErrCategoryNameTooLong)
	})
}

func TestCategoryRenameAndSetActive(t *testing.T) {
	c, err := NewCategory(vo.NewCategoryCodeFromStored("clothing"), "Clothes")
	require.NoError(t, err)

	t.Run("rename valid", func(t *testing.T) {
		require.NoError(t, c.Rename("  Apparel  "))
		assert.Equal(t, "Apparel", c.Name)
	})

	t.Run("rename invalid", func(t *testing.T) {
		assert.ErrorIs(t, c.Rename(""), domain.ErrEmptyCategoryName)
	})

	t.Run("set active", func(t *testing.T) {
		c.SetActive(false)
		assert.False(t, c.Active)
		c.SetActive(true)
		assert.True(t, c.Active)
	})
}
