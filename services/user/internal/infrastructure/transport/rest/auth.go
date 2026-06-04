package rest

import (
	"context"
	"net/http"
	"strings"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
)

const bearerPrefix = "Bearer "

type identityCtxKey struct{}

type Middleware struct {
	account port.Account
}

func NewMiddleware(account port.Account) *Middleware {
	return &Middleware{account: account}
}


func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, domain.ErrInvalidToken)
			return
		}

		identity, err := m.account.ValidateToken(token)
		if err != nil {
			writeError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), identityCtxKey{}, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


func IdentityFrom(ctx context.Context) (model.Identity, bool) {
	identity, ok := ctx.Value(identityCtxKey{}).(model.Identity)
	return identity, ok
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, bearerPrefix))
}
