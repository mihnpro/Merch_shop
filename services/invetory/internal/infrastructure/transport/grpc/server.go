package grpc

import (
	inventorypb "github.com/mihnpro/Merch_shop/services/invetory/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/service"
)

type GRPCServer struct {
	write service.InventoryService
	read  query.InventoryReadService
	inventorypb.UnimplementedInventoryServiceServer
}

func NewGRPCServer(write service.InventoryService, read query.InventoryReadService) *GRPCServer {
	return &GRPCServer{write: write, read: read}
}
