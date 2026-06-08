package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

const (
	bearerPrefix     = "Bearer "
	accessCookieName = "access_token"
)

func (m *Middleware) Authorize(w http.ResponseWriter, r *http.Request) bool {
	token, ok := accessToken(r)
	if !ok {
		unauthorized(w)
		return false
	}

	if err := m.verify.VerifyAccess(token); err != nil {
		unauthorized(w)
		return false
	}
	r.Header.Set("Authorization", bearerPrefix+token)
	return true
}

func accessToken(r *http.Request) (string, bool) {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, bearerPrefix) {
		if token := strings.TrimSpace(strings.TrimPrefix(h, bearerPrefix)); token != "" {
			return token, true
		}
	}
	if c, err := r.Cookie(accessCookieName); err == nil && c.Value != "" {
		return c.Value, true
	}
	return "", false
}

func unauthorized(w http.ResponseWriter) {
	writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "unauthenticated")
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    code,
		"message": message,
	})
}
