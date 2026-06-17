package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthorization(t *testing.T) {
	t.Run("get cart without token → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodGet, "/cart", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("add item without token → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodPost, "/cart/items", addCartItemBody{ProductID: uuid.NewString(), Quantity: 1})
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})
}
