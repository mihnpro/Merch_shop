package service

import (
	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/valueobject"
)

type AdjustStockInput struct {
	OperationID uuid.UUID
	ProductID   uuid.UUID
	Delta       int
	Reason      string
}

type StockFactory struct{}

func NewStockFactory() *StockFactory { return &StockFactory{} }

func (f *StockFactory) NewAdjustment(in AdjustStockInput) (*model.StockAdjustment, error) {
	delta, err := vo.NewDelta(in.Delta)
	if err != nil {
		return nil, err
	}
	reason, err := vo.NewReason(in.Reason)
	if err != nil {
		return nil, err
	}
	return model.NewStockAdjustment(in.OperationID, in.ProductID, delta, reason)
}
