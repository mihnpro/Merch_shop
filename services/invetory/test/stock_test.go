package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStock(t *testing.T) {
	admin := adminClient(t)
	productID := newProductID()
	status, raw := adjust(t, admin, productID, 40, uuid.NewString())
	requireStatus(t, http.StatusOK, status, raw)

	t.Run("returns available for adjusted product", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/stock?product_ids="+productID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		list := decode[listStockView](t, raw)
		require.Len(t, list.Items, 1)
		assert.Equal(t, productID, list.Items[0].ProductID)
		assert.Equal(t, 40, list.Items[0].Available)
	})

	t.Run("unknown product reports zero", func(t *testing.T) {
		unknown := uuid.NewString()
		status, raw, err := admin.do(http.MethodGet, "/stock?product_ids="+unknown, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		list := decode[listStockView](t, raw)
		require.Len(t, list.Items, 1)
		assert.Equal(t, 0, list.Items[0].Available)
	})

	t.Run("authenticated non-admin can read stock", func(t *testing.T) {
		user := newUserClient(t)
		status, raw, err := user.do(http.MethodGet, "/stock?product_ids="+productID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
	})
}

func TestAdminListInventory(t *testing.T) {
	admin := adminClient(t)
	productID := newProductID()
	status, raw := adjust(t, admin, productID, 15, uuid.NewString())
	requireStatus(t, http.StatusOK, status, raw)

	status, raw, err := admin.do(http.MethodGet, "/admin/inventory", nil)
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)

	list := decode[listStockView](t, raw)
	found := false
	for _, it := range list.Items {
		if it.ProductID == productID {
			found = true
			assert.Equal(t, 15, it.Available)
		}
	}
	assert.True(t, found, "adjusted product not present in admin inventory listing")
}
