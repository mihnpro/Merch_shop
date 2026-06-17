package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
)

func TestNewStock(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		s, err := NewStock(uuid.New())
		require.NoError(t, err)
		assert.Equal(t, 0, s.Available)
		assert.Equal(t, 1, s.Version)
	})

	t.Run("nil product id", func(t *testing.T) {
		_, err := NewStock(uuid.Nil)
		assert.ErrorIs(t, err, domain.ErrInvalidProductID)
	})
}

func TestStockAdjust(t *testing.T) {
	s, err := NewStock(uuid.New())
	require.NoError(t, err)

	t.Run("increase", func(t *testing.T) {
		require.NoError(t, s.Adjust(10))
		assert.Equal(t, 10, s.Available)
	})

	t.Run("decrease within bounds", func(t *testing.T) {
		require.NoError(t, s.Adjust(-4))
		assert.Equal(t, 6, s.Available)
	})

	t.Run("decrease below zero rejected", func(t *testing.T) {
		err := s.Adjust(-100)
		assert.ErrorIs(t, err, domain.ErrInsufficientStock)
		assert.Equal(t, 6, s.Available)
	})
}

func TestStockReserveReleaseCanReserve(t *testing.T) {
	s, err := NewStock(uuid.New())
	require.NoError(t, err)
	require.NoError(t, s.Adjust(10))

	t.Run("reserve within available", func(t *testing.T) {
		require.NoError(t, s.Reserve(7))
		assert.Equal(t, 3, s.Available)
	})

	t.Run("reserve beyond available rejected", func(t *testing.T) {
		err := s.Reserve(5)
		assert.ErrorIs(t, err, domain.ErrInsufficientStock)
		assert.Equal(t, 3, s.Available)
	})

	t.Run("release restores", func(t *testing.T) {
		s.Release(2)
		assert.Equal(t, 5, s.Available)
	})

	t.Run("can reserve", func(t *testing.T) {
		assert.True(t, s.CanReserve(5))
		assert.False(t, s.CanReserve(6))
	})
}
