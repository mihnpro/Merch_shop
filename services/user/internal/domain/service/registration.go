package service

import (
	"context"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type RegistrationInput struct {
	Login       string
	FirstName   string
	LastName    string
	Patronymic  string
	Email       string
	PhoneNumber string
}

type Registrar struct {
	users repository.UserRepository
}

func NewRegistrar(users repository.UserRepository) *Registrar {
	return &Registrar{users: users}
}

func (r *Registrar) Register(ctx context.Context, in RegistrationInput, hash vo.PasswordHash) (*model.User, error) {
	login, err := vo.NewLogin(in.Login)
	if err != nil {
		return nil, err
	}
	email, err := vo.NewEmail(in.Email)
	if err != nil {
		return nil, err
	}
	phone, err := vo.NewPhoneNumber(in.PhoneNumber)
	if err != nil {
		return nil, err
	}

	exists, err := r.users.ExistsByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrUserAlreadyExists
	}

	return model.NewUser(login, hash, in.FirstName, in.LastName, in.Patronymic, email, phone)
}
