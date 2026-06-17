package valueobject

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
)

func TestNewQuantity(t *testing.T) {
	t.Run("valid bounds", func(t *testing.T) {
		for _, v := range []int{1, 100, 9999} {
			q, err := NewQuantity(v)
			require.NoErrorf(t, err, "qty %d", v)
			assert.Equal(t, v, q.Int())
		}
	})

	for name, v := range map[string]int{"zero": 0, "negative": -1, "above max": 10000} {
		t.Run("invalid "+name, func(t *testing.T) {
			_, err := NewQuantity(v)
			assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
		})
	}

	t.Run("from stored skips validation", func(t *testing.T) {
		assert.Equal(t, 50000, NewQuantityFromStored(50000).Int())
	})
}

func TestQuantityJSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		out, err := json.Marshal(NewQuantityFromStored(7))
		require.NoError(t, err)
		assert.Equal(t, "7", string(out))
	})

	t.Run("unmarshal", func(t *testing.T) {
		var q Quantity
		require.NoError(t, json.Unmarshal([]byte("42"), &q))
		assert.Equal(t, 42, q.Int())
	})
}
