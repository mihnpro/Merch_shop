package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/valueobject"
)

func TestNewStockAdjustment(t *testing.T) {
	delta := vo.NewDeltaFromStored(5)
	reason := vo.NewReasonFromStored("restock")

	t.Run("valid", func(t *testing.T) {
		a, err := NewStockAdjustment(uuid.New(), uuid.New(), delta, reason)
		require.NoError(t, err)
		assert.Equal(t, 5, a.Delta.Int())
		assert.Equal(t, "restock", a.Reason.String())
	})

	t.Run("nil operation id", func(t *testing.T) {
		_, err := NewStockAdjustment(uuid.Nil, uuid.New(), delta, reason)
		assert.ErrorIs(t, err, domain.ErrInvalidOperationID)
	})

	t.Run("nil product id", func(t *testing.T) {
		_, err := NewStockAdjustment(uuid.New(), uuid.Nil, delta, reason)
		assert.ErrorIs(t, err, domain.ErrInvalidProductID)
	})
}
