# cart_e2e — E2E тесты cart-сервиса (Go)

Чёрный-ящик e2e-набор для cart-сервиса. Тесты бьют по REST через **gateway**
(`http://localhost:8080`) и покрывают корзину: добавление (с учётом products +
inventory), мерж, обновление, удаление, очистку, ошибки (нет на складе, товар
неактивен, товар не найден), авторизацию.

Отдельный Go-модуль (`github.com/mihnpro/Merch_shop/services/cart/test`). Добавление
в корзину дёргает products (цена/имя) и inventory (наличие), поэтому e2e сначала
готовит товар через gateway:

1. admin создаёт категорию и товар (products);
2. admin пополняет склад товара (inventory adjust);
3. юзер кладёт товар в корзину (`POST /cart/items`).

Токен/админ — через user-сервис (общий `JWT_ACCESS_SECRET`). Баллы не нужны.

Тонкость: cart проверяет наличие **до** товара, поэтому для кейса «товар не
найден → 404» e2e заводит сток на случайный product_id (без создания товара) —
inventory проходит, а products возвращает NotFound.

## Предусловия

Стек поднимается **вручную** (Docker): `gateway`, `user`, `products`, `inventory`,
`cart` (+ Postgres/Redis). Postgres-БД доступны с хоста:

- user `:5433` (`merch_users`), products `:5434` (`merch_service`),
  inventory `:5435` (`merch_service`), cart `:5436` (`merch_cart`).

`task up`.

## Этапы

1. **Setup** (`TestMain`): readiness (login→401, GET /cart→401), подключение к 4 БД, seed админа.
2. **Tests** — `auth_test.go`, `cart_test.go`, `errors_test.go`.
3. **Teardown** — удаляет корзины (cart-БД, cart_items по cascade),
   products/categories (products-БД), stock (inventory-БД), users (user-БД).

## Запуск

```bash
cd services/cart/test
GOFLAGS=-mod=mod GOPROXY=off GOTOOLCHAIN=go1.25.10 \
  POSTGRES_PASSWORD=changeme \
  go test ./... -v -count=1
```

Или из корня: `task cart:e2e:go`.

## Конфигурация (env, значения по умолчанию)

| Переменная                | Default                  |
|---------------------------|--------------------------|
| `GW`                      | `http://localhost:8080`  |
| `POSTGRES_USER`           | `postgres`               |
| `POSTGRES_PASSWORD`       | `changeme`               |
| `USER_POSTGRES_PORT`      | `5433` (`merch_users`)   |
| `PRODUCTS_POSTGRES_PORT`  | `5434` (`merch_service`) |
| `INVENTORY_POSTGRES_PORT` | `5435` (`merch_service`) |
| `CART_POSTGRES_PORT`      | `5436` (`merch_cart`)    |
