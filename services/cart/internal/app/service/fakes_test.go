package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/cart/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"
)

var errBoom = errors.New("boom")

type fakeCartRepo struct {
	cart       *model.Cart
	getCartErr error

	itemByProduct *model.CartItem
	itemByID      *model.CartItem
	itemByIDErr   error

	items        []*model.CartItem
	insertResult *model.CartItem
	insertErr    error

	updateQtyErr error
	updatedID    uuid.UUID
	updatedQty   int

	deleteErr    error
	deleteCalled bool
	clearErr     error
	clearCalled  bool
}

func newFakeCartRepo() *fakeCartRepo {
	return &fakeCartRepo{cart: &model.Cart{ID: uuid.New(), UserID: uuid.New()}}
}

func (f *fakeCartRepo) GetOrCreateCart(_ context.Context, _ uuid.UUID) (*model.Cart, error) {
	return f.cart, nil
}

func (f *fakeCartRepo) GetCartByUserID(_ context.Context, _ uuid.UUID) (*model.Cart, error) {
	if f.getCartErr != nil {
		return nil, f.getCartErr
	}
	return f.cart, nil
}

func (f *fakeCartRepo) GetCartItems(_ context.Context, _ uuid.UUID) ([]*model.CartItem, error) {
	return f.items, nil
}

func (f *fakeCartRepo) GetItemByProduct(_ context.Context, _, _ uuid.UUID) (*model.CartItem, error) {
	if f.itemByProduct != nil {
		return f.itemByProduct, nil
	}
	return nil, domain.ErrItemNotFound
}

func (f *fakeCartRepo) GetItemByID(_ context.Context, _ uuid.UUID) (*model.CartItem, error) {
	if f.itemByIDErr != nil {
		return nil, f.itemByIDErr
	}
	return f.itemByID, nil
}

func (f *fakeCartRepo) InsertItem(_ context.Context, item *model.CartItem) (*model.CartItem, error) {
	if f.insertErr != nil {
		return nil, f.insertErr
	}
	if f.insertResult != nil {
		return f.insertResult, nil
	}
	return item, nil
}

func (f *fakeCartRepo) UpdateItemQty(_ context.Context, itemID uuid.UUID, qty int) error {
	f.updatedID, f.updatedQty = itemID, qty
	return f.updateQtyErr
}

func (f *fakeCartRepo) DeleteItem(_ context.Context, _, _ uuid.UUID) error {
	f.deleteCalled = true
	return f.deleteErr
}

func (f *fakeCartRepo) ClearCart(_ context.Context, _ uuid.UUID) error {
	f.clearCalled = true
	return f.clearErr
}

func (f *fakeCartRepo) TouchCart(_ context.Context, _ uuid.UUID) error { return nil }

type fakeCache struct {
	items            []*model.CartItem
	hit              bool
	setCalled        bool
	invalidateCalled int
}

func (f *fakeCache) TryFromCache(_ context.Context, _ string) ([]*model.CartItem, bool) {
	return f.items, f.hit
}

func (f *fakeCache) SetCache(_ context.Context, _ string, items []*model.CartItem) {
	f.setCalled = true
	f.items = items
}

func (f *fakeCache) InvalidateCache(_ context.Context, _ string) {
	f.invalidateCalled++
}

type fakeProducts struct {
	info dto.ProductInfo
	err  error
}

func (f *fakeProducts) GetProduct(_ context.Context, _ string) (dto.ProductInfo, error) {
	if f.err != nil {
		return dto.ProductInfo{}, f.err
	}
	return f.info, nil
}

type fakeInventory struct {
	available int
	err       error
}

func (f *fakeInventory) CheckStock(_ context.Context, _ string, qty int) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return qty <= f.available, nil
}

func activeProduct() dto.ProductInfo {
	return dto.ProductInfo{Name: "Hoodie", PricePoints: 100, Active: true}
}
