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

-- Справочник категорий товаров.
CREATE TABLE categories (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT         NOT NULL UNIQUE
                             CHECK (code ~ '^[a-z_]{1,30}$'),
    name        TEXT         NOT NULL
                             CHECK (length(name) BETWEEN 1 AND 100),
    active      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_categories_touch
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- Товары каталога.
CREATE TABLE products (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT         NOT NULL
                               CHECK (length(name) BETWEEN 1 AND 500),
    description   TEXT         NOT NULL
                               CHECK (length(description) BETWEEN 1 AND 5000),
    price_points  BIGINT       NOT NULL
                               CHECK (price_points > 0),
    category_id   UUID         NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    -- Необязательный набор размеров товара: массив кодов (XS, S, M, ...).
    -- Отдельного справочника размеров нет — осознанная денормализация.
    sizes         TEXT[]       NOT NULL DEFAULT '{}'
                               CHECK (array_position(sizes, NULL) IS NULL),
    photo_key     TEXT         CHECK (photo_key IS NULL
                                      OR photo_key ~ '^products/[0-9a-fA-F-]{36}\.(jpg|jpeg|png|webp)$'),
    active        BOOLEAN      NOT NULL DEFAULT TRUE,
    version       INT          NOT NULL DEFAULT 1
                               CHECK (version >= 1),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_products_touch
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- Каталог: листинг активных товаров по категории.
CREATE INDEX products_category_active_idx
    ON products (category_id) WHERE active;

-- Keyset-пагинация: стабильный порядок по времени создания.
CREATE INDEX products_created_idx
    ON products (created_at DESC, id DESC);

-- Seed: базовые категории.
INSERT INTO categories (code, name) VALUES
    ('clothing',    'Одежда'),
    ('accessories', 'Аксессуары');

-- +goose Down

DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;

DROP FUNCTION IF EXISTS touch_updated_at();
