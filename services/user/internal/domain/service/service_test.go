package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)


type fakeUserRepo struct {
	existing  map[string]bool
	existsErr error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{existing: make(map[string]bool)}
}

func (f *fakeUserRepo) CreateUser(ctx context.Context, u *model.User) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (f *fakeUserRepo) GetUserByLogin(ctx context.Context, login vo.Login) (*model.User, error) {
	return nil, domain.ErrUserNotFound
}

func (f *fakeUserRepo) ExistsByLogin(ctx context.Context, login vo.Login) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.existing[login.String()], nil
}



func validRegistrationInput() RegistrationInput {
	return RegistrationInput{
		Login:       "testuser",
		FirstName:   "John",
		LastName:    "Doe",
		Patronymic:  "Smith",
		Email:       "john@example.com",
		PhoneNumber: "+1234567890",
	}
}

var errBoom = errors.New("boom")
