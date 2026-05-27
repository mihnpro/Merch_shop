# Merch Store MVP — Архитектурная документация

| | |
|---|---|
| **Версия** | 1.1.0 MVP |
| **Стек** | Go · gRPC · Apache Kafka · Kubernetes · PostgreSQL · Redis · MinIO (S3) |
| **Тип проекта** | Внутренний магазин мерча за корпоративные баллы |
| **Архитектура** | Микросервисная (7 сервисов) |
| **Срок** | 1 месяц |

---

## Содержание

1. [Обзор и скоуп MVP](#1-обзор-и-скоуп-mvp)
2. [Сервисы и ответственности](#2-сервисы-и-ответственности)
3. [Полный сценарий покупки](#3-полный-сценарий-покупки)
4. [Админка — управление и начисление баллов](#4-админка--управление-и-начисление-баллов)
5. [Хранение фото — MinIO и presigned URL](#5-хранение-фото--minio-и-presigned-url)
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

**Merch Store** — внутренняя платформа, где сотрудники могут заказывать корпоративный мерч за баллы. Баллы начисляет HR через админ-панель (за достижения, бонусы, выслугу).

### 1.1 Что входит в MVP

- Каталог товаров с фильтрами по категориям и размерам, с фото из S3
- Корзина с TTL (24 часа)
- Оформление заказа с проверкой остатков и баланса баллов
- Резервирование остатков (защита от двойной продажи)
- Аутентификация через корпоративный SSO (JWT с ролями `user` и `admin`)
- Просмотр истории своих заказов
- **Админ-панель**: управление каталогом, загрузка фото в S3, начисление баллов сотрудникам, просмотр всех заказов, audit-лог

### 1.2 Что **не** входит в MVP

- Уведомления (Slack, email) — статус виден в личном кабинете
- Автоматический фулфилмент — выгрузка заказов в Excel
- Возвраты и сложная история транзакций
- Промокоды, скидки, акции

### 1.3 Ключевые архитектурные принципы

- **gRPC между сервисами** — синхронные вызовы там, где пользователь ждёт ответа
- **Kafka для асинхронности** — одно событие `order.created`
- **DB per service** — каждый сервис со своей PostgreSQL-схемой, Cart использует Redis
- **API Gateway как единая точка входа** для сотрудников
- **Admin Service как изолированный слой** — все админские write-операции проходят только через него, с обязательной записью в audit-log
- **S3 для бинарных файлов** — фото товаров хранятся в MinIO, загрузка через presigned URL напрямую из браузера

---

## 2. Сервисы и ответственности

В MVP **7 сервисов**:

| # | Сервис | Что делает | Хранилище |
|---|---|---|---|
| 1 | **API Gateway** | Принимает запросы от сотрудников, проверяет JWT, проксирует в нужный сервис | — |
| 2 | **Admin Service** | Все админские операции: управление каталогом, начисление баллов, audit | PostgreSQL (audit_log) |
| 3 | **Product Service** | Каталог: список товаров, поиск, детали, photo_key | PostgreSQL |
| 4 | **Cart Service** | Корзина пользователя с TTL 24h | Redis |
| 5 | **Order Service** | Создание заказа, история, статусы | PostgreSQL |
| 6 | **User Service** | Профиль сотрудника, баланс баллов, начисление/списание | PostgreSQL |
| 7 | **Inventory Service** | Остатки товаров, резервирование при заказе | PostgreSQL |

Плюс инфраструктурный компонент: **MinIO** — S3-совместимое хранилище для фото товаров.

### 2.1 Кто кого вызывает

**Пользовательские потоки (через Gateway):**

| Источник | Цель | Метод | Зачем |
|---|---|---|---|
| Gateway | все сервисы | различные | Проксирование запросов сотрудника |
| Cart | Product | `GetProduct` | Получить цену и название |
| Cart | Inventory | `CheckStock` | Проверить наличие при добавлении |
| Order | Cart | `GetCart`, `ClearCart` | Взять корзину и очистить |
| Order | User | `GetBalance`, `DeductPoints` | Проверить и списать баллы |
| Order | Kafka (producer) | `order.created` | Публикация события |
| Inventory | Kafka (consumer) | `order.created` | Резервирование остатков |

**Админские потоки (через Admin Service):**

| Источник | Цель | Метод | Зачем |
|---|---|---|---|
| Admin | Product | `CreateProduct`, `UpdateProduct`, `DeactivateProduct` | Управление каталогом |
| Admin | Inventory | `AdjustStock` | Изменение остатков |
| Admin | User | `GrantPoints` | Начисление баллов сотруднику |
| Admin | Order | `ListAllOrders`, `ExportOrders` | Просмотр и выгрузка всех заказов |
| Admin | MinIO | `PresignPutObject` | Генерация ссылок для загрузки фото |

Важно: `Admin Service` сам не хранит товары или баллы. Он лишь проверяет роль, пишет в `audit_log` и вызывает соответствующий доменный сервис.

---

## 3. Полный сценарий покупки

### Шаг 1. Авторизация
Сотрудник заходит, проходит через SSO, получает JWT с ролью `user`.

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
    Cart сохраняет в Redis: cart:{user_id} → JSON

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

### 4.1 Почему отдельный сервис

`admin-service` выделен в отдельный сервис, чтобы:
- **Изолировать опасные операции.** Методы вроде `GrantPoints` потенциально могут привести к злоупотреблению. Их код полностью изолирован от пользовательских путей
- **Иметь единый audit-log.** Каждое админское действие записывается в `audit_log` ДО вызова доменного сервиса
- **Не дублировать логику.** Сам admin-service не хранит товары, баллы или заказы — он вызывает Product, User, Inventory, Order через gRPC

### 4.2 gRPC API

```protobuf
syntax = "proto3";
package admin.v1;

service AdminService {
  rpc CreateProduct(CreateProductRequest) returns (Product);
  rpc UpdateProduct(UpdateProductRequest) returns (Product);
  rpc DeactivateProduct(DeactivateProductRequest) returns (Empty);
  
  rpc AdjustStock(AdjustStockRequest) returns (StockInfo);
  
  rpc GrantPoints(GrantPointsRequest) returns (Balance);
  
  rpc ListAllOrders(ListAllOrdersRequest) returns (ListOrdersResponse);
  rpc ExportOrders(ExportOrdersRequest) returns (ExportResponse);
  
  rpc GetUploadURL(GetUploadURLRequest) returns (UploadURL);
  
  rpc GetAuditLog(GetAuditLogRequest) returns (AuditLogResponse);
}

message CreateProductRequest {
  string name = 1;
  string description = 2;
  int64 price_points = 3;
  string category = 4;
  repeated string sizes = 5;
  string photo_key = 6;
}

message GrantPointsRequest {
  string user_id = 1;
  int64 amount = 2;
  string reason = 3;
}

message GetUploadURLRequest {
  string filename = 1;
  string content_type = 2;
}

message UploadURL {
  string upload_url = 1;
  string photo_key = 2;
  google.protobuf.Timestamp expires_at = 3;
}
```

### 4.3 Реализация GrantPoints с audit

```go
func (s *AdminService) GrantPoints(
    ctx context.Context, req *pb.GrantPointsRequest,
) (*pb.Balance, error) {
    adminID := ctx.Value("admin_id").(string)

    auditID := uuid.New()
    if err := s.audit.Log(ctx, AuditEntry{
        ID:         auditID,
        AdminID:    adminID,
        Action:     "points.grant",
        TargetType: "user",
        TargetID:   req.UserId,
        Payload:    map[string]any{"amount": req.Amount, "reason": req.Reason},
    }); err != nil {
        return nil, status.Error(codes.Internal, "failed to write audit log")
    }

    balance, err := s.userClient.AddPoints(ctx, &userpb.AddPointsRequest{
        UserId:    req.UserId,
        Amount:    req.Amount,
        Reason:    req.Reason,
        AuditId:   auditID.String(),
    })
    if err != nil {
        s.audit.MarkFailed(ctx, auditID, err.Error())
        return nil, err
    }

    return &pb.Balance{Points: balance.Points}, nil
}
```

### 4.4 Сценарии работы админа

**Добавление товара:**
1. Админ нажимает «Добавить товар», выбирает фото
2. Браузер запрашивает у admin-service presigned URL для загрузки в S3
3. Браузер загружает файл напрямую в S3
4. Браузер вызывает `CreateProduct` с `photo_key`
5. admin-service пишет audit и вызывает `Product.CreateProduct`

**Начисление баллов:**
1. Админ открывает профиль сотрудника, вводит сумму и причину
2. admin-service пишет audit и вызывает `User.AddPoints`
3. Возвращается новый баланс

**Просмотр заказов:**
1. admin-service вызывает `Order.ListAllOrders` с фильтрами
2. По кнопке «Экспорт» — формирует Excel-файл

---

## 5. Хранение фото — MinIO и presigned URL

### 5.1 Почему MinIO

MinIO — S3-совместимое хранилище с открытым исходным кодом. Подходит идеально:
- Локально разворачивается через `docker-compose` одной строкой
- В Kubernetes — через StatefulSet
- Тот же API, что у AWS S3 / Yandex Object Storage — при переезде в облако код не меняется
- Используется тот же Go-клиент `minio-go`

### 5.2 Паттерн загрузки — presigned URL

Файлы **не** проходят через сервисы. Браузер загружает напрямую в S3, используя временную подписанную ссылку:

```
1. Браузер → admin-service: GetUploadURL(filename, content_type)
2. admin-service → MinIO: PresignPutObject() — генерация ссылки на 5 минут
3. admin-service → Браузер: { upload_url, photo_key }
4. Браузер → MinIO: PUT файл по upload_url (напрямую!)
5. Браузер → admin-service: CreateProduct({ ..., photo_key })
```

**Преимущества:**
- Сервисы не буферизуют мегабайты в памяти
- Не нужно настраивать большие `client_max_body_size` в Ingress
- Загрузка масштабируется отдельно от backend

### 5.3 Реализация GetUploadURL

```go
import "github.com/minio/minio-go/v7"

func (s *AdminService) GetUploadURL(
    ctx context.Context, req *pb.GetUploadURLRequest,
) (*pb.UploadURL, error) {
    if !isAllowedContentType(req.ContentType) {
        return nil, status.Error(codes.InvalidArgument, "only jpeg, png, webp allowed")
    }

    ext := mimeToExt(req.ContentType)
    key := fmt.Sprintf("products/%s%s", uuid.New(), ext)

    url, err := s.s3.PresignedPutObject(
        ctx,
        "merch-products",
        key,
        5*time.Minute,
    )
    if err != nil {
        return nil, status.Error(codes.Internal, "failed to generate URL")
    }

    return &pb.UploadURL{
        UploadUrl: url.String(),
        PhotoKey:  key,
        ExpiresAt: timestamppb.New(time.Now().Add(5 * time.Minute)),
    }, nil
}

func isAllowedContentType(ct string) bool {
    return ct == "image/jpeg" || ct == "image/png" || ct == "image/webp"
}
```

### 5.4 Bucket'ы и доступ

| Bucket | Доступ | Назначение |
|---|---|---|
| `merch-products` | public-read | Фото товаров |

Bucket `merch-products` настроен на публичное чтение — любой пользователь может посмотреть фото товара по прямой ссылке. Это упрощает фронт: ему не нужны presigned GET URL для каждого товара.

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

`Create/Update/Deactivate` доступны **только** для `admin-service` — это обеспечивается NetworkPolicy в Kubernetes.

### 6.2 UserService

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc GetBalance(GetBalanceRequest) returns (Balance);
  rpc DeductPoints(DeductPointsRequest) returns (Balance);
  rpc AddPoints(AddPointsRequest) returns (Balance);
}

message AddPointsRequest {
  string user_id = 1;
  int64 amount = 2;
  string reason = 3;
  string audit_id = 4;
}
```

`AddPoints` использует `audit_id` для идемпотентности — повторный вызов с тем же ID не приведёт к двойному начислению.

### 6.3 InventoryService

```protobuf
service InventoryService {
  rpc CheckStock(CheckStockRequest) returns (StockInfo);
  rpc ReserveStock(ReserveStockRequest) returns (Reservation);
  rpc ReleaseReserve(ReleaseReserveRequest) returns (Empty);
  rpc AdjustStock(AdjustStockRequest) returns (StockInfo);
}

message AdjustStockRequest {
  string product_id = 1;
  string size = 2;
  int32 delta = 3;
  string audit_id = 4;
}
```

(остальные сервисы — Cart, Order — без изменений из предыдущей версии)

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

### 8.2 Cart Service (Redis)

```
Key:   cart:{user_id}
Type:  String (JSON)
TTL:   86400 (24 часа)
```

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
CREATE TABLE users (
    id          UUID PRIMARY KEY,
    email       TEXT UNIQUE NOT NULL,
    full_name   TEXT NOT NULL,
    department  TEXT
);

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
    audit_id    UUID,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(order_id, user_id),
    UNIQUE(audit_id)
);
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

### 8.6 Admin Service — audit_log

```sql
CREATE TABLE audit_log (
    id          UUID PRIMARY KEY,
    admin_id    UUID NOT NULL,
    action      TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id   TEXT NOT NULL,
    payload     JSONB NOT NULL,
    status      TEXT DEFAULT 'success',
    error       TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_admin   ON audit_log(admin_id, created_at DESC);
CREATE INDEX idx_audit_target  ON audit_log(target_type, target_id);
CREATE INDEX idx_audit_action  ON audit_log(action, created_at DESC);
```

Примеры записей:

| action | target_type | target_id | payload |
|---|---|---|---|
| `product.create` | `product` | `p123` | `{"name": "T-shirt", "price": 500}` |
| `points.grant` | `user` | `u456` | `{"amount": 1000, "reason": "Q4 bonus"}` |
| `stock.adjust` | `product` | `p123` | `{"size": "L", "delta": 50}` |

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
│   ├── admin/v1/
│   └── events/v1/
│
├── services/
│   ├── api-gateway/
│   ├── admin-service/
│   ├── product-service/
│   ├── cart-service/
│   ├── order-service/
│   ├── user-service/
│   └── inventory-service/
│
├── pkg/
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
| admin-service | 1 | 100m | 128Mi | 2 |
| product-service | 2 | 100m | 128Mi | 6 |
| cart-service | 2 | 100m | 128Mi | 6 |
| order-service | 2 | 150m | 256Mi | 8 |
| user-service | 2 | 100m | 128Mi | 4 |
| inventory-service | 2 | 100m | 128Mi | 6 |
| minio | 1 | 200m | 512Mi | — |

### 10.3 NetworkPolicy для admin-service

Критически важно ограничить, кто может вызывать admin-service:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: admin-service-policy
  namespace: merch-prod
spec:
  podSelector:
    matchLabels:
      app: admin-service
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: admin-gateway
```

Точно так же — write-методы Product, User, Inventory должны быть доступны только из admin-service:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: product-write-policy
  namespace: merch-prod
spec:
  podSelector:
    matchLabels:
      app: product-service
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: api-gateway
    - podSelector:
        matchLabels:
          app: admin-service
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
- `merch_points_granted_total{admin_id}`
- `merch_products_created_total{admin_id}`

### 11.2 Audit-аналитика

Поверх таблицы `audit_log` можно строить дашборды:
- Кто из админов сколько баллов начислил за месяц
- Топ-10 наиболее активных админов
- Аномалии — например, разовое начисление > 10000 баллов

---

## 12. План реализации по неделям

### Неделя 1 — Фундамент и листовые сервисы

- День 1-2: monorepo, buf, Makefile, docker-compose (Postgres × 6, Redis, Kafka, MinIO)
- День 3-4: product-service (proto, БД, gRPC)
- День 5-6: user-service (proto, БД с points_transactions)
- День 7: inventory-service (proto, БД, атомарное резервирование)

### Неделя 2 — Корзина, заказы, админка

- День 8-9: cart-service (Redis, gRPC-клиенты)
- День 10-11: order-service (оркестрация, Outbox)
- День 12: Kafka producer/consumer
- День 13: api-gateway (JWT, маршрутизация)
- День 14: admin-service (CRUD каталога, GrantPoints, audit_log, MinIO presigned URL)

### Неделя 3 — Kubernetes

- День 15-16: Dockerfile (multi-stage, distroless)
- День 17-18: K8s манифесты, Strimzi Kafka, MinIO StatefulSet
- День 19: NetworkPolicy для admin-service и write-методов
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

### ADR-004: Redis для корзины
TTL, O(1) операции, нет миграций.

### ADR-005: Choreography Saga
Хореография через события вместо центрального оркестратора.

### ADR-006: Outbox для надёжной публикации в Kafka
Гарантия at-least-once доставки.

### ADR-007: Идемпотентность через order_id и audit_id
Защита от повторной обработки.

### ADR-008: Отдельный admin-service вместо админских эндпоинтов в Gateway

**Контекст**: админские операции (начисление баллов, изменение каталога) опасны — ошибка в проверке прав = катастрофа.

**Решение**: вынести админскую функциональность в отдельный сервис с собственной БД (audit_log) и NetworkPolicy.

**Альтернативы**:
- Класть админские методы в Gateway — отвергнуто: смешивание ответственностей, общий код, риск ошибки в маршрутизации

**Последствия**: +1 сервис, но получаем изоляцию, audit, чёткую границу безопасности.

### ADR-009: MinIO (S3) для фото товаров

**Контекст**: фото товаров нельзя хранить в БД или файловой системе подов.

**Решение**: MinIO как S3-совместимое хранилище, presigned URL для загрузки напрямую из браузера.

**Альтернативы**:
- Загрузка через сервис — отвергнуто: сервисы буферизуют файлы, нагрузка на CPU/RAM
- AWS S3 / Yandex Object Storage — отвергнуто для MVP: лишняя зависимость от внешнего провайдера, но при необходимости можно переехать без изменений кода

**Последствия**: один Go-клиент (`minio-go`), легко поднимается локально и в k8s, при переезде в облако код не меняется.

---

*Merch Store MVP Architecture v1.1*