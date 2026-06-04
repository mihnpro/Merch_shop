package grpc

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userpb "github.com/mihnpro/Merch_shop/services/user_customer/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

func TestNewGRPCError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{"already exists", domain.ErrUserAlreadyExists, codes.AlreadyExists},
		{"invalid credentials", domain.ErrInvalidCredentials, codes.Unauthenticated},
		{"invalid token", domain.ErrInvalidToken, codes.Unauthenticated},
		{"not found", domain.ErrUserNotFound, codes.NotFound},
		{"invalid input", domain.ErrInvalidInput, codes.InvalidArgument},
		{"empty request", domain.ErrEmptyRequest, codes.InvalidArgument},
		{"empty login", domain.ErrEmptyLogin, codes.InvalidArgument},
		{"weak password", domain.ErrWeakPassword, codes.InvalidArgument},
		{"empty first name", domain.ErrEmptyFirstName, codes.InvalidArgument},
		{"empty last name", domain.ErrEmptyLastName, codes.InvalidArgument},
		{"invalid email", domain.ErrInvalidEmailFormat, codes.InvalidArgument},
		{"invalid phone", domain.ErrInvalidPhoneFormat, codes.InvalidArgument},
		{"unknown error falls back to internal", errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGRPCError(tt.err)
			assert.Equal(t, tt.wantCode, status.Code(got))
		})
	}
}

func TestNewGRPCError_WrappedError(t *testing.T) {
	wrapped := errors.Join(errors.New("context"), domain.ErrUserNotFound)
	assert.Equal(t, codes.NotFound, status.Code(NewGRPCError(wrapped)))
}

func TestToRegisterInput(t *testing.T) {
	req := &userpb.RegisterRequest{
		Login:       "testuser",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Patronymic:  "Smith",
		Email:       "john@example.com",
		PhoneNumber: "+1234567890",
	}
	got := toRegisterInput(req)

	assert.Equal(t, dto.RegisterInput{
		Login:       "testuser",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Patronymic:  "Smith",
		Email:       "john@example.com",
		PhoneNumber: "+1234567890",
	}, got)
}

func TestToRegisterInput_NilSafe(t *testing.T) {
	assert.NotPanics(t, func() { toRegisterInput(nil) })
	assert.Equal(t, dto.RegisterInput{}, toRegisterInput(nil))
}

func TestToLoginInput(t *testing.T) {
	got := toLoginInput(&userpb.LoginRequest{Login: "testuser", Password: "secret"})
	assert.Equal(t, dto.LoginInput{Login: "testuser", Password: "secret"}, got)
}

func TestToUserProto(t *testing.T) {
	created := time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC)
	v := dto.UserView{
		ID:          "id-1",
		Login:       "testuser",
		FirstName:   "John",
		LastName:    "Doe",
		Patronymic:  "Smith",
		Email:       "john@example.com",
		PhoneNumber: "+1234567890",
		CreatedAt:   created,
	}
	got := toUserProto(v)

	assert.Equal(t, "id-1", got.GetId())
	assert.Equal(t, "testuser", got.GetLogin())
	assert.Equal(t, "john@example.com", got.GetEmail())
	assert.Equal(t, "+1234567890", got.GetPhoneNumber())
	assert.Equal(t, created, got.GetCreatedAt().AsTime())
	assert.Nil(t, got.GetUpdatedAt()) 
}

func TestToTokenPairProto(t *testing.T) {
	accessExp := time.Date(2026, 6, 3, 11, 0, 0, 0, time.UTC)
	p := port.TokenPair{
		AccessToken:     "a",
		RefreshToken:    "r",
		AccessExpiresAt: accessExp,
	}
	got := toTokenPairProto(p)

	assert.Equal(t, "a", got.GetAccessToken())
	assert.Equal(t, "r", got.GetRefreshToken())
	assert.Equal(t, accessExp, got.GetAccessExpiresAt().AsTime())
	assert.Nil(t, got.GetRefreshExpiresAt())
}

func TestToTimestamp(t *testing.T) {
	t.Run("zero time maps to nil", func(t *testing.T) {
		assert.Nil(t, toTimestamp(time.Time{}))
	})
	t.Run("non-zero time is preserved", func(t *testing.T) {
		now := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
		ts := toTimestamp(now)
		assert.NotNil(t, ts)
		assert.Equal(t, now, ts.AsTime())
	})
}
