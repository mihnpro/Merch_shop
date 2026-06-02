package grpc

import (
	userpb "github.com/mihnpro/Merch_shop/services/user_customer/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/service"
)

type GRPCServer struct {
	service service.AuthQueryService
	userpb.UnimplementedUserServiceServer
}

func NewGRPCServer(service service.AuthQueryService) *GRPCServer {
	return &GRPCServer{service: service}
}
