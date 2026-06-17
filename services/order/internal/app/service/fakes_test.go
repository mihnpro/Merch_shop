package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/order/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/repository"
)

var errBoom = errors.New("boom")

type fakeOrderRepo struct {
	byID         map[uuid.UUID]*model.Order
	createErr    error
	createdOrder *model.Order
	createdPayld []byte
	getErr       error
	updateErr    error
	updatedID    uuid.UUID
	updatedTo    model.OrderStatus
	updatedRsn   string
	updateCalls  int
	listResult   []*model.Order
	listToken    string
	listErr      error
	lastFilter   repository.ListOrdersFilter
	analytics    repository.Analytics
	analyticsErr error
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{byID: make(map[uuid.UUID]*model.Order)}
}

func (f *fakeOrderRepo) CreateOrder(_ context.Context, order *model.Order, payload []byte) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createdOrder = order
	f.createdPayld = payload
	f.byID[order.ID] = order
	return nil
}

func (f *fakeOrderRepo) GetOrderByID(_ context.Context, id uuid.UUID) (*model.Order, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if o, ok := f.byID[id]; ok {
		return o, nil
	}
	return nil, domain.ErrOrderNotFound
}

func (f *fakeOrderRepo) UpdateOrderStatus(_ context.Context, id uuid.UUID, status model.OrderStatus, reason string) error {
	f.updateCalls++
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updatedID, f.updatedTo, f.updatedRsn = id, status, reason
	return nil
}

func (f *fakeOrderRepo) ListOrders(_ context.Context, filter repository.ListOrdersFilter) ([]*model.Order, string, error) {
	f.lastFilter = filter
	if f.listErr != nil {
		return nil, "", f.listErr
	}
	return f.listResult, f.listToken, nil
}

func (f *fakeOrderRepo) GetAnalytics(_ context.Context, _ time.Time) (repository.Analytics, error) {
	if f.analyticsErr != nil {
		return repository.Analytics{}, f.analyticsErr
	}
	return f.analytics, nil
}

type fakeCart struct {
	items       []port.CartItem
	getErr      error
	clearErr    error
	clearCalled bool
}

func (f *fakeCart) GetCart(_ context.Context, _ string) ([]port.CartItem, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.items, nil
}

func (f *fakeCart) ClearCart(_ context.Context, _ string) error {
	f.clearCalled = true
	return f.clearErr
}

type fakeUser struct {
	balance      int64
	balanceErr   error
	deductErr    error
	deductCalls  int
	deductAmount int64
	addCalls     int
	addAmount    int64
	addReasons   []string
}

func (f *fakeUser) GetBalance(_ context.Context, _ string) (int64, error) {
	if f.balanceErr != nil {
		return 0, f.balanceErr
	}
	return f.balance, nil
}

func (f *fakeUser) DeductPoints(_ context.Context, _ string, amount int64, _, _, _ string) error {
	f.deductCalls++
	f.deductAmount = amount
	return f.deductErr
}

func (f *fakeUser) AddPoints(_ context.Context, _ string, amount int64, _, _, reason string) error {
	f.addCalls++
	f.addAmount = amount
	f.addReasons = append(f.addReasons, reason)
	return nil
}

type fakeInventory struct {
	releaseCalled bool
	releaseReason string
}

func (f *fakeInventory) ReleaseReserve(_ context.Context, _, reason string) error {
	f.releaseCalled = true
	f.releaseReason = reason
	return nil
}

func cartItem(productID uuid.UUID, qty int, price int64) port.CartItem {
	return port.CartItem{
		ItemID:      uuid.NewString(),
		ProductID:   productID.String(),
		ProductName: "Item",
		Quantity:    qty,
		PricePoints: price,
	}
}
