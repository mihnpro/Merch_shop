package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

func NewGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "invalid token")
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, domain.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input")
	case errors.Is(err, domain.ErrEmptyRequest):
		return status.Error(codes.InvalidArgument, "empty request")
	case errors.Is(err, domain.ErrEmptyLogin):
		return status.Error(codes.InvalidArgument, "login cannot be empty")
	case errors.Is(err, domain.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	case errors.Is(err, domain.ErrEmptyFirstName):
		return status.Error(codes.InvalidArgument, "first name cannot be empty")
	case errors.Is(err, domain.ErrEmptyLastName):
		return status.Error(codes.InvalidArgument, "last name cannot be empty")
	case errors.Is(err, domain.ErrInvalidEmailFormat):
		return status.Error(codes.InvalidArgument, "invalid email format")
	case errors.Is(err, domain.ErrInvalidPhoneFormat):
		return status.Error(codes.InvalidArgument, "invalid phone number format")
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, "permission denied")
	case errors.Is(err, domain.ErrInsufficientBalance):
		return status.Error(codes.FailedPrecondition, "insufficient points balance")
	case errors.Is(err, domain.ErrOperationAlreadyDone):
		return status.Error(codes.AlreadyExists, "operation already processed")
	case errors.Is(err, domain.ErrUserBlocked):
		return status.Error(codes.PermissionDenied, "user is blocked")
	case errors.Is(err, domain.ErrInvalidRole):
		return status.Error(codes.InvalidArgument, "invalid role")
	case errors.Is(err, domain.ErrInvalidStatus):
		return status.Error(codes.InvalidArgument, "invalid status")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
