package e2e

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	trackMu      sync.Mutex
	createdUsers []string
)

func trackUser(id string) {
	if id == "" {
		return
	}
	trackMu.Lock()
	createdUsers = append(createdUsers, id)
	trackMu.Unlock()
}

func connect(cfg pgConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.dsn())
	if err != nil {
		return nil, fmt.Errorf("connect postgres at %s:%s/%s (is the stack up?): %w", cfg.Host, cfg.Port, cfg.Database, err)
	}
	return db, nil
}

func promoteToAdmin(userDB *sqlx.DB, login string) error {
	res, err := userDB.Exec(`UPDATE users SET role = 'admin' WHERE lower(login) = lower($1)`, login)
	if err != nil {
		return fmt.Errorf("promote %q to admin: %w", login, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("promote %q to admin: no row updated", login)
	}
	return nil
}

func cleanup(userDB *sqlx.DB) error {
	trackMu.Lock()
	users := append([]string(nil), createdUsers...)
	trackMu.Unlock()
	if len(users) == 0 {
		return nil
	}
	_, _ = userDB.Exec(`DELETE FROM points_transactions WHERE user_id = ANY($1)`, pq.Array(users))
	_, _ = userDB.Exec(`DELETE FROM points_balance WHERE user_id = ANY($1)`, pq.Array(users))
	if _, err := userDB.Exec(`DELETE FROM users WHERE id = ANY($1)`, pq.Array(users)); err != nil {
		return fmt.Errorf("delete users: %w", err)
	}
	return nil
}
