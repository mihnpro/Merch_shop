-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE carts (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cart_items (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id             UUID         NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id          UUID         NOT NULL,
    quantity            INT          NOT NULL CHECK (quantity > 0),
    price_at_add        INT          NOT NULL CHECK (price_at_add > 0),
    product_name_at_add VARCHAR(255) NOT NULL,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT cart_items_unique UNIQUE (cart_id, product_id)
);

CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION touch_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER trg_carts_updated_at
    BEFORE UPDATE ON carts
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_cart_items_updated_at
    BEFORE UPDATE ON cart_items
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- +goose Down

DROP TRIGGER IF EXISTS trg_cart_items_updated_at ON cart_items;
DROP TRIGGER IF EXISTS trg_carts_updated_at ON carts;
DROP FUNCTION IF EXISTS touch_updated_at();
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;
