package service

import (
	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/cart/internal/domain/valueobject"
)

type NewCartItemInput struct {
	CartID      uuid.UUID
	ProductID   uuid.UUID
	Quantity    int
	PriceAtAdd  int
	ProductName string
}

type CartFactory struct{}

func NewCartFactory() *CartFactory {
	return &CartFactory{}
}

func (f *CartFactory) NewCartItem(in NewCartItemInput) (*model.CartItem, error) {
	qty, err := vo.NewQuantity(in.Quantity)
	if err != nil {
		return nil, err
	}
	return &model.CartItem{
		ID:               uuid.New(),
		CartID:           in.CartID,
		ProductID:        in.ProductID,
		Quantity:         qty,
		PriceAtAdd:       in.PriceAtAdd,
		ProductNameAtAdd: in.ProductName,
	}, nil
}
