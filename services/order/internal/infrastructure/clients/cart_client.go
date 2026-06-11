package clients

import (
	"context"
	"fmt"

	cartpb "github.com/mihnpro/Merch_shop/services/order/api/clients/cart/ClientPublic"
	"github.com/mihnpro/Merch_shop/services/order/internal/app/port"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type CartGRPCClient struct {
	client cartpb.CartServiceClient
}

func NewCartClient(addr string) (*CartGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial cart service: %w", err)
	}
	return &CartGRPCClient{client: cartpb.NewCartServiceClient(conn)}, nil
}

func (c *CartGRPCClient) GetCart(ctx context.Context, userID string) ([]port.CartItem, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-user-id", userID)
	resp, err := c.client.GetCart(ctx, &cartpb.GetCartRequest{})
	if err != nil {
		return nil, err
	}
	items := make([]port.CartItem, 0, len(resp.Items))
	for _, it := range resp.Items {
		items = append(items, port.CartItem{
			ItemID:      it.ItemId,
			ProductID:   it.ProductId,
			ProductName: it.ProductName,
			Quantity:    int(it.Quantity),
			PricePoints: int64(it.PricePerUnit),
		})
	}
	return items, nil
}

func (c *CartGRPCClient) ClearCart(ctx context.Context, userID string) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-user-id", userID)
	_, err := c.client.ClearCart(ctx, &cartpb.ClearCartRequest{})
	return err
}
