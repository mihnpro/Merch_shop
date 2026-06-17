# products_e2e — E2E тесты products-сервиса (Go)

Чёрный-ящик e2e-набор для `products`-сервиса. Тесты бьют по REST через **gateway**
(`http://localhost:8080`, путь client → gateway → product gRPC/HTTP) и покрывают
каталог: категории, товары, авторизацию.

Отдельный Go-модуль (`github.com/mihnpro/Merch_shop/services/products/test`). Все
эндпоинты products требуют JWT (чтение — любой авторизованный, запись — только
admin), поэтому токен получается через **user**-сервис (`/api/v1/auth/*`), а
admin засевается промоутом в user-БД (тот же `JWT_ACCESS_SECRET` на весь стек).

## Предусловия

Стек поднимается **вручную** (Docker). Нужны: `gateway` (`:8080`), `user`,
`product`, и обе Postgres-БД, доступные с хоста:

- user-БД на `localhost:5433` (db `merch_users`) — сидинг/очистка админа;
- products-БД на `localhost:5434` (db `merch_service`) — очистка созданных товаров/категорий.

`task up` (или как минимум `user`, `products`, `gateway`).

## Этапы

1. **Setup** (`TestMain`): `waitForReady` (login → 401 и GET /products → 401),
   подключение к обеим БД, `seedAdmin` (register через REST + промоут в user-БД).
2. **Tests** — `auth_test.go`, `categories_test.go`, `products_test.go`.
3. **Teardown** — удаляет созданные товары/категории (products-БД) и
   пользователей (user-БД); seed-категории `clothing`/`accessories` не трогаются.

## Запуск

```bash
cd services/products/test
GOFLAGS=-mod=mod GOPROXY=off GOTOOLCHAIN=go1.25.10 \
  POSTGRES_PASSWORD=changeme \
  go test ./... -v -count=1
```

Или из корня репозитория: `task products:e2e:go`.

## Конфигурация (env, со значениями по умолчанию)

| Переменная               | Default                 | Назначение                  |
|--------------------------|-------------------------|-----------------------------|
| `GW`                     | `http://localhost:8080` | базовый URL gateway         |
| `POSTGRES_USER`          | `postgres`              | пользователь обеих БД        |
| `POSTGRES_PASSWORD`      | `changeme`              | пароль обеих БД              |
| `USER_POSTGRES_PORT`     | `5433`                  | порт user-БД                 |
| `USER_POSTGRES_DB`       | `merch_users`           | имя user-БД                  |
| `PRODUCTS_POSTGRES_PORT` | `5434`                  | порт products-БД             |
| `PRODUCTS_POSTGRES_DB`   | `merch_service`         | имя products-БД              |
