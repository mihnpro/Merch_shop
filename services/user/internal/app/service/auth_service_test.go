package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

func TestAuthService_Register(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		view, err := h.authSvc.Register(context.Background(), validRegisterInput())

		assert.NoError(t, err)
		assert.Equal(t, "testuser", view.Login)
		assert.Equal(t, "john@example.com", view.Email)
		assert.Equal(t, h.repo.createID.String(), view.ID)
		assert.Len(t, h.repo.created, 1)
	})

	t.Run("duplicate login", func(t *testing.T) {
		h := newHarness()
		h.repo.existing["testuser"] = true

		_, err := h.authSvc.Register(context.Background(), validRegisterInput())
		assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
		assert.Empty(t, h.repo.created)
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			mutate  func(in *dto.RegisterInput)
			wantErr error
		}{
			{"empty login", func(in *dto.RegisterInput) { in.Login = "   " }, domain.ErrEmptyLogin},
			{"invalid email", func(in *dto.RegisterInput) { in.Email = "invalid-email" }, domain.ErrInvalidEmailFormat},
			{"invalid phone", func(in *dto.RegisterInput) { in.PhoneNumber = "abc" }, domain.ErrInvalidPhoneFormat},
			{"weak password", func(in *dto.RegisterInput) { in.Password = "123" }, domain.ErrWeakPassword},
			{"empty first name", func(in *dto.RegisterInput) { in.FirstName = "" }, domain.ErrEmptyFirstName},
			{"empty last name", func(in *dto.RegisterInput) { in.LastName = "" }, domain.ErrEmptyLastName},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h := newHarness()
				in := validRegisterInput()
				tt.mutate(&in)

				_, err := h.authSvc.Register(context.Background(), in)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, h.repo.created)
			})
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "testuser")
		h.account.verify = true
		h.account.tokens = port.TokenPair{AccessToken: "a", RefreshToken: "r"}

		res, err := h.authSvc.Login(context.Background(), dto.LoginInput{Login: "testuser", Password: "password123"})
		assert.NoError(t, err)
		assert.Equal(t, "r", res.Tokens.RefreshToken)
		assert.Equal(t, u.ID.String(), h.tokens.stored["r"])
	})

	t.Run("empty credentials", func(t *testing.T) {
		h := newHarness()
		_, err := h.authSvc.Login(context.Background(), dto.LoginInput{Login: "  ", Password: ""})
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("unknown user maps to invalid credentials", func(t *testing.T) {
		h := newHarness()
		_, err := h.authSvc.Login(context.Background(), dto.LoginInput{Login: "ghost", Password: "password123"})
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("wrong password", func(t *testing.T) {
		h := newHarness()
		seedUser(h, "testuser")
		h.account.verify = false

		_, err := h.authSvc.Login(context.Background(), dto.LoginInput{Login: "testuser", Password: "wrong"})
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("blocked user", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "testuser")
		u.Status = vo.StatusBlocked
		h.account.verify = true

		_, err := h.authSvc.Login(context.Background(), dto.LoginInput{Login: "testuser", Password: "password123"})
		assert.ErrorIs(t, err, domain.ErrUserBlocked)
	})

	t.Run("token store error", func(t *testing.T) {
		h := newHarness()
		seedUser(h, "testuser")
		h.account.verify = true
		h.account.tokens = port.TokenPair{AccessToken: "a", RefreshToken: "r"}
		h.tokens.storeErr = errBoom

		_, err := h.authSvc.Login(context.Background(), dto.LoginInput{Login: "testuser", Password: "password123"})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestAuthService_Logout(t *testing.T) {
	t.Run("empty token", func(t *testing.T) {
		h := newHarness()
		err := h.authSvc.Logout(context.Background(), "  ")
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("deletes token", func(t *testing.T) {
		h := newHarness()
		h.tokens.stored["r"] = "user-1"

		err := h.authSvc.Logout(context.Background(), "r")
		assert.NoError(t, err)
		assert.NotContains(t, h.tokens.stored, "r")
	})
}

func TestAuthService_Refresh(t *testing.T) {
	t.Run("empty token", func(t *testing.T) {
		h := newHarness()
		_, err := h.authSvc.Refresh(context.Background(), "  ")
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("invalid token", func(t *testing.T) {
		h := newHarness()
		h.account.validateErr = domain.ErrInvalidToken
		_, err := h.authSvc.Refresh(context.Background(), "bad-token")
		assert.ErrorIs(t, err, domain.ErrInvalidToken)
	})

	t.Run("token not in store", func(t *testing.T) {
		h := newHarness()
		h.account.validateIdentity = model.Identity{UserID: uuid.New(), Role: "user"}
		h.tokens.getUserErr = domain.ErrInvalidToken

		_, err := h.authSvc.Refresh(context.Background(), "expired-token")
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		uid := uuid.New()
		h.account.validateIdentity = model.Identity{UserID: uid, Role: "user"}
		h.account.tokens = port.TokenPair{AccessToken: "new-a", RefreshToken: "new-r"}
		h.tokens.stored["old-r"] = uid.String()

		pair, err := h.authSvc.Refresh(context.Background(), "old-r")
		assert.NoError(t, err)
		assert.Equal(t, "new-r", pair.RefreshToken)
		assert.NotContains(t, h.tokens.stored, "old-r")
		assert.Equal(t, uid.String(), h.tokens.stored["new-r"])
	})

	t.Run("delete error", func(t *testing.T) {
		h := newHarness()
		h.account.validateIdentity = model.Identity{UserID: uuid.New(), Role: "user"}
		h.tokens.stored["old-r"] = uuid.New().String()
		h.tokens.deleteErr = errBoom

		_, err := h.authSvc.Refresh(context.Background(), "old-r")
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("generate error", func(t *testing.T) {
		h := newHarness()
		h.account.validateIdentity = model.Identity{UserID: uuid.New(), Role: "user"}
		h.tokens.stored["old-r"] = uuid.New().String()
		h.account.genErr = errBoom

		_, err := h.authSvc.Refresh(context.Background(), "old-r")
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("store error after generate", func(t *testing.T) {
		h := newHarness()
		h.account.validateIdentity = model.Identity{UserID: uuid.New(), Role: "user"}
		h.account.tokens = port.TokenPair{AccessToken: "a", RefreshToken: "new-r"}
		h.tokens.stored["old-r"] = uuid.New().String()
		h.tokens.storeErr = errBoom

		_, err := h.authSvc.Refresh(context.Background(), "old-r")
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestAuthService_Me(t *testing.T) {
	t.Run("empty token", func(t *testing.T) {
		h := newHarness()
		_, err := h.authSvc.Me(context.Background(), "  ")
		assert.ErrorIs(t, err, domain.ErrInvalidToken)
	})

	t.Run("invalid token", func(t *testing.T) {
		h := newHarness()
		h.account.validateErr = domain.ErrInvalidToken
		_, err := h.authSvc.Me(context.Background(), "bad")
		assert.ErrorIs(t, err, domain.ErrInvalidToken)
	})

	t.Run("user not found", func(t *testing.T) {
		h := newHarness()
		h.account.validateIdentity = model.Identity{UserID: uuid.New(), Role: "user"}
		_, err := h.authSvc.Me(context.Background(), "token")
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("blocked user", func(t *testing.T) {
		h := newHarness()
		uid := uuid.New()
		h.account.validateIdentity = model.Identity{UserID: uid, Role: "user"}
		u := &model.User{
			ID:     uid,
			Status: vo.StatusBlocked,
			Role:   vo.RoleUser,
			Login:  vo.NewLoginFromStored("blocked"),
			Email:  vo.NewEmailFromStored("b@x.com"),
		}
		h.repo.byID[uid] = u

		_, err := h.authSvc.Me(context.Background(), "token")
		assert.ErrorIs(t, err, domain.ErrUserBlocked)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		uid := uuid.New()
		h.account.validateIdentity = model.Identity{UserID: uid, Role: "admin"}
		u := &model.User{
			ID:     uid,
			Status: vo.StatusActive,
			Role:   vo.RoleAdmin,
			Login:  vo.NewLoginFromStored("admin"),
			Email:  vo.NewEmailFromStored("a@x.com"),
		}
		h.repo.byID[uid] = u

		me, err := h.authSvc.Me(context.Background(), "token")
		assert.NoError(t, err)
		assert.Equal(t, uid.String(), me.UserID)
		assert.Equal(t, "admin", me.Role)
	})
}
