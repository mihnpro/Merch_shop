package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"
)

type StockRepository interface {
	List(ctx context.Context) ([]*model.Stock, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) (*model.Stock, error)
	Adjust(ctx context.Context, adj *model.StockAdjustment) (*model.Stock, error)
}
