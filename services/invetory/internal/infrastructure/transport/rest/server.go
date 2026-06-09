package rest

import (
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/service"
)

type Server struct {
	write service.InventoryService
	read  query.InventoryReadService
	log   *zap.Logger
}

func NewServer(write service.InventoryService, read query.InventoryReadService, log *zap.Logger) *Server {
	return &Server{write: write, read: read, log: log}
}
