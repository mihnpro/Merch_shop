package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthorization(t *testing.T) {
	t.Run("get stock without token → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodGet, "/stock?product_ids="+uuid.NewString(), nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("non-admin cannot list inventory → 403", func(t *testing.T) {
		user := newUserClient(t)
		status, raw, err := user.do(http.MethodGet, "/admin/inventory", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})

	t.Run("non-admin cannot adjust → 403", func(t *testing.T) {
		user := newUserClient(t)
		status, raw, err := user.do(http.MethodPost, "/admin/inventory/adjust", adjustStockBody{
			ProductID: uuid.NewString(), Delta: 1, OperationID: uuid.NewString(), Reason: "x",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})

	t.Run("get stock without product_ids → 400", func(t *testing.T) {
		admin := adminClient(t)
		status, raw, err := admin.do(http.MethodGet, "/stock", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})
}
