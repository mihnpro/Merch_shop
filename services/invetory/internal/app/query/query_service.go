package query

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/repository"
)

type InventoryReadService interface {
	CheckStock(ctx context.Context, in dto.CheckStockInput) (dto.CheckStockView, error)
	ListStock(ctx context.Context) ([]dto.StockView, error)
}

type inventoryReadService struct {
	stock repository.StockRepository
}

func NewInventoryReadService(stock repository.StockRepository) InventoryReadService {
	return &inventoryReadService{stock: stock}
}

func (s *inventoryReadService) CheckStock(ctx context.Context, in dto.CheckStockInput) (dto.CheckStockView, error) {
	id, err := uuid.Parse(in.ProductID)
	if err != nil {
		return dto.CheckStockView{}, domain.ErrInvalidProductID
	}

	need := in.Qty
	if need < 1 {
		need = 1
	}

	st, err := s.stock.GetByProductID(ctx, id)
	if errors.Is(err, domain.ErrStockNotFound) {
		return dto.CheckStockView{ProductID: in.ProductID, Available: 0, InStock: false}, nil
	}
	if err != nil {
		return dto.CheckStockView{}, err
	}

	return dto.CheckStockView{
		ProductID: st.ProductID.String(),
		Available: st.Available,
		InStock:   st.Available >= need,
	}, nil
}

func (s *inventoryReadService) ListStock(ctx context.Context) ([]dto.StockView, error) {
	stocks, err := s.stock.List(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]dto.StockView, 0, len(stocks))
	for _, st := range stocks {
		views = append(views, dto.ToStockView(st))
	}
	return views, nil
}
