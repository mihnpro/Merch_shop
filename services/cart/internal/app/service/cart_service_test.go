package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/cart/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/cart/internal/domain/valueobject"
)

type deps struct {
	repo      *fakeCartRepo
	cache     *fakeCache
	products  *fakeProducts
	inventory *fakeInventory
	svc       CartService
}

func newDeps() *deps {
	repo := newFakeCartRepo()
	cache := &fakeCache{}
	products := &fakeProducts{info: activeProduct()}
	inventory := &fakeInventory{available: 1000}
	return &deps{
		repo:      repo,
		cache:     cache,
		products:  products,
		inventory: inventory,
		svc:       NewCartService(repo, products, inventory, cache, zap.NewNop()),
	}
}

func addInput(productID string, qty int) dto.AddItemInput {
	return dto.AddItemInput{UserID: uuid.NewString(), ProductID: productID, Quantity: qty}
}

func TestAddItem(t *testing.T) {
	ctx := context.Background()

	t.Run("new item inserted", func(t *testing.T) {
		d := newDeps()
		view, err := d.svc.AddItem(ctx, addInput(uuid.NewString(), 2))
		require.NoError(t, err)
		assert.Equal(t, 2, view.Quantity)
		assert.Equal(t, 100, view.PricePerUnit)
		assert.Equal(t, 200, view.LineTotal)
		assert.Equal(t, 1, d.cache.invalidateCalled)
	})

	t.Run("existing item merges quantity", func(t *testing.T) {
		d := newDeps()
		d.repo.itemByProduct = &model.CartItem{
			ID: uuid.New(), CartID: d.repo.cart.ID, ProductID: uuid.New(),
			Quantity: vo.NewQuantityFromStored(3), PriceAtAdd: 100, ProductNameAtAdd: "Hoodie",
		}
		view, err := d.svc.AddItem(ctx, addInput(uuid.NewString(), 2))
		require.NoError(t, err)
		assert.Equal(t, 5, view.Quantity)
		assert.Equal(t, 5, d.repo.updatedQty)
	})

	t.Run("out of stock", func(t *testing.T) {
		d := newDeps()
		d.inventory.available = 0
		_, err := d.svc.AddItem(ctx, addInput(uuid.NewString(), 1))
		assert.ErrorIs(t, err, domain.ErrOutOfStock)
	})

	t.Run("out of stock on merge", func(t *testing.T) {
		d := newDeps()
		d.inventory.available = 4
		d.repo.itemByProduct = &model.CartItem{
			ID: uuid.New(), CartID: d.repo.cart.ID, Quantity: vo.NewQuantityFromStored(3), PriceAtAdd: 100,
		}
		_, err := d.svc.AddItem(ctx, addInput(uuid.NewString(), 2))
		assert.ErrorIs(t, err, domain.ErrOutOfStock)
	})

	t.Run("inactive product", func(t *testing.T) {
		d := newDeps()
		d.products.info = dto.ProductInfo{Name: "X", PricePoints: 10, Active: false}
		_, err := d.svc.AddItem(ctx, addInput(uuid.NewString(), 1))
		assert.ErrorIs(t, err, domain.ErrProductInactive)
	})

	t.Run("product lookup error propagates", func(t *testing.T) {
		d := newDeps()
		d.products.err = domain.ErrProductNotFound
		_, err := d.svc.AddItem(ctx, addInput(uuid.NewString(), 1))
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})

	t.Run("empty input", func(t *testing.T) {
		d := newDeps()
		_, err := d.svc.AddItem(ctx, dto.AddItemInput{UserID: "", ProductID: uuid.NewString(), Quantity: 1})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("invalid user id", func(t *testing.T) {
		d := newDeps()
		_, err := d.svc.AddItem(ctx, dto.AddItemInput{UserID: "bad", ProductID: uuid.NewString(), Quantity: 1})
		assert.ErrorIs(t, err, domain.ErrInvalidUserID)
	})

	t.Run("invalid product id", func(t *testing.T) {
		d := newDeps()
		_, err := d.svc.AddItem(ctx, dto.AddItemInput{UserID: uuid.NewString(), ProductID: "bad", Quantity: 1})
		assert.ErrorIs(t, err, domain.ErrInvalidProductID)
	})
}

func TestUpdateItem(t *testing.T) {
	ctx := context.Background()

	seedItem := func(d *deps, qty int) *model.CartItem {
		item := &model.CartItem{
			ID: uuid.New(), CartID: d.repo.cart.ID, ProductID: uuid.New(),
			Quantity: vo.NewQuantityFromStored(qty), PriceAtAdd: 100,
		}
		d.repo.itemByID = item
		return item
	}

	t.Run("zero quantity removes item", func(t *testing.T) {
		d := newDeps()
		item := seedItem(d, 2)
		_, err := d.svc.UpdateItem(ctx, dto.UpdateItemInput{UserID: uuid.NewString(), ItemID: item.ID.String(), NewQuantity: 0})
		require.NoError(t, err)
		assert.True(t, d.repo.deleteCalled)
	})

	t.Run("negative quantity rejected", func(t *testing.T) {
		d := newDeps()
		_, err := d.svc.UpdateItem(ctx, dto.UpdateItemInput{UserID: uuid.NewString(), ItemID: uuid.NewString(), NewQuantity: -1})
		assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
	})

	t.Run("empty input", func(t *testing.T) {
		d := newDeps()
		_, err := d.svc.UpdateItem(ctx, dto.UpdateItemInput{UserID: uuid.NewString(), ItemID: "", NewQuantity: 1})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("item not found", func(t *testing.T) {
		d := newDeps()
		d.repo.itemByIDErr = domain.ErrItemNotFound
		_, err := d.svc.UpdateItem(ctx, dto.UpdateItemInput{UserID: uuid.NewString(), ItemID: uuid.NewString(), NewQuantity: 1})
		assert.ErrorIs(t, err, domain.ErrItemNotFound)
	})

	t.Run("item belongs to another cart", func(t *testing.T) {
		d := newDeps()
		item := seedItem(d, 2)
		item.CartID = uuid.New()
		_, err := d.svc.UpdateItem(ctx, dto.UpdateItemInput{UserID: uuid.NewString(), ItemID: item.ID.String(), NewQuantity: 3})
		assert.ErrorIs(t, err, domain.ErrItemNotFound)
	})

	t.Run("increase checks stock", func(t *testing.T) {
		d := newDeps()
		d.inventory.available = 3
		item := seedItem(d, 2)
		_, err := d.svc.UpdateItem(ctx, dto.UpdateItemInput{UserID: uuid.NewString(), ItemID: item.ID.String(), NewQuantity: 5})
		assert.ErrorIs(t, err, domain.ErrOutOfStock)
	})

	t.Run("decrease skips stock check", func(t *testing.T) {
		d := newDeps()
		d.inventory.available = 0
		item := seedItem(d, 5)
		view, err := d.svc.UpdateItem(ctx, dto.UpdateItemInput{UserID: uuid.NewString(), ItemID: item.ID.String(), NewQuantity: 2})
		require.NoError(t, err)
		assert.Equal(t, 2, view.Quantity)
		assert.Equal(t, 2, d.repo.updatedQty)
	})
}

func TestRemoveItem(t *testing.T) {
	ctx := context.Background()

	t.Run("success deletes", func(t *testing.T) {
		d := newDeps()
		require.NoError(t, d.svc.RemoveItem(ctx, dto.RemoveItemInput{UserID: uuid.NewString(), ItemID: uuid.NewString()}))
		assert.True(t, d.repo.deleteCalled)
		assert.Equal(t, 1, d.cache.invalidateCalled)
	})

	t.Run("idempotent when cart missing", func(t *testing.T) {
		d := newDeps()
		d.repo.getCartErr = domain.ErrCartNotFound
		require.NoError(t, d.svc.RemoveItem(ctx, dto.RemoveItemInput{UserID: uuid.NewString(), ItemID: uuid.NewString()}))
		assert.False(t, d.repo.deleteCalled)
	})

	t.Run("invalid input", func(t *testing.T) {
		d := newDeps()
		assert.ErrorIs(t, d.svc.RemoveItem(ctx, dto.RemoveItemInput{UserID: "", ItemID: "x"}), domain.ErrInvalidInput)
	})
}

func TestClearCart(t *testing.T) {
	ctx := context.Background()

	t.Run("success clears", func(t *testing.T) {
		d := newDeps()
		require.NoError(t, d.svc.ClearCart(ctx, dto.ClearCartInput{UserID: uuid.NewString()}))
		assert.True(t, d.repo.clearCalled)
		assert.Equal(t, 1, d.cache.invalidateCalled)
	})

	t.Run("invalid user id", func(t *testing.T) {
		d := newDeps()
		assert.ErrorIs(t, d.svc.ClearCart(ctx, dto.ClearCartInput{UserID: "bad"}), domain.ErrInvalidUserID)
	})
}
