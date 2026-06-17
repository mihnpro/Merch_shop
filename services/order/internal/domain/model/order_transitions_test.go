package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
)

func TestOrderStatusCanTransitionTo(t *testing.T) {
	allowed := map[OrderStatus][]OrderStatus{
		StatusPending:       {StatusConfirmed, StatusCancelled},
		StatusConfirmed:     {StatusReadyToPickup, StatusCancelled},
		StatusReadyToPickup: {StatusDelivered},
	}
	all := []OrderStatus{StatusPending, StatusConfirmed, StatusReadyToPickup, StatusDelivered, StatusCancelled}

	isAllowed := func(from, to OrderStatus) bool {
		for _, a := range allowed[from] {
			if a == to {
				return true
			}
		}
		return false
	}

	for _, from := range all {
		for _, to := range all {
			want := isAllowed(from, to)
			assert.Equalf(t, want, from.CanTransitionTo(to), "%s -> %s", from, to)
		}
	}
}

func TestTransitionToTerminalStates(t *testing.T) {
	t.Run("delivered is terminal", func(t *testing.T) {
		assert.False(t, StatusDelivered.CanTransitionTo(StatusCancelled))
		assert.False(t, StatusDelivered.CanTransitionTo(StatusPending))
	})

	t.Run("cancelled is terminal", func(t *testing.T) {
		assert.False(t, StatusCancelled.CanTransitionTo(StatusConfirmed))
	})
}

func TestTransitionToMutatesStatus(t *testing.T) {
	o := &Order{Status: StatusPending}

	t.Run("valid transition updates status", func(t *testing.T) {
		require.NoError(t, o.TransitionTo(StatusConfirmed))
		assert.Equal(t, StatusConfirmed, o.Status)
	})

	t.Run("invalid transition rejected", func(t *testing.T) {
		err := o.TransitionTo(StatusPending)
		assert.ErrorIs(t, err, domain.ErrInvalidStatusChange)
		assert.Equal(t, StatusConfirmed, o.Status)
	})
}
