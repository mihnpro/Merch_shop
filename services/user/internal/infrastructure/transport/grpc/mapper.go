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
	u := &userpb.User{
		Id:          v.ID,
		Login:       v.Login,
		FirstName:   v.FirstName,
		LastName:    v.LastName,
		Patronymic:  v.Patronymic,
		Email:       v.Email,
		PhoneNumber: v.PhoneNumber,
		Role:        v.Role,
		Status:      toUserStatusEnum(v.Status),
		CreatedAt:   toTimestamp(v.CreatedAt),
		UpdatedAt:   toTimestamp(v.UpdatedAt),
	}
	if v.LastLoginAt != nil {
		u.LastLoginAt = toTimestamp(*v.LastLoginAt)
	}
	return u
}

func toTokenPairProto(t port.TokenPair) *userpb.TokenPair {
	return &userpb.TokenPair{
		AccessToken:      t.AccessToken,
		RefreshToken:     t.RefreshToken,
		AccessExpiresAt:  toTimestamp(t.AccessExpiresAt),
		RefreshExpiresAt: toTimestamp(t.RefreshExpiresAt),
	}
}

func toBalanceProto(b dto.BalanceView) *userpb.Balance {
	return &userpb.Balance{
		Points:    b.Points,
		UpdatedAt: toTimestamp(b.UpdatedAt),
	}
}

func toTransactionProto(t dto.TransactionView) *userpb.PointsTransaction {
	return &userpb.PointsTransaction{
		Id:          t.ID,
		UserId:      t.UserID,
		OperationId: t.OperationID,
		OrderId:     t.OrderID,
		Amount:      t.Amount,
		Reason:      t.Reason,
		CreatedAt:   toTimestamp(t.CreatedAt),
	}
}

func toUserStatusEnum(status string) userpb.UserStatus {
	switch status {
	case "blocked":
		return userpb.UserStatus_USER_STATUS_BLOCKED
	case "active":
		return userpb.UserStatus_USER_STATUS_ACTIVE
	default:
		return userpb.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

func toUserStatusString(s userpb.UserStatus) string {
	switch s {
	case userpb.UserStatus_USER_STATUS_ACTIVE:
		return "active"
	case userpb.UserStatus_USER_STATUS_BLOCKED:
		return "blocked"
	default:
		return ""
	}
}

func toTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
