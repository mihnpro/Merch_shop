# inventory_e2e — E2E тесты inventory-сервиса (Go)

Чёрный-ящик e2e-набор для inventory-сервиса (каталог сервиса — `services/invetory`,
исторический тайпо). Тесты бьют по REST через **gateway** (`http://localhost:8080`)
и покрывают остатки: корректировки склада (adjust), чтение остатков, авторизацию.

Отдельный Go-модуль (`github.com/mihnpro/Merch_shop/services/invetory/test`). Все
эндпоинты требуют JWT (чтение `/stock` — любой авторизованный, `/admin/inventory*`
— admin), поэтому токен берётся через **user**-сервис, а admin засевается промоутом
в user-БД (общий `JWT_ACCESS_SECRET` на весь стек). inventory не имеет FK на
products, поэтому товарами в тестах выступают случайные UUID.

## Предусловия

Стек поднимается **вручную** (Docker): `gateway` (`:8080`), `user`, `inventory`,
обе Postgres-БД доступны с хоста:

- user-БД на `localhost:5433` (db `merch_users`) — сидинг/очистка админа;
- inventory-БД на `localhost:5435` (db `merch_service`) — очистка созданных остатков.

`task up` (или как минимум `user`, `inventory`, `gateway`).

## Этапы

1. **Setup** (`TestMain`): `waitForReady` (login → 401 и GET /stock → 401),
   подключение к обеим БД, `seedAdmin`.
2. **Tests** — `auth_test.go`, `adjust_test.go`, `stock_test.go`.
3. **Teardown** — удаляет созданные `stock_adjustments`/`stock` (inventory-БД) и
   пользователей (user-БД).

## Запуск

```bash
cd services/invetory/test
GOFLAGS=-mod=mod GOPROXY=off GOTOOLCHAIN=go1.25.10 \
  POSTGRES_PASSWORD=changeme \
  go test ./... -v -count=1
```

Или из корня: `task inventory:e2e:go`.

## Конфигурация (env, значения по умолчанию)

| Переменная                | Default                 | Назначение            |
|---------------------------|-------------------------|-----------------------|
| `GW`                      | `http://localhost:8080` | базовый URL gateway   |
| `POSTGRES_USER`           | `postgres`              | пользователь БД        |
| `POSTGRES_PASSWORD`       | `changeme`              | пароль БД              |
| `USER_POSTGRES_PORT`      | `5433`                  | порт user-БД           |
| `USER_POSTGRES_DB`        | `merch_users`           | имя user-БД            |
| `INVENTORY_POSTGRES_PORT` | `5435`                  | порт inventory-БД      |
| `INVENTORY_POSTGRES_DB`   | `merch_service`         | имя inventory-БД       |
