package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/mihnpro/Merch_shop/services/media/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/media/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
)

func AdminOnlyStreamInterceptor(v port.Verifier) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		id, err := identityFromContext(ss.Context(), v)
		if err != nil {
			return NewGRPCError(err)
		}
		if !id.IsAdmin() {
			return NewGRPCError(domain.ErrForbidden)
		}
		return handler(srv, ss)
	}
}

func identityFromContext(ctx context.Context, v port.Verifier) (dto.Identity, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return dto.Identity{}, domain.ErrUnauthenticated
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return dto.Identity{}, domain.ErrUnauthenticated
	}
	token := strings.TrimSpace(vals[0])
	if len(token) >= 7 && strings.EqualFold(token[:7], "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	return v.ParseAccess(token)
}
