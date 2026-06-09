package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/valueobject"
)


type StockAdjustment struct {
	ID          uuid.UUID
	OperationID uuid.UUID
	ProductID   uuid.UUID
	Delta       vo.Delta
	Reason      vo.Reason
	CreatedAt   time.Time
}

func NewStockAdjustment(operationID, productID uuid.UUID, delta vo.Delta, reason vo.Reason) (*StockAdjustment, error) {
	a := &StockAdjustment{
		OperationID: operationID,
		ProductID:   productID,
		Delta:       delta,
		Reason:      reason,
	}
	if err := a.Validate(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *StockAdjustment) Validate() error {
	if a.OperationID == uuid.Nil {
		return domain.ErrInvalidOperationID
	}
	if a.ProductID == uuid.Nil {
		return domain.ErrInvalidProductID
	}
	return nil
}
