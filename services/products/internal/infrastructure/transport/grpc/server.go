package grpc

import (
	productpb "github.com/mihnpro/Merch_shop/services/products/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/products/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/products/internal/app/service"
)

type GRPCServer struct {
	write service.CatalogService
	read  query.CatalogReadService
	productpb.UnimplementedProductServiceServer
}

func NewGRPCServer(write service.CatalogService, read query.CatalogReadService) *GRPCServer {
	return &GRPCServer{write: write, read: read}
}
