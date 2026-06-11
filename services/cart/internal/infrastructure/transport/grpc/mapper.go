package grpc

import (
	cartpb "github.com/mihnpro/Merch_shop/services/cart/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/cart/internal/app/dto"
)

func toCartItemProto(v dto.CartItemView) *cartpb.CartItem {
	return &cartpb.CartItem{
		ItemId:       v.ID,
		ProductId:    v.ProductID,
		ProductName:  v.ProductName,
		Quantity:     int32(v.Quantity),
		PricePerUnit: int32(v.PricePerUnit),
		LineTotal:    int32(v.LineTotal),
	}
}

func toGetCartResponseProto(v dto.CartView) *cartpb.GetCartResponse {
	items := make([]*cartpb.CartItem, 0, len(v.Items))
	for _, item := range v.Items {
		items = append(items, toCartItemProto(item))
	}
	return &cartpb.GetCartResponse{
		Items:     items,
		Total:     int32(v.Total),
		ItemCount: int32(v.ItemCount),
	}
}
