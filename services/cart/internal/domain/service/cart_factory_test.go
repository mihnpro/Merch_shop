package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
)

func TestNewCartItem(t *testing.T) {
	f := NewCartFactory()
	cartID := uuid.New()
	productID := uuid.New()

	t.Run("valid populates all fields", func(t *testing.T) {
		item, err := f.NewCartItem(NewCartItemInput{
			CartID: cartID, ProductID: productID, Quantity: 3, PriceAtAdd: 150, ProductName: "Hoodie",
		})
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, item.ID)
		assert.Equal(t, cartID, item.CartID)
		assert.Equal(t, productID, item.ProductID)
		assert.Equal(t, 3, item.Quantity.Int())
		assert.Equal(t, 150, item.PriceAtAdd)
		assert.Equal(t, "Hoodie", item.ProductNameAtAdd)
	})

	t.Run("invalid quantity rejected", func(t *testing.T) {
		_, err := f.NewCartItem(NewCartItemInput{CartID: cartID, ProductID: productID, Quantity: 0, PriceAtAdd: 10})
		assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
	})
}
