package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
	domainsvc "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/service"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type UserService interface {
	GetUser(ctx context.Context, userID string) (dto.UserView, error)
	ChangePassword(ctx context.Context, in dto.ChangePasswordInput) error
	GetBalance(ctx context.Context, userID string) (dto.BalanceView, error)
	GetTransactions(ctx context.Context, in dto.GetTransactionsInput) ([]dto.TransactionView, string, error)
	DeductPoints(ctx context.Context, in dto.DeductPointsInput) (dto.BalanceView, error)
	AddPoints(ctx context.Context, in dto.AddPointsInput) (dto.BalanceView, error)
}

type userService struct {
	users   repository.UserRepository
	balance repository.BalanceRepository
	points  *domainsvc.PointsManager
	account port.Account
}

func NewUserService(users repository.UserRepository, balance repository.BalanceRepository, points *domainsvc.PointsManager, account port.Account) UserService {
	return &userService{users: users, balance: balance, points: points, account: account}
}

func (s *userService) GetUser(ctx context.Context, userID string) (dto.UserView, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return dto.UserView{}, domain.ErrInvalidInput
	}
	u, err := s.users.GetUserByID(ctx, id)
	if err != nil {
		return dto.UserView{}, err
	}
	return toUserView(u), nil
}

func (s *userService) ChangePassword(ctx context.Context, in dto.ChangePasswordInput) error {
	if strings.TrimSpace(in.NewPassword) == "" {
		return domain.ErrWeakPassword
	}
	id, err := parseUUID(in.UserID)
	if err != nil {
		return domain.ErrInvalidInput
	}
	u, err := s.users.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if !s.account.VerifyPassword(vo.PlainPasswordFromRaw(in.OldPassword), u.PasswordHash) {
		return domain.ErrInvalidCredentials
	}
	plain, err := vo.NewPlainPassword(in.NewPassword)
	if err != nil {
		return err
	}
	hash, err := s.account.HashPassword(plain)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, id, hash)
}

func (s *userService) GetBalance(ctx context.Context, userID string) (dto.BalanceView, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return dto.BalanceView{}, domain.ErrInvalidInput
	}
	bal, err := s.balance.GetBalance(ctx, id)
	if err != nil {
		return dto.BalanceView{}, err
	}
	return dto.BalanceView{Points: bal.Points, UpdatedAt: bal.UpdatedAt}, nil
}

func (s *userService) GetTransactions(ctx context.Context, in dto.GetTransactionsInput) ([]dto.TransactionView, string, error) {
	id, err := parseUUID(in.UserID)
	if err != nil {
		return nil, "", domain.ErrInvalidInput
	}
	txs, nextToken, err := s.balance.GetTransactions(ctx, id, in.PageSize, in.PageToken)
	if err != nil {
		return nil, "", err
	}
	return toTransactionViews(txs), nextToken, nil
}

func (s *userService) DeductPoints(ctx context.Context, in dto.DeductPointsInput) (dto.BalanceView, error) {
	cmd, err := buildPointsCommand(in.UserID, in.OperationID, in.OrderID, in.Reason)
	if err != nil {
		return dto.BalanceView{}, err
	}
	bal, err := s.points.Debit(ctx, cmd, in.Amount)
	if err != nil {
		return dto.BalanceView{}, err
	}
	return dto.BalanceView{Points: bal.Points, UpdatedAt: bal.UpdatedAt}, nil
}

func (s *userService) AddPoints(ctx context.Context, in dto.AddPointsInput) (dto.BalanceView, error) {
	cmd, err := buildPointsCommand(in.UserID, in.OperationID, in.OrderID, in.Reason)
	if err != nil {
		return dto.BalanceView{}, err
	}
	bal, err := s.points.Credit(ctx, cmd, in.Amount)
	if err != nil {
		return dto.BalanceView{}, err
	}
	return dto.BalanceView{Points: bal.Points, UpdatedAt: bal.UpdatedAt}, nil
}

func buildPointsCommand(userID, operationID, orderID, reason string) (repository.ApplyPointsCommand, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return repository.ApplyPointsCommand{}, domain.ErrInvalidInput
	}
	opID, err := parseUUID(operationID)
	if err != nil {
		return repository.ApplyPointsCommand{}, domain.ErrInvalidInput
	}
	cmd := repository.ApplyPointsCommand{UserID: uid, OperationID: opID, Reason: reason}
	if orderID != "" {
		oid, err := parseUUID(orderID)
		if err != nil {
			return repository.ApplyPointsCommand{}, domain.ErrInvalidInput
		}
		cmd.OrderID = &oid
	}
	return cmd, nil
}

func toTransactionViews(txs []model.PointsTransaction) []dto.TransactionView {
	views := make([]dto.TransactionView, 0, len(txs))
	for _, t := range txs {
		v := dto.TransactionView{
			ID:          t.ID.String(),
			UserID:      t.UserID.String(),
			OperationID: t.OperationID.String(),
			Amount:      t.Amount,
			Reason:      t.Reason,
			CreatedAt:   t.CreatedAt,
		}
		if t.OrderID != nil {
			v.OrderID = t.OrderID.String()
		}
		views = append(views, v)
	}
	return views
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
