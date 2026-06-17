package e2e

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	trackMu         sync.Mutex
	createdUsers    []string
	createdProducts []string
)

func trackUser(id string)    { track(&createdUsers, id) }
func trackProduct(id string) { track(&createdProducts, id) }

func track(dst *[]string, id string) {
	if id == "" {
		return
	}
	trackMu.Lock()
	*dst = append(*dst, id)
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

func cleanup(userDB, inventoryDB *sqlx.DB) error {
	trackMu.Lock()
	users := append([]string(nil), createdUsers...)
	products := append([]string(nil), createdProducts...)
	trackMu.Unlock()

	var errs []error

	if len(products) > 0 {
		if _, err := inventoryDB.Exec(`DELETE FROM stock_adjustments WHERE product_id = ANY($1)`, pq.Array(products)); err != nil {
			errs = append(errs, fmt.Errorf("delete stock_adjustments: %w", err))
		}
		if _, err := inventoryDB.Exec(`DELETE FROM stock WHERE product_id = ANY($1)`, pq.Array(products)); err != nil {
			errs = append(errs, fmt.Errorf("delete stock: %w", err))
		}
	}
	if len(users) > 0 {
		_, _ = userDB.Exec(`DELETE FROM points_transactions WHERE user_id = ANY($1)`, pq.Array(users))
		_, _ = userDB.Exec(`DELETE FROM points_balance WHERE user_id = ANY($1)`, pq.Array(users))
		if _, err := userDB.Exec(`DELETE FROM users WHERE id = ANY($1)`, pq.Array(users)); err != nil {
			errs = append(errs, fmt.Errorf("delete users: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
