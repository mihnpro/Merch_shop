package rest

import (
	"net/http"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

const roleAdmin = "admin"

func (m *Middleware) Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}
