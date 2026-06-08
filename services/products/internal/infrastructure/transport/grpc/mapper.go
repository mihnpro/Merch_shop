package grpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	productpb "github.com/mihnpro/Merch_shop/services/products/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/products/internal/app/dto"
)

func toCreateProductInput(req *productpb.CreateProductRequest) dto.CreateProductInput {
	return dto.CreateProductInput{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		PricePoints: req.GetPricePoints(),
		CategoryID:  req.GetCategoryId(),
		PhotoKeys:   req.GetPhotoKeys(),
	}
}

func toUpdateProductInput(req *productpb.UpdateProductRequest) dto.UpdateProductInput {
	return dto.UpdateProductInput{
		ProductID:   req.GetProductId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		PricePoints: req.GetPricePoints(),
		CategoryID:  req.GetCategoryId(),
		PhotoKeys:   req.GetPhotoKeys(),
		Active:      req.GetActive(),
		Version:     int(req.GetVersion()),
	}
}

func toListProductsInput(req *productpb.ListProductsRequest) dto.ListProductsInput {
	return dto.ListProductsInput{
		CategoryID: req.GetCategoryId(),
		Search:     req.GetSearch(),
		ActiveOnly: req.GetActiveOnly(),
		PageSize:   int(req.GetPageSize()),
		PageToken:  req.GetPageToken(),
	}
}

func toCreateCategoryInput(req *productpb.CreateCategoryRequest) dto.CreateCategoryInput {
	return dto.CreateCategoryInput{
		Code: req.GetCode(),
		Name: req.GetName(),
	}
}

func toUpdateCategoryInput(req *productpb.UpdateCategoryRequest) dto.UpdateCategoryInput {
	return dto.UpdateCategoryInput{
		ID:     req.GetId(),
		Name:   req.GetName(),
		Active: req.GetActive(),
	}
}


func toCategoryProto(v dto.CategoryView) *productpb.Category {
	return &productpb.Category{
		Id:        v.ID,
		Code:      v.Code,
		Name:      v.Name,
		Active:    v.Active,
		CreatedAt: toTimestamp(v.CreatedAt),
		UpdatedAt: toTimestamp(v.UpdatedAt),
	}
}

func toProductProto(v dto.ProductView) *productpb.Product {
	return &productpb.Product{
		Id:          v.ID,
		Name:        v.Name,
		Description: v.Description,
		PricePoints: v.PricePoints,
		Category:    toCategoryProto(v.Category),
		PhotoKeys:   v.PhotoKeys,
		Active:      v.Active,
		Version:     int32(v.Version),
		CreatedAt:   toTimestamp(v.CreatedAt),
		UpdatedAt:   toTimestamp(v.UpdatedAt),
	}
}

func toListProductsProto(res dto.ListProductsResult) *productpb.ListProductsResponse {
	products := make([]*productpb.Product, 0, len(res.Products))
	for _, p := range res.Products {
		products = append(products, toProductProto(p))
	}
	return &productpb.ListProductsResponse{
		Products:      products,
		NextPageToken: res.NextPageToken,
	}
}

func toCategoriesProto(views []dto.CategoryView) *productpb.ListCategoriesResponse {
	categories := make([]*productpb.Category, 0, len(views))
	for _, c := range views {
		categories = append(categories, toCategoryProto(c))
	}
	return &productpb.ListCategoriesResponse{Categories: categories}
}

func toTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
