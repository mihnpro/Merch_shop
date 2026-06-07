package rest

import (
	"net/http"
	"strings"
	"time"
)

const (
	accessCookieName  = "access_token"
	refreshCookieName = "refresh_token"
	refreshCookiePath = "/api/v1/auth"
)


type CookieConfig struct {
	Secure     bool
	SameSite   http.SameSite
	Domain     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}


func ParseSameSite(s string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func (s *Server) setAuthCookies(w http.ResponseWriter, access, refresh string) {
	http.SetCookie(w, s.cookie(accessCookieName, access, "/", int(s.cookies.AccessTTL.Seconds())))
	http.SetCookie(w, s.cookie(refreshCookieName, refresh, refreshCookiePath, int(s.cookies.RefreshTTL.Seconds())))
}

func (s *Server) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, s.cookie(accessCookieName, "", "/", -1))
	http.SetCookie(w, s.cookie(refreshCookieName, "", refreshCookiePath, -1))
}

func (s *Server) cookie(name, value, path string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   s.cookies.Domain,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   s.cookies.Secure,
		SameSite: s.cookies.SameSite,
	}
}

func readRefreshCookie(r *http.Request) string {
	c, err := r.Cookie(refreshCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
