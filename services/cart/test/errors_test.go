package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestAddItemErrors(t *testing.T) {
	admin := adminClient(t)

	t.Run("out of stock → 409", func(t *testing.T) {
		user := newUserClient(t)
		product, _ := createProduct(t, admin, 100)
		stockUp(t, admin, product.ID, 1)
		status, raw := addToCart(t, user, product.ID, 5)
		requireStatus(t, http.StatusConflict, status, raw)
	})

	t.Run("inactive product → 409", func(t *testing.T) {
		user := newUserClient(t)
		product, catID := createProduct(t, admin, 100)
		stockUp(t, admin, product.ID, 50)
		deactivateProduct(t, admin, product, catID)
		status, raw := addToCart(t, user, product.ID, 1)
		requireStatus(t, http.StatusConflict, status, raw)
	})

	t.Run("product not found → 404", func(t *testing.T) {
		user := newUserClient(t)
		phantom := uuid.NewString()
		trackProduct(phantom)
		stockUp(t, admin, phantom, 50)
		status, raw := addToCart(t, user, phantom, 1)
		requireStatus(t, http.StatusNotFound, status, raw)
	})
}
