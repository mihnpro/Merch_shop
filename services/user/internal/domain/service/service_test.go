package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
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

func (f *fakeUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return nil, domain.ErrUserNotFound
}

func (f *fakeUserRepo) UpdateUserStatus(ctx context.Context, id uuid.UUID, status vo.Status) (*model.User, error) {
	return nil, domain.ErrUserNotFound
}

func (f *fakeUserRepo) UpdateUserRole(ctx context.Context, id uuid.UUID, role vo.Role) (*model.User, error) {
	return nil, domain.ErrUserNotFound
}

func (f *fakeUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, hash vo.PasswordHash) error {
	return nil
}

func (f *fakeUserRepo) ListUsers(ctx context.Context, flt repository.ListUsersFilter) ([]*model.User, string, error) {
	return nil, "", nil
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
