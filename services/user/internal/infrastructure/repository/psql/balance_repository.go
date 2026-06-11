package psql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
)

type balanceRepository struct {
	db *sqlx.DB
}

func NewBalanceRepository(db *sqlx.DB) repository.BalanceRepository {
	return &balanceRepository{db: db}
}

func (r *balanceRepository) withTx(ctx context.Context, fn func(tx *sqlx.Tx) error) (err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *balanceRepository) EnsureBalance(ctx context.Context, userID uuid.UUID) error {
	return ensureBalance(ctx, r.db, userID)
}

func (r *balanceRepository) GetBalance(ctx context.Context, userID uuid.UUID) (model.PointsBalance, error) {
	if err := r.EnsureBalance(ctx, userID); err != nil {
		return model.PointsBalance{}, err
	}
	row, err := getBalanceRow(ctx, r.db, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PointsBalance{UserID: userID}, nil
		}
		return model.PointsBalance{}, err
	}
	return row.toModel(), nil
}

func (r *balanceRepository) Apply(ctx context.Context, cmd repository.ApplyPointsCommand, mutate func(*model.PointsBalance) error) (model.PointsBalance, error) {
	if err := r.EnsureBalance(ctx, cmd.UserID); err != nil {
		return model.PointsBalance{}, err
	}

	var (
		result    model.PointsBalance
		duplicate bool
	)
	err := r.withTx(ctx, func(tx *sqlx.Tx) error {
		row, err := lockBalanceRow(ctx, tx, cmd.UserID)
		if err != nil {
			return err
		}

		bal := row.toModel()
		before := bal.Points
		if err := mutate(&bal); err != nil {
			return err
		}
		delta := bal.Points - before

		inserted, err := insertPointsTx(ctx, tx, cmd, delta)
		if err != nil {
			return err
		}
		if !inserted {
			duplicate = true
			return nil
		}

		updated, err := updateBalanceRow(ctx, tx, cmd.UserID, delta)
		if err != nil {
			return err
		}
		result = updated.toModel()
		return nil
	})
	if err != nil {
		return model.PointsBalance{}, err
	}

	if duplicate {
		return r.GetBalance(ctx, cmd.UserID)
	}
	return result, nil
}

func (r *balanceRepository) GetTransactions(ctx context.Context, userID uuid.UUID, limit int, cursor string) ([]model.PointsTransaction, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	rows, err := selectTransactions(ctx, r.db, userID, limit, cursor)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(rows) > limit {
		rows = rows[:limit]
		last := rows[limit-1]
		nextToken = encodeTxCursor(last.CreatedAt, last.ID)
	}

	var txs []model.PointsTransaction
	for _, row := range rows {
		txs = append(txs, row.toModel())
	}
	return txs, nextToken, nil
}
