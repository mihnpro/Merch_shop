package rest

import (
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/service"
)

type Server struct {
	auth    service.AuthService
	user    service.UserService
	admin   service.AdminService
	log     *zap.Logger
	cookies CookieConfig
}

func NewServer(auth service.AuthService, user service.UserService, admin service.AdminService, log *zap.Logger, cookies CookieConfig) *Server {
	return &Server{auth: auth, user: user, admin: admin, log: log, cookies: cookies}
}
