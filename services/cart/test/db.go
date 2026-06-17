package e2e

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	trackMu           sync.Mutex
	createdUsers      []string
	createdProducts   []string
	createdCategories []string
)

func trackUser(id string)     { track(&createdUsers, id) }
func trackProduct(id string)  { track(&createdProducts, id) }
func trackCategory(id string) { track(&createdCategories, id) }

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

func cleanup(dbs map[string]*sqlx.DB) error {
	trackMu.Lock()
	users := append([]string(nil), createdUsers...)
	products := append([]string(nil), createdProducts...)
	categories := append([]string(nil), createdCategories...)
	trackMu.Unlock()

	var errs []error
	exec := func(db *sqlx.DB, ids []string, query string) {
		if db == nil || len(ids) == 0 {
			return
		}
		if _, err := db.Exec(query, pq.Array(ids)); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", query, err))
		}
	}
	exec(dbs["cart"], users, `DELETE FROM carts WHERE user_id = ANY($1)`)

	if inv := dbs["inventory"]; inv != nil {
		exec(inv, products, `DELETE FROM stock_adjustments WHERE product_id = ANY($1)`)
		exec(inv, products, `DELETE FROM stock WHERE product_id = ANY($1)`)
	}
	if prod := dbs["products"]; prod != nil {
		exec(prod, products, `DELETE FROM product_photos WHERE product_id = ANY($1)`)
		exec(prod, products, `DELETE FROM products WHERE id = ANY($1)`)
		exec(prod, categories, `DELETE FROM categories WHERE id = ANY($1)`)
	}
	if user := dbs["user"]; user != nil {
		exec(user, users, `DELETE FROM points_transactions WHERE user_id = ANY($1)`)
		exec(user, users, `DELETE FROM points_balance WHERE user_id = ANY($1)`)
		exec(user, users, `DELETE FROM users WHERE id = ANY($1)`)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
