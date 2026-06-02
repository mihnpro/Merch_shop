package middleware

import (
	"github.com/mihnpro/Merch_shop/services/gateway/internal/domain"
	"net/http"
)

func (m *Middleware) RequireAdminRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims := domain.ClaimsFrom(ctx)
		if claims == nil {
			unauthorized(w)
			return
		}
		if !claims.Role.IsAdmin() {
			forbidden(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}
