package clients

import (
	"context"
	"fmt"

	inventorypb "github.com/mihnpro/Merch_shop/services/order/api/clients/inventory/ClientPublic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InventoryGRPCClient struct {
	client inventorypb.InventoryServiceClient
}

func NewInventoryClient(addr string) (*InventoryGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial inventory service: %w", err)
	}
	return &InventoryGRPCClient{client: inventorypb.NewInventoryServiceClient(conn)}, nil
}

func (c *InventoryGRPCClient) ReleaseReserve(ctx context.Context, orderID, reason string) error {
	_, err := c.client.ReleaseReserve(ctx, &inventorypb.ReleaseReserveRequest{
		OrderId: orderID,
		Reason:  reason,
	})
	return err
}
