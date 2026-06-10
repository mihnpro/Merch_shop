package clients

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	productpb "github.com/mihnpro/Merch_shop/services/cart/api/client/product"

	"github.com/mihnpro/Merch_shop/services/cart/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
)

type ProductClient struct {
	client productpb.ProductServiceClient
}

func NewProductClient(addr string) (*ProductClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &ProductClient{client: productpb.NewProductServiceClient(conn)}, nil
}

func (c *ProductClient) GetProduct(ctx context.Context, productID string) (dto.ProductInfo, error) {
	resp, err := c.client.GetProduct(ctx, &productpb.GetProductRequest{ProductId: productID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return dto.ProductInfo{}, domain.ErrProductNotFound
			}
		}
		return dto.ProductInfo{}, err
	}
	return dto.ProductInfo{
		Name:        resp.GetName(),
		PricePoints: int(resp.GetPricePoints()),
		Active:      resp.GetActive(),
	}, nil
}
