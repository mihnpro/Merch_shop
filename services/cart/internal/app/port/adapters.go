package port

import (
	"context"

	"github.com/mihnpro/Merch_shop/services/cart/internal/app/dto"
)

type ProductAdapter interface {
	GetProduct(ctx context.Context, productID string) (dto.ProductInfo, error)
}

type InventoryAdapter interface {
	CheckStock(ctx context.Context, productID string, qty int) (bool, error)
}
