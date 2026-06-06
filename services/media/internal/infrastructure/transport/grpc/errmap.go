package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
)

func NewGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEmptyRequest):
		return status.Error(codes.InvalidArgument, "empty request")
	case errors.Is(err, domain.ErrInvalidContentType):
		return status.Error(codes.InvalidArgument, "unsupported content type: allowed image/jpeg, image/png, image/webp")
	case errors.Is(err, domain.ErrFileTooLarge):
		return status.Error(codes.InvalidArgument, "file too large: max 5 MB")
	case errors.Is(err, domain.ErrEmptyFile):
		return status.Error(codes.InvalidArgument, "empty file")
	case errors.Is(err, domain.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "unauthenticated")
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "invalid token")
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, "admin role required")
	case errors.Is(err, domain.ErrStorageUnavailable):
		return status.Error(codes.Unavailable, "storage unavailable")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
