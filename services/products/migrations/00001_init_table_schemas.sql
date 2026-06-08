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
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(30)   NOT NULL UNIQUE
                              CHECK (code ~ '^[a-z_]{1,30}$'),
    name        VARCHAR(100)  NOT NULL
                              CHECK (length(name) BETWEEN 1 AND 100),
    active      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_categories_touch
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- Товары каталога.
CREATE TABLE products (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(500)  NOT NULL
                                CHECK (length(name) BETWEEN 1 AND 500),
    description   TEXT          NOT NULL
                                CHECK (length(description) BETWEEN 1 AND 5000),
    price_points  BIGINT        NOT NULL
                                CHECK (price_points > 0),
    category_id   UUID          NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    active        BOOLEAN       NOT NULL DEFAULT TRUE,
    version       INT           NOT NULL DEFAULT 1
                                CHECK (version >= 1),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
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

-- Фотографии товара: нормализованная связь один-ко-многим.
CREATE TABLE product_photos (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id  UUID          NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    photo_key   VARCHAR(100)  NOT NULL
                              CHECK (photo_key ~ '^products/[0-9a-fA-F-]{36}\.(jpg|jpeg|png|webp)$'),
    position    INT           NOT NULL DEFAULT 0
                              CHECK (position >= 0),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (product_id, position)
);

-- Выборка фото товара в порядке отображения (обложка = position 0).
CREATE INDEX product_photos_product_idx
    ON product_photos (product_id, position);

-- Seed: базовые категории.
INSERT INTO categories (code, name) VALUES
    ('clothing',    'Одежда'),
    ('accessories', 'Аксессуары');

-- +goose Down

DROP TABLE IF EXISTS product_photos;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;

DROP FUNCTION IF EXISTS touch_updated_at();
