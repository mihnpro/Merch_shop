package dto

import "github.com/mihnpro/Merch_shop/services/order/internal/domain/model"

func ToOrderView(o *model.Order) OrderView {
	items := make([]OrderItemView, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, OrderItemView{
			ID:          it.ID.String(),
			ProductID:   it.ProductID.String(),
			ProductName: it.ProductName,
			Quantity:    it.Quantity,
			PricePoints: it.PricePoints,
		})
	}
	return OrderView{
		ID:              o.ID.String(),
		UserID:          o.UserID.String(),
		Status:          string(o.Status),
		TotalPoints:     o.TotalPoints,
		DeliveryAddress: o.DeliveryAddress,
		Items:           items,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
}
