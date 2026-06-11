package psql

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
)

type balanceRow struct {
	UserID    uuid.UUID `db:"user_id"`
	Points    int64     `db:"points"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r balanceRow) toModel() model.PointsBalance {
	return model.PointsBalance{UserID: r.UserID, Points: r.Points, UpdatedAt: r.UpdatedAt}
}

type txRow struct {
	ID          uuid.UUID     `db:"id"`
	UserID      uuid.UUID     `db:"user_id"`
	OperationID uuid.UUID     `db:"operation_id"`
	OrderID     uuid.NullUUID `db:"order_id"`
	Amount      int64         `db:"amount"`
	Reason      string        `db:"reason"`
	CreatedAt   time.Time     `db:"created_at"`
}

func (r txRow) toModel() model.PointsTransaction {
	t := model.PointsTransaction{
		ID:          r.ID,
		UserID:      r.UserID,
		OperationID: r.OperationID,
		Amount:      r.Amount,
		Reason:      r.Reason,
		CreatedAt:   r.CreatedAt,
	}
	if r.OrderID.Valid {
		id := r.OrderID.UUID
		t.OrderID = &id
	}
	return t
}

const (
	ensureBalanceSQL  = `INSERT INTO points_balance (user_id, points) VALUES ($1, 0) ON CONFLICT (user_id) DO NOTHING`
	selectBalanceSQL  = `SELECT user_id, points, updated_at FROM points_balance WHERE user_id = $1`
	lockBalanceSQL    = `SELECT user_id, points, updated_at FROM points_balance WHERE user_id = $1 FOR UPDATE`
	updateBalanceSQL  = `UPDATE points_balance SET points = points + $2 WHERE user_id = $1 RETURNING user_id, points, updated_at`
	insertPointsTxSQL = `INSERT INTO points_transactions (user_id, operation_id, order_id, amount, reason)
	 VALUES ($1, $2, $3, $4, $5) ON CONFLICT (operation_id) DO NOTHING`
)


func ensureBalance(ctx context.Context, exec sqlx.ExecerContext, userID uuid.UUID) error {
	_, err := exec.ExecContext(ctx, ensureBalanceSQL, userID)
	return err
}

func getBalanceRow(ctx context.Context, q sqlx.QueryerContext, userID uuid.UUID) (balanceRow, error) {
	var row balanceRow
	err := sqlx.GetContext(ctx, q, &row, selectBalanceSQL, userID)
	return row, err
}

func lockBalanceRow(ctx context.Context, q sqlx.QueryerContext, userID uuid.UUID) (balanceRow, error) {
	var row balanceRow
	err := sqlx.GetContext(ctx, q, &row, lockBalanceSQL, userID)
	return row, err
}

func updateBalanceRow(ctx context.Context, q sqlx.QueryerContext, userID uuid.UUID, delta int64) (balanceRow, error) {
	var row balanceRow
	err := sqlx.GetContext(ctx, q, &row, updateBalanceSQL, userID, delta)
	return row, err
}

func insertPointsTx(ctx context.Context, exec sqlx.ExecerContext, cmd repository.ApplyPointsCommand, delta int64) (bool, error) {
	res, err := exec.ExecContext(ctx, insertPointsTxSQL,
		cmd.UserID, cmd.OperationID, cmd.OrderID, delta, cmd.Reason)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func selectTransactions(ctx context.Context, db *sqlx.DB, userID uuid.UUID, limit int, cursor string) ([]txRow, error) {
	args := []any{userID}
	cond := ""
	if cursor != "" {
		if cur, err := decodeTxCursor(cursor); err == nil {
			cond = fmt.Sprintf(`AND (t.created_at, t.id) < ($%d, $%d)`, len(args)+1, len(args)+2)
			args = append(args, cur.CreatedAt, cur.ID)
		}
	}

	q := fmt.Sprintf(`
		SELECT t.id, t.user_id, t.operation_id, t.order_id, t.amount, t.reason, t.created_at
		FROM points_transactions t
		WHERE t.user_id = $1 %s
		ORDER BY t.created_at DESC, t.id DESC LIMIT $%d`, cond, len(args)+1)
	args = append(args, limit+1)

	var rows []txRow
	if err := db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	return rows, nil
}


type txCursor struct {
	CreatedAt time.Time `json:"ca"`
	ID        uuid.UUID `json:"id"`
}

func encodeTxCursor(createdAt time.Time, id uuid.UUID) string {
	b, _ := json.Marshal(txCursor{CreatedAt: createdAt, ID: id})
	return base64.StdEncoding.EncodeToString(b)
}

func decodeTxCursor(token string) (txCursor, error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return txCursor{}, err
	}
	var cur txCursor
	return cur, json.Unmarshal(b, &cur)
}
