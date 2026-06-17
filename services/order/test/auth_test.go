package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthorization(t *testing.T) {
	t.Run("create order without token → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodPost, "/orders", createOrderBody{DeliveryAddress: "x"})
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("list orders without token → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodGet, "/orders", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("non-admin cannot list all orders → 403", func(t *testing.T) {
		user, _ := newUserClient(t)
		status, raw, err := user.do(http.MethodGet, "/admin/orders", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})

	t.Run("non-admin cannot update status → 403", func(t *testing.T) {
		user, _ := newUserClient(t)
		status, raw, err := user.do(http.MethodPut, "/admin/orders/"+uuid.NewString()+"/status", updateOrderStatusBody{Status: "confirmed"})
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})

	t.Run("non-admin cannot read analytics → 403", func(t *testing.T) {
		user, _ := newUserClient(t)
		status, raw, err := user.do(http.MethodGet, "/admin/analytics?period=day", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})
}
