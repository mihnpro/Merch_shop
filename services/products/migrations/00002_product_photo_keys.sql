-- +goose Up

ALTER TABLE products
    ADD COLUMN photo_keys TEXT[] NOT NULL DEFAULT '{}'
        CHECK (array_position(photo_keys, NULL) IS NULL);

UPDATE products
    SET photo_keys = ARRAY[photo_key]
    WHERE photo_key IS NOT NULL AND photo_key <> '';

ALTER TABLE products DROP COLUMN photo_key;

-- +goose Down

ALTER TABLE products
    ADD COLUMN photo_key TEXT
        CHECK (photo_key IS NULL
               OR photo_key ~ '^products/[0-9a-fA-F-]{36}\.(jpg|jpeg|png|webp)$');

UPDATE products
    SET photo_key = photo_keys[1]
    WHERE array_length(photo_keys, 1) >= 1;

ALTER TABLE products DROP COLUMN photo_keys;
