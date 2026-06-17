package valueobject

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
)

func TestNewCategoryCode(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c, err := NewCategoryCode("  clothing  ")
		require.NoError(t, err)
		assert.Equal(t, "clothing", c.String())
	})

	t.Run("underscores allowed", func(t *testing.T) {
		c, err := NewCategoryCode("home_decor")
		require.NoError(t, err)
		assert.Equal(t, "home_decor", c.String())
	})

	for name, raw := range map[string]string{
		"empty":      "",
		"uppercase":  "Clothing",
		"digits":     "cat1",
		"dash":       "home-decor",
		"whitespace": "   ",
		"too long":   strings.Repeat("a", 31),
	} {
		t.Run("invalid "+name, func(t *testing.T) {
			_, err := NewCategoryCode(raw)
			assert.ErrorIs(t, err, domain.ErrInvalidCategoryCode)
		})
	}

	t.Run("max length 30 ok", func(t *testing.T) {
		_, err := NewCategoryCode(strings.Repeat("a", 30))
		assert.NoError(t, err)
	})

	t.Run("from stored skips validation", func(t *testing.T) {
		assert.Equal(t, "INVALID!", NewCategoryCodeFromStored("INVALID!").String())
	})
}

func TestNewPricePoints(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		p, err := NewPricePoints(100)
		require.NoError(t, err)
		assert.Equal(t, int64(100), p.Int64())
	})

	for name, v := range map[string]int64{"zero": 0, "negative": -5} {
		t.Run("invalid "+name, func(t *testing.T) {
			_, err := NewPricePoints(v)
			assert.ErrorIs(t, err, domain.ErrInvalidPrice)
		})
	}

	t.Run("from stored skips validation", func(t *testing.T) {
		assert.Equal(t, int64(-1), NewPricePointsFromStored(-1).Int64())
	})
}

func TestNewPhotoKey(t *testing.T) {
	valid := "products/123e4567-e89b-12d3-a456-426614174000.jpg"

	t.Run("valid", func(t *testing.T) {
		k, err := NewPhotoKey(valid)
		require.NoError(t, err)
		assert.Equal(t, valid, k.String())
		assert.False(t, k.IsEmpty())
	})

	t.Run("empty is allowed and marked empty", func(t *testing.T) {
		k, err := NewPhotoKey("   ")
		require.NoError(t, err)
		assert.True(t, k.IsEmpty())
		assert.Equal(t, "", k.String())
	})

	for name, raw := range map[string]string{
		"no prefix":     "123e4567-e89b-12d3-a456-426614174000.jpg",
		"bad extension": "products/123e4567-e89b-12d3-a456-426614174000.gif",
		"missing uuid":  "products/.jpg",
		"arbitrary":     "not-a-key",
	} {
		t.Run("invalid "+name, func(t *testing.T) {
			_, err := NewPhotoKey(raw)
			assert.ErrorIs(t, err, domain.ErrInvalidPhotoKey)
		})
	}

	t.Run("png and webp accepted", func(t *testing.T) {
		for _, ext := range []string{"jpeg", "png", "webp"} {
			_, err := NewPhotoKey("products/123e4567-e89b-12d3-a456-426614174000." + ext)
			assert.NoErrorf(t, err, "ext %s", ext)
		}
	})
}
