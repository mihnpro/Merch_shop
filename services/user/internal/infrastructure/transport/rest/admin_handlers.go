package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)


type balanceResponse struct {
	Points    int64     `json:"points"`
	UpdatedAt time.Time `json:"updated_at"`
}

type transactionResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	OperationID string    `json:"operation_id"`
	OrderID     string    `json:"order_id,omitempty"`
	Amount      int64     `json:"amount"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

type transactionsResponse struct {
	Transactions  []transactionResponse `json:"transactions"`
	NextPageToken string                `json:"next_page_token"`
}

type listUsersResponse struct {
	Users         []userResponse `json:"users"`
	NextPageToken string         `json:"next_page_token"`
}

type resetPasswordResponse struct {
	NewPassword string `json:"new_password"`
}

func toBalanceResponse(v dto.BalanceView) balanceResponse {
	return balanceResponse{Points: v.Points, UpdatedAt: v.UpdatedAt}
}

func toTransactionsResponse(txs []dto.TransactionView, nextToken string) transactionsResponse {
	items := make([]transactionResponse, 0, len(txs))
	for _, t := range txs {
		items = append(items, transactionResponse{
			ID:          t.ID,
			UserID:      t.UserID,
			OperationID: t.OperationID,
			OrderID:     t.OrderID,
			Amount:      t.Amount,
			Reason:      t.Reason,
			CreatedAt:   t.CreatedAt,
		})
	}
	return transactionsResponse{Transactions: items, NextPageToken: nextToken}
}

func toListUsersResponse(views []dto.UserView, nextToken string) listUsersResponse {
	users := make([]userResponse, 0, len(views))
	for _, v := range views {
		users = append(users, toUserResponse(v))
	}
	return listUsersResponse{Users: users, NextPageToken: nextToken}
}

// ─── Balance (self) ───────────────────────────────────────────────────────────

func (s *Server) GetMyBalance(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	bal, err := s.user.GetBalance(r.Context(), identity.UserID.String())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBalanceResponse(bal))
}

func (s *Server) GetMyTransactions(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	q := r.URL.Query()
	txs, nextToken, err := s.user.GetTransactions(r.Context(), dto.GetTransactionsInput{
		UserID:    identity.UserID.String(),
		PageSize:  queryInt(q, "page_size", 25),
		PageToken: q.Get("page_token"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toTransactionsResponse(txs, nextToken))
}

// ─── Admin: users ─────────────────────────────────────────────────────────────

func (s *Server) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	in := dto.ListUsersInput{
		Search:    q.Get("search"),
		RoleCode:  q.Get("role"),
		Status:    q.Get("status"),
		PageSize:  queryInt(q, "page_size", 25),
		PageToken: q.Get("page_token"),
	}
	users, nextToken, err := s.admin.ListUsers(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toListUsersResponse(users, nextToken))
}

func (s *Server) AdminGetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	view, err := s.user.GetUser(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(view))
}

type grantPointsRequest struct {
	Amount      int64  `json:"amount"`
	OperationID string `json:"operation_id"`
	Reason      string `json:"reason"`
}

func (s *Server) AdminGrantPoints(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	var req grantPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "INVALID_JSON", Message: "invalid request body"})
		return
	}
	bal, err := s.admin.GrantPoints(r.Context(), dto.GrantPointsInput{
		UserID:      userID,
		Amount:      req.Amount,
		OperationID: req.OperationID,
		Reason:      req.Reason,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBalanceResponse(bal))
}

func (s *Server) AdminResetPassword(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	newPass, err := s.admin.ResetPassword(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resetPasswordResponse{NewPassword: newPass})
}

type blockUserRequest struct {
	Blocked bool   `json:"blocked"`
	Reason  string `json:"reason"`
}

func (s *Server) AdminBlockUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	var req blockUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "INVALID_JSON", Message: "invalid request body"})
		return
	}
	view, err := s.admin.BlockUser(r.Context(), userID, req.Blocked)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(view))
}

type changeRoleRequest struct {
	Role string `json:"role"`
}

func (s *Server) AdminChangeRole(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	var req changeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "INVALID_JSON", Message: "invalid request body"})
		return
	}
	view, err := s.admin.ChangeRole(r.Context(), userID, req.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(view))
}

func (s *Server) AdminGetUserTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	q := r.URL.Query()
	txs, nextToken, err := s.admin.GetTransactions(r.Context(), dto.GetTransactionsInput{
		UserID:    userID,
		PageSize:  queryInt(q, "page_size", 25),
		PageToken: q.Get("page_token"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toTransactionsResponse(txs, nextToken))
}

func queryInt(q interface{ Get(string) string }, key string, def int) int {
	var n int
	if _, err := fmt.Sscanf(q.Get(key), "%d", &n); err != nil || n <= 0 {
		return def
	}
	return n
}
