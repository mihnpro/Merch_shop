package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProduct(t *testing.T) {
	admin := adminClient(t)
	cat := createCategory(t, admin)

	t.Run("happy path", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPost, "/admin/products", createProductBody{
			Name: "Hoodie", Description: "warm", PricePoints: 500, CategoryID: cat.ID,
			PhotoKeys: []string{photoKey()},
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusCreated, status, raw)
		p := decode[productView](t, raw)
		trackProduct(p.ID)
		assert.Equal(t, "Hoodie", p.Name)
		assert.Equal(t, int64(500), p.PricePoints)
		assert.Equal(t, cat.ID, p.Category.ID)
		assert.True(t, p.Active)
		assert.Equal(t, 1, p.Version)
		assert.Len(t, p.PhotoKeys, 1)
	})

	t.Run("invalid price → 400", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPost, "/admin/products", createProductBody{
			Name: "x", Description: "y", PricePoints: 0, CategoryID: cat.ID,
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("invalid photo key → 400", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPost, "/admin/products", createProductBody{
			Name: "x", Description: "y", PricePoints: 100, CategoryID: cat.ID, PhotoKeys: []string{"bad-key"},
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("nonexistent category → 404", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPost, "/admin/products", createProductBody{
			Name: "x", Description: "y", PricePoints: 100, CategoryID: uuid.NewString(),
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusNotFound, status, raw)
	})
}

func TestGetProduct(t *testing.T) {
	admin := adminClient(t)
	cat := createCategory(t, admin)
	created := createProduct(t, admin, cat.ID)

	t.Run("by id → 200", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/products/"+created.ID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, created.ID, decode[productView](t, raw).ID)
	})

	t.Run("nonexistent → 404", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/products/"+uuid.NewString(), nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusNotFound, status, raw)
	})

	t.Run("malformed id → 404", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/products/not-a-uuid", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusNotFound, status, raw)
	})
}

func TestListProducts(t *testing.T) {
	admin := adminClient(t)
	cat := createCategory(t, admin)
	created := createProduct(t, admin, cat.ID)

	t.Run("filter by category returns our product", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/products?category_id="+cat.ID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		list := decode[listProductsView](t, raw)
		require.NotEmpty(t, list.Products)

		found := false
		for _, p := range list.Products {
			if p.ID == created.ID {
				found = true
			}
			assert.Equal(t, cat.ID, p.Category.ID)
		}
		assert.True(t, found, "created product not in category listing")
	})
}

func TestUpdateProductOptimisticLocking(t *testing.T) {
	admin := adminClient(t)
	cat := createCategory(t, admin)
	created := createProduct(t, admin, cat.ID)

	body := updateProductBody{
		Name: "Updated", Description: "updated desc", PricePoints: 999,
		CategoryID: cat.ID, Active: true, Version: created.Version,
	}

	t.Run("update with correct version → 200", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/products/"+created.ID, body)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		updated := decode[productView](t, raw)
		assert.Equal(t, "Updated", updated.Name)
		assert.Equal(t, int64(999), updated.PricePoints)
		assert.Greater(t, updated.Version, created.Version)
	})

	t.Run("update with stale version → 409", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/products/"+created.ID, body)
		require.NoError(t, err)
		requireStatus(t, http.StatusConflict, status, raw)
	})
}

func TestDeactivateProduct(t *testing.T) {
	admin := adminClient(t)
	cat := createCategory(t, admin)
	created := createProduct(t, admin, cat.ID)

	t.Run("deactivate → 204", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodDelete, "/admin/products/"+created.ID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusNoContent, status, raw)
	})

	t.Run("excluded from active-only listing", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/products?category_id="+cat.ID+"&active_only=true", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		for _, p := range decode[listProductsView](t, raw).Products {
			assert.NotEqual(t, created.ID, p.ID, "deactivated product still listed")
		}
	})

	t.Run("still retrievable by id with active=false", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/products/"+created.ID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		assert.False(t, decode[productView](t, raw).Active)
	})
}
