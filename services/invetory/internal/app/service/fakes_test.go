package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/valueobject"
)

var errBoom = errors.New("boom")

type fakeStockRepo struct {
	listErr      error
	list         []*model.Stock
	getErr       error
	byID         map[uuid.UUID]*model.Stock
	getManyErr   error
	many         []*model.Stock
	adjustErr    error
	adjustResult *model.Stock
	adjusted     *model.StockAdjustment
}

func newFakeStockRepo() *fakeStockRepo {
	return &fakeStockRepo{byID: make(map[uuid.UUID]*model.Stock)}
}

func (f *fakeStockRepo) List(_ context.Context) ([]*model.Stock, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.list, nil
}

func (f *fakeStockRepo) GetByProductID(_ context.Context, id uuid.UUID) (*model.Stock, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if s, ok := f.byID[id]; ok {
		return s, nil
	}
	return nil, domain.ErrStockNotFound
}

func (f *fakeStockRepo) GetByProductIDs(_ context.Context, _ []uuid.UUID) ([]*model.Stock, error) {
	if f.getManyErr != nil {
		return nil, f.getManyErr
	}
	return f.many, nil
}

func (f *fakeStockRepo) Adjust(_ context.Context, adj *model.StockAdjustment) (*model.Stock, error) {
	f.adjusted = adj
	if f.adjustErr != nil {
		return nil, f.adjustErr
	}
	if f.adjustResult != nil {
		return f.adjustResult, nil
	}
	return &model.Stock{ProductID: adj.ProductID, Available: adj.Delta.Int(), Version: 1}, nil
}

type fakeReservationRepo struct {
	getErr        error
	getResult     *model.Reservation
	reserveErr    error
	reserveResult *model.Reservation
	releaseErr    error
	releaseResult *model.Reservation
}

func (f *fakeReservationRepo) GetByOrderID(_ context.Context, _ uuid.UUID) (*model.Reservation, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getResult, nil
}

func (f *fakeReservationRepo) Reserve(_ context.Context, r *model.Reservation) (*model.Reservation, error) {
	if f.reserveErr != nil {
		return nil, f.reserveErr
	}
	if f.reserveResult != nil {
		return f.reserveResult, nil
	}
	return r, nil
}

func (f *fakeReservationRepo) Release(_ context.Context, _ uuid.UUID, _ string) (*model.Reservation, error) {
	if f.releaseErr != nil {
		return nil, f.releaseErr
	}
	return f.releaseResult, nil
}

type fakeOrderNotifier struct {
	calls      int
	lastStatus string
	lastReason string
	err        error
}

func (f *fakeOrderNotifier) UpdateOrderStatus(_ context.Context, _, newStatus, reason string) error {
	f.calls++
	f.lastStatus = newStatus
	f.lastReason = reason
	return f.err
}

func sampleStock(productID uuid.UUID, available int) *model.Stock {
	return &model.Stock{ProductID: productID, Available: available, Version: 1}
}

func sampleReservation(orderID uuid.UUID) *model.Reservation {
	r, _ := model.NewReservation(orderID, []model.ReservationItem{
		{ProductID: uuid.New(), Qty: vo.NewQuantityFromStored(2)},
	})
	r.ID = uuid.New()
	return r
}
