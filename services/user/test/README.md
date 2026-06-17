# user_e2e — E2E тесты user-сервиса (Go)

Чёрный-ящик e2e-набор для `user`-сервиса. Тесты бьют по REST через **gateway**
(`http://localhost:8080`, путь client → gateway → user gRPC) и проверяют все
сценарии: auth, профиль (`/me`), admin.

Отдельный Go-модуль (`github.com/mihnpro/Merch_shop/services/user/test`), не
импортирует код сервиса — общается только по HTTP и напрямую с Postgres (для
сидинга админа и очистки).

## Предусловия

Стек поднимается **вручную** (Docker). Нужны:

- `gateway` (`:8080`) + `user` (gRPC) + `postgres` + `redis` — например `task up`
  или как минимум `task user:up && task gateway:up`;
- Postgres доступен с хоста на `localhost:5433` (порт из `services/user/docker-compose.yml`).

## Этапы

1. **Setup** (`TestMain` в `main_test.go`):
   - `waitForReady` — поллит `POST /api/v1/auth/login` пока стек не ответит (401);
   - `connectDB` — подключение к Postgres;
   - `seedAdmin` — регистрирует юзера через REST и промоутит до `admin` SQL-ом.
2. **Tests** — `auth_test.go`, `me_test.go`, `admin_test.go`.
3. **Teardown** — удаляет всех созданных за прогон юзеров (children-first из-за
   FK `ON DELETE RESTRICT`).

## Запуск

```bash
cd services/user/test
GOFLAGS=-mod=mod GOPROXY=off GOTOOLCHAIN=go1.25.10 \
  POSTGRES_PASSWORD=changeme \
  go test ./... -v -count=1
```

Или через Taskfile из корня репозитория:

```bash
task user:e2e:go
```

## Конфигурация (env, со значениями по умолчанию)

| Переменная          | Default                  | Назначение                          |
|---------------------|--------------------------|-------------------------------------|
| `GW`                | `http://localhost:8080`  | базовый URL gateway                 |
| `POSTGRES_HOST`     | `localhost`              | хост Postgres для сидинга/очистки    |
| `POSTGRES_PORT`     | `5433`                   | порт Postgres (публикуется compose)  |
| `POSTGRES_USER`     | `postgres`               | пользователь БД                      |
| `POSTGRES_PASSWORD` | `changeme`               | пароль БД                            |
| `POSTGRES_DB`       | `merch_users`            | имя БД                               |
| `POSTGRES_SSLMODE`  | `disable`                | sslmode                              |

> Хост клиента и API — всегда `localhost` (HttpOnly `SameSite=Lax` cookie
> возвращаются только при совпадении хоста).
