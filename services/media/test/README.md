# media_e2e — E2E тесты media-сервиса (Go)

Чёрный-ящик e2e-набор для media-сервиса. Тесты бьют по REST через **gateway**
(`http://localhost:8080`) и покрывают единственный эндпоинт загрузки фото
(`POST /api/v1/admin/media/photos`, multipart, только admin): авторизацию,
валидацию (content-type, пустой файл, отсутствие поля `file`) и успешную загрузку.

Отдельный Go-модуль (`github.com/mihnpro/Merch_shop/services/media/test`). media
хранит файлы в MinIO и не использует SQL. Эндпоинт под `admin`, поэтому токен/
админ берутся из user-сервиса (общий `JWT_ACCESS_SECRET`). Загрузка — прямая
(multipart → media → MinIO), не presigned.

## Предусловия

Стек поднимается **вручную** (Docker): `gateway`, `user`, `media` (+ **MinIO**,
от которого media зависит) и user-Postgres на `localhost:5433` (`merch_users`)
для сидинга/очистки админа. `task up`.

## Этапы

1. **Setup** (`TestMain`): readiness (login→401 и POST /admin/media/photos→401),
   подключение к user-БД, seed админа.
2. **Tests** — `upload_test.go` (auth + загрузка/валидация).
3. **Teardown** — удаляет созданных пользователей (user-БД). Загруженные объекты
   в MinIO не чистятся (мелкие тестовые картинки, bucket public-read).

## Запуск

```bash
cd services/media/test
GOFLAGS=-mod=mod GOPROXY=off GOTOOLCHAIN=go1.25.10 \
  POSTGRES_PASSWORD=changeme \
  go test ./... -v -count=1
```

Или из корня: `task media:e2e:go`.

## Конфигурация (env, значения по умолчанию)

| Переменная          | Default                 |
|---------------------|-------------------------|
| `GW`                | `http://localhost:8080` |
| `POSTGRES_USER`     | `postgres`              |
| `POSTGRES_PASSWORD` | `changeme`              |
| `USER_POSTGRES_PORT`| `5433` (`merch_users`)  |
