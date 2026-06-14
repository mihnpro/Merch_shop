package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
	domainsvc "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/service"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type AdminService interface {
	GrantPoints(ctx context.Context, in dto.GrantPointsInput) (dto.BalanceView, error)
	ResetPassword(ctx context.Context, userID string) (string, error)
	BlockUser(ctx context.Context, userID string, blocked bool) (dto.UserView, error)
	ChangeRole(ctx context.Context, userID, role string) (dto.UserView, error)
	ListUsers(ctx context.Context, in dto.ListUsersInput) ([]dto.UserView, string, error)
	GetTransactions(ctx context.Context, in dto.GetTransactionsInput) ([]dto.TransactionView, string, error)
	NewUsersStats(ctx context.Context, period string) (dto.UsersStatsView, error)
}

type adminService struct {
	users   repository.UserRepository
	balance repository.BalanceRepository
	points  *domainsvc.PointsManager
	account port.Account
}

func NewAdminService(users repository.UserRepository, balance repository.BalanceRepository, points *domainsvc.PointsManager, account port.Account) AdminService {
	return &adminService{users: users, balance: balance, points: points, account: account}
}

func (s *adminService) GrantPoints(ctx context.Context, in dto.GrantPointsInput) (dto.BalanceView, error) {
	userID, err := parseUUID(in.UserID)
	if err != nil {
		return dto.BalanceView{}, domain.ErrInvalidInput
	}
	if _, err := s.users.GetUserByID(ctx, userID); err != nil {
		return dto.BalanceView{}, err
	}
	opID, err := parseUUID(in.OperationID)
	if err != nil {
		return dto.BalanceView{}, domain.ErrInvalidInput
	}
	bal, err := s.points.Credit(ctx, repository.ApplyPointsCommand{
		UserID:      userID,
		OperationID: opID,
		Reason:      in.Reason,
	}, in.Amount)
	if err != nil {
		return dto.BalanceView{}, err
	}
	return dto.BalanceView{Points: bal.Points, UpdatedAt: bal.UpdatedAt}, nil
}

func (s *adminService) ResetPassword(ctx context.Context, userID string) (string, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return "", domain.ErrInvalidInput
	}
	if _, err := s.users.GetUserByID(ctx, id); err != nil {
		return "", err
	}
	newPass, err := generatePassword(12)
	if err != nil {
		return "", err
	}
	plain, err := vo.NewPlainPassword(newPass)
	if err != nil {
		return "", err
	}
	hash, err := s.account.HashPassword(plain)
	if err != nil {
		return "", err
	}
	if err := s.users.UpdatePassword(ctx, id, hash); err != nil {
		return "", err
	}
	return newPass, nil
}

func (s *adminService) BlockUser(ctx context.Context, userID string, blocked bool) (dto.UserView, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return dto.UserView{}, domain.ErrInvalidInput
	}
	status := vo.StatusActive
	if blocked {
		status = vo.StatusBlocked
	}
	u, err := s.users.UpdateUserStatus(ctx, id, status)
	if err != nil {
		return dto.UserView{}, err
	}
	return toUserView(u), nil
}

func (s *adminService) ChangeRole(ctx context.Context, userID, role string) (dto.UserView, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return dto.UserView{}, domain.ErrInvalidInput
	}
	roleVO, err := vo.NewRole(role)
	if err != nil {
		return dto.UserView{}, err
	}
	u, err := s.users.UpdateUserRole(ctx, id, roleVO)
	if err != nil {
		return dto.UserView{}, err
	}
	return toUserView(u), nil
}

func (s *adminService) ListUsers(ctx context.Context, in dto.ListUsersInput) ([]dto.UserView, string, error) {
	users, nextToken, err := s.users.ListUsers(ctx, repository.ListUsersFilter{
		Search:    in.Search,
		RoleCode:  in.RoleCode,
		Status:    in.Status,
		PageSize:  in.PageSize,
		PageToken: in.PageToken,
	})
	if err != nil {
		return nil, "", err
	}
	views := make([]dto.UserView, 0, len(users))
	for _, u := range users {
		views = append(views, toUserView(u))
	}
	return views, nextToken, nil
}

func (s *adminService) GetTransactions(ctx context.Context, in dto.GetTransactionsInput) ([]dto.TransactionView, string, error) {
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

func (s *adminService) NewUsersStats(ctx context.Context, period string) (dto.UsersStatsView, error) {
	n, err := s.users.CountUsersCreatedSince(ctx, periodSince(period))
	if err != nil {
		return dto.UsersStatsView{}, err
	}
	return dto.UsersStatsView{NewUsers: n}, nil
}
func periodSince(period string) time.Time {
	now := time.Now()
	switch period {
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, -1, 0)
	default:
		return now.AddDate(0, 0, -1)
	}
}

func generatePassword(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:length], nil
}
