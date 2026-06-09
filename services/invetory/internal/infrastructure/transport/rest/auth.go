package rest

import (
	"context"
	"net/http"
	"strings"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"
)

const (
	bearerPrefix = "Bearer "
	roleAdmin    = "admin"
)

type identityCtxKey struct{}

type Middleware struct {
	tokens port.TokenValidator
}

func NewMiddleware(tokens port.TokenValidator) *Middleware {
	return &Middleware{tokens: tokens}
}

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, domain.ErrInvalidToken)
			return
		}

		identity, err := m.tokens.ValidateToken(token)
		if err != nil {
			writeError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), identityCtxKey{}, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) Admin(next http.Handler) http.Handler {
	return m.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := IdentityFrom(r.Context())
		if !ok {
			writeError(w, domain.ErrInvalidToken)
			return
		}
		if identity.Role != roleAdmin {
			writeError(w, domain.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
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
