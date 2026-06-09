package grpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	inventorypb "github.com/mihnpro/Merch_shop/services/invetory/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/dto"
)

func toReserveStockInput(req *inventorypb.ReserveStockRequest) dto.ReserveStockInput {
	items := make([]dto.ReserveItemInput, 0, len(req.GetItems()))
	for _, it := range req.GetItems() {
		items = append(items, dto.ReserveItemInput{
			ProductID: it.GetProductId(),
			Qty:       int(it.GetQty()),
		})
	}
	return dto.ReserveStockInput{
		OrderID: req.GetOrderId(),
		Items:   items,
	}
}

func toCheckStockProto(v dto.CheckStockView) *inventorypb.CheckStockResponse {
	return &inventorypb.CheckStockResponse{
		ProductId: v.ProductID,
		Available: int32(v.Available),
		InStock:   v.InStock,
	}
}

func toReservationProto(v dto.ReservationView) *inventorypb.Reservation {
	items := make([]*inventorypb.ReservationItem, 0, len(v.Items))
	for _, it := range v.Items {
		items = append(items, &inventorypb.ReservationItem{
			ProductId: it.ProductID,
			Qty:       int32(it.Qty),
		})
	}
	return &inventorypb.Reservation{
		Id:             v.ID,
		OrderId:        v.OrderID,
		Items:          items,
		ReleasedReason: v.ReleasedReason,
		ReleasedAt:     toTimestampPtr(v.ReleasedAt),
		CreatedAt:      toTimestamp(v.CreatedAt),
		UpdatedAt:      toTimestamp(v.UpdatedAt),
	}
}

func toAdjustStockProto(v dto.StockView) *inventorypb.AdjustStockResponse {
	return &inventorypb.AdjustStockResponse{
		ProductId: v.ProductID,
		Available: int32(v.Available),
		Version:   int32(v.Version),
	}
}

func toTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func toTimestampPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return toTimestamp(*t)
}
