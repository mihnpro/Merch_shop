package rest

import (
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/products/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/products/internal/app/service"
)

type Server struct {
	write service.CatalogService
	read  query.CatalogReadService
	log   *zap.Logger
}

func NewServer(write service.CatalogService, read query.CatalogReadService, log *zap.Logger) *Server {
	return &Server{write: write, read: read, log: log}
}
