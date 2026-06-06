package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)


type fakeUserRepo struct {
	existing  map[string]bool 
	createID  uuid.UUID
	created   []*model.User
	existsErr error
	createErr error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{existing: make(map[string]bool), createID: uuid.New()}
}

func (f *fakeUserRepo) CreateUser(ctx context.Context, u *model.User) (uuid.UUID, error) {
	if f.createErr != nil {
		return uuid.Nil, f.createErr
	}
	f.created = append(f.created, u)
	return f.createID, nil
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


type fakeReader struct {
	users map[string]*model.User
}

func newFakeReader() *fakeReader {
	return &fakeReader{users: make(map[string]*model.User)}
}

func (f *fakeReader) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	u, ok := f.users[login]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}


type fakeAccount struct {
	verify  bool 
	tokens  port.TokenPair
	hashErr error
	genErr  error
}

func (f *fakeAccount) HashPassword(p vo.PlainPassword) (vo.PasswordHash, error) {
	if f.hashErr != nil {
		return vo.PasswordHash{}, f.hashErr
	}
	return vo.NewPasswordHash("hashed:" + p.Reveal()), nil
}

func (f *fakeAccount) VerifyPassword(p vo.PlainPassword, h vo.PasswordHash) bool { return f.verify }

func (f *fakeAccount) GenerateTokens(id model.Identity) (port.TokenPair, error) {
	if f.genErr != nil {
		return port.TokenPair{}, f.genErr
	}
	return f.tokens, nil
}

func (f *fakeAccount) ValidateToken(token string) (model.Identity, error) {
	return model.Identity{}, nil
}


type fakeTokenStore struct {
	stored   map[string]string 
	storeErr error
}

func newFakeTokenStore() *fakeTokenStore {
	return &fakeTokenStore{stored: make(map[string]string)}
}

func (f *fakeTokenStore) Store(ctx context.Context, userID, token string, ttl time.Duration) error {
	if f.storeErr != nil {
		return f.storeErr
	}
	f.stored[token] = userID
	return nil
}

func (f *fakeTokenStore) GetUserID(ctx context.Context, token string) (string, error) {
	id, ok := f.stored[token]
	if !ok {
		return "", domain.ErrInvalidToken
	}
	return id, nil
}

func (f *fakeTokenStore) Delete(ctx context.Context, token string) error {
	delete(f.stored, token)
	return nil
}

func (f *fakeTokenStore) DeleteByUserID(ctx context.Context, userID string) error { return nil }

type harness struct {
	repo    *fakeUserRepo
	reader  *fakeReader
	account *fakeAccount
	tokens  *fakeTokenStore
	svc     AuthService
}

func newHarness() *harness {
	repo := newFakeUserRepo()
	reader := newFakeReader()
	account := &fakeAccount{}
	tokens := newFakeTokenStore()
	return &harness{
		repo:    repo,
		reader:  reader,
		account: account,
		tokens:  tokens,
		svc:     NewAuthService(repo, reader, account, tokens, time.Hour),
	}
}


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

func TestAuthService_Register(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		view, err := h.svc.Register(context.Background(), validRegisterInput())

		assert.NoError(t, err)
		assert.Equal(t, "testuser", view.Login)
		assert.Equal(t, "john@example.com", view.Email)
		assert.Equal(t, h.repo.createID.String(), view.ID)
		assert.Len(t, h.repo.created, 1) 
	})

	t.Run("duplicate login", func(t *testing.T) {
		h := newHarness()
		h.repo.existing["testuser"] = true

		_, err := h.svc.Register(context.Background(), validRegisterInput())
		assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
		assert.Empty(t, h.repo.created) 
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			mutate  func(in *dto.RegisterInput)
			wantErr error
		}{
			{"empty login", func(in *dto.RegisterInput) { in.Login = "   " }, domain.ErrEmptyLogin},
			{"invalid email", func(in *dto.RegisterInput) { in.Email = "invalid-email" }, domain.ErrInvalidEmailFormat},
			{"invalid phone", func(in *dto.RegisterInput) { in.PhoneNumber = "abc" }, domain.ErrInvalidPhoneFormat},
			{"weak password", func(in *dto.RegisterInput) { in.Password = "123" }, domain.ErrWeakPassword},
			{"empty first name", func(in *dto.RegisterInput) { in.FirstName = "" }, domain.ErrEmptyFirstName},
			{"empty last name", func(in *dto.RegisterInput) { in.LastName = "" }, domain.ErrEmptyLastName},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h := newHarness()
				in := validRegisterInput()
				tt.mutate(&in)

				_, err := h.svc.Register(context.Background(), in)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, h.repo.created)
			})
		}
	})
}

func TestAuthService_Login(t *testing.T) {

	seedUser := func(h *harness) *model.User {
		u := &model.User{
			ID:           uuid.New(),
			Login:        vo.NewLoginFromStored("testuser"),
			PasswordHash: vo.NewPasswordHash("hash"),
			Email:        vo.NewEmailFromStored("john@example.com"),
			Role:         "user",
		}
		h.reader.users["testuser"] = u
		return u
	}

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h)
		h.account.verify = true
		h.account.tokens = port.TokenPair{AccessToken: "a", RefreshToken: "r"}

		res, err := h.svc.Login(context.Background(), dto.LoginInput{Login: "testuser", Password: "password123"})
		assert.NoError(t, err)
		assert.Equal(t, "r", res.Tokens.RefreshToken)
		assert.Equal(t, u.ID.String(), h.tokens.stored["r"]) // refresh-токен сохранён под id юзера
	})

	t.Run("empty credentials", func(t *testing.T) {
		h := newHarness()
		_, err := h.svc.Login(context.Background(), dto.LoginInput{Login: "  ", Password: ""})
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("unknown user maps to invalid credentials", func(t *testing.T) {
		h := newHarness() 
		_, err := h.svc.Login(context.Background(), dto.LoginInput{Login: "ghost", Password: "password123"})
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("wrong password", func(t *testing.T) {
		h := newHarness()
		seedUser(h)
		h.account.verify = false

		_, err := h.svc.Login(context.Background(), dto.LoginInput{Login: "testuser", Password: "wrong"})
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})
}

func TestAuthService_Logout(t *testing.T) {
	t.Run("empty token", func(t *testing.T) {
		h := newHarness()
		err := h.svc.Logout(context.Background(), "  ")
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("deletes token", func(t *testing.T) {
		h := newHarness()
		h.tokens.stored["r"] = "user-1"

		err := h.svc.Logout(context.Background(), "r")
		assert.NoError(t, err)
		assert.NotContains(t, h.tokens.stored, "r")
	})
}
