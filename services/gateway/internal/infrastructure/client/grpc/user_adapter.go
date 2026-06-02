package grpcclient

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/domain"
	userpb "github.com/mihnpro/Merch_shop/services/user_customer/api/server/AccountInternal"
)

type userAdapter struct {
	client userpb.UserServiceClient
}

func NewUserAdapter(client userpb.UserServiceClient) port.UserAdapter {
	return &userAdapter{client: client}
}

func (a *userAdapter) Register(ctx context.Context, in dto.RegisterInput) (dto.UserView, error) {
	resp, err := a.client.Register(ctx, &userpb.RegisterRequest{
		Login:       in.Login,
		Password:    in.Password,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Patronymic:  in.Patronymic,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
	})
	if err != nil {
		return dto.UserView{}, mapErr(err)
	}
	return toUserView(resp.GetUser()), nil
}

func (a *userAdapter) Login(ctx context.Context, in dto.LoginInput) (dto.AuthResult, error) {
	resp, err := a.client.Login(ctx, &userpb.LoginRequest{
		Login:    in.Login,
		Password: in.Password,
	})
	if err != nil {
		return dto.AuthResult{}, mapErr(err)
	}
	return dto.AuthResult{
		User:   toUserView(resp.GetUser()),
		Tokens: toTokenPair(resp.GetTokens()),
	}, nil
}

func (a *userAdapter) Refresh(ctx context.Context, refreshToken string) (dto.TokenPair, error) {
	resp, err := a.client.Refresh(ctx, &userpb.RefreshRequest{RefreshToken: refreshToken})
	if err != nil {
		return dto.TokenPair{}, mapErr(err)
	}
	return toTokenPair(resp), nil
}

func (a *userAdapter) Logout(ctx context.Context, refreshToken string) error {
	if _, err := a.client.Logout(ctx, &userpb.LogoutRequest{RefreshToken: refreshToken}); err != nil {
		return mapErr(err)
	}
	return nil
}

func toUserView(u *userpb.User) dto.UserView {
	if u == nil {
		return dto.UserView{}
	}
	return dto.UserView{
		ID:          u.GetId(),
		Login:       u.GetLogin(),
		FirstName:   u.GetFirstName(),
		LastName:    u.GetLastName(),
		Patronymic:  u.GetPatronymic(),
		Email:       u.GetEmail(),
		PhoneNumber: u.GetPhoneNumber(),
	}
}

func toTokenPair(t *userpb.TokenPair) dto.TokenPair {
	if t == nil {
		return dto.TokenPair{}
	}
	return dto.TokenPair{
		AccessToken:  t.GetAccessToken(),
		RefreshToken: t.GetRefreshToken(),
	}
}

func mapErr(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return domain.ErrInternal
	}
	switch st.Code() {
	case codes.OK:
		return nil
	case codes.InvalidArgument:
		return domain.ErrBadRequest
	case codes.Unauthenticated:
		return domain.ErrUnauthenticated
	case codes.PermissionDenied:
		return domain.ErrForbidden
	case codes.NotFound:
		return domain.ErrNotFound
	case codes.AlreadyExists:
		return domain.ErrAlreadyExists
	case codes.FailedPrecondition:
		return domain.ErrFailedPrecondition
	case codes.ResourceExhausted:
		return domain.ErrRateLimited
	case codes.Unavailable:
		return domain.ErrUnavailable
	case codes.DeadlineExceeded:
		return domain.ErrTimeout
	default:
		return domain.ErrInternal
	}
}
