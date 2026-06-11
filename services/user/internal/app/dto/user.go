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
	Role        string
	Status      string
	LastLoginAt *time.Time
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

type BalanceView struct {
	Points    int64
	UpdatedAt time.Time
}

type TransactionView struct {
	ID          string
	UserID      string
	OperationID string
	OrderID     string
	Amount      int64
	Reason      string
	CreatedAt   time.Time
}

type DeductPointsInput struct {
	UserID      string
	Amount      int64
	OperationID string
	OrderID     string
	Reason      string
}

type AddPointsInput struct {
	UserID      string
	Amount      int64
	OperationID string
	OrderID     string
	Reason      string
}

type GrantPointsInput struct {
	UserID      string
	Amount      int64
	OperationID string
	Reason      string
}

type ChangePasswordInput struct {
	UserID      string
	OldPassword string
	NewPassword string
}

type ListUsersInput struct {
	Search    string
	RoleCode  string
	Status    string
	PageSize  int
	PageToken string
}

type GetTransactionsInput struct {
	UserID    string
	PageSize  int
	PageToken string
}
