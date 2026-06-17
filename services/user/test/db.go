package e2e

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	createdMu    sync.Mutex
	createdUsers []string
)

func trackUser(id string) {
	if id == "" {
		return
	}
	createdMu.Lock()
	createdUsers = append(createdUsers, id)
	createdMu.Unlock()
}

func connectDB(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.dsn())
	if err != nil {
		return nil, fmt.Errorf(
			"connect postgres at %s:%s (is the stack up? expected host port 5433): %w",
			cfg.PGHost, cfg.PGPort, err,
		)
	}
	return db, nil
}

func promoteToAdmin(db *sqlx.DB, login string) error {
	res, err := db.Exec(`UPDATE users SET role = 'admin' WHERE lower(login) = lower($1)`, login)
	if err != nil {
		return fmt.Errorf("promote %q to admin: %w", login, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("promote %q to admin: no row updated", login)
	}
	return nil
}

func cleanup(db *sqlx.DB) error {
	createdMu.Lock()
	ids := append([]string(nil), createdUsers...)
	createdMu.Unlock()
	if len(ids) == 0 {
		return nil
	}

	if _, err := db.Exec(`DELETE FROM points_transactions WHERE user_id = ANY($1)`, pq.Array(ids)); err != nil {
		return fmt.Errorf("delete transactions: %w", err)
	}
	if _, err := db.Exec(`DELETE FROM points_balance WHERE user_id = ANY($1)`, pq.Array(ids)); err != nil {
		return fmt.Errorf("delete balances: %w", err)
	}
	if _, err := db.Exec(`DELETE FROM users WHERE id = ANY($1)`, pq.Array(ids)); err != nil {
		return fmt.Errorf("delete users: %w", err)
	}
	return nil
}
