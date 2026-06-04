package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

const bearerPrefix = "Bearer "

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.isPublic(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, bearerPrefix) {
			unauthorized(w)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(h, bearerPrefix))
		if token == "" {
			unauthorized(w)
			return
		}

		if err := m.verify.VerifyAccess(token); err != nil {
			unauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) isPublic(path string) bool {
	for _, p := range m.publicPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    "UNAUTHENTICATED",
		"message": "unauthenticated",
	})
}
