package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

type registerRequest struct {
	Login       string `json:"login"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Patronymic  string `json:"patronymic,omitempty"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type userResponse struct {
	ID          string     `json:"id"`
	Login       string     `json:"login"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Patronymic  string     `json:"patronymic,omitempty"`
	Email       string     `json:"email"`
	PhoneNumber string     `json:"phone_number,omitempty"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	User userResponse `json:"user"`
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "INVALID_JSON", Message: "invalid request body"})
		return
	}

	view, err := s.auth.Register(r.Context(), dto.RegisterInput{
		Login:       req.Login,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Patronymic:  req.Patronymic,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(view))
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "INVALID_JSON", Message: "invalid request body"})
		return
	}

	res, err := s.auth.Login(r.Context(), dto.LoginInput{Login: req.Login, Password: req.Password})
	if err != nil {
		writeError(w, err)
		return
	}

	s.setAuthCookies(w, res.Tokens.AccessToken, res.Tokens.RefreshToken)
	writeJSON(w, http.StatusOK, loginResponse{User: toUserResponse(res.User)})
}

func (s *Server) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := readRefreshCookie(r)
	if refreshToken == "" {
		writeError(w, domain.ErrInvalidToken)
		return
	}

	tokens, err := s.auth.Refresh(r.Context(), refreshToken)
	if err != nil {
		writeError(w, err)
		return
	}

	s.setAuthCookies(w, tokens.AccessToken, tokens.RefreshToken)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	if refreshToken := readRefreshCookie(r); refreshToken != "" {
		if err := s.auth.Logout(r.Context(), refreshToken); err != nil {
			writeError(w, err)
			return
		}
	}

	s.clearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

type meResponse struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}

	writeJSON(w, http.StatusOK, meResponse{UserID: identity.UserID.String(), Role: identity.Role})
}

func toUserResponse(v dto.UserView) userResponse {
	resp := userResponse{
		ID:          v.ID,
		Login:       v.Login,
		FirstName:   v.FirstName,
		LastName:    v.LastName,
		Patronymic:  v.Patronymic,
		Email:       v.Email,
		PhoneNumber: v.PhoneNumber,
		Role:        v.Role,
		Status:      v.Status,
		LastLoginAt: v.LastLoginAt,
	}
	if !v.CreatedAt.IsZero() {
		resp.CreatedAt = &v.CreatedAt
	}
	if !v.UpdatedAt.IsZero() {
		resp.UpdatedAt = &v.UpdatedAt
	}
	return resp
}
