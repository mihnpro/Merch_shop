package grpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	userpb "github.com/mihnpro/Merch_shop/services/user_customer/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
)

func toRegisterInput(req *userpb.RegisterRequest) dto.RegisterInput {
	return dto.RegisterInput{
		Login:       req.GetLogin(),
		Password:    req.GetPassword(),
		FirstName:   req.GetFirstName(),
		LastName:    req.GetLastName(),
		Patronymic:  req.GetPatronymic(),
		Email:       req.GetEmail(),
		PhoneNumber: req.GetPhoneNumber(),
	}
}

func toLoginInput(req *userpb.LoginRequest) dto.LoginInput {
	return dto.LoginInput{
		Login:    req.GetLogin(),
		Password: req.GetPassword(),
	}
}

func toUserProto(v dto.UserView) *userpb.User {
	return &userpb.User{
		Id:          v.ID,
		Login:       v.Login,
		FirstName:   v.FirstName,
		LastName:    v.LastName,
		Patronymic:  v.Patronymic,
		Email:       v.Email,
		PhoneNumber: v.PhoneNumber,
		CreatedAt:   toTimestamp(v.CreatedAt),
		UpdatedAt:   toTimestamp(v.UpdatedAt),
	}
}

func toTokenPairProto(t port.TokenPair) *userpb.TokenPair {
	return &userpb.TokenPair{
		AccessToken:      t.AccessToken,
		RefreshToken:     t.RefreshToken,
		AccessExpiresAt:  toTimestamp(t.AccessExpiresAt),
		RefreshExpiresAt: toTimestamp(t.RefreshExpiresAt),
	}
}

func toTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
