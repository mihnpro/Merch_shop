-- +goose Up
ALTER TABLE orders ADD COLUMN note TEXT
    CHECK (note IS NULL OR length(note) <= 1000);

-- +goose Down
ALTER TABLE orders DROP COLUMN IF EXISTS note;
