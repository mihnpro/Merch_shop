-- +goose Up

ALTER TABLE orders ADD COLUMN cancel_reason TEXT
    CHECK (cancel_reason IS NULL OR length(cancel_reason) <= 1000);

-- +goose Down

ALTER TABLE orders DROP COLUMN IF EXISTS cancel_reason;
