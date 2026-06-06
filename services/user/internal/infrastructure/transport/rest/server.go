package rest

import (
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/service"
)

type Server struct {
	auth service.AuthService
	log  *zap.Logger
}

func NewServer(auth service.AuthService, log *zap.Logger) *Server {
	return &Server{auth: auth, log: log}
}
