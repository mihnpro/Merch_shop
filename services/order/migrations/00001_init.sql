-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION touch_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TYPE order_status AS ENUM ('pending', 'confirmed', 'ready_to_pickup', 'delivered', 'cancelled');

CREATE TABLE orders (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID         NOT NULL,
    total_points     BIGINT       NOT NULL CHECK (total_points > 0),
    status           order_status NOT NULL DEFAULT 'pending',
    note             TEXT         CHECK (note IS NULL OR length(note) <= 1000),
    cancel_reason    TEXT         CHECK (cancel_reason IS NULL OR length(cancel_reason) <= 1000),
    delivery_address TEXT         NOT NULL CHECK (length(delivery_address) > 0),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX orders_user_id_idx ON orders (user_id, created_at DESC);
CREATE INDEX orders_status_idx  ON orders (status);

CREATE TRIGGER trg_orders_touch
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TABLE order_items (
    id           UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID    NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id   UUID    NOT NULL,
    product_name TEXT    NOT NULL,
    quantity     INT     NOT NULL CHECK (quantity > 0),
    price_points BIGINT  NOT NULL CHECK (price_points > 0)
);

CREATE INDEX order_items_order_id_idx ON order_items (order_id);

CREATE TABLE outbox (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT        NOT NULL,
    payload    JSONB       NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at    TIMESTAMPTZ
);

CREATE INDEX outbox_unsent_idx ON outbox (created_at) WHERE sent_at IS NULL;

-- +goose Down

DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TYPE  IF EXISTS order_status;
DROP FUNCTION IF EXISTS touch_updated_at();
