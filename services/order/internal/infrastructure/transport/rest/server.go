package rest

import (
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/order/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/order/internal/app/service"
)

type Server struct {
	orders service.OrderService
	query  query.OrderQueryService
	log    *zap.Logger
}

func NewServer(orders service.OrderService, query query.OrderQueryService, log *zap.Logger) *Server {
	return &Server{orders: orders, query: query, log: log}
}
