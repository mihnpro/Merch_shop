# order_e2e — E2E тесты order-сервиса (Go)

Чёрный-ящик e2e-набор для order-сервиса. Тесты бьют по REST через **gateway**
(`http://localhost:8080`) и проверяют полный жизненный цикл заказа: создание
(сага через cart/products/inventory/user), просмотр, переходы статусов, отмену с
возвратом баллов, админ-листинг и аналитику.

Отдельный Go-модуль (`github.com/mihnpro/Merch_shop/services/order/test`). Создание
заказа — это распределённая сага, поэтому e2e перед `POST /orders` сам готовит
предусловия через gateway:

1. admin создаёт категорию и товар (products) с известной ценой;
2. admin пополняет склад товара (inventory adjust);
3. регистрируется обычный юзер, ему начисляются баллы (user grant-points);
4. юзер кладёт товар в корзину (`POST /cart/items` — cart дёргает products+inventory);
5. юзер создаёт заказ (`POST /orders` — order читает корзину, списывает баллы, чистит корзину).

Токен/админ — через user-сервис (общий `JWT_ACCESS_SECRET`).

## Предусловия

Стек поднимается **вручную** (Docker): `gateway`, `user`, `products`, `inventory`,
`cart`, `order` (+ их Postgres, Redis, Kafka). Postgres-БД доступны с хоста:

- user `:5433` (`merch_users`), products `:5434` (`merch_service`),
  inventory `:5435` (`merch_service`), order `:5437` (`merch_order`).

`task up`.

## Этапы

1. **Setup** (`TestMain`): readiness (login→401, GET /orders→401), подключение к 4 БД, seed админа.
2. **Tests** — `auth_test.go`, `order_lifecycle_test.go`, `errors_test.go`.
3. **Teardown** — удаляет созданные orders/order_items/outbox (order-БД),
   products/categories (products-БД), stock (inventory-БД), users (user-БД).
   Корзина очищается самим заказом.

## Запуск

```bash
cd services/order/test
GOFLAGS=-mod=mod GOPROXY=off GOTOOLCHAIN=go1.25.10 \
  POSTGRES_PASSWORD=changeme \
  go test ./... -v -count=1
```

Или из корня: `task order:e2e:go`.

## Конфигурация (env, значения по умолчанию)

| Переменная                | Default                 |
|---------------------------|-------------------------|
| `GW`                      | `http://localhost:8080` |
| `POSTGRES_USER`           | `postgres`              |
| `POSTGRES_PASSWORD`       | `changeme`              |
| `USER_POSTGRES_PORT`      | `5433` (`merch_users`)  |
| `PRODUCTS_POSTGRES_PORT`  | `5434` (`merch_service`)|
| `INVENTORY_POSTGRES_PORT` | `5435` (`merch_service`)|
| `ORDER_POSTGRES_PORT`     | `5437` (`merch_order`)  |

> Пароли Postgres в стеке различаются по сервисам: order-БД использует `secret`
> (дефолт `ORDER_POSTGRES_PASSWORD`), остальные — `changeme` (`POSTGRES_PASSWORD`).
> Переопределяются через `*_POSTGRES_PASSWORD`.
