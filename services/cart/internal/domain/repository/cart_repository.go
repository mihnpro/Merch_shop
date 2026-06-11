package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"
)

type CartRepository interface {
	GetOrCreateCart(ctx context.Context, userID uuid.UUID) (*model.Cart, error)
	GetCartByUserID(ctx context.Context, userID uuid.UUID) (*model.Cart, error)
	GetCartItems(ctx context.Context, cartID uuid.UUID) ([]*model.CartItem, error)
	GetItemByProduct(ctx context.Context, cartID, productID uuid.UUID) (*model.CartItem, error)
	GetItemByID(ctx context.Context, itemID uuid.UUID) (*model.CartItem, error)
	InsertItem(ctx context.Context, item *model.CartItem) (*model.CartItem, error)
	UpdateItemQty(ctx context.Context, itemID uuid.UUID, qty int) error
	DeleteItem(ctx context.Context, itemID, cartID uuid.UUID) error
	ClearCart(ctx context.Context, cartID uuid.UUID) error
	TouchCart(ctx context.Context, cartID uuid.UUID) error
}
