# Merch Store — Диаграммы

---

## 1. Архитектура сервисов

```mermaid
flowchart TB
    Client([Браузер])
    GW["API Gateway\nJWT · роль"]
    KF[[Kafka]]
    MN[(MinIO S3)]

    subgraph ps_box [Product Service]
        PS[Product]
        PS_DB[(PostgreSQL)]
    end

    subgraph cs_box [Cart Service]
        CS[Cart]
        CS_DB[(PostgreSQL)]
        CS_RD[(Redis)]
    end

    subgraph os_box [Order Service]
        OS[Order]
        OS_DB[(PostgreSQL)]
    end

    subgraph us_box [User Service]
        US[User]
        US_DB[(PostgreSQL)]
        US_RD[(Redis)]
    end

    subgraph is_box [Inventory Service]
        IS[Inventory]
        IS_DB[(PostgreSQL)]
    end

    subgraph ms_box [Media Service]
        MS[Media]
    end

    Client --> GW
    GW --> PS & CS & OS & US & IS & MS

    CS --> PS & IS
    OS --> CS & US
    OS --> KF --> IS
    MS --> MN

    PS --- PS_DB
    CS --- CS_DB & CS_RD
    OS --- OS_DB
    US --- US_DB & US_RD
    IS --- IS_DB
```

---

## 2. Sequence: Добавление товара в корзину

```mermaid
sequenceDiagram
    actor Пользователь
    participant GW as API Gateway
    participant Cart as Cart Service
    participant Inv as Inventory Service
    participant Prod as Product Service

    Пользователь->>GW: AddItem(product_id, size, qty)
    GW->>Cart: проксирует (JWT ✓)

    Cart->>Inv: CheckStock(product_id, size)

    alt нет в наличии
        Inv-->>Cart: OUT_OF_STOCK
        Cart-->>Пользователь: ошибка
    else есть
        Inv-->>Cart: available
        Cart->>Prod: GetProduct(product_id)
        Prod-->>Cart: цена + название (снапшот)
        Cart->>Cart: UPSERT в PostgreSQL
        Cart->>Cart: обновить Redis-кеш
        Cart-->>Пользователь: корзина обновлена
    end
```

---

## 3. Sequence: Оформление заказа

```mermaid
sequenceDiagram
    actor Пользователь
    participant GW as API Gateway
    participant OS as Order Service
    participant CS as Cart Service
    participant US as User Service
    participant KF as Kafka
    participant IS as Inventory Service

    Пользователь->>GW: CreateOrder(address)
    GW->>OS: проксирует (JWT ✓)

    OS->>CS: GetCart(user_id)
    CS-->>OS: items[]

    OS->>US: GetBalance(user_id)
    US-->>OS: balance

    alt баллов не хватает
        OS-->>Пользователь: INSUFFICIENT_POINTS
    else баллов достаточно
        OS->>US: DeductPoints(order_id, amount)
        US-->>OS: ok

        OS->>OS: сохранить заказ — pending
        OS->>KF: order.created (Outbox)
        OS->>CS: ClearCart(user_id)
        OS-->>Пользователь: Order{id, status: pending}

        KF->>IS: order.created
        IS->>IS: зарезервировать остатки
        IS->>OS: UpdateStatus(confirmed)
    end
```

---

## 4. Sequence: Загрузка фото и создание товара (admin)

```mermaid
sequenceDiagram
    actor Админ
    participant GW as API Gateway
    participant MS as Media Service
    participant MN as MinIO
    participant PS as Product Service

    Админ->>GW: UploadPhoto(filename, content_type, bytes)
    GW->>MS: проксирует (role=admin ✓)
    MS->>MS: валидация content-type и размера (≤ 5 МБ)
    MS->>MS: ключ products/{uuid}.{ext}
    MS->>MN: PutObject(key, stream)
    MN-->>MS: 200 OK
    MS-->>Админ: {photo_key}

    Админ->>GW: CreateProduct({..., photo_key})
    GW->>PS: проксирует (role=admin ✓)
    PS->>PS: сохранить в БД
    PS-->>Админ: Product{id, ...}
```

---

## 5. Sequence: Вход и выдача токенов

```mermaid
sequenceDiagram
    actor Пользователь
    participant GW as API Gateway
    participant US as User Service
    participant RD as Redis

    Пользователь->>GW: Login(login, password)
    GW->>US: проксирует

    US->>US: найти пользователя, проверить bcrypt

    alt неверные данные
        US-->>Пользователь: 401 Unauthorized
    else успех
        US->>US: сгенерировать access_token (15 мин) + refresh_token (30 дней)
        US->>RD: SET refresh:{jti} → user_id  EX 30d
        US-->>Пользователь: {access_token, refresh_token, user}
    end
```

