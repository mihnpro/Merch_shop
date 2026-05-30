# Merch Store MVP — Архитектурная документация

| | |
|---|---|
| **Версия** | 1.1.0 MVP |
| **Стек** | Go · gRPC · Apache Kafka · Kubernetes · PostgreSQL · Redis · MinIO (S3) |
| **Тип проекта** | Обособленный магазин мерча с собственной авторизацией |
| **Архитектура** | Микросервисная (7 сервисов) |
| **Срок** | 1 месяц |

---

## Содержание

1. [Обзор и скоуп MVP](#1-обзор-и-скоуп-mvp)
2. [Сервисы и ответственности](#2-сервисы-и-ответственности)
3. [Полный сценарий покупки](#3-полный-сценарий-покупки)
4. [Админка — управление и начисление баллов](#4-админка--управление-и-начисление-баллов)
5. [Хранение фото — MinIO и Media Service](#5-хранение-фото--minio-и-media-service)
6. [gRPC API](#6-grpc-api)
7. [Kafka — событие order.created](#7-kafka--событие-ordercreated)
8. [Схема данных](#8-схема-данных)
9. [Структура репозитория](#9-структура-репозитория)
10. [Kubernetes — деплой](#10-kubernetes--деплой)
11. [Наблюдаемость](#11-наблюдаемость)
12. [План реализации по неделям](#12-план-реализации-по-неделям)
13. [Архитектурные решения (ADR)](#13-архитектурные-решения-adr)

---

## 1. Обзор и скоуп MVP

**Merch Store** — обособленная платформа, где пользователи могут заказывать мерч за внутренние баллы. Регистрация и вход — собственные, через логин и пароль. Баллы начисляет админ через админ-панель. Никаких интеграций с внешними системами.

### 1.1 Что входит в MVP

- Каталог товаров с фильтрами по категориям и размерам, с фото из S3
- Корзина с постоянным хранением в PostgreSQL (Redis-кеш с TTL 30 мин, авто-очистки по времени нет)
- Оформление заказа с проверкой остатков и баланса баллов
- Резервирование остатков (защита от двойной продажи)
- Регистрация и вход через логин и пароль, JWT-токены (access + refresh) с ролями `user` и `admin`
- Просмотр истории своих заказов
- **Админ-панель**: управление каталогом, загрузка фото в S3, начисление баллов пользователям, просмотр всех заказов

### 1.2 Что **не** входит в MVP

- Уведомления (Slack, email) — статус виден в личном кабинете
- Автоматический фулфилмент — выгрузка заказов в Excel
- Возвраты и сложная история транзакций
- Промокоды, скидки, акции

### 1.3 Ключевые архитектурные принципы

- **gRPC между сервисами** — синхронные вызовы там, где пользователь ждёт ответа
- **Kafka для асинхронности** — одно событие `order.created`
- **DB per service** — каждый сервис со своей PostgreSQL-схемой; Cart хранит корзины в PostgreSQL и использует Redis как write-through кеш
- **API Gateway как единая точка входа** для пользователей
- **Role-based access** — роль `admin` / `user` содержится в JWT; каждый сервис самостоятельно проверяет права на admin-методах через общий интерцептор из `pkg/auth`
- **S3 для бинарных файлов** — фото товаров хранятся в MinIO; админ грузит фото через Media Service (multipart/streaming), который сам пишет в MinIO. Браузер с MinIO напрямую не общается

---

## 2. Сервисы и ответственности

В MVP **7 сервисов**:

| # | Сервис | Что делает | Хранилище |
|---|---|---|---|
| 1 | **API Gateway** | Принимает запросы, проверяет JWT, маршрутизирует по роли; `/admin/*` доступен только с `role=admin` | — |
| 2 | **Product Service** | Каталог: список товаров, поиск, детали; admin-методы: создание, редактирование, деактивация | PostgreSQL |
| 3 | **Cart Service** | Корзина пользователя (постоянное хранение, кеш с TTL 30 мин) | PostgreSQL + Redis (cache) |
| 4 | **Order Service** | Создание заказа, история, статусы; admin-методы: просмотр всех заказов, экспорт, смена статуса | PostgreSQL |
| 5 | **User Service** | Профиль, баланс баллов; admin-методы: начисление баллов, управление пользователями | PostgreSQL |
| 6 | **Inventory Service** | Остатки товаров, резервирование; admin-метод: AdjustStock | PostgreSQL |
| 7 | **Media Service** | Приём загружаемых фото от админа (multipart/streaming), валидация формата и размера, запись в MinIO через PutObject; только для `role=admin` | — (MinIO SDK) |

Плюс инфраструктурный компонент: **MinIO** — S3-совместимое хранилище для фото товаров.

### 2.1 Кто кого вызывает

**Все потоки идут через единый Gateway**, который проверяет JWT и маршрутизирует запросы. Методы, помеченные как admin-only, доступны только при `role=admin` в токене.

| Источник | Цель | Метод | Зачем |
|---|---|---|---|
| Gateway | все сервисы | различные | Проксирование запросов |
| Cart | Product | `GetProduct` | Получить цену и название |
| Cart | Inventory | `CheckStock` | Проверить наличие при добавлении |
| Order | Cart | `GetCart`, `ClearCart` | Взять корзину и очистить |
| Order | User | `GetBalance`, `DeductPoints` | Проверить и списать баллы |
| Order | Kafka (producer) | `order.created` | Публикация события |
| Inventory | Kafka (consumer) | `order.created` | Резервирование остатков |
| Media | MinIO | `PutObject` (stream) | Запись загруженного фото в bucket (вызывается из Media.UploadPhoto, admin-only) |

---

## 3. Полный сценарий покупки

### Шаг 1. Авторизация
Пользователь регистрируется (или входит) через логин и пароль, получает JWT-токен с ролью `user`.

### Шаг 2. Просмотр каталога
```
Client → Gateway → Product.ListProducts(category="t-shirts")
       ← ListProductsResponse{items: [{..., photo_key: "products/abc.jpg"}]}

Браузер сам собирает URL фото:
       https://minio.merch.local/merch-products/products/abc.jpg
```

### Шаг 3. Добавление в корзину
```
Client → Gateway → Cart.AddItem(user_id, product_id, size, qty)

  внутри Cart Service:
    Cart → Inventory.CheckStock(product_id, size)
    Cart → Product.GetProduct(product_id)
    Cart в транзакции пишет позицию в Postgres (carts, cart_items)
    Cart обновляет JSON-снапшот в Redis: cart:{user_id}, TTL 30 мин
    (если Redis недоступен — операция всё равно успешна, инкремент метрики cache_write_failed)

← Cart{items: [...], total_points}
```

### Шаг 4. Оформление заказа
```
Client → Gateway → Order.CreateOrder(user_id, delivery_address)

  внутри Order Service:
    1. Order → Cart.GetCart(user_id)
    2. Order → User.GetBalance(user_id)
    3. Order → User.DeductPoints(user_id, amount, order_id)
    4. Order сохраняет заказ в БД (статус "pending")
    5. Order публикует в Kafka order.created
    6. Order → Cart.ClearCart(user_id)

← Order{id, status: "pending"}
```

### Шаг 5. Асинхронная обработка
```
Inventory ← Kafka.consume("order.created")
Inventory резервирует остатки
Inventory → Order.UpdateStatus(order_id, "confirmed")
```

### Обработка ошибок

| Сбой | Что делаем |
|---|---|
| `CheckStock` вернул 0 | `AddItem` возвращает `OUT_OF_STOCK`, баллы не списаны |
| Недостаточно баллов | `CreateOrder` возвращает `INSUFFICIENT_POINTS` |
| Inventory не смог резервировать | Меняет статус на `cancelled`, баллы возвращаются через `AddPoints` |
| Kafka недоступен | Outbox pattern — событие сохраняется в БД, отдельный воркер досылает |

---

## 4. Админка — управление и начисление баллов

### 4.1 Подход: роль в JWT

Права проверяются непосредственно в каждом доменном сервисе. При входе User Service выдаёт JWT с полем `role: "admin"` или `role: "user"`. Gateway проксирует все запросы, проверяя JWT; маршруты `/admin/*` отклоняются с 403 без токена с нужной ролью.

Каждый admin-метод на стороне gRPC-сервиса проверяет роль через общий интерцептор из `pkg/auth`:

```go
func AdminOnly(ctx context.Context) error {
    role, _ := ctx.Value("role").(string)
    if role != "admin" {
        return status.Error(codes.PermissionDenied, "admin only")
    }
    return nil
}
```

### 4.2 Admin-методы по сервисам

**Product Service** — управление каталогом:
- `CreateProduct` — создать товар
- `UpdateProduct` — обновить поля
- `DeactivateProduct` — мягкое удаление (active=false)

**User Service** — управление пользователями:
- `GrantPoints` — начислить баллы
- `ListUsers` — список пользователей
- `BlockUser` — заблокировать учётную запись
- `ChangeUserRole` — изменить роль
- `ResetUserPassword` — сбросить пароль

**Order Service** — управление заказами:
- `ListAllOrders` — все заказы с фильтрами
- `ExportOrders` — выгрузка в Excel
- `UpdateStatus` — сменить статус заказа

**Inventory Service** — управление остатками:
- `AdjustStock` — изменить остаток по дельте

**Media Service** — загрузка фото:
- `UploadPhoto` — принять файл (multipart/streaming), валидировать формат и размер, записать в MinIO, вернуть `photo_key` (только `role=admin`)

### 4.3 Сценарии работы админа

**Добавление товара:**
1. Админ нажимает «Добавить товар», выбирает фото
2. Браузер шлёт `Gateway → media-service.UploadPhoto(filename, content_type, bytes)`
3. Media Service пишет файл в MinIO через `PutObject` и возвращает `photo_key`
4. Браузер вызывает `Gateway → product-service.CreateProduct` с `photo_key`

**Начисление баллов:**
1. Админ открывает профиль пользователя, вводит сумму и причину
2. `Gateway → user-service.GrantPoints` — баланс увеличивается
3. Возвращается новый баланс

**Просмотр заказов:**
1. `Gateway → order-service.ListAllOrders` с фильтрами
2. По кнопке «Экспорт» — `order-service.ExportOrders` формирует Excel-файл

---

## 5. Хранение фото — MinIO и Media Service

### 5.1 Почему MinIO

MinIO — S3-совместимое хранилище с открытым исходным кодом. Подходит идеально:
- Локально разворачивается через `docker-compose` одной строкой
- В Kubernetes — через StatefulSet
- Тот же API, что у AWS S3 / Yandex Object Storage — при переезде в облако код не меняется
- Используется тот же Go-клиент `minio-go`

### 5.2 Паттерн загрузки — через Media Service

Файл **проходит через Media Service**. Браузер с MinIO напрямую не общается — у MinIO нет публичного PUT-эндпоинта, а NetworkPolicy запрещает остальным сервисам сетевой доступ к MinIO на запись.

```
1. Браузер → Gateway → media-service: UploadPhoto(filename, content_type, bytes)
   (gRPC streaming или HTTP multipart, чанки по 64 КБ)
2. media-service: AdminOnly(JWT), validate content-type и размер ≤ 5 МБ
3. media-service: ключ = products/{uuid}.{ext}
4. media-service → MinIO: PutObject(key, stream)
5. media-service → Браузер: { photo_key }
6. Браузер → Gateway → product-service: CreateProduct({ ..., photo_key })
```

**Преимущества:**
- Единая точка валидации формата, размера и роли
- MinIO не выставлен наружу — поверхность атаки минимальна
- Streaming в Media Service устраняет проблему буферизации мегабайт целиком; лимит запроса на стороне сервиса (6 МБ) исключает DoS
- Имена объектов контролируются сервисом — клиент не может перезаписать чужой ключ

### 5.3 Реализация UploadPhoto

Метод живёт в `media-service`. Проверяет роль через интерцептор (`pkg/auth.AdminOnly`), валидирует формат и размер, пишет в MinIO потоком:

```go
import "github.com/minio/minio-go/v7"

const maxPhotoSize = 5 * 1024 * 1024 // 5 МБ

func (s *MediaService) UploadPhoto(stream pb.Media_UploadPhotoServer) error {
    ctx := stream.Context()
    if err := auth.AdminOnly(ctx); err != nil {
        return err
    }

    // Первый чанк содержит метаданные
    head, err := stream.Recv()
    if err != nil {
        return status.Error(codes.InvalidArgument, "missing metadata")
    }
    meta := head.GetMetadata()
    if !isAllowedContentType(meta.ContentType) {
        return status.Error(codes.InvalidArgument, "only jpeg, png, webp allowed")
    }

    ext := mimeToExt(meta.ContentType)
    key := fmt.Sprintf("products/%s%s", uuid.New(), ext)

    // Считаем размер на лету, чтобы не буферизовать
    pr, pw := io.Pipe()
    go func() {
        defer pw.Close()
        var total int64
        for {
            chunk, err := stream.Recv()
            if err == io.EOF {
                return
            }
            if err != nil {
                pw.CloseWithError(err)
                return
            }
            total += int64(len(chunk.GetData()))
            if total > maxPhotoSize {
                pw.CloseWithError(status.Error(codes.InvalidArgument, "file too large"))
                return
            }
            if _, err := pw.Write(chunk.GetData()); err != nil {
                pw.CloseWithError(err)
                return
            }
        }
    }()

    _, err = s.s3.PutObject(ctx, "merch-products", key, pr, -1,
        minio.PutObjectOptions{ContentType: meta.ContentType})
    if err != nil {
        return status.Error(codes.Internal, "upload failed")
    }

    return stream.SendAndClose(&pb.PhotoKey{PhotoKey: key})
}

func isAllowedContentType(ct string) bool {
    return ct == "image/jpeg" || ct == "image/png" || ct == "image/webp"
}
```

### 5.4 Bucket'ы и доступ

| Bucket | Доступ на чтение | Доступ на запись | Назначение |
|---|---|---|---|
| `merch-products` | public-read (GET для всех браузеров) | только Media Service (NetworkPolicy + IAM-аккаунт сервиса) | Фото товаров |

Bucket `merch-products` настроен на публичное чтение — любой пользователь может посмотреть фото товара по прямой ссылке. Это упрощает фронт: ему не нужны подписанные GET URL для каждого товара. На запись MinIO принимает только запросы от Media Service.

### 5.5 Что хранит Product Service

Сервис **не хранит файл**, только ключ:

```sql
ALTER TABLE products ADD COLUMN photo_key TEXT;
```

При выдаче товара клиент сам собирает URL: `https://{S3_PUBLIC_HOST}/merch-products/{photo_key}`. Хост передаётся фронту через переменную окружения.

### 5.6 Деплой MinIO в Kubernetes

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: minio
  namespace: merch-prod
spec:
  serviceName: minio
  replicas: 1
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: minio/minio:latest
        args: ["server", "/data", "--console-address", ":9001"]
        ports:
        - containerPort: 9000
          name: s3-api
        - containerPort: 9001
          name: console
        env:
        - name: MINIO_ROOT_USER
          valueFrom:
            secretKeyRef:
              name: minio-credentials
              key: user
        - name: MINIO_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: minio-credentials
              key: password
        volumeMounts:
        - name: data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ReadWriteOnce]
      resources:
        requests:
          storage: 10Gi
```

---

## 6. gRPC API

### 6.1 ProductService

```protobuf
service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
  rpc CreateProduct(CreateProductRequest) returns (Product);
  rpc UpdateProduct(UpdateProductRequest) returns (Product);
  rpc DeactivateProduct(DeactivateProductRequest) returns (Empty);
}

message Product {
  string id = 1;
  string name = 2;
  string description = 3;
  int64 price_points = 4;
  string category = 5;
  repeated string sizes = 6;
  string photo_key = 7;
  bool active = 8;
}
```

`Create/Update/Deactivate` доступны только с `role=admin` в JWT — проверяется через `pkg/auth.AdminOnly` в gRPC-интерцепторе.

### 6.2 UserService

UserService отвечает не только за профиль и баллы, но и за регистрацию и аутентификацию — это собственная система авторизации, без внешних провайдеров.

```protobuf
service UserService {
  // пользовательские методы
  rpc Register(RegisterRequest) returns (AuthResponse);
  rpc Login(LoginRequest) returns (AuthResponse);
  rpc Refresh(RefreshRequest) returns (AuthResponse);
  rpc Logout(LogoutRequest) returns (Empty);
  rpc ChangePassword(ChangePasswordRequest) returns (Empty);
  rpc GetUser(GetUserRequest) returns (User);
  rpc GetBalance(GetBalanceRequest) returns (Balance);
  rpc DeductPoints(DeductPointsRequest) returns (Balance);
  rpc AddPoints(AddPointsRequest) returns (Balance);
  // admin-only
  rpc GrantPoints(GrantPointsRequest) returns (Balance);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  rpc BlockUser(BlockUserRequest) returns (Empty);
  rpc ChangeUserRole(ChangeUserRoleRequest) returns (Empty);
  rpc ResetUserPassword(ResetUserPasswordRequest) returns (ResetPasswordResponse);
}

message RegisterRequest {
  string login = 1;
  string password = 2;
  string full_name = 3;
}

message LoginRequest {
  string login = 1;
  string password = 2;
}

message AuthResponse {
  string access_token = 1;
  string refresh_token = 2;
  User user = 3;
}

message AddPointsRequest {
  string user_id = 1;
  int64 amount = 2;
  string reason = 3;
}

message GrantPointsRequest {
  string user_id = 1;
  int64 amount = 2;
  string reason = 3;
}
```

**Хеширование паролей**: bcrypt с cost=12 (баланс безопасности и скорости).

**JWT-токены**:
- `access_token` — TTL 15 минут, содержит `user_id`, `role` (`user` / `admin`), `exp`. Подписан HS256 или RS256
- `refresh_token` — TTL 30 дней, хранится в Redis в множестве `user_refresh_tokens:{user_id}` для возможности отзыва
- При смене пароля все refresh-токены пользователя удаляются из Redis

### 6.3 InventoryService

```protobuf
service InventoryService {
  // пользовательские методы
  rpc CheckStock(CheckStockRequest) returns (StockInfo);
  rpc ReserveStock(ReserveStockRequest) returns (Reservation);
  rpc ReleaseReserve(ReleaseReserveRequest) returns (Empty);
  // admin-only
  rpc AdjustStock(AdjustStockRequest) returns (StockInfo);
}

message AdjustStockRequest {
  string product_id = 1;
  string size = 2;
  int32 delta = 3;
  string reason = 4;
}
```

### 6.4 OrderService

```protobuf
service OrderService {
  // пользовательские методы
  rpc CreateOrder(CreateOrderRequest) returns (Order);
  rpc GetOrder(GetOrderRequest) returns (Order);
  rpc ListUserOrders(ListUserOrdersRequest) returns (ListOrdersResponse);
  // admin-only
  rpc ListAllOrders(ListAllOrdersRequest) returns (ListOrdersResponse);
  rpc ExportOrders(ExportOrdersRequest) returns (ExportResponse);
  rpc UpdateStatus(UpdateStatusRequest) returns (Order);
}
```

### 6.5 MediaService

```protobuf
service MediaService {
  // admin-only; gRPC client-streaming
  rpc UploadPhoto(stream UploadPhotoRequest) returns (PhotoKey);
}

message UploadPhotoRequest {
  // первый message — метаданные, последующие — чанки
  oneof payload {
    UploadMetadata metadata = 1;
    bytes data = 2;
  }
}

message UploadMetadata {
  string filename = 1;
  string content_type = 2; // image/jpeg | image/png | image/webp
}

message PhotoKey {
  string photo_key = 1; // например products/{uuid}.jpg
}
```

---

## 7. Kafka — событие order.created

В MVP **один топик** — `order.created`.

```protobuf
message OrderCreatedEvent {
  string order_id = 1;
  string user_id = 2;
  repeated OrderItemEvent items = 3;
  int64 total_points = 4;
  string delivery_address = 5;
  google.protobuf.Timestamp created_at = 6;
}
```

Конфигурация: 3 партиции, replication factor 2, retention 7 дней.

Producer (Order Service): `acks=all`, `enable.idempotence=true`.
Consumer (Inventory Service): `group.id=inventory-service`, ручной коммит, идемпотентная обработка по `order_id`.

### Transactional Outbox

Для гарантии доставки используется Outbox-паттерн: событие пишется в таблицу `outbox` в одной транзакции с заказом, отдельная горутина досылает в Kafka.

---

## 8. Схема данных

### 8.1 Product Service

```sql
CREATE TABLE products (
    id            UUID PRIMARY KEY,
    name          TEXT NOT NULL,
    description   TEXT,
    price_points  BIGINT NOT NULL CHECK (price_points >= 0),
    category      TEXT NOT NULL,
    sizes         TEXT[] NOT NULL,
    photo_key     TEXT,
    active        BOOLEAN DEFAULT true,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_products_category ON products(category) WHERE active = true;
```

### 8.2 Cart Service (PostgreSQL + Redis cache)

PostgreSQL — источник истины:

```sql
CREATE TABLE carts (
    user_id     UUID PRIMARY KEY,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cart_items (
    cart_user_id UUID NOT NULL REFERENCES carts(user_id) ON DELETE CASCADE,
    product_id   UUID NOT NULL,
    size         TEXT NOT NULL,
    qty          INT  NOT NULL CHECK (qty > 0),
    price_points BIGINT NOT NULL,
    product_name TEXT NOT NULL,
    PRIMARY KEY (cart_user_id, product_id, size)
);

CREATE INDEX idx_carts_updated_at ON carts(updated_at);
```

`price_points` и `product_name` снэпшотятся в момент добавления — изменение каталога админом не ломает UX уже собранной корзины.

Redis — write-through кеш:

```
Key:   cart:{user_id}
Type:  String (JSON-снапшот всей корзины)
TTL:   1800 (30 минут)
```

**Поток операций:**
- `AddItem/UpdateQty/RemoveItem` — транзакция в Postgres (UPSERT `carts.updated_at` + изменение `cart_items`), после COMMIT — `SET cart:{user_id} {json} EX 1800`. При недоступности Redis запись в кеш пропускается, операция считается успешной (graceful degradation).
- `GetCart` — `GET cart:{user_id}`; при miss или недоступности Redis собирает корзину из Postgres и best-effort обновляет кеш.
- `ClearCart` — `DELETE FROM carts WHERE user_id = $1` (cascade), затем `DEL cart:{user_id}`.

**Cleanup-воркер:** отдельная горутина раз в час выполняет `DELETE FROM carts WHERE updated_at < NOW() - INTERVAL '24 hours'` — реализует MVP-правило про жизнь корзины 24ч как политику, а не как свойство хранилища.

### 8.3 Order Service

```sql
CREATE TYPE order_status AS ENUM ('pending', 'confirmed', 'cancelled');

CREATE TABLE orders (
    id               UUID PRIMARY KEY,
    user_id          UUID NOT NULL,
    total_points     BIGINT NOT NULL,
    delivery_address TEXT NOT NULL,
    status           order_status NOT NULL DEFAULT 'pending',
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE order_items (
    id            UUID PRIMARY KEY,
    order_id      UUID NOT NULL REFERENCES orders(id),
    product_id    UUID NOT NULL,
    product_name  TEXT NOT NULL,
    size          TEXT NOT NULL,
    qty           INT NOT NULL,
    price_points  BIGINT NOT NULL
);

CREATE TABLE outbox (
    id          UUID PRIMARY KEY,
    aggregate   TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    payload     BYTEA NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    sent_at     TIMESTAMPTZ
);
```

### 8.4 User Service

```sql
CREATE TYPE user_role AS ENUM ('user', 'admin');
CREATE TYPE user_status AS ENUM ('active', 'blocked');

CREATE TABLE users (
    id              UUID PRIMARY KEY,
    login           TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    full_name       TEXT NOT NULL,
    role            user_role NOT NULL DEFAULT 'user',
    status          user_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_login ON users(login);

CREATE TABLE points_balance (
    user_id     UUID PRIMARY KEY REFERENCES users(id),
    points      BIGINT NOT NULL DEFAULT 0 CHECK (points >= 0)
);

CREATE TABLE points_transactions (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id),
    amount      BIGINT NOT NULL,
    reason      TEXT NOT NULL,
    order_id    UUID,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(order_id, user_id)
);
```

Refresh-токены хранятся не в PostgreSQL, а в Redis с TTL:

```
Key:   refresh:{token_jti}
Type:  String
Value: user_id
TTL:   30 дней

Key:   user_refresh_tokens:{user_id}
Type:  Set (для отзыва всех токенов пользователя одной операцией)
```

### 8.5 Inventory Service

```sql
CREATE TABLE stock (
    product_id    UUID NOT NULL,
    size          TEXT NOT NULL,
    available     INT NOT NULL CHECK (available >= 0),
    PRIMARY KEY (product_id, size)
);

CREATE TABLE reservations (
    id           UUID PRIMARY KEY,
    order_id     UUID UNIQUE NOT NULL,
    items        JSONB NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);
```

---

## 9. Структура репозитория

```
merch-store/
├── proto/
│   ├── product/v1/
│   ├── cart/v1/
│   ├── order/v1/
│   ├── user/v1/
│   ├── inventory/v1/
│   ├── media/v1/
│   └── events/v1/
│
├── services/
│   ├── api-gateway/
│   ├── product-service/
│   ├── cart-service/
│   ├── order-service/
│   ├── user-service/
│   ├── inventory-service/
│   └── media-service/
│
├── pkg/
│   ├── auth/
│   ├── kafka/
│   ├── grpc/
│   ├── s3/
│   ├── config/
│   └── logger/
│
├── k8s/
│   ├── base/
│   │   ├── postgres.yaml
│   │   ├── redis.yaml
│   │   ├── kafka.yaml
│   │   ├── minio.yaml
│   │   └── services/
│   └── overlays/
│       ├── dev/
│       └── prod/
│
├── docker-compose.yml
├── buf.yaml
└── Makefile
```

---

## 10. Kubernetes — деплой

### 10.1 Неймспейсы

- `merch-prod` — основное окружение
- `kafka` — Strimzi Kafka cluster
- `monitoring` — Prometheus, Grafana, Jaeger

### 10.2 Ресурсы

| Сервис | Replicas | CPU req | Mem req | HPA max |
|---|---|---|---|---|
| api-gateway | 2 | 200m | 256Mi | 8 |
| product-service | 2 | 100m | 128Mi | 6 |
| cart-service | 2 | 100m | 128Mi | 6 |
| order-service | 2 | 150m | 256Mi | 8 |
| user-service | 2 | 100m | 128Mi | 4 |
| inventory-service | 2 | 100m | 128Mi | 6 |
| media-service | 2 | 150m | 128Mi | 4 |
| minio | 1 | 200m | 512Mi | — |

### 10.3 NetworkPolicy

Все запросы идут через единый Gateway, который проверяет JWT и роль. Доменные сервисы принимают входящий трафик только от Gateway, Cart и Order (межсервисные вызовы):

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-gateway-and-internal
  namespace: merch-prod
spec:
  podSelector: {}
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: api-gateway
    - podSelector:
        matchLabels:
          app: cart-service
    - podSelector:
        matchLabels:
          app: order-service
```

---

## 11. Наблюдаемость

### 11.1 Метрики

Технические:
- `grpc_server_handled_total{service, method, code}`
- `grpc_server_handling_seconds{service, method}`
- `kafka_consumer_lag{topic, partition}`
- `s3_upload_url_generated_total`
- `s3_objects_uploaded_total`

Бизнес:
- `merch_orders_created_total`
- `merch_points_granted_total`
- `merch_products_created_total`

---

## 12. План реализации по неделям

### Неделя 1 — Фундамент и листовые сервисы

- День 1-2: monorepo, buf, Makefile, docker-compose (Postgres × 7, Redis, Kafka, MinIO)
- День 3-4: product-service (proto, БД, gRPC)
- День 5-6: user-service (proto, БД с points_transactions)
- День 7: inventory-service (proto, БД, атомарное резервирование)

### Неделя 2 — Корзина, заказы, админка

- День 8-9: cart-service (Postgres-схема, Redis-кеш с TTL, gRPC-клиенты)
- День 10-11: order-service (оркестрация, Outbox)
- День 12: Kafka producer/consumer
- День 13: api-gateway (JWT, маршрутизация по роли, `/admin/*` guard)
- День 14: media-service (UploadPhoto gRPC streaming, MinIO SDK PutObject), admin-методы в domain-сервисах, интерцептор `pkg/auth`

### Неделя 3 — Kubernetes

- День 15-16: Dockerfile (multi-stage, distroless)
- День 17-18: K8s манифесты, Strimzi Kafka, MinIO StatefulSet
- День 19: NetworkPolicy для межсервисного трафика
- День 20: HPA, Ingress + TLS
- День 21: Kustomize, smoke-тесты

### Неделя 4 — Наблюдаемость и доводка

- День 22-23: OpenTelemetry, Jaeger
- День 24: Prometheus + Grafana, бизнес-дашборды
- День 25: Loki для логов, alert'ы
- День 26-27: e2e-тесты, нагрузочное тестирование (k6)
- День 28: финал, документация, демо

---

## 13. Архитектурные решения (ADR)

### ADR-001: gRPC + Protobuf для синхронных вызовов
Бинарная сериализация, строгий контракт, HTTP/2 multiplexing.

### ADR-002: Kafka для асинхронности
Гарантия доставки, replay, fan-out.

### ADR-003: DB per service
Независимость сервисов, изоляция отказов.

### ADR-004: PostgreSQL + Redis cache для корзины

**Контекст**: корзина — данные на грани между «хочется быстро» и «нельзя терять». Полностью хранить в Redis (как было в первой версии) рискованно: падение Redis или истечение TTL у активного пользователя приводит к потере собранной корзины.

**Решение**: PostgreSQL как источник истины (`carts`, `cart_items`), Redis как write-through кеш с TTL 30 мин. Автоматической очистки корзины по времени нет — содержимое в PostgreSQL живёт до явных действий пользователя (оформление заказа, очистка, удаление позиций). При истечении TTL в Redis запись на следующем обращении подгружается из PostgreSQL.

**Альтернативы**:
- Только Redis — отвергнуто: риск потери данных при сбое или TTL
- Только Postgres — отвергнуто: каждый AddItem/GetCart на горячем пути упирается в БД

**Последствия**: +1 Postgres-инстанс, чуть сложнее логика записи, но получаем надёжность и сохраняем O(1) на чтении из кеша. При недоступности Redis сервис продолжает работать на Postgres (graceful degradation).

### ADR-005: Choreography Saga
Хореография через события вместо центрального оркестратора.

### ADR-006: Outbox для надёжной публикации в Kafka
Гарантия at-least-once доставки.

### ADR-007: Идемпотентность через order_id и operation_id
Защита от повторной обработки.

### ADR-008: Роль admin в JWT

**Контекст**: нужен способ ограничить доступ к опасным операциям (начисление баллов, изменение каталога) без избыточной сложности.

**Решение**: роль `admin` кладётся в JWT при входе; каждый сервис проверяет роль через общий интерцептор из `pkg/auth`. Админские операции выполняются непосредственно теми же доменными сервисами, что и пользовательские — отдельного сервиса для админки нет.

**Альтернативы**:
- Проверка роли только в Gateway — отвергнуто: не защищает от прямых обращений к сервисам внутри кластера

**Последствия**: проще деплой и поддержка; меньше точек отказа.

### ADR-009: MinIO (S3) для фото товаров и загрузка через Media Service

**Контекст**: фото товаров нельзя хранить в БД или файловой системе подов; нужен путь загрузки от админа, не выставляющий MinIO в публичный интернет.

**Решение**: MinIO как S3-совместимое хранилище. Загрузка идёт `браузер → Media Service (gRPC streaming / HTTP multipart) → MinIO.PutObject`. MinIO не имеет публичного PUT-эндпоинта; NetworkPolicy ограничивает запись в bucket только Media Service. Чтение фото клиентами идёт по public-read GET напрямую к MinIO (или CDN), без сервиса.

**Альтернативы**:
- Presigned URL для PUT и прямая загрузка браузером в MinIO — отвергнуто: расширяет поверхность атаки на MinIO, усложняет валидацию формата/размера, требует публичного PUT-эндпоинта; для админа сэкономленного оверхеда мало
- AWS S3 / Yandex Object Storage — отвергнуто для MVP: лишняя зависимость от внешнего провайдера; при необходимости можно переехать без изменений кода (тот же S3 API)
- Хранить фото в Postgres (bytea) — отвергнуто: раздувание БД, медленные бэкапы, плохая отдача через CDN

**Последствия**: один Go-клиент (`minio-go`), легко поднимается локально и в k8s; единая точка валидации формата/размера/роли в Media Service; streaming-чтение исключает буферизацию мегабайт в памяти, лимит запроса 6 МБ устраняет DoS.

---

*Merch Store MVP Architecture v1.1*