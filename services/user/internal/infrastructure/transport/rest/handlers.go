package rest

import (
	"encoding/json"
	"net/http"

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
	ID          string `json:"id"`
	Login       string `json:"login"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Patronymic  string `json:"patronymic,omitempty"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type tokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type loginResponse struct {
	User   userResponse   `json:"user"`
	Tokens tokensResponse `json:"tokens"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
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

	writeJSON(w, http.StatusOK, loginResponse{
		User: toUserResponse(res.User),
		Tokens: tokensResponse{
			AccessToken:  res.Tokens.AccessToken,
			RefreshToken: res.Tokens.RefreshToken,
		},
	})
}

func (s *Server) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "INVALID_JSON", Message: "invalid request body"})
		return
	}

	tokens, err := s.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "INVALID_JSON", Message: "invalid request body"})
		return
	}

	if err := s.auth.Logout(r.Context(), req.RefreshToken); err != nil {
		writeError(w, err)
		return
	}

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
	return userResponse{
		ID:          v.ID,
		Login:       v.Login,
		FirstName:   v.FirstName,
		LastName:    v.LastName,
		Patronymic:  v.Patronymic,
		Email:       v.Email,
		PhoneNumber: v.PhoneNumber,
	}
}
