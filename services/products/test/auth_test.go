package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthorization(t *testing.T) {
	t.Run("list products without token → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodGet, "/products", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("list categories without token → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodGet, "/categories", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("non-admin cannot create product → 403", func(t *testing.T) {
		user := newUserClient(t)
		status, raw, err := user.do(http.MethodPost, "/admin/products", createProductBody{
			Name: "x", Description: "y", PricePoints: 1, CategoryID: "00000000-0000-0000-0000-000000000000",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})

	t.Run("non-admin cannot create category → 403", func(t *testing.T) {
		user := newUserClient(t)
		status, raw, err := user.do(http.MethodPost, "/admin/categories", createCategoryBody{Code: uniqueCode(), Name: "x"})
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})
}
