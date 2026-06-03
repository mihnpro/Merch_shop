package account

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

func newTestAccount() *account {
	return &account{
		accessSecret:  "access-secret",
		refreshSecret: "refresh-secret",
		accessTTL:     time.Minute,
		refreshTTL:    time.Hour,
	}
}

func TestAccount_HashAndVerifyPassword(t *testing.T) {
	a := newTestAccount()
	pwd := vo.PlainPasswordFromRaw("password123")

	hash, err := a.HashPassword(pwd)
	require.NoError(t, err)
	assert.NotEmpty(t, hash.String())
	assert.NotEqual(t, "password123", hash.String())

	t.Run("correct password verifies", func(t *testing.T) {
		assert.True(t, a.VerifyPassword(pwd, hash))
	})

	t.Run("wrong password rejected", func(t *testing.T) {
		assert.False(t, a.VerifyPassword(vo.PlainPasswordFromRaw("wrong-password"), hash))
	})

	t.Run("same password hashes to different values (salt)", func(t *testing.T) {
		hash2, err := a.HashPassword(pwd)
		require.NoError(t, err)
		assert.NotEqual(t, hash.String(), hash2.String())
		assert.True(t, a.VerifyPassword(pwd, hash2))
	})
}

func TestAccount_GenerateAndValidateToken(t *testing.T) {
	a := newTestAccount()
	identity := model.Identity{
		UserID: uuid.New(),
		Email:  "john@example.com",
		Role:   "customer",
	}

	pair, err := a.GenerateTokens(identity)
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)
	assert.True(t, pair.AccessExpiresAt.After(pair.AccessExpiresAt.Add(-a.accessTTL)))

	t.Run("access token round-trips to identity", func(t *testing.T) {
		got, err := a.ValidateToken(pair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, identity, got)
	})

	t.Run("refresh token round-trips to identity", func(t *testing.T) {
		got, err := a.ValidateToken(pair.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, identity, got)
	})

	t.Run("garbage token rejected", func(t *testing.T) {
		_, err := a.ValidateToken("not-a-jwt")
		assert.ErrorIs(t, err, domain.ErrInvalidToken)
	})

	t.Run("token signed with another secret rejected", func(t *testing.T) {
		other := &account{
			accessSecret:  "different-secret",
			refreshSecret: "different-refresh",
			accessTTL:     time.Minute,
			refreshTTL:    time.Hour,
		}
		foreign, err := other.GenerateTokens(identity)
		require.NoError(t, err)

		_, err = a.ValidateToken(foreign.AccessToken)
		assert.ErrorIs(t, err, domain.ErrInvalidToken)
	})

	t.Run("expired token rejected", func(t *testing.T) {
		expiredAcc := &account{
			accessSecret:  "access-secret",
			refreshSecret: "refresh-secret",
			accessTTL:     -time.Minute,
			refreshTTL:    -time.Minute,
		}
		expired, err := expiredAcc.GenerateTokens(identity)
		require.NoError(t, err)

		_, err = expiredAcc.ValidateToken(expired.AccessToken)
		assert.ErrorIs(t, err, domain.ErrInvalidToken)
	})
}
