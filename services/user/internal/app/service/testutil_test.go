package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
	domainsvc "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/service"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type fakeUserRepo struct {
	existing  map[string]bool
	createID  uuid.UUID
	created   []*model.User
	existsErr error
	createErr error

	byID    map[uuid.UUID]*model.User
	byLogin map[string]*model.User

	getUserErr       error
	createUserErr    error
	updateStatusErr  error
	updateRoleErr    error
	updateProfileErr error
	updatePassErr    error
	listUsersErr     error
	listUsers        []*model.User
	listUsersToken   string
	countSince       int
	countSinceErr    error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		existing:  make(map[string]bool),
		byID:      make(map[uuid.UUID]*model.User),
		byLogin:   make(map[string]*model.User),
		createID:  uuid.New(),
		listUsers: []*model.User{},
	}
}

func (f *fakeUserRepo) CreateUser(_ context.Context, u *model.User) (uuid.UUID, error) {
	if f.createUserErr != nil {
		return uuid.Nil, f.createUserErr
	}
	f.created = append(f.created, u)
	return f.createID, nil
}

func (f *fakeUserRepo) GetUserByLogin(_ context.Context, login vo.Login) (*model.User, error) {
	if u, ok := f.byLogin[login.String()]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (f *fakeUserRepo) ExistsByLogin(_ context.Context, login vo.Login) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.existing[login.String()], nil
}

func (f *fakeUserRepo) GetUserByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	if f.getUserErr != nil {
		return nil, f.getUserErr
	}
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (f *fakeUserRepo) UpdateUserStatus(_ context.Context, id uuid.UUID, status vo.Status) (*model.User, error) {
	if f.updateStatusErr != nil {
		return nil, f.updateStatusErr
	}
	u, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	u.Status = status
	return u, nil
}

func (f *fakeUserRepo) UpdateUserRole(_ context.Context, id uuid.UUID, role vo.Role) (*model.User, error) {
	if f.updateRoleErr != nil {
		return nil, f.updateRoleErr
	}
	u, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	u.Role = role
	return u, nil
}

func (f *fakeUserRepo) UpdateUserProfile(_ context.Context, id uuid.UUID, p repository.ProfileUpdate) (*model.User, error) {
	if f.updateProfileErr != nil {
		return nil, f.updateProfileErr
	}
	u, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	u.FirstName = p.FirstName
	u.LastName = p.LastName
	u.Patronymic = p.Patronymic
	u.Email = vo.NewEmailFromStored(p.Email)
	u.Phone = vo.NewPhoneNumberFromStored(p.PhoneNumber)
	return u, nil
}

func (f *fakeUserRepo) UpdatePassword(_ context.Context, _ uuid.UUID, _ vo.PasswordHash) error {
	return f.updatePassErr
}

func (f *fakeUserRepo) ListUsers(_ context.Context, _ repository.ListUsersFilter) ([]*model.User, string, error) {
	if f.listUsersErr != nil {
		return nil, "", f.listUsersErr
	}
	return f.listUsers, f.listUsersToken, nil
}

func (f *fakeUserRepo) CountUsersCreatedSince(_ context.Context, _ time.Time) (int, error) {
	if f.countSinceErr != nil {
		return 0, f.countSinceErr
	}
	return f.countSince, nil
}

type fakeReader struct {
	users map[string]*model.User
}

func newFakeReader() *fakeReader {
	return &fakeReader{users: make(map[string]*model.User)}
}

func (f *fakeReader) GetUserByLogin(_ context.Context, login string) (*model.User, error) {
	u, ok := f.users[login]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

// --- fakeAccount ---

type fakeAccount struct {
	verify           bool
	tokens           port.TokenPair
	hashErr          error
	genErr           error
	validateErr      error
	validateIdentity model.Identity
}

func (f *fakeAccount) HashPassword(p vo.PlainPassword) (vo.PasswordHash, error) {
	if f.hashErr != nil {
		return vo.PasswordHash{}, f.hashErr
	}
	return vo.NewPasswordHash("hashed:" + p.Reveal()), nil
}

func (f *fakeAccount) VerifyPassword(_ vo.PlainPassword, _ vo.PasswordHash) bool {
	return f.verify
}

func (f *fakeAccount) GenerateTokens(_ model.Identity) (port.TokenPair, error) {
	if f.genErr != nil {
		return port.TokenPair{}, f.genErr
	}
	return f.tokens, nil
}

func (f *fakeAccount) ValidateToken(_ string) (model.Identity, error) {
	if f.validateErr != nil {
		return model.Identity{}, f.validateErr
	}
	return f.validateIdentity, nil
}

type fakeTokenStore struct {
	stored             map[string]string
	storeErr           error
	getUserErr         error
	deleteErr          error
	deleteByUserErr    error
	deleteByUserCalled bool
}

func newFakeTokenStore() *fakeTokenStore {
	return &fakeTokenStore{stored: make(map[string]string)}
}

func (f *fakeTokenStore) Store(_ context.Context, userID, token string, _ time.Duration) error {
	if f.storeErr != nil {
		return f.storeErr
	}
	f.stored[token] = userID
	return nil
}

func (f *fakeTokenStore) GetUserID(_ context.Context, token string) (string, error) {
	if f.getUserErr != nil {
		return "", f.getUserErr
	}
	id, ok := f.stored[token]
	if !ok {
		return "", domain.ErrInvalidToken
	}
	return id, nil
}

func (f *fakeTokenStore) Delete(_ context.Context, token string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.stored, token)
	return nil
}

