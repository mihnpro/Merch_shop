package grpc

import (
	userpb "github.com/mihnpro/Merch_shop/services/user_customer/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/service"
)

type GRPCServer struct {
	auth  service.AuthService
	user  service.UserService
	admin service.AdminService
	userpb.UnimplementedUserServiceServer
}

func NewGRPCServer(auth service.AuthService, user service.UserService, admin service.AdminService) *GRPCServer {
	return &GRPCServer{auth: auth, user: user, admin: admin}
}
