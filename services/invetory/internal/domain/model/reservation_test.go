package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/valueobject"
)

func validItems() []ReservationItem {
	return []ReservationItem{{ProductID: uuid.New(), Qty: vo.NewQuantityFromStored(2)}}
}

func TestNewReservation(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		r, err := NewReservation(uuid.New(), validItems())
		require.NoError(t, err)
		assert.False(t, r.IsReleased())
	})

	t.Run("nil order id", func(t *testing.T) {
		_, err := NewReservation(uuid.Nil, validItems())
		assert.ErrorIs(t, err, domain.ErrInvalidOrderID)
	})

	t.Run("empty items", func(t *testing.T) {
		_, err := NewReservation(uuid.New(), nil)
		assert.ErrorIs(t, err, domain.ErrEmptyReservation)
	})

	t.Run("item with nil product id", func(t *testing.T) {
		items := []ReservationItem{{ProductID: uuid.Nil, Qty: vo.NewQuantityFromStored(1)}}
		_, err := NewReservation(uuid.New(), items)
		assert.ErrorIs(t, err, domain.ErrInvalidProductID)
	})

	t.Run("item with zero quantity", func(t *testing.T) {
		items := []ReservationItem{{ProductID: uuid.New(), Qty: vo.NewQuantityFromStored(0)}}
		_, err := NewReservation(uuid.New(), items)
		assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
	})
}

func TestReservationRelease(t *testing.T) {
	r, err := NewReservation(uuid.New(), validItems())
	require.NoError(t, err)

	t.Run("first release sets reason", func(t *testing.T) {
		require.NoError(t, r.Release("order cancelled"))
		assert.Equal(t, "order cancelled", r.ReleasedReason)
	})

	t.Run("release when already released is rejected", func(t *testing.T) {
		now := time.Now()
		r.ReleasedAt = &now
		assert.ErrorIs(t, r.Release("again"), domain.ErrAlreadyReleased)
	})
}