func (f *fakeTokenStore) DeleteByUserID(_ context.Context, _ string) error {
	if f.deleteByUserErr != nil {
		return f.deleteByUserErr
	}
	f.deleteByUserCalled = true
	return nil
}

type fakeBalanceRepo struct {
	balances    map[uuid.UUID]*model.PointsBalance
	applyErr    error
	getTransErr error
	txs         []model.PointsTransaction
	txsToken    string
}

func newFakeBalanceRepo() *fakeBalanceRepo {
	return &fakeBalanceRepo{balances: make(map[uuid.UUID]*model.PointsBalance)}
}

func (f *fakeBalanceRepo) EnsureBalance(_ context.Context, userID uuid.UUID) error {
	if _, ok := f.balances[userID]; !ok {
		f.balances[userID] = &model.PointsBalance{UserID: userID, Points: 0, UpdatedAt: time.Now()}
	}
	return nil
}

func (f *fakeBalanceRepo) GetBalance(_ context.Context, userID uuid.UUID) (model.PointsBalance, error) {
	b, ok := f.balances[userID]
	if !ok {
		return model.PointsBalance{}, nil
	}
	return *b, nil
}

func (f *fakeBalanceRepo) Apply(_ context.Context, cmd repository.ApplyPointsCommand, mutate func(*model.PointsBalance) error) (model.PointsBalance, error) {
	if f.applyErr != nil {
		return model.PointsBalance{}, f.applyErr
	}
	b, ok := f.balances[cmd.UserID]
	if !ok {
		b = &model.PointsBalance{UserID: cmd.UserID, Points: 0, UpdatedAt: time.Now()}
		f.balances[cmd.UserID] = b
	}
	if err := mutate(b); err != nil {
		return model.PointsBalance{}, err
	}
	return *b, nil
}

func (f *fakeBalanceRepo) GetTransactions(_ context.Context, userID uuid.UUID, _ int, _ string) ([]model.PointsTransaction, string, error) {
	if f.getTransErr != nil {
		return nil, "", f.getTransErr
	}
	return f.txs, f.txsToken, nil
}

type harness struct {
	repo    *fakeUserRepo
	reader  *fakeReader
	account *fakeAccount
	tokens  *fakeTokenStore
	balance *fakeBalanceRepo

	authSvc   AuthService
	userSvc   UserService
	adminSvc  AdminService
	pointsSvc *pointsManagerAdapter
}

func newHarness() *harness {
	repo := newFakeUserRepo()
	reader := newFakeReader()
	account := &fakeAccount{}
	tokens := newFakeTokenStore()
	balance := newFakeBalanceRepo()

	h := &harness{
		repo:    repo,
		reader:  reader,
		account: account,
		tokens:  tokens,
		balance: balance,
	}
	h.authSvc = NewAuthService(repo, reader, account, tokens, time.Hour)
	h.userSvc = NewUserService(repo, balance, domainsvc.NewPointsManager(balance), account)
	h.adminSvc = NewAdminService(repo, balance, domainsvc.NewPointsManager(balance), account, tokens)
	return h
}


type pointsManagerAdapter struct {
	balances *fakeBalanceRepo
}

func (m *pointsManagerAdapter) Credit(ctx context.Context, cmd repository.ApplyPointsCommand, amount int64) (model.PointsBalance, error) {
	return m.balances.Apply(ctx, cmd, func(b *model.PointsBalance) error {
		return b.Add(amount)
	})
}

func (m *pointsManagerAdapter) Debit(ctx context.Context, cmd repository.ApplyPointsCommand, amount int64) (model.PointsBalance, error) {
	return m.balances.Apply(ctx, cmd, func(b *model.PointsBalance) error {
		return b.Deduct(amount)
	})
}

var errBoom = errors.New("boom")

func validRegisterInput() dto.RegisterInput {
	return dto.RegisterInput{
		Login:       "testuser",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Patronymic:  "Smith",
		Email:       "john@example.com",
		PhoneNumber: "+1234567890",
	}
}

func seedUser(h *harness, login string) *model.User {
	u := &model.User{
		ID:           uuid.New(),
		Login:        vo.NewLoginFromStored(login),
		PasswordHash: vo.NewPasswordHash("hash"),
		Email:        vo.NewEmailFromStored(login + "@example.com"),
		Role:         vo.RoleUser,
		Status:       vo.StatusActive,
	}
	h.repo.byID[u.ID] = u
	h.repo.byLogin[login] = u
	h.reader.users[login] = u
	return u
}
