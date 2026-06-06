package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

func TestRegistrar_Register(t *testing.T) {
	hash := vo.NewPasswordHash("hashed:password123")

	t.Run("happy path builds unsaved user with given hash", func(t *testing.T) {
		repo := newFakeUserRepo()
		r := NewRegistrar(repo)

		user, err := r.Register(context.Background(), validRegistrationInput(), hash)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Login.String())
		assert.Equal(t, "john@example.com", user.Email.String())
		assert.Equal(t, "hashed:password123", user.PasswordHash.String())
		assert.Equal(t, uuid.Nil, user.ID) 
	})

	t.Run("duplicate login", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.existing["testuser"] = true
		r := NewRegistrar(repo)

		user, err := r.Register(context.Background(), validRegistrationInput(), hash)

		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	})

	t.Run("validation errors abort before existence check", func(t *testing.T) {
		tests := []struct {
			name    string
			mutate  func(in *RegistrationInput)
			wantErr error
		}{
			{"empty login", func(in *RegistrationInput) { in.Login = "   " }, domain.ErrEmptyLogin},
			{"invalid email", func(in *RegistrationInput) { in.Email = "invalid-email" }, domain.ErrInvalidEmailFormat},
			{"invalid phone", func(in *RegistrationInput) { in.PhoneNumber = "abc" }, domain.ErrInvalidPhoneFormat},
			{"empty first name", func(in *RegistrationInput) { in.FirstName = "" }, domain.ErrEmptyFirstName},
			{"empty last name", func(in *RegistrationInput) { in.LastName = "" }, domain.ErrEmptyLastName},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := NewRegistrar(newFakeUserRepo())
				in := validRegistrationInput()
				tt.mutate(&in)

				user, err := r.Register(context.Background(), in, hash)
				assert.Nil(t, user)
				assert.ErrorIs(t, err, tt.wantErr)
			})
		}
	})

	t.Run("propagates existence-check error", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.existsErr = errBoom
		r := NewRegistrar(repo)

		_, err := r.Register(context.Background(), validRegistrationInput(), hash)
		assert.ErrorIs(t, err, errBoom)
	})
}
