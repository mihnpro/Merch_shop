package dto

import (
	"time"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
)

type RegisterInput struct {
	Login       string
	Password    string
	FirstName   string
	LastName    string
	Patronymic  string
	Email       string
	PhoneNumber string
}

type LoginInput struct {
	Login    string
	Password string
}

type UserView struct {
	ID          string
	Login       string
	FirstName   string
	LastName    string
	Patronymic  string
	Email       string
	PhoneNumber string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AuthResult struct {
	User   UserView
	Tokens port.TokenPair
}

type Me struct {
	UserID string
	Role   string
}
