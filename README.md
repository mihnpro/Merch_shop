# Merch Store MVP — Архитектурная документация

| | |
|---|---|
| **Версия** | 1.0.0 MVP |
| **Стек** | Go · gRPC · Apache Kafka · Kubernetes · PostgreSQL · Redis |
| **Тип проекта** | Внутренний магазин мерча за корпоративные баллы |
| **Архитектура** | Микросервисная (6 сервисов) |
| **Срок** | 1 месяц |

---

## Содержание

1. [Обзор и скоуп MVP](#1-обзор-и-скоуп-mvp)
2. [Сервисы и ответственности](#2-сервисы-и-ответственности)
3. [Полный сценарий покупки](#3-полный-сценарий-покупки)
4. [gRPC API — все методы и связи](#4-grpc-api--все-методы-и-связи)
5. [Kafka — событие order.created](#5-kafka--событие-ordercreated)
6. [Схема данных по сервисам](#6-схема-данных-по-сервисам)
7. [Структура репозитория](#7-структура-репозитория)
8. [Kubernetes — деплой](#8-kubernetes--деплой)
9. [Наблюдаемость](#9-наблюдаемость)
10. [План реализации по неделям](#10-план-реализации-по-неделям)
11. [Архитектурные решения (ADR)](#11-архитектурные-решения-adr)
12. [Что вне MVP — план v2](#12-что-вне-mvp--план-v2)

---

## 1. Обзор и скоуп MVP

**Merch Store** — внутренняя платформа, где сотрудники могут заказывать корпоративный мерч (футболки, кружки, толстовки) за корпоративные баллы, которые начисляются за достижения, выслугу, KPI.

### 1.1 Что входит в MVP

- Каталог товаров с фильтрами по категориям и размерам
- Корзина с TTL (24 часа)
- Оформление заказа с проверкой остатков и баланса баллов
- Резервирование остатков (защита от двойной продажи последней футболки)
- Аутентификация через корпоративный SSO (JWT)
- Просмотр истории своих заказов

### 1.2 Что **не** входит в MVP

- Уведомления (Slack, email) — заказы видны в интерфейсе, статус обновляется при перезагрузке
- Автоматический фулфилмент — оплаченные заказы выгружаются админом вручную в Excel
- Возвраты и сложная история транзакций — только начисление и списание
- Промокоды, скидки, акции
- Отдельный payment-service — баллы списывает `user-service` через `DeductPoints`

### 1.3 Ключевые архитектурные принципы

- **gRPC между сервисами** — синхронные вызовы там, где пользователь ждёт ответа (проверка остатков, баланса)
- **Kafka для асинхронности** — одно событие `order.created`, чтобы развязать создание заказа от резервирования остатков
- **DB per service** — каждый сервис со своей PostgreSQL-схемой, Cart использует Redis
- **API Gateway как единая точка входа** — клиент работает только с gateway, никогда напрямую с сервисами

---

## 2. Сервисы и ответственности

В MVP **6 сервисов**:

| # | Сервис | Что делает | Хранилище |
|---|---|---|---|
| 1 | **API Gateway** | Принимает HTTP/gRPC от клиента, проверяет JWT, проксирует в нужный сервис | — |
| 2 | **Product Service** | Каталог: список товаров, поиск, детали | PostgreSQL |
| 3 | **Cart Service** | Корзина пользователя с TTL 24h | Redis |
| 4 | **Order Service** | Создание заказа, история, статусы | PostgreSQL |
| 5 | **User Service** | Профиль сотрудника, баланс баллов, начисление/списание | PostgreSQL |
| 6 | **Inventory Service** | Остатки товаров, резервирование при заказе | PostgreSQL |

### 2.1 Кто кого вызывает (граф зависимостей)

| Источник | Цель | Метод | Зачем |
|---|---|---|---|
| Gateway | все 5 сервисов | различные | Проксирование клиентских запросов |
| Cart | Product | `GetProduct` | Получить цену и название для отображения корзины |
| Cart | Inventory | `CheckStock` | Проверить наличие при добавлении товара |
| Order | Cart | `GetCart`, `ClearCart` | Взять содержимое корзины и очистить после оформления |
| Order | User | `GetBalance`, `DeductPoints` | Проверить и списать баллы |
| Inventory | Kafka (consumer) | `order.created` | Резервирование остатков асинхронно |
| Order | Kafka (producer) | `order.created` | Публикация события после успешного создания |

Важно: `Product`, `User` и `Inventory` **никого не вызывают синхронно** — только отвечают на gRPC. Это листовые сервисы без зависимостей, их легко тестировать и масштабировать.

---

## 3. Полный сценарий покупки

Самый важный сценарий — пройти его шаг за шагом, чтобы понять весь поток.

### Шаг 1. Авторизация
Сотрудник заходит в магазин, проходит через корпоративный SSO, получает JWT-токен. Все последующие запросы идут с этим токеном.

### Шаг 2. Просмотр каталога
```
Client → Gateway → Product.ListProducts(category="t-shirts", page=1)
       ← Product.ListProductsResponse{items: [...], total: 24}
```

### Шаг 3. Добавление в корзину
```
Client → Gateway → Cart.AddItem(user_id, product_id="p123", size="L", qty=1)

  внутри Cart Service:
    Cart → Inventory.CheckStock(product_id="p123", size="L")
         ← StockInfo{available: 12}    [OK, в наличии]
    Cart → Product.GetProduct(product_id="p123")
         ← Product{name: "...", price_points: 500, ...}
    Cart сохраняет в Redis: cart:{user_id} → {items, total_points}

← Cart{items: [...], total_points: 500, expires_at: "..."}
```

### Шаг 4. Оформление заказа (самый интересный шаг)
```
Client → Gateway → Order.CreateOrder(user_id, delivery_address)

  внутри Order Service (всё это происходит в одной транзакции Order'а):
    1. Order → Cart.GetCart(user_id)
            ← Cart{items: [...], total_points: 500}

    2. Order → User.GetBalance(user_id)
            ← Balance{points: 1200}      [хватает]

    3. Order → User.DeductPoints(user_id, amount=500, order_id="o456")
            ← Balance{remaining_points: 700}   [списано]

    4. Order сохраняет заказ в свою БД (orders, order_items) со статусом "pending"

    5. Order → Kafka.publish("order.created", {order_id, user_id, items, total_points})

    6. Order → Cart.ClearCart(user_id)
            ← Empty

← Order{id: "o456", status: "pending", created_at: "..."}
```

### Шаг 5. Асинхронная обработка
```
Параллельно, в фоне:
  Inventory ← Kafka.consume("order.created")
  Inventory резервирует остатки по каждой позиции (UPDATE stock SET qty = qty - 1)
  Inventory обновляет статус заказа через gRPC: Order.UpdateStatus(order_id, "confirmed")
```

### Что произойдёт если что-то пойдёт не так

| Сбой | Что делаем |
|---|---|
| `Inventory.CheckStock` вернул 0 | `Cart.AddItem` возвращает ошибку `OUT_OF_STOCK`, баллы не списаны |
| `User.GetBalance` показал недостаточно | `Order.CreateOrder` возвращает `INSUFFICIENT_POINTS`, ничего не списано |
| `Inventory` не смог зарезервировать после Kafka | Публикует событие `order.cancelled` (в MVP — просто меняет статус через gRPC), `Order` возвращает баллы через `User.AddPoints` |
| Kafka недоступен | `Order.CreateOrder` падает на шаге 5 — баллы уже списаны. Решение: **Transactional Outbox pattern** (сохраняем событие в таблицу `outbox`, отдельный воркер досылает в Kafka) |

---

## 4. gRPC API — все методы и связи

Все proto-файлы лежат в `proto/{service}/v1/`. Кодогенерация через `buf`.

### 4.1 ProductService

**proto/product/v1/product.proto**

```protobuf
syntax = "proto3";
package product.v1;

option go_package = "merch-store/proto/product/v1;productv1";

service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
}

message Product {
  string id = 1;
  string name = 2;
  string description = 3;
  int64 price_points = 4;
  string category = 5;          // "t-shirts", "mugs", "hoodies"
  repeated string sizes = 6;    // ["S", "M", "L", "XL"]
  string photo_url = 7;
  bool active = 8;
}

message GetProductRequest {
  string product_id = 1;
}

message ListProductsRequest {
  int32 page = 1;
  int32 limit = 2;
  string category = 3;          // optional фильтр
}

message ListProductsResponse {
  repeated Product items = 1;
  int32 total = 2;
}
```

| Метод | Кто вызывает | Когда |
|---|---|---|
| `GetProduct` | Cart, Order, Gateway | Получить детали товара |
| `ListProducts` | Gateway | Показать каталог пользователю |

### 4.2 CartService

**proto/cart/v1/cart.proto**

```protobuf
syntax = "proto3";
package cart.v1;

service CartService {
  rpc AddItem(AddItemRequest) returns (Cart);
  rpc RemoveItem(RemoveItemRequest) returns (Cart);
  rpc GetCart(GetCartRequest) returns (Cart);
  rpc ClearCart(ClearCartRequest) returns (Empty);
}

message Cart {
  string user_id = 1;
  repeated CartItem items = 2;
  int64 total_points = 3;
  google.protobuf.Timestamp expires_at = 4;
}

message CartItem {
  string product_id = 1;
  string product_name = 2;       // снепшот для UI
  string size = 3;
  int32 qty = 4;
  int64 price_points = 5;        // снепшот цены на момент добавления
}

message AddItemRequest {
  string user_id = 1;
  string product_id = 2;
  string size = 3;
  int32 qty = 4;
}
```

| Метод | Кто вызывает | Когда |
|---|---|---|
| `AddItem` | Gateway | Пользователь нажал «В корзину» |
| `RemoveItem` | Gateway | Пользователь удалил позицию |
| `GetCart` | Gateway, **Order** | Показать корзину / получить для создания заказа |
| `ClearCart` | **Order** | После успешного оформления |

### 4.3 OrderService

**proto/order/v1/order.proto**

```protobuf
syntax = "proto3";
package order.v1;

service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (Order);
  rpc GetOrder(GetOrderRequest) returns (Order);
  rpc ListUserOrders(ListUserOrdersRequest) returns (ListUserOrdersResponse);
  rpc UpdateStatus(UpdateStatusRequest) returns (Order);
}

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;       // создан, ждёт резерва остатков
  ORDER_STATUS_CONFIRMED = 2;     // остатки зарезервированы
  ORDER_STATUS_CANCELLED = 3;     // не удалось зарезервировать или отменён
}

message Order {
  string id = 1;
  string user_id = 2;
  repeated OrderItem items = 3;
  int64 total_points = 4;
  string delivery_address = 5;
  OrderStatus status = 6;
  google.protobuf.Timestamp created_at = 7;
}

message OrderItem {
  string product_id = 1;
  string product_name = 2;
  string size = 3;
  int32 qty = 4;
  int64 price_points = 5;
}

message CreateOrderRequest {
  string user_id = 1;
  string delivery_address = 2;
}
```

| Метод | Кто вызывает | Когда |
|---|---|---|
| `CreateOrder` | Gateway | Пользователь нажал «Оформить» |
| `GetOrder` | Gateway | Просмотр конкретного заказа |
| `ListUserOrders` | Gateway | История заказов пользователя |
| `UpdateStatus` | **Inventory** | После резерва / при отмене |

### 4.4 UserService

**proto/user/v1/user.proto**

```protobuf
syntax = "proto3";
package user.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc GetBalance(GetBalanceRequest) returns (Balance);
  rpc DeductPoints(DeductPointsRequest) returns (Balance);
  rpc AddPoints(AddPointsRequest) returns (Balance);
}

message User {
  string id = 1;
  string email = 2;
  string full_name = 3;
  string department = 4;
}

message Balance {
  string user_id = 1;
  int64 points = 2;
}

message DeductPointsRequest {
  string user_id = 1;
  int64 amount = 2;
  string order_id = 3;            // для идемпотентности
  string reason = 4;
}
```

| Метод | Кто вызывает | Когда |
|---|---|---|
| `GetUser` | Gateway | Получение профиля |
| `GetBalance` | Gateway, **Order** | Показать баланс / проверить перед списанием |
| `DeductPoints` | **Order** | Списать при создании заказа |
| `AddPoints` | **Order** | Вернуть при отмене (компенсация) |

`DeductPoints` использует `order_id` как ключ идемпотентности — повторный вызов с тем же `order_id` не приведёт к двойному списанию.

### 4.5 InventoryService

**proto/inventory/v1/inventory.proto**

```protobuf
syntax = "proto3";
package inventory.v1;

service InventoryService {
  rpc CheckStock(CheckStockRequest) returns (StockInfo);
  rpc ReserveStock(ReserveStockRequest) returns (Reservation);
  rpc ReleaseReserve(ReleaseReserveRequest) returns (Empty);
}

message StockInfo {
  string product_id = 1;
  string size = 2;
  int32 available = 3;
}

message ReserveStockRequest {
  string order_id = 1;            // ключ идемпотентности
  repeated ReserveItem items = 2;
}

message ReserveItem {
  string product_id = 1;
  string size = 2;
  int32 qty = 3;
}

message Reservation {
  string id = 1;
  string order_id = 2;
  google.protobuf.Timestamp expires_at = 3;
}
```

| Метод | Кто вызывает | Когда |
|---|---|---|
| `CheckStock` | **Cart** | Проверка при `AddItem` |
| `ReserveStock` | внутренний | Используется обработчиком Kafka-события |
| `ReleaseReserve` | внутренний | При отмене заказа |

---

## 5. Kafka — событие order.created

В MVP **один топик** — `order.created`. Этого достаточно: всё остальное обрабатывается синхронно через gRPC.

### 5.1 Структура события

```protobuf
syntax = "proto3";
package events.v1;

message OrderCreatedEvent {
  string order_id = 1;
  string user_id = 2;
  repeated OrderItemEvent items = 3;
  int64 total_points = 4;
  string delivery_address = 5;
  google.protobuf.Timestamp created_at = 6;
}

message OrderItemEvent {
  string product_id = 1;
  string size = 2;
  int32 qty = 3;
}
```

### 5.2 Конфигурация топика

| Параметр | Значение | Зачем |
|---|---|---|
| partitions | 3 | Возможность горизонтального масштабирования consumer'ов |
| replication.factor | 2 | Отказоустойчивость (в проде — 3) |
| retention.ms | 604800000 (7 дней) | Возможность replay в случае инцидентов |
| cleanup.policy | delete | Логи удаляются по retention |

### 5.3 Producer (в Order Service)

Используется `confluent-kafka-go` или `segmentio/kafka-go`. Конфигурация:

```go
config := &kafka.ConfigMap{
    "bootstrap.servers":  os.Getenv("KAFKA_BROKERS"),
    "acks":               "all",            // ждём подтверждения от всех реплик
    "enable.idempotence": true,             // защита от дубликатов
    "retries":            10,
    "retry.backoff.ms":   100,
}
```

### 5.4 Consumer (в Inventory Service)

```go
config := &kafka.ConfigMap{
    "bootstrap.servers":   os.Getenv("KAFKA_BROKERS"),
    "group.id":            "inventory-service",
    "auto.offset.reset":   "earliest",
    "enable.auto.commit":  false,           // ручной коммит после успешной обработки
}
```

Обработка должна быть **идемпотентной** — если consumer упал после обработки, но до коммита, при перезапуске тот же event придёт снова. Используем `order_id` для дедупликации.

### 5.5 Transactional Outbox Pattern

Чтобы избежать ситуации «заказ записан в БД, но Kafka недоступен», `order-service` использует Outbox-паттерн:

1. В транзакции вставляем запись в таблицу `orders` **и** в таблицу `outbox` (событие)
2. Отдельная горутина читает `outbox` и шлёт в Kafka, помечая отправленные
3. При сбое — повторяет, гарантируя at-least-once доставку

```sql
CREATE TABLE outbox (
    id          UUID PRIMARY KEY,
    aggregate   TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    payload     BYTEA NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    sent_at     TIMESTAMPTZ
);
```

---

## 6. Схема данных по сервисам

Каждый сервис имеет **собственную БД**. Никакого shared schema, никаких JOIN между сервисами.

### 6.1 Product Service (PostgreSQL)

```sql
CREATE TABLE products (
    id            UUID PRIMARY KEY,
    name          TEXT NOT NULL,
    description   TEXT,
    price_points  BIGINT NOT NULL CHECK (price_points >= 0),
    category      TEXT NOT NULL,
    sizes         TEXT[] NOT NULL,
    photo_url     TEXT,
    active        BOOLEAN DEFAULT true,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_products_category ON products(category) WHERE active = true;
CREATE INDEX idx_products_active   ON products(active);
```

### 6.2 Cart Service (Redis)

Корзина хранится как один JSON-объект на пользователя:

```
Key:   cart:{user_id}
Type:  String (JSON)
TTL:   86400 (24 часа)

Value:
{
  "user_id": "u123",
  "items": [
    {"product_id": "p1", "name": "T-shirt Black", "size": "L", "qty": 2, "price_points": 500}
  ],
  "total_points": 1000,
  "updated_at": "2026-05-26T10:00:00Z"
}
```

При каждой операции (`AddItem`, `RemoveItem`) TTL обновляется, чтобы активная корзина не пропадала.

### 6.3 Order Service (PostgreSQL)

```sql
CREATE TYPE order_status AS ENUM ('pending', 'confirmed', 'cancelled');

CREATE TABLE orders (
    id               UUID PRIMARY KEY,
    user_id          UUID NOT NULL,
    total_points     BIGINT NOT NULL,
    delivery_address TEXT NOT NULL,
    status           order_status NOT NULL DEFAULT 'pending',
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE order_items (
    id            UUID PRIMARY KEY,
    order_id      UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id    UUID NOT NULL,
    product_name  TEXT NOT NULL,        -- снепшот, на случай удаления товара
    size          TEXT NOT NULL,
    qty           INT NOT NULL,
    price_points  BIGINT NOT NULL
);

CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);
CREATE INDEX idx_orders_status        ON orders(status);

CREATE TABLE outbox (
    id          UUID PRIMARY KEY,
    aggregate   TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    payload     BYTEA NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    sent_at     TIMESTAMPTZ
);

CREATE INDEX idx_outbox_unsent ON outbox(created_at) WHERE sent_at IS NULL;
```

### 6.4 User Service (PostgreSQL)

```sql
CREATE TABLE users (
    id          UUID PRIMARY KEY,
    email       TEXT UNIQUE NOT NULL,
    full_name   TEXT NOT NULL,
    department  TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE points_balance (
    user_id     UUID PRIMARY KEY REFERENCES users(id),
    points      BIGINT NOT NULL DEFAULT 0 CHECK (points >= 0),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE points_transactions (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id),
    amount      BIGINT NOT NULL,          -- положительное = начисление, отрицательное = списание
    reason      TEXT NOT NULL,
    order_id    UUID,                     -- NULL для начислений, заполнено для списаний
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(order_id, user_id)             -- идемпотентность DeductPoints
);

CREATE INDEX idx_tx_user ON points_transactions(user_id, created_at DESC);
```

`DeductPoints` работает так: в одной транзакции вставляет запись в `points_transactions` (с уникальным ключом по `order_id`) и обновляет `points_balance.points`. Если запись с таким `order_id` уже есть — возвращает текущий баланс, не списывая повторно.

### 6.5 Inventory Service (PostgreSQL)

```sql
CREATE TABLE stock (
    product_id    UUID NOT NULL,
    size          TEXT NOT NULL,
    available     INT NOT NULL CHECK (available >= 0),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (product_id, size)
);

CREATE TABLE reservations (
    id           UUID PRIMARY KEY,
    order_id     UUID UNIQUE NOT NULL,    -- идемпотентность
    items        JSONB NOT NULL,          -- [{product_id, size, qty}]
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_reservations_order ON reservations(order_id);
```

Резервирование — это атомарный `UPDATE stock SET available = available - $qty WHERE product_id = $1 AND size = $2 AND available >= $qty RETURNING available`. Если `UPDATE` затронул 0 строк — товара нет.

---

## 7. Структура репозитория

Monorepo с разделением на сервисы и общий код:

```
merch-store/
├── proto/                              # gRPC контракты
│   ├── product/v1/product.proto
│   ├── cart/v1/cart.proto
│   ├── order/v1/order.proto
│   ├── user/v1/user.proto
│   ├── inventory/v1/inventory.proto
│   └── events/v1/order_events.proto   # Kafka-сообщения
│
├── services/
│   ├── api-gateway/
│   ├── product-service/
│   ├── cart-service/
│   ├── order-service/
│   ├── user-service/
│   └── inventory-service/
│
├── pkg/                                # Переиспользуемый код
│   ├── kafka/                          # Producer/Consumer helpers
│   ├── grpc/                           # Interceptors (auth, logging, tracing, metrics)
│   ├── config/                         # Viper loader
│   ├── database/                       # pgx pool helper
│   └── logger/                         # Zap setup
│
├── k8s/
│   ├── base/                           # Базовые манифесты
│   │   ├── product-service.yaml
│   │   ├── cart-service.yaml
│   │   ├── order-service.yaml
│   │   ├── user-service.yaml
│   │   ├── inventory-service.yaml
│   │   ├── api-gateway.yaml
│   │   ├── postgres.yaml
│   │   ├── redis.yaml
│   │   └── kafka.yaml
│   └── overlays/
│       ├── dev/
│       └── prod/
│
├── docker-compose.yml                  # Локальная разработка
├── buf.yaml                            # Конфиг buf для proto-генерации
├── buf.gen.yaml
└── Makefile
```

### 7.1 Структура отдельного сервиса (на примере order-service)

Используем чистую архитектуру (хендлер → сервис → репозиторий):

```
order-service/
├── cmd/
│   └── main.go                         # Entry point, DI wiring
│
├── internal/
│   ├── handler/                        # gRPC хендлеры
│   │   └── order_handler.go            # реализует pb.OrderServiceServer
│   │
│   ├── service/                        # Бизнес-логика
│   │   └── order_service.go            # CreateOrder со всей оркестрацией
│   │
│   ├── repository/                     # Работа с БД
│   │   ├── order_repo.go
│   │   └── outbox_repo.go
│   │
│   ├── client/                         # gRPC клиенты к другим сервисам
│   │   ├── cart_client.go
│   │   └── user_client.go
│   │
│   ├── kafka/                          # Outbox dispatcher
│   │   └── outbox_publisher.go
│   │
│   └── config/
│       └── config.go
│
├── migrations/                         # SQL миграции (goose/golang-migrate)
│   ├── 001_create_orders.up.sql
│   └── 001_create_orders.down.sql
│
├── Dockerfile
└── config.yaml
```

### 7.2 Makefile — основные команды

```makefile
.PHONY: proto build test docker up down

proto:
	buf generate

build:
	@for svc in product-service cart-service order-service user-service inventory-service api-gateway; do \
		go build -o ./bin/$$svc ./services/$$svc/cmd; \
	done

test:
	go test ./...

docker:
	@for svc in product-service cart-service order-service user-service inventory-service api-gateway; do \
		docker build -t merch/$$svc:latest -f services/$$svc/Dockerfile .; \
	done

up:
	docker-compose up -d

down:
	docker-compose down -v

k8s-deploy:
	kubectl apply -k k8s/overlays/dev
```

---

## 8. Kubernetes — деплой

### 8.1 Структура неймспейсов

- `merch-dev` — окружение для разработки и интеграционных тестов
- `merch-prod` — продакшен
- `kafka` — Strimzi Kafka Operator + Kafka cluster
- `monitoring` — Prometheus, Grafana, Jaeger

### 8.2 Пример Deployment для order-service

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-service
  namespace: merch-prod
  labels:
    app: order-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: order-service
  template:
    metadata:
      labels:
        app: order-service
    spec:
      containers:
      - name: order-service
        image: registry.local/merch/order-service:v1.0.0
        ports:
        - containerPort: 50051
          name: grpc
        - containerPort: 9090
          name: metrics
        env:
        - name: KAFKA_BROKERS
          valueFrom:
            configMapKeyRef:
              name: kafka-config
              key: brokers
        - name: DB_DSN
          valueFrom:
            secretKeyRef:
              name: order-db
              key: dsn
        - name: CART_SERVICE_ADDR
          value: "cart-service:50051"
        - name: USER_SERVICE_ADDR
          value: "user-service:50051"
        resources:
          requests:
            cpu: 150m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
        readinessProbe:
          grpc:
            port: 50051
          initialDelaySeconds: 5
          periodSeconds: 5
        livenessProbe:
          grpc:
            port: 50051
          periodSeconds: 10
          failureThreshold: 3
---
apiVersion: v1
kind: Service
metadata:
  name: order-service
  namespace: merch-prod
spec:
  selector:
    app: order-service
  ports:
  - port: 50051
    targetPort: 50051
    name: grpc
  - port: 9090
    targetPort: 9090
    name: metrics
```

### 8.3 HPA по метрикам

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: order-service-hpa
  namespace: merch-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: order-service
  minReplicas: 2
  maxReplicas: 8
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### 8.4 Ingress для API Gateway

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: merch-gateway
  namespace: merch-prod
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt
spec:
  tls:
  - hosts:
    - merch.company.com
    secretName: merch-tls
  rules:
  - host: merch.company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-gateway
            port:
              number: 8080
```

---

## 9. Наблюдаемость

### 9.1 Метрики (Prometheus)

Каждый сервис экспонирует `/metrics` на порту 9090.

Технические:
- `grpc_server_handled_total{service, method, code}` — количество запросов
- `grpc_server_handling_seconds{service, method}` — latency (histogram)
- `kafka_consumer_lag{topic, partition}` — отставание consumer'ов
- `db_query_duration_seconds{service, query_type}` — время БД-запросов

Бизнес-метрики:
- `merch_orders_created_total` — счётчик созданных заказов
- `merch_orders_cancelled_total{reason}` — отмены с причинами
- `merch_cart_items_added_total` — добавления в корзину

### 9.2 Distributed tracing (Jaeger)

OpenTelemetry SDK + автоматические interceptor'ы для gRPC и пропагация trace_id через Kafka headers.

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    "google.golang.org/grpc"
)

server := grpc.NewServer(
    grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
)

conn, _ := grpc.Dial(addr,
    grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
)
```

Это позволяет в Jaeger увидеть полный trace одного `CreateOrder`: gateway → order-service → cart-service → product-service → user-service → kafka → inventory-service.

### 9.3 Логи (Zap)

Структурированные JSON-логи. Обязательные поля:

```json
{
  "level": "info",
  "timestamp": "2026-05-26T10:00:00.000Z",
  "service": "order-service",
  "trace_id": "abc123",
  "span_id": "def456",
  "message": "order created",
  "order_id": "o789",
  "user_id": "u123",
  "total_points": 500
}
```

Сбор через Fluent Bit → Loki, просмотр через Grafana.

---

## 10. План реализации по неделям

### Неделя 1 — Фундамент и листовые сервисы

**Цель**: рабочий локальный стек с двумя сервисами без зависимостей.

- День 1-2: настройка monorepo, `buf` для proto, Makefile, `docker-compose.yml` (Postgres × 5, Redis, Kafka, Zookeeper)
- День 3-4: `product-service` — proto, БД, миграции, gRPC server, базовые методы, unit-тесты
- День 5-6: `user-service` — proto, БД с `points_transactions`, gRPC server, идемпотентность `DeductPoints`
- День 7: `inventory-service` — proto, БД, атомарное резервирование через SQL

**Результат недели**: три сервиса запускаются в docker-compose, можно вызвать их через `grpcurl`.

### Неделя 2 — Корзина, заказы, Kafka

**Цель**: полный пользовательский сценарий покупки.

- День 8-9: `cart-service` — Redis-репозиторий, gRPC server, вызов `Product.GetProduct` и `Inventory.CheckStock` через gRPC-клиенты
- День 10-12: `order-service` — gRPC server, оркестрация (Cart → User → save → Kafka → ClearCart), Outbox-паттерн
- День 13: Kafka producer в `order-service`, consumer в `inventory-service`, обработка `order.created`
- День 14: `api-gateway` — gRPC-gateway, JWT-middleware, маршрутизация

**Результат недели**: полный e2e сценарий покупки работает локально.

### Неделя 3 — Kubernetes

**Цель**: всё развёрнуто в k8s-кластере.

- День 15-16: Dockerfile для каждого сервиса (multi-stage, distroless), сборка через Makefile, локальный registry
- День 17: K8s манифесты — Deployment, Service, ConfigMap, Secret для каждого сервиса
- День 18: Strimzi Kafka Operator в кластере, создание топика `order.created`
- День 19: PostgreSQL операторы (или StatefulSet), Redis Deployment
- День 20: Ingress + cert-manager для gateway, HPA для всех сервисов
- День 21: Kustomize overlays для dev/prod, smoke-тесты в кластере

**Результат недели**: проект развёрнут в k8s, доступен через домен.

### Неделя 4 — Наблюдаемость и доводка

**Цель**: продакшен-готовность.

- День 22-23: OpenTelemetry interceptors во всех сервисах, Jaeger в кластере, проверка end-to-end trace
- День 24: Prometheus + Grafana, dashboard'ы для бизнес- и технических метрик
- День 25: ELK/Loki для логов, alert'ы в Grafana (consumer lag, error rate)
- День 26-27: e2e-тесты сценария покупки, нагрузочное тестирование через k6 (target: 100 RPS)
- День 28: финальная проверка, документация, демо

**Результат**: проект готов к презентации с метриками и графиками.

---

## 11. Архитектурные решения (ADR)

### ADR-001: gRPC + Protobuf для синхронных вызовов

**Контекст**: нужен межсервисный протокол с типизацией и низкой задержкой.

**Решение**: gRPC поверх HTTP/2 с Protobuf вместо REST/JSON.

**Альтернативы**:
- REST/JSON — отвергнут: больше latency, нет строгого контракта, ручной парсинг ошибок
- GraphQL — отвергнут: избыточен для server-to-server, оверхед на парсинг queries

**Последствия**: нужен `buf` для генерации, сложнее отлаживать (бинарный формат), но быстрее на ~3-5x и контракт строго типизирован.

### ADR-002: Kafka для асинхронных событий

**Контекст**: после создания заказа нужно зарезервировать остатки, но клиента не нужно заставлять ждать.

**Решение**: один топик `order.created`, `inventory-service` слушает.

**Альтернативы**:
- RabbitMQ — отвергнут: нет replay событий, сложнее масштабировать consumer'ов
- Делать резерв синхронно через gRPC — отвергнут: блокирует ответ пользователю, при сбое Inventory падает весь сценарий

**Последствия**: добавляется eventual consistency (между созданием заказа и резервом есть лаг ~100ms), нужен Outbox для гарантии доставки, но zero coupling — Inventory может упасть, заказ всё равно создастся.

### ADR-003: DB per service

**Решение**: каждый сервис имеет собственную PostgreSQL-схему (в проде — отдельный instance).

**Альтернативы**:
- Shared DB — отвергнуто: связанность по схеме, миграция одного сервиса ломает другой, общая точка отказа

**Последствия**: нет JOIN между сервисами, нужны снепшоты данных (например, `product_name` в `order_items`), но сервисы независимы в развитии.

### ADR-004: Redis для корзины

**Контекст**: корзина — это временные данные, читаются часто, не требуют сильной консистентности.

**Решение**: Redis с TTL 24 часа, один ключ на пользователя.

**Альтернативы**:
- PostgreSQL — отвергнуто: лишние миграции, нужна задача чистки старых корзин
- In-memory в сервисе — отвергнуто: теряется при рестарте пода

**Последствия**: данные не персистентны (теряются при падении Redis), но MVP это устраивает — пользователь просто пересоберёт корзину.

### ADR-005: Хореография (без оркестратора) для оформления заказа

**Контекст**: создание заказа затрагивает 4 сервиса (Cart, User, Order, Inventory).

**Решение**: `order-service` сам оркестрирует синхронные вызовы Cart и User, а Inventory подписан на Kafka. Без отдельного Saga-оркестратора.

**Альтернативы**:
- Saga Orchestrator (отдельный сервис) — отвергнуто: overkill для MVP, добавляет ещё один сервис без острой необходимости

**Последствия**: логика заказа сосредоточена в `order-service`, что приемлемо для MVP. При росте сложности можно вынести в оркестратор.

### ADR-006: Outbox для надёжной публикации в Kafka

**Контекст**: ситуация «заказ сохранён в БД, но Kafka недоступна» приведёт к рассинхрону.

**Решение**: в одной транзакции с заказом записываем событие в таблицу `outbox`, отдельный воркер досылает в Kafka.

**Последствия**: гарантия at-least-once доставки, потребители должны быть идемпотентны (что они и есть благодаря `order_id`).

### ADR-007: Идемпотентность через order_id

**Решение**: `DeductPoints` и `ReserveStock` идемпотентны по `order_id` (уникальный индекс в БД).

**Последствия**: можно безопасно retry'ить вызовы, повторный запрос с тем же `order_id` не приведёт к двойному списанию или резерву.

---

## 12. Что вне MVP — план v2

После запуска MVP и его стабилизации, следующие итерации:

### v1.1 — Уведомления
- `notification-service` (Kafka consumer)
- Топики `notification.send` (универсальный) и `order.status_changed`
- Каналы доставки: email (SMTP), Slack (webhook)

### v1.2 — Фулфилмент
- `fulfillment-service` — интеграция со складом или типографией
- Топик `order.confirmed`, статусы `in_production` → `shipped` → `delivered`
- Webhook для обновлений от поставщика

### v1.3 — Возвраты и история
- `payment-service` как отдельный сервис с полной историей транзакций
- Поток возврата: отмена → возврат баллов → снятие резерва остатков
- UI истории баллов и заказов

### v1.4 — Промокоды и скидки
- `promo-service` для управления промокодами
- Расширение `Order.CreateOrder` опциональным `promo_code`
- Расчёт скидок в `order-service`

### v1.5 — Аналитика и админка
- Отдельный read-model в ClickHouse, populated через Kafka events
- Дашборды для HR/PR: какие товары популярны, по отделам, по периодам
- Админ-UI для управления каталогом и остатками

---

*Merch Store MVP Architecture v1.0 — документ актуален на 26.05.2026*