---

## 6. Sequence: Обновление access-токена

```mermaid
sequenceDiagram
    actor Клиент
    participant GW as API Gateway
    participant US as User Service
    participant RD as Redis

    Клиент->>GW: любой запрос
    GW-->>Клиент: 401 (access_token истёк)

    Клиент->>GW: Refresh(refresh_token)
    GW->>US: проксирует

    US->>RD: GET refresh:{jti}

    alt токен не найден или отозван
        RD-->>US: nil
        US-->>Клиент: 401 — нужно войти заново
    else токен валиден
        RD-->>US: user_id
        US->>US: выпустить новую пару токенов
        US->>RD: DEL старый jti · SET новый jti
        US-->>Клиент: {access_token, refresh_token}
        Клиент->>GW: повторить оригинальный запрос
    end
```

---

## 7. Sequence: Отмена заказа администратором (компенсация)

```mermaid
sequenceDiagram
    actor Админ
    participant GW as API Gateway
    participant OS as Order Service
    participant US as User Service
    participant IS as Inventory Service

    Админ->>GW: UpdateStatus(order_id, cancelled, reason)
    GW->>OS: проксирует (role=admin ✓)

    OS->>OS: проверить допустимость перехода статуса

    OS->>US: AddPoints(user_id, total_points, order_id)
    US->>US: вернуть баллы (идемпотентно по order_id)
    US-->>OS: ok

    OS->>IS: ReleaseReserve(order_id)
    IS->>IS: снять резерв остатков
    IS-->>OS: ok

    OS->>OS: статус → cancelled
    OS-->>Админ: Order{status: cancelled}
```

---

## 8. Sequence: Начисление баллов (admin)

```mermaid
sequenceDiagram
    actor Админ
    participant GW as API Gateway
    participant US as User Service

    Админ->>GW: GrantPoints(user_id, amount, reason)
    GW->>US: проксирует (role=admin ✓)

    alt пользователь не найден
        US-->>Админ: NOT_FOUND
    else
        US->>US: INSERT points_transactions
        US->>US: UPDATE points_balance + amount
        US-->>Админ: Balance{points}
    end
```

---

## 9. Состояния заказа

```mermaid
stateDiagram-v2
    direction LR

    [*] --> pending : CreateOrder

    pending --> confirmed : Inventory зарезервировал остатки
    pending --> cancelled : Inventory — нет в наличии

    confirmed --> ready_to_pickup : Админ — готов к выдаче
    confirmed --> cancelled : Админ — отмена

    ready_to_pickup --> delivered : Админ — получен

    delivered --> [*]
    cancelled --> [*]

    note right of cancelled : Баллы возвращаются\nРезерв снимается
```

---

## 10. Схема сущностей по сервисам

Каждый сервис имеет собственную изолированную базу данных. Связи между сервисами — логические, на уровне приложения (по ID).

### User Service DB

```mermaid
erDiagram
    USERS {
        uuid id PK
        text login
        text password_hash
        enum role
        enum status
    }
    POINTS_BALANCE {
        uuid user_id PK
        bigint points
    }
    POINTS_TRANSACTIONS {
        uuid id PK
        uuid user_id FK
        bigint amount
        text reason
        uuid order_id
    }

    USERS ||--|| POINTS_BALANCE : ""
    USERS ||--o{ POINTS_TRANSACTIONS : ""
```

### Product Service DB

```mermaid
erDiagram
    PRODUCTS {
        uuid id PK
        text name
        bigint price_points
        text category
        text[] sizes
        text photo_key
        bool active
    }
```

### Cart Service DB

```mermaid
erDiagram
    CARTS {
        uuid user_id PK
        timestamptz updated_at
    }
    CART_ITEMS {
        uuid cart_user_id FK
        uuid product_id
        text size
        int qty
        bigint price_points
        text product_name
    }

    CARTS ||--o{ CART_ITEMS : ""
```

### Order Service DB

```mermaid
erDiagram
    ORDERS {
        uuid id PK
        uuid user_id
        bigint total_points
        enum status
        text delivery_address
    }
    ORDER_ITEMS {
        uuid id PK
        uuid order_id FK
        uuid product_id
        text size
        int qty
        bigint price_points
    }
    OUTBOX {
        uuid id PK
        text event_type
        bytea payload
        timestamptz sent_at
    }

    ORDERS ||--o{ ORDER_ITEMS : ""
    ORDERS ||--o| OUTBOX : ""
```

### Inventory Service DB

```mermaid
erDiagram
    STOCK {
        uuid product_id PK
        text size PK
        int available
    }
    RESERVATIONS {
        uuid id PK
        uuid order_id
        jsonb items
    }
```
