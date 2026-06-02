package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/domain"
)

const bearerPrefix = "Bearer "

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, bearerPrefix) {
			unauthorized(w)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(h, bearerPrefix))

		claims, err := m.verify.VerifyAccess(token)
		if err != nil {
			unauthorized(w)
			return
		}

		ctx := domain.WithClaims(r.Context(), claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    "UNAUTHENTICATED",
		"message": "unauthenticated",
	})
}

func forbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    "PERMISSION_DENIED",
		"message": "forbidden",
	})
}
