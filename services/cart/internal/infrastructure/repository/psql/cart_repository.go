package psql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/cart/internal/domain/valueobject"
)

type cartRepository struct {
	db *sqlx.DB
}

func NewCartRepository(db *sqlx.DB) repository.CartRepository {
	return &cartRepository{db: db}
}


type cartRow struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *cartRow) toModel() *model.Cart {
	return &model.Cart{
		ID:        r.ID,
		UserID:    r.UserID,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

type cartItemRow struct {
	ID               uuid.UUID `db:"id"`
	CartID           uuid.UUID `db:"cart_id"`
	ProductID        uuid.UUID `db:"product_id"`
	Quantity         int       `db:"quantity"`
	PriceAtAdd       int       `db:"price_at_add"`
	ProductNameAtAdd string    `db:"product_name_at_add"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

func (r *cartItemRow) toModel() *model.CartItem {
	return &model.CartItem{
		ID:               r.ID,
		CartID:           r.CartID,
		ProductID:        r.ProductID,
		Quantity:         vo.NewQuantityFromStored(r.Quantity),
		PriceAtAdd:       r.PriceAtAdd,
		ProductNameAtAdd: r.ProductNameAtAdd,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}


const upsertCart = `
	INSERT INTO carts (user_id)
	VALUES ($1)
	ON CONFLICT (user_id) DO UPDATE SET updated_at = carts.updated_at
	RETURNING id, user_id, created_at, updated_at`

const selectCartByUserID = `
	SELECT id, user_id, created_at, updated_at
	FROM carts WHERE user_id = $1`

const selectCartItems = `
	SELECT id, cart_id, product_id, quantity, price_at_add, product_name_at_add, created_at, updated_at
	FROM cart_items WHERE cart_id = $1
	ORDER BY created_at`

const selectItemByProduct = `
	SELECT id, cart_id, product_id, quantity, price_at_add, product_name_at_add, created_at, updated_at
	FROM cart_items WHERE cart_id = $1 AND product_id = $2`

const selectItemByID = `
	SELECT id, cart_id, product_id, quantity, price_at_add, product_name_at_add, created_at, updated_at
	FROM cart_items WHERE id = $1`

const insertItem = `
	INSERT INTO cart_items (id, cart_id, product_id, quantity, price_at_add, product_name_at_add)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, cart_id, product_id, quantity, price_at_add, product_name_at_add, created_at, updated_at`

const updateItemQty = `
	UPDATE cart_items SET quantity = $2, updated_at = NOW() WHERE id = $1`

const deleteItem = `
	DELETE FROM cart_items WHERE id = $1 AND cart_id = $2`

const clearCart = `
	DELETE FROM cart_items WHERE cart_id = $1`

const touchCart = `
	UPDATE carts SET updated_at = NOW() WHERE id = $1`


func (r *cartRepository) GetOrCreateCart(ctx context.Context, userID uuid.UUID) (*model.Cart, error) {
	var row cartRow
	if err := r.db.QueryRowxContext(ctx, upsertCart, userID).StructScan(&row); err != nil {
		return nil, err
	}
	return row.toModel(), nil
}

func (r *cartRepository) GetCartByUserID(ctx context.Context, userID uuid.UUID) (*model.Cart, error) {
	var row cartRow
	if err := r.db.GetContext(ctx, &row, selectCartByUserID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCartNotFound
		}
		return nil, err
	}
	return row.toModel(), nil
}

func (r *cartRepository) GetCartItems(ctx context.Context, cartID uuid.UUID) ([]*model.CartItem, error) {
	var rows []cartItemRow
	if err := r.db.SelectContext(ctx, &rows, selectCartItems, cartID); err != nil {
		return nil, err
	}
	items := make([]*model.CartItem, 0, len(rows))
	for i := range rows {
		items = append(items, rows[i].toModel())
	}
	return items, nil
}

func (r *cartRepository) GetItemByProduct(ctx context.Context, cartID, productID uuid.UUID) (*model.CartItem, error) {
	var row cartItemRow
	if err := r.db.GetContext(ctx, &row, selectItemByProduct, cartID, productID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrItemNotFound
		}
		return nil, err
	}
	return row.toModel(), nil
}

func (r *cartRepository) GetItemByID(ctx context.Context, itemID uuid.UUID) (*model.CartItem, error) {
	var row cartItemRow
	if err := r.db.GetContext(ctx, &row, selectItemByID, itemID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrItemNotFound
		}
		return nil, err
	}
	return row.toModel(), nil
}

func (r *cartRepository) InsertItem(ctx context.Context, item *model.CartItem) (*model.CartItem, error) {
	var row cartItemRow
	err := r.db.QueryRowxContext(ctx, insertItem,
		item.ID, item.CartID, item.ProductID,
		item.Quantity.Int(), item.PriceAtAdd, item.ProductNameAtAdd,
	).StructScan(&row)
	if err != nil {
		return nil, err
	}
	return row.toModel(), nil
}

func (r *cartRepository) UpdateItemQty(ctx context.Context, itemID uuid.UUID, qty int) error {
	_, err := r.db.ExecContext(ctx, updateItemQty, itemID, qty)
	return err
}

func (r *cartRepository) DeleteItem(ctx context.Context, itemID, cartID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, deleteItem, itemID, cartID)
	return err
}

func (r *cartRepository) ClearCart(ctx context.Context, cartID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, clearCart, cartID)
	return err
}

func (r *cartRepository) TouchCart(ctx context.Context, cartID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, touchCart, cartID)
	return err
}
