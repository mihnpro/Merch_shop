package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
)

func NewGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrCartNotFound):
		return status.Error(codes.NotFound, "cart not found")
	case errors.Is(err, domain.ErrItemNotFound):
		return status.Error(codes.NotFound, "cart item not found")
	case errors.Is(err, domain.ErrOutOfStock):
		return status.Error(codes.FailedPrecondition, "out of stock")
	case errors.Is(err, domain.ErrProductInactive):
		return status.Error(codes.FailedPrecondition, "product is inactive")
	case errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, "product not found")
	case errors.Is(err, domain.ErrInvalidQuantity):
		return status.Error(codes.InvalidArgument, "invalid quantity")
	case errors.Is(err, domain.ErrInvalidSize):
		return status.Error(codes.InvalidArgument, "invalid size")
	case errors.Is(err, domain.ErrInvalidUserID),
		errors.Is(err, domain.ErrInvalidItemID),
		errors.Is(err, domain.ErrInvalidProductID),
		errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrEmptyRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "invalid token")
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, "permission denied")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
