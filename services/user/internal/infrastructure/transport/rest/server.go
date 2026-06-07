package rest

import (
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/service"
)

type Server struct {
	auth    service.AuthService
	log     *zap.Logger
	cookies CookieConfig
}

func NewServer(auth service.AuthService, log *zap.Logger, cookies CookieConfig) *Server {
	return &Server{auth: auth, log: log, cookies: cookies}
}
