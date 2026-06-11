-- +goose Up

-- Расширение для gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION touch_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Остатки по товару. Размеры в MVP не используются: ключ — только product_id.
CREATE TABLE stock (
    product_id  UUID          PRIMARY KEY,
    available   INT           NOT NULL DEFAULT 0
                              CHECK (available >= 0),
    version     INT           NOT NULL DEFAULT 1
                              CHECK (version >= 1),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_stock_touch
    BEFORE UPDATE ON stock
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- Резервы под заказ. order_id — ключ идемпотентности ReserveStock (см. TDR §17.7).
CREATE TABLE reservations (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id         UUID         NOT NULL UNIQUE,
    items            JSONB        NOT NULL
                                  CHECK (jsonb_array_length(items) >= 1),
    released_reason  TEXT                  CHECK (length(released_reason) <= 500),
    released_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_reservations_touch
    BEFORE UPDATE ON reservations
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- Активные резервы: отмена и дашборды (резерв «жив», пока released_at IS NULL).
CREATE INDEX reservations_active_idx
    ON reservations (created_at) WHERE released_at IS NULL;

-- Журнал корректировок остатков админом. operation_id — ключ идемпотентности AdjustStock.
CREATE TABLE stock_adjustments (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    operation_id  UUID         NOT NULL UNIQUE,
    product_id    UUID         NOT NULL,
    delta         INT          NOT NULL
                               CHECK (delta <> 0),
    reason        TEXT         NOT NULL
                               CHECK (length(reason) BETWEEN 1 AND 500),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- История корректировок по товару (аудит, «последние сверху»).
CREATE INDEX stock_adjustments_product_idx
    ON stock_adjustments (product_id, created_at DESC);

-- +goose Down

DROP TABLE IF EXISTS stock_adjustments;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS stock;

DROP FUNCTION IF EXISTS touch_updated_at();
