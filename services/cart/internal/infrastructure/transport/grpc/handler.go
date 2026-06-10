package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"

	cartpb "github.com/mihnpro/Merch_shop/services/cart/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/cart/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
)

const metaUserID = "x-user-id"

func (g *GRPCServer) GetCart(ctx context.Context, req *cartpb.GetCartRequest) (*cartpb.GetCartResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	userID, err := userIDFromMeta(ctx)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	view, err := g.read.GetCart(ctx, dto.GetCartInput{UserID: userID})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toGetCartResponseProto(view), nil
}

func (g *GRPCServer) AddItem(ctx context.Context, req *cartpb.AddItemRequest) (*cartpb.AddItemResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	userID, err := userIDFromMeta(ctx)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	item, err := g.write.AddItem(ctx, dto.AddItemInput{
		UserID:    userID,
		ProductID: req.GetProductId(),
		Quantity:  int(req.GetQuantity()),
	})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return &cartpb.AddItemResponse{Item: toCartItemProto(item)}, nil
}

func (g *GRPCServer) UpdateItem(ctx context.Context, req *cartpb.UpdateItemRequest) (*cartpb.UpdateItemResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	userID, err := userIDFromMeta(ctx)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	item, err := g.write.UpdateItem(ctx, dto.UpdateItemInput{
		UserID:      userID,
		ItemID:      req.GetItemId(),
		NewQuantity: int(req.GetNewQuantity()),
	})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return &cartpb.UpdateItemResponse{Item: toCartItemProto(item)}, nil
}

func (g *GRPCServer) RemoveItem(ctx context.Context, req *cartpb.RemoveItemRequest) (*cartpb.RemoveItemResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	userID, err := userIDFromMeta(ctx)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	if err := g.write.RemoveItem(ctx, dto.RemoveItemInput{
		UserID: userID,
		ItemID: req.GetItemId(),
	}); err != nil {
		return nil, NewGRPCError(err)
	}
	return &cartpb.RemoveItemResponse{}, nil
}

func (g *GRPCServer) ClearCart(ctx context.Context, req *cartpb.ClearCartRequest) (*cartpb.ClearCartResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	userID, err := userIDFromMeta(ctx)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	if err := g.write.ClearCart(ctx, dto.ClearCartInput{UserID: userID}); err != nil {
		return nil, NewGRPCError(err)
	}
	return &cartpb.ClearCartResponse{}, nil
}

func userIDFromMeta(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", domain.ErrInvalidToken
	}
	vals := md.Get(metaUserID)
	if len(vals) == 0 || vals[0] == "" {
		return "", domain.ErrInvalidToken
	}
	return vals[0], nil
}
