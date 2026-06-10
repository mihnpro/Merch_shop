package grpc

import (
	cartpb "github.com/mihnpro/Merch_shop/services/cart/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/cart/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/cart/internal/app/service"
)

type GRPCServer struct {
	write service.CartService
	read  query.CartQueryService
	cartpb.UnimplementedCartServiceServer
}

func NewGRPCServer(write service.CartService, read query.CartQueryService) *GRPCServer {
	return &GRPCServer{write: write, read: read}
}
