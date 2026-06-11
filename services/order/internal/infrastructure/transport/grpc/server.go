package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderpb "github.com/mihnpro/Merch_shop/services/order/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/order/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/order/internal/app/service"
)

type GRPCServer struct {
	orders service.OrderService
	orderpb.UnimplementedOrderServiceServer
}

func NewGRPCServer(orders service.OrderService) *GRPCServer {
	return &GRPCServer{orders: orders}
}

func (s *GRPCServer) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.OrderResponse, error) {
	newStatus := protoStatusToString(req.NewStatus)
	if newStatus == "" {
		return nil, status.Error(codes.InvalidArgument, "unknown order status")
	}
	view, err := s.orders.UpdateOrderStatus(ctx, dto.UpdateOrderStatusInput{
		OrderID:   req.OrderId,
		NewStatus: newStatus,
		Reason:    req.Reason,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderpb.OrderResponse{Order: toOrderProto(view)}, nil
}

func protoStatusToString(s orderpb.OrderStatus) string {
	switch s {
	case orderpb.OrderStatus_ORDER_STATUS_PENDING:
		return "pending"
	case orderpb.OrderStatus_ORDER_STATUS_CONFIRMED:
		return "confirmed"
	case orderpb.OrderStatus_ORDER_STATUS_READY_TO_PICKUP:
		return "ready_to_pickup"
	case orderpb.OrderStatus_ORDER_STATUS_DELIVERED:
		return "delivered"
	case orderpb.OrderStatus_ORDER_STATUS_CANCELLED:
		return "cancelled"
	default:
		return ""
	}
}

func toOrderProto(v dto.OrderView) *orderpb.Order {
	items := make([]*orderpb.OrderItem, 0, len(v.Items))
	for _, it := range v.Items {
		items = append(items, &orderpb.OrderItem{
			Id:          it.ID,
			ProductId:   it.ProductID,
			ProductName: it.ProductName,
			Quantity:    int32(it.Quantity),
			PricePoints: it.PricePoints,
		})
	}
	return &orderpb.Order{
		Id:              v.ID,
		UserId:          v.UserID,
		Status:          stringToProtoStatus(v.Status),
		TotalPoints:     v.TotalPoints,
		DeliveryAddress: v.DeliveryAddress,
		Items:           items,
	}
}

func stringToProtoStatus(s string) orderpb.OrderStatus {
	switch strings.ToLower(s) {
	case "pending":
		return orderpb.OrderStatus_ORDER_STATUS_PENDING
	case "confirmed":
		return orderpb.OrderStatus_ORDER_STATUS_CONFIRMED
	case "ready_to_pickup":
		return orderpb.OrderStatus_ORDER_STATUS_READY_TO_PICKUP
	case "delivered":
		return orderpb.OrderStatus_ORDER_STATUS_DELIVERED
	case "cancelled":
		return orderpb.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderpb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
