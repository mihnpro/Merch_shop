package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	userpb "github.com/mihnpro/Merch_shop/services/user_customer/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

func (g *GRPCServer) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.AuthResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.service.Register(ctx, toRegisterInput(req))
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return &userpb.AuthResponse{User: toUserProto(view)}, nil
}

func (g *GRPCServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.AuthResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	result, err := g.service.Login(ctx, toLoginInput(req))
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return &userpb.AuthResponse{
		User:   toUserProto(result.User),
		Tokens: toTokenPairProto(result.Tokens),
	}, nil
}

func (g *GRPCServer) Refresh(ctx context.Context, req *userpb.RefreshRequest) (*userpb.TokenPair, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	tokens, err := g.service.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toTokenPairProto(tokens), nil
}

func (g *GRPCServer) Logout(ctx context.Context, req *userpb.LogoutRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	if err := g.service.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, NewGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}
