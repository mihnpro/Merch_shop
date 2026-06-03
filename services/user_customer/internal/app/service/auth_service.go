package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type AuthService interface {
	Register(ctx context.Context, in dto.RegisterInput) (dto.UserView, error)
	Login(ctx context.Context, in dto.LoginInput) (dto.AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (port.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	users      repository.UserRepository
	query      query.AuthReadService
	account    port.Account
	tokens     port.TokenStore
	refreshTTL time.Duration
}

func NewAuthService(
	users repository.UserRepository,
	reader query.AuthReadService,
	account port.Account,
	tokens port.TokenStore,
	refreshTTL time.Duration,
) AuthService {
	return &authService{
		users:      users,
		query:      reader,
		account:    account,
		tokens:     tokens,
		refreshTTL: refreshTTL,
	}
}

func (s *authService) Register(ctx context.Context, in dto.RegisterInput) (dto.UserView, error) {
	login, err := vo.NewLogin(in.Login)
	if err != nil {
		return dto.UserView{}, err
	}
	email, err := vo.NewEmail(in.Email)
	if err != nil {
		return dto.UserView{}, err
	}
	phone, err := vo.NewPhoneNumber(in.PhoneNumber)
	if err != nil {
		return dto.UserView{}, err
	}
	password, err := vo.NewPlainPassword(in.Password)
	if err != nil {
		return dto.UserView{}, err
	}

	exists, err := s.users.ExistsByLogin(ctx, login)
	if err != nil {
		return dto.UserView{}, err
	}
	if exists {
		return dto.UserView{}, domain.ErrUserAlreadyExists
	}

	hash, err := s.account.HashPassword(password)
	if err != nil {
		return dto.UserView{}, err
	}

	user, err := model.NewUser(login, hash, in.FirstName, in.LastName, in.Patronymic, email, phone)
	if err != nil {
		return dto.UserView{}, err
	}

	id, err := s.users.CreateUser(ctx, user)
	if err != nil {
		return dto.UserView{}, err
	}
	user.ID = id

	return toUserView(user), nil
}

func (s *authService) Login(ctx context.Context, in dto.LoginInput) (dto.AuthResult, error) {
	if strings.TrimSpace(in.Login) == "" || in.Password == "" {
		return dto.AuthResult{}, domain.ErrInvalidCredentials
	}

	user, err := s.query.GetUserByLogin(ctx, in.Login)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return dto.AuthResult{}, domain.ErrInvalidCredentials
		}
		return dto.AuthResult{}, err
	}

	if !s.account.VerifyPassword(vo.PlainPasswordFromRaw(in.Password), user.PasswordHash) {
		return dto.AuthResult{}, domain.ErrInvalidCredentials
	}

	tokens, err := s.account.GenerateTokens(user.Identity())
	if err != nil {
		return dto.AuthResult{}, err
	}

	if err := s.tokens.Store(ctx, user.ID.String(), tokens.RefreshToken, s.refreshTTL); err != nil {
		return dto.AuthResult{}, err
	}

	return dto.AuthResult{User: toUserView(user), Tokens: tokens}, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (port.TokenPair, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return port.TokenPair{}, domain.ErrInvalidInput
	}

	identity, err := s.account.ValidateToken(refreshToken)
	if err != nil {
		return port.TokenPair{}, err
	}

	if _, err := s.tokens.GetUserID(ctx, refreshToken); err != nil {
		return port.TokenPair{}, domain.ErrInvalidCredentials
	}

	if err := s.tokens.Delete(ctx, refreshToken); err != nil {
		return port.TokenPair{}, err
	}

	tokens, err := s.account.GenerateTokens(identity)
	if err != nil {
		return port.TokenPair{}, err
	}

	if err := s.tokens.Store(ctx, identity.UserID.String(), tokens.RefreshToken, s.refreshTTL); err != nil {
		return port.TokenPair{}, err
	}

	return tokens, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return domain.ErrInvalidInput
	}
	return s.tokens.Delete(ctx, refreshToken)
}

func toUserView(u *model.User) dto.UserView {
	return dto.UserView{
		ID:          u.ID.String(),
		Login:       u.Login.String(),
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Patronymic:  u.Patronymic,
		Email:       u.Email.String(),
		PhoneNumber: u.Phone.String(),
		Status:      u.Status,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
