package dto

import "github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"

func ToCartItemView(item *model.CartItem) CartItemView {
	lineTotal := item.Quantity.Int() * item.PriceAtAdd
	return CartItemView{
		ID:           item.ID.String(),
		ProductID:    item.ProductID.String(),
		ProductName:  item.ProductNameAtAdd,
		Quantity:     item.Quantity.Int(),
		PricePerUnit: item.PriceAtAdd,
		LineTotal:    lineTotal,
	}
}


func ToCartView(items []*model.CartItem) CartView {
	views := make([]CartItemView, 0, len(items))
	total := 0
	for _, item := range items {
		v := ToCartItemView(item)
		views = append(views, v)
		total += v.LineTotal
	}
	return CartView{
		Items:     views,
		Total:     total,
		ItemCount: len(views),
	}
}
