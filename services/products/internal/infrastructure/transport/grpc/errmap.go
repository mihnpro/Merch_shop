package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
)

func NewGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, "product not found")
	case errors.Is(err, domain.ErrCategoryNotFound):
		return status.Error(codes.NotFound, "category not found")
	case errors.Is(err, domain.ErrVersionConflict):
		return status.Error(codes.FailedPrecondition, "product was modified by another request")
	case errors.Is(err, domain.ErrCategoryNotActive):
		return status.Error(codes.FailedPrecondition, "category is not active")
	case errors.Is(err, domain.ErrCategoryAlreadyExists):
		return status.Error(codes.AlreadyExists, "category with this code already exists")
	case errors.Is(err, domain.ErrInvalidPrice):
		return status.Error(codes.InvalidArgument, "price must be a positive number of points")
	case errors.Is(err, domain.ErrInvalidPhotoKey):
		return status.Error(codes.InvalidArgument, "invalid photo key format")
	case errors.Is(err, domain.ErrInvalidCategoryCode):
		return status.Error(codes.InvalidArgument, "invalid category code")
	case errors.Is(err, domain.ErrEmptyName):
		return status.Error(codes.InvalidArgument, "product name cannot be empty")
	case errors.Is(err, domain.ErrNameTooLong):
		return status.Error(codes.InvalidArgument, "product name is too long")
	case errors.Is(err, domain.ErrEmptyDescription):
		return status.Error(codes.InvalidArgument, "product description cannot be empty")
	case errors.Is(err, domain.ErrDescriptionTooLong):
		return status.Error(codes.InvalidArgument, "product description is too long")
	case errors.Is(err, domain.ErrEmptyCategoryName):
		return status.Error(codes.InvalidArgument, "category name cannot be empty")
	case errors.Is(err, domain.ErrCategoryNameTooLong):
		return status.Error(codes.InvalidArgument, "category name is too long")
	case errors.Is(err, domain.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input")
	case errors.Is(err, domain.ErrEmptyRequest):
		return status.Error(codes.InvalidArgument, "empty request")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
