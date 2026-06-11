package rest

import (
	"net/http"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

func (m *Middleware) Admin(next http.Handler) http.Handler {
	return m.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := IdentityFrom(r.Context())
		if !ok {
			writeError(w, domain.ErrInvalidToken)
			return
		}
		if !vo.NewRoleFromStored(identity.Role).IsAdmin() {
			writeError(w, domain.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}
