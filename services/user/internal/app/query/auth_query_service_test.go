package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type mockUserRepo struct {
	users    map[string]*model.User
	getCalls int 
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*model.User)}
}

func (m *mockUserRepo) GetUserByLogin(ctx context.Context, login vo.Login) (*model.User, error) {
	m.getCalls++
	u, ok := m.users[login.String()]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) CreateUser(ctx context.Context, u *model.User) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *mockUserRepo) ExistsByLogin(ctx context.Context, login vo.Login) (bool, error) {
	_, ok := m.users[login.String()]
	return ok, nil
}

func TestAuthReadService_GetUserByLogin(t *testing.T) {
	existing := &model.User{ID: uuid.New(), FirstName: "John"}

	tests := []struct {
		name        string
		login       string
		seed        *model.User 
		wantErr     error
		wantUser    bool
		wantRepoHit bool 
	}{
		{
			name:        "existing user",
			login:       "testuser",
			seed:        existing,
			wantErr:     nil,
			wantUser:    true,
			wantRepoHit: true,
		},
		{
			name:        "user not found",
			login:       "ghost",
			seed:        nil,
			wantErr:     domain.ErrUserNotFound,
			wantUser:    false,
			wantRepoHit: true,
		},
		{
			name:        "empty login is rejected before repo",
			login:       "   ",
			seed:        nil,
			wantErr:     domain.ErrEmptyLogin,
			wantUser:    false,
			wantRepoHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepo()
			if tt.seed != nil {
				repo.users[tt.login] = tt.seed
			}
			svc := NewAuthReadService(repo)

			got, err := svc.GetUserByLogin(context.Background(), tt.login)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantUser && got != tt.seed {
				t.Errorf("got user %+v, want %+v", got, tt.seed)
			}
			if !tt.wantUser && got != nil {
				t.Errorf("expected nil user, got %+v", got)
			}
			if repoHit := repo.getCalls > 0; repoHit != tt.wantRepoHit {
				t.Errorf("repo called = %v, want %v", repoHit, tt.wantRepoHit)
			}
		})
	}
}
