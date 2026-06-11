package repository

import (
	"context"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"
)

type CacheRepository interface {
	TryFromCache(ctx context.Context, userID string) ([]*model.CartItem, bool)
	SetCache(ctx context.Context, userID string, items []*model.CartItem)
	InvalidateCache(ctx context.Context, userID string)
}
