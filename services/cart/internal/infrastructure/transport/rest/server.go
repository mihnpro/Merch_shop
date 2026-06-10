package rest

import (
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/cart/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/cart/internal/app/service"
)

type Server struct {
	write service.CartService
	read  query.CartQueryService
	log   *zap.Logger
}

func NewServer(write service.CartService, read query.CartQueryService, log *zap.Logger) *Server {
	return &Server{write: write, read: read, log: log}
}
