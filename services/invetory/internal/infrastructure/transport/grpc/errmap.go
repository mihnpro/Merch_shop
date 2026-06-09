package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
)

func NewGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrStockNotFound):
		return status.Error(codes.NotFound, "stock not found")
	case errors.Is(err, domain.ErrReservationNotFound):
		return status.Error(codes.NotFound, "reservation not found")
	case errors.Is(err, domain.ErrInsufficientStock):
		return status.Error(codes.FailedPrecondition, "insufficient stock")
	case errors.Is(err, domain.ErrVersionConflict):
		return status.Error(codes.FailedPrecondition, "stock was modified by another request")
	case errors.Is(err, domain.ErrInvalidProductID):
		return status.Error(codes.InvalidArgument, "invalid product id")
	case errors.Is(err, domain.ErrInvalidOrderID):
		return status.Error(codes.InvalidArgument, "invalid order id")
	case errors.Is(err, domain.ErrInvalidOperationID):
		return status.Error(codes.InvalidArgument, "invalid operation id")
	case errors.Is(err, domain.ErrInvalidQuantity):
		return status.Error(codes.InvalidArgument, "quantity must be between 1 and 100")
	case errors.Is(err, domain.ErrZeroDelta):
		return status.Error(codes.InvalidArgument, "adjustment delta must be non-zero")
	case errors.Is(err, domain.ErrEmptyReason):
		return status.Error(codes.InvalidArgument, "reason cannot be empty")
	case errors.Is(err, domain.ErrReasonTooLong):
		return status.Error(codes.InvalidArgument, "reason is too long")
	case errors.Is(err, domain.ErrEmptyReservation):
		return status.Error(codes.InvalidArgument, "reservation must contain at least one item")
	case errors.Is(err, domain.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input")
	case errors.Is(err, domain.ErrEmptyRequest):
		return status.Error(codes.InvalidArgument, "empty request")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
