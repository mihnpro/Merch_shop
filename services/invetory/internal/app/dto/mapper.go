package dto

import "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"

func ToStockView(s *model.Stock) StockView {
	return StockView{
		ProductID: s.ProductID.String(),
		Available: s.Available,
		Version:   s.Version,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func ToReservationView(r *model.Reservation) ReservationView {
	items := make([]ReservationItemView, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, ReservationItemView{
			ProductID: it.ProductID.String(),
			Qty:       it.Qty.Int(),
		})
	}
	return ReservationView{
		ID:             r.ID.String(),
		OrderID:        r.OrderID.String(),
		Items:          items,
		ReleasedReason: r.ReleasedReason,
		ReleasedAt:     r.ReleasedAt,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}
