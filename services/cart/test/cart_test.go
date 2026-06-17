package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEmptyCart(t *testing.T) {
	user := newUserClient(t)
	status, raw, err := user.do(http.MethodGet, "/cart", nil)
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
	view := decode[cartView](t, raw)
	assert.Equal(t, 0, view.ItemCount)
	assert.Equal(t, 0, view.Total)
	assert.NotNil(t, view.Items)
}

func TestAddItem(t *testing.T) {
	admin := adminClient(t)
	user := newUserClient(t)
	product, _ := createProduct(t, admin, 100)
	stockUp(t, admin, product.ID, 50)

	t.Run("happy path", func(t *testing.T) {
		status, raw := addToCart(t, user, product.ID, 2)
		requireStatus(t, http.StatusCreated, status, raw)
		item := decode[cartItemView](t, raw)
		assert.Equal(t, product.ID, item.ProductID)
		assert.Equal(t, 2, item.Quantity)
		assert.Equal(t, 100, item.PricePerUnit)
		assert.Equal(t, 200, item.LineTotal)
	})

	t.Run("cart reflects item", func(t *testing.T) {
		status, raw, err := user.do(http.MethodGet, "/cart", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		view := decode[cartView](t, raw)
		assert.Equal(t, 1, view.ItemCount)
		assert.Equal(t, 200, view.Total)
	})

	t.Run("adding same product merges quantity", func(t *testing.T) {
		status, raw := addToCart(t, user, product.ID, 3)
		requireStatus(t, http.StatusCreated, status, raw)
		assert.Equal(t, 5, decode[cartItemView](t, raw).Quantity)
	})
}

func TestUpdateAndRemoveItem(t *testing.T) {
	admin := adminClient(t)
	user := newUserClient(t)
	product, _ := createProduct(t, admin, 40)
	stockUp(t, admin, product.ID, 50)

	status, raw := addToCart(t, user, product.ID, 2)
	requireStatus(t, http.StatusCreated, status, raw)
	itemID := decode[cartItemView](t, raw).ID

	t.Run("update quantity", func(t *testing.T) {
		status, raw, err := user.do(http.MethodPatch, "/cart/items/"+itemID, updateCartItemBody{Quantity: 4})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, 4, decode[cartItemView](t, raw).Quantity)
	})

	t.Run("remove item", func(t *testing.T) {
		status, raw, err := user.do(http.MethodDelete, "/cart/items/"+itemID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusNoContent, status, raw)

		status, raw, err = user.do(http.MethodGet, "/cart", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, 0, decode[cartView](t, raw).ItemCount)
	})
}

func TestClearCart(t *testing.T) {
	admin := adminClient(t)
	user := newUserClient(t)
	product, _ := createProduct(t, admin, 25)
	stockUp(t, admin, product.ID, 50)

	status, raw := addToCart(t, user, product.ID, 2)
	requireStatus(t, http.StatusCreated, status, raw)

	status, raw, err := user.do(http.MethodDelete, "/cart", nil)
	require.NoError(t, err)
	requireStatus(t, http.StatusNoContent, status, raw)

	status, raw, err = user.do(http.MethodGet, "/cart", nil)
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
	assert.Equal(t, 0, decode[cartView](t, raw).ItemCount)
}
