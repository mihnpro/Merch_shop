package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	productpb "github.com/mihnpro/Merch_shop/services/products/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
)


func (g *GRPCServer) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	res, err := g.read.ListProducts(ctx, toListProductsInput(req))
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toListProductsProto(res), nil
}

func (g *GRPCServer) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.Product, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.read.GetProduct(ctx, req.GetProductId())
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toProductProto(view), nil
}

func (g *GRPCServer) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.Product, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.write.CreateProduct(ctx, toCreateProductInput(req))
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toProductProto(view), nil
}

func (g *GRPCServer) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.Product, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.write.UpdateProduct(ctx, toUpdateProductInput(req))
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toProductProto(view), nil
}

func (g *GRPCServer) DeactivateProduct(ctx context.Context, req *productpb.DeactivateProductRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	if err := g.write.DeactivateProduct(ctx, req.GetProductId()); err != nil {
		return nil, NewGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}


func (g *GRPCServer) ListCategories(ctx context.Context, req *productpb.ListCategoriesRequest) (*productpb.ListCategoriesResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	views, err := g.read.ListCategories(ctx, req.GetActiveOnly())
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toCategoriesProto(views), nil
}

func (g *GRPCServer) CreateCategory(ctx context.Context, req *productpb.CreateCategoryRequest) (*productpb.Category, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.write.CreateCategory(ctx, toCreateCategoryInput(req))
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toCategoryProto(view), nil
}

func (g *GRPCServer) UpdateCategory(ctx context.Context, req *productpb.UpdateCategoryRequest) (*productpb.Category, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.write.UpdateCategory(ctx, toUpdateCategoryInput(req))
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toCategoryProto(view), nil
}
