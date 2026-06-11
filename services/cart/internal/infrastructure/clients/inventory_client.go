package clients

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	inventorypb "github.com/mihnpro/Merch_shop/services/cart/api/client/inventory"
)

type InventoryClient struct {
	client inventorypb.InventoryServiceClient
}

func NewInventoryClient(addr string) (*InventoryClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &InventoryClient{client: inventorypb.NewInventoryServiceClient(conn)}, nil
}

func (c *InventoryClient) CheckStock(ctx context.Context, productID string, qty int) (bool, error) {
	resp, err := c.client.CheckStock(ctx, &inventorypb.CheckStockRequest{
		ProductId: productID,
		Qty:       int32(qty),
	})
	if err != nil {
		return false, err
	}
	return resp.GetInStock(), nil
}
