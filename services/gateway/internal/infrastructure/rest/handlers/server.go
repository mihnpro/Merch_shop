package handlers

import (
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/service"
)

type Server struct {
	auth service.AuthQueryService
	log  *zap.Logger
}

func NewServer(auth service.AuthQueryService, log *zap.Logger) *Server {
	return &Server{auth: auth, log: log}
}
