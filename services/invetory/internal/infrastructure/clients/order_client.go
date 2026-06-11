package clients

import (
	"context"
	"fmt"

	orderpb "github.com/mihnpro/Merch_shop/services/invetory/api/client/ClientPublic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient struct {
	client orderpb.OrderServiceClient
}

func NewOrderClient(addr string) (*OrderClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial order service: %w", err)
	}
	return &OrderClient{client: orderpb.NewOrderServiceClient(conn)}, nil
}

func (c *OrderClient) UpdateOrderStatus(ctx context.Context, orderID, newStatus, reason string) error {
	status := stringToProtoStatus(newStatus)
	_, err := c.client.UpdateOrderStatus(ctx, &orderpb.UpdateOrderStatusRequest{
		OrderId:   orderID,
		NewStatus: status,
		Reason:    reason,
	})
	return err
}

func stringToProtoStatus(s string) orderpb.OrderStatus {
	switch s {
	case "confirmed":
		return orderpb.OrderStatus_ORDER_STATUS_CONFIRMED
	case "cancelled":
		return orderpb.OrderStatus_ORDER_STATUS_CANCELLED
	case "ready_to_pickup":
		return orderpb.OrderStatus_ORDER_STATUS_READY_TO_PICKUP
	case "delivered":
		return orderpb.OrderStatus_ORDER_STATUS_DELIVERED
	default:
		return orderpb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
