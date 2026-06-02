package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/contract"
)

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	var req contract.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "INVALID_JSON", Message: "invalid request body"})
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
	var req contract.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "INVALID_JSON", Message: "invalid request body"})
		return
	}

	res, err := s.auth.Login(r.Context(), dto.LoginInput{Login: req.Login, Password: req.Password})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, contract.LoginResponse{
		User: toUserResponse(res.User),
		Tokens: contract.Tokens{
			AccessToken:  res.Tokens.AccessToken,
			RefreshToken: res.Tokens.RefreshToken,
		},
	})
}

func (s *Server) Refresh(w http.ResponseWriter, r *http.Request) {
	var req contract.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "INVALID_JSON", Message: "invalid request body"})
		return
	}

	tokens, err := s.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, contract.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	var req contract.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "INVALID_JSON", Message: "invalid request body"})
		return
	}

	if err := s.auth.Logout(r.Context(), req.RefreshToken); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toUserResponse(v dto.UserView) contract.UserResponse {
	return contract.UserResponse{
		ID:          v.ID,
		Login:       v.Login,
		FirstName:   v.FirstName,
		LastName:    v.LastName,
		Patronymic:  v.Patronymic,
		Email:       v.Email,
		PhoneNumber: v.PhoneNumber,
	}
}
