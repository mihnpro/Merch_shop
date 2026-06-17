package query

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"
)

var errBoom = errors.New("boom")

type fakeStockRepo struct {
	listErr    error
	list       []*model.Stock
	getErr     error
	byID       map[uuid.UUID]*model.Stock
	getManyErr error
	many       []*model.Stock
	lastMany   []uuid.UUID
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

func (f *fakeStockRepo) GetByProductIDs(_ context.Context, ids []uuid.UUID) ([]*model.Stock, error) {
	f.lastMany = ids
	if f.getManyErr != nil {
		return nil, f.getManyErr
	}
	return f.many, nil
}

func (f *fakeStockRepo) Adjust(_ context.Context, _ *model.StockAdjustment) (*model.Stock, error) {
	return nil, nil
}

func stockAt(productID uuid.UUID, available int) *model.Stock {
	return &model.Stock{ProductID: productID, Available: available, Version: 1}
}
