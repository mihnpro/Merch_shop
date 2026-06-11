package clients

import (
	"context"
	"fmt"

	userpb "github.com/mihnpro/Merch_shop/services/order/api/clients/user/ClientPublic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserGRPCClient struct {
	client userpb.UserServiceClient
}

func NewUserClient(addr string) (*UserGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial user service: %w", err)
	}
	return &UserGRPCClient{client: userpb.NewUserServiceClient(conn)}, nil
}

func (c *UserGRPCClient) GetBalance(ctx context.Context, userID string) (int64, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-user-id", userID)
	resp, err := c.client.GetBalance(ctx, &emptypb.Empty{})
	if err != nil {
		return 0, err
	}
	return resp.Points, nil
}

func (c *UserGRPCClient) DeductPoints(ctx context.Context, userID string, amount int64, operationID, orderID, reason string) error {
	_, err := c.client.DeductPoints(ctx, &userpb.DeductPointsRequest{
		UserId:      userID,
		Amount:      amount,
		OperationId: operationID,
		OrderId:     orderID,
		Reason:      reason,
	})
	return err
}

func (c *UserGRPCClient) AddPoints(ctx context.Context, userID string, amount int64, operationID, orderID, reason string) error {
	_, err := c.client.AddPoints(ctx, &userpb.AddPointsRequest{
		UserId:      userID,
		Amount:      amount,
		OperationId: operationID,
		OrderId:     orderID,
		Reason:      reason,
	})
	return err
}
