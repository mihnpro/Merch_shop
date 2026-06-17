package query

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/cart/internal/domain/valueobject"
)

type fakeCartRepo struct {
	cart          *model.Cart
	getCartErr    error
	getCartCalled bool
	items         []*model.CartItem
	getItemsErr   error
}

func newFakeCartRepo() *fakeCartRepo {
	return &fakeCartRepo{cart: &model.Cart{ID: uuid.New(), UserID: uuid.New()}}
}

func (f *fakeCartRepo) GetOrCreateCart(_ context.Context, _ uuid.UUID) (*model.Cart, error) {
	return f.cart, nil
}

func (f *fakeCartRepo) GetCartByUserID(_ context.Context, _ uuid.UUID) (*model.Cart, error) {
	f.getCartCalled = true
	if f.getCartErr != nil {
		return nil, f.getCartErr
	}
	return f.cart, nil
}

func (f *fakeCartRepo) GetCartItems(_ context.Context, _ uuid.UUID) ([]*model.CartItem, error) {
	if f.getItemsErr != nil {
		return nil, f.getItemsErr
	}
	return f.items, nil
}

func (f *fakeCartRepo) GetItemByProduct(_ context.Context, _, _ uuid.UUID) (*model.CartItem, error) {
	return nil, domain.ErrItemNotFound
}

func (f *fakeCartRepo) GetItemByID(_ context.Context, _ uuid.UUID) (*model.CartItem, error) {
	return nil, domain.ErrItemNotFound
}

func (f *fakeCartRepo) InsertItem(_ context.Context, item *model.CartItem) (*model.CartItem, error) {
	return item, nil
}

func (f *fakeCartRepo) UpdateItemQty(_ context.Context, _ uuid.UUID, _ int) error { return nil }
func (f *fakeCartRepo) DeleteItem(_ context.Context, _, _ uuid.UUID) error        { return nil }
func (f *fakeCartRepo) ClearCart(_ context.Context, _ uuid.UUID) error            { return nil }
func (f *fakeCartRepo) TouchCart(_ context.Context, _ uuid.UUID) error            { return nil }

type fakeCache struct {
	items     []*model.CartItem
	hit       bool
	setCalled bool
}

func (f *fakeCache) TryFromCache(_ context.Context, _ string) ([]*model.CartItem, bool) {
	return f.items, f.hit
}

func (f *fakeCache) SetCache(_ context.Context, _ string, items []*model.CartItem) {
	f.setCalled = true
	f.items = items
}

func (f *fakeCache) InvalidateCache(_ context.Context, _ string) {}

func item(qty, price int) *model.CartItem {
	return &model.CartItem{
		ID: uuid.New(), CartID: uuid.New(), ProductID: uuid.New(),
		Quantity: vo.NewQuantityFromStored(qty), PriceAtAdd: price, ProductNameAtAdd: "Item",
	}
}
