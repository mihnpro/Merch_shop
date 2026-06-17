package query

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/repository"
)

var errBoom = errors.New("boom")

type fakeOrderRepo struct {
	byID         map[uuid.UUID]*model.Order
	getErr       error
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

func (f *fakeOrderRepo) CreateOrder(_ context.Context, _ *model.Order, _ []byte) error { return nil }

func (f *fakeOrderRepo) GetOrderByID(_ context.Context, id uuid.UUID) (*model.Order, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if o, ok := f.byID[id]; ok {
		return o, nil
	}
	return nil, domain.ErrOrderNotFound
}

func (f *fakeOrderRepo) UpdateOrderStatus(_ context.Context, _ uuid.UUID, _ model.OrderStatus, _ string) error {
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

func sampleOrder(userID uuid.UUID) *model.Order {
	o, _ := model.NewOrder(userID, []model.OrderItem{
		{ProductID: uuid.New(), ProductName: "Item", Quantity: 1, PricePoints: 100},
	}, "addr", nil)
	return o
}
