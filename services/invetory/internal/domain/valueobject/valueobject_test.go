package valueobject

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
)

func TestNewQuantity(t *testing.T) {
	t.Run("valid bounds", func(t *testing.T) {
		for _, v := range []int{1, 50, 100} {
			q, err := NewQuantity(v)
			require.NoErrorf(t, err, "qty %d", v)
			assert.Equal(t, v, q.Int())
		}
	})

	for name, v := range map[string]int{"zero": 0, "negative": -1, "above max": 101} {
		t.Run("invalid "+name, func(t *testing.T) {
			_, err := NewQuantity(v)
			assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
		})
	}

	t.Run("from stored skips validation", func(t *testing.T) {
		assert.Equal(t, 999, NewQuantityFromStored(999).Int())
	})
}

func TestNewDelta(t *testing.T) {
	t.Run("valid non-zero", func(t *testing.T) {
		for _, v := range []int{1, -5, 100} {
			d, err := NewDelta(v)
			require.NoErrorf(t, err, "delta %d", v)
			assert.Equal(t, v, d.Int())
		}
	})

	t.Run("zero is invalid", func(t *testing.T) {
		_, err := NewDelta(0)
		assert.ErrorIs(t, err, domain.ErrZeroDelta)
	})

	t.Run("from stored skips validation", func(t *testing.T) {
		assert.Equal(t, 0, NewDeltaFromStored(0).Int())
	})
}

func TestNewReason(t *testing.T) {
	t.Run("valid trims", func(t *testing.T) {
		r, err := NewReason("  restock  ")
		require.NoError(t, err)
		assert.Equal(t, "restock", r.String())
	})

	t.Run("empty", func(t *testing.T) {
		_, err := NewReason("   ")
		assert.ErrorIs(t, err, domain.ErrEmptyReason)
	})

	t.Run("too long", func(t *testing.T) {
		_, err := NewReason(strings.Repeat("a", 501))
		assert.ErrorIs(t, err, domain.ErrReasonTooLong)
	})

	t.Run("max length ok", func(t *testing.T) {
		_, err := NewReason(strings.Repeat("a", 500))
		assert.NoError(t, err)
	})

	t.Run("from stored skips validation", func(t *testing.T) {
		assert.Equal(t, "", NewReasonFromStored("").String())
	})
}
