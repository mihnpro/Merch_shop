package rest

import (
	"net/http"
	"strings"

	"github.com/mihnpro/Merch_shop/services/media/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
)

const bearerPrefix = "Bearer "

type Middleware struct {
	verify port.Verifier
}

func NewMiddleware(verify port.Verifier) *Middleware {
	return &Middleware{verify: verify}
}

func (m *Middleware) AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := m.verify.ParseAccess(bearerToken(r))
		if err != nil {
			writeError(w, err)
			return
		}
		if !id.IsAdmin() {
			writeError(w, domain.ErrForbidden)
			return
		}
		next(w, r)
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, bearerPrefix))
}
