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

CREATE TYPE user_status AS ENUM ('active', 'blocked');

CREATE TABLE users (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    login               TEXT         NOT NULL,
    password_hash       TEXT         NOT NULL,
    first_name          TEXT         NOT NULL
                                     CHECK (length(first_name) BETWEEN 1 AND 100),
    last_name           TEXT         NOT NULL
                                     CHECK (length(last_name) BETWEEN 1 AND 100),
    patronymic          TEXT         CHECK (patronymic IS NULL OR length(patronymic) BETWEEN 1 AND 100),
    email               TEXT         CHECK (email IS NULL OR email ~* '^[^@]+@[^@]+\.[^@]+$'),
    phone_number        TEXT         CHECK (phone_number IS NULL OR phone_number ~ '^\+?[1-9][0-9]{6,14}$'),
    role                TEXT         NOT NULL DEFAULT 'user'
                                     CHECK (role IN ('user', 'admin')),
    status              user_status  NOT NULL DEFAULT 'active',
    failed_login_count  INT          NOT NULL DEFAULT 0
                                     CHECK (failed_login_count >= 0),
    locked_until        TIMESTAMPTZ,
    last_login_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX users_login_lower_uk ON users (lower(login));

CREATE INDEX users_status_blocked_idx ON users (status) WHERE status = 'blocked';

CREATE TRIGGER trg_users_touch
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TABLE points_balance (
    user_id     UUID         PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
    points      BIGINT       NOT NULL DEFAULT 0
                             CHECK (points >= 0),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_points_balance_touch
    BEFORE UPDATE ON points_balance
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE TABLE points_transactions (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    operation_id  UUID         NOT NULL UNIQUE,
    order_id      UUID,
    amount        BIGINT       NOT NULL CHECK (amount <> 0),
    reason        TEXT         NOT NULL
                               CHECK (length(reason) BETWEEN 1 AND 500),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);


CREATE INDEX points_tx_user_created_idx
    ON points_transactions (user_id, created_at DESC);

-- +goose Down

DROP TABLE IF EXISTS points_transactions;
DROP TABLE IF EXISTS points_balance;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_status;

DROP FUNCTION IF EXISTS touch_updated_at();
