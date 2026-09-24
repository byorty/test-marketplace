# Delivery Service — Specification Kit

## 1. User Stories

### US-1: Создание доставки (employee)

**Как** работник (employee),
**Я хочу** создать доставку для заказа,
**Чтобы** инициировать процесс доставки товара клиенту.

### US-2: Просмотр доставки (customer, employee)

**Как** клиент (customer),
**Я хочу** посмотреть статус и детали своей доставки,
**Чтобы** знать, когда и куда забирать заказ.

**Как** работник (employee),
**Я хочу** посмотреть детали любой доставки,
**Чтобы** отслеживать и управлять процессом доставки.

### US-3: Получение QR-кода (customer)

**Как** клиент (customer),
**Я хочу** получить QR-код для получения заказа,
**Чтобы** предъявить его на пункте выдачи.

### US-4: Верификация QR-кода и завершение доставки (employee)

**Как** работник (employee),
**Я хочу** отсканировать и верифицировать QR-код клиента,
**Чтобы** подтвердить выдачу заказа и завершить доставку.

### US-5: Перенос доставки (employee)

**Как** работник (employee),
**Я хочу** перенести дату доставки на другой день,
**Чтобы** скорректировать плановую дату при изменении обстоятельств.

### US-6: Изменение статуса доставки (employee)

**Как** работник (employee),
**Я хочу** перевести доставку в следующий статус,
**Чтобы** отражать реальное состояние доставки в системе.

### US-7: Информация об опоздании (customer, employee)

**Как** клиент (customer),
**Я хочу** видеть, что доставка просрочена,
**Чтобы** знать о задержке.

**Как** работник (employee),
**Я хочу** видеть просроченные доставки,
**Чтобы** принимать меры.

---

## 2. Acceptance Criteria

### AC-1: Создание доставки

- Сервер создаёт доставку с статусом `PENDING`.
- Если заказ не найден или не в статусе `PAID`, возвращается ошибка.
- Если доставка для данного заказа уже существует, возвращается ошибка.
- Возвращается полная информация о доставке, включая ID.
- Плановая дата доставки (`estimated_delivery_date`) обязательна.
- Адрес пункта выдачи (`pickup_address`) обязателен.

### AC-2: Просмотр доставки (customer)

- Клиент видит только свои доставки (по `user_id` из JWT-claims).
- Возвращается: статус, плановая дата, флаг опоздания, адрес пункта выдачи.
- При запросе чужой доставки возвращается 403 Forbidden.

### AC-2a: Просмотр доставки (employee)

- Работник видит любую доставку.
- Возвращается полная информация, включая историю переносов.

### AC-3: Получение QR-кода

- QR-код доступен только для доставки в статусе `READY_FOR_PICKUP`.
- QR-код доступен только владельцу доставки (customer).
- QR-код содержит криптографически подписанный токен.
- QR-код имеет срок действия (настраиваемый, по умолчанию 24 часа).
- Повторный запрос генерирует новый QR-код (старый инвалидируется).

### AC-4: Верификация QR-кода

- Верификация доступна только работнику (employee).
- При успешной верификации статус доставки переходит в `DELIVERED`.
- При невалидном QR-коде возвращается ошибка.
- При просроченном QR-коде возвращается ошибка.
- При QR-коде для доставки не в статусе `READY_FOR_PICKUP` возвращается ошибка.
- QR-код одноразовый: повторная верификация того же QR невозможна.

### AC-5: Перенос доставки

- Перенос возможен только если доставка в статусе `PENDING` или `IN_TRANSIT`.
- Предыдущая дата, новая дата и время изменения сохраняются.
- Причина переноса опциональна.
- Плановая дата доставки обновляется.
- Флаг опоздания пересчитывается.

### AC-6: Изменение статуса

- Переход допустим только по разрешённым путям (см. State Machine).
- При недопустимом переходе возвращается 400.
- При переводе в `DELIVERED` через верификацию QR, статус и QR проверяются атомарно.

### AC-7: Опоздание

- Флаг `is_late` автоматически устанавливается, когда `now > estimated_delivery_date` и статус не является терминальным (`DELIVERED`, `CANCELLED`).
- Флаг пересчитывается при каждом запросе и при изменении `estimated_delivery_date`.
- При переносе на более позднюю дату флаг может быть снят.

---

## 3. Business Rules

### BR-1: Уникальность доставки на заказ

Для одного заказа может существовать только одна активная доставка (не `CANCELLED` и не `DELIVERED`).

### BR-2: Заказ должен существовать и быть оплачен

Создание доставки возможно только для заказа в статусе `PAID`. Delivery-service проверяет статус через вызов order-service.

### BR-3: Принадлежность доставки клиенту

Клиент может просматривать и получать QR только для доставок, принадлежащих его заказам. Проверка `user_id` из JWT-claims совпадает с `user_id` заказа.

### BR-4: QR-код доступен только в статусе READY_FOR_PICKUP

Генерация QR-кода возможна только когда доставка находится в статусе `READY_FOR_PICKUP`.

### BR-5: Терминальные статусы неизменны

Из статусов `DELIVERED` и `CANCELLED` невозможны переходы в другие статусы.

### BR-6: История переносов immutable

Записи о переносах нельзя изменять или удалять. Каждая запись содержит предыдущую дату, новую дату, время изменения и опциональную причину.

### BR-7: Опоздание определяется автоматически

Флаг `is_late` не устанавливается вручную. Он вычисляется на основе текущего времени и `estimated_delivery_date`.

### BR-8: Перенос обновляет estimated_delivery_date

При переносе доставки поле `estimated_delivery_date` обновляется на новую дату, а старое значение сохраняется в истории переносов.

### BR-9: Перенос невозможен для терминальных статусов

Перенос доставки в статусе `DELIVERED` или `CANCELLED` невозможен.

---

## 4. Domain Model

### 4.1 Delivery

```
Delivery
├── id                  UUID, PK
├── order_id            UUID, NOT NULL, UNIQUE (одна активная доставка на заказ)
├── status              VARCHAR(30), NOT NULL
├── pickup_address      TEXT, NOT NULL
├── estimated_delivery_date  TIMESTAMP, NOT NULL
├── is_late             BOOLEAN, NOT NULL, DEFAULT FALSE
├── qr_nonce            UUID, NULLABLE (для инвалидации предыдущего QR)
├── created_at          TIMESTAMP, NOT NULL
└── updated_at          TIMESTAMP, NOT NULL
```

**Статусы (DeliveryStatus):**

| Статус             | Описание                            |
| ------------------ | ----------------------------------- |
| `PENDING`          | Создана, ожидает назначения курьера |
| `IN_TRANSIT`       | В пути к пункту выдачи              |
| `READY_FOR_PICKUP` | Прибыла, ожидает клиента            |
| `DELIVERED`        | Клиент забрал заказ                 |
| `CANCELLED`        | Доставка отменена                   |

### 4.2 DeliveryReschedule

```
DeliveryReschedule
├── id              UUID, PK
├── delivery_id     UUID, NOT NULL, FK → deliveries(id)
├── previous_date   TIMESTAMP, NOT NULL
├── new_date        TIMESTAMP, NOT NULL
├── reason          TEXT, NULLABLE
└── created_at      TIMESTAMP, NOT NULL
```

### 4.3 QR Token (не персистентная модель)

```
QRToken (передаётся в QR-коде, не хранится в БД целиком)
├── delivery_id     UUID
├── order_id        UUID
├── user_id         UUID
├── nonce           UUID (совпадает с delivery.qr_nonce)
├── iat             int64 (unix timestamp, момент выдачи)
└── exp             int64 (unix timestamp, срок действия)
```

**Подпись:** HMAC-SHA256 над JSON-сериализованным payload с использованием серверного секрета.

**Формат токена:** `base64url(JSON(payload)) + "." + base64url(HMAC-SHA256(payload, secret_key))`

### 4.4 Интерфейсы (Go)

```go
type DeliveryRepository interface {
    Create(ctx context.Context, delivery *Delivery) error
    GetByID(ctx context.Context, id uuid.UUID) (*Delivery, error)
    GetByOrderID(ctx context.Context, orderID uuid.UUID) (*Delivery, error)
    List(ctx context.Context, filter DeliveryListFilter) (*DeliveryList, error)
    Update(ctx context.Context, delivery *Delivery) error
    CreateReschedule(ctx context.Context, reschedule *DeliveryReschedule) error
    ListReschedules(ctx context.Context, deliveryID uuid.UUID) ([]DeliveryReschedule, error)
}

type DeliveryService interface {
    Create(ctx context.Context, input *DeliveryCreateInput) (*Delivery, error)
    GetByID(ctx context.Context, userID, id uuid.UUID) (*Delivery, error)
    List(ctx context.Context, userID uuid.UUID, role string, filter DeliveryListFilter) (*DeliveryList, error)
    Reschedule(ctx context.Context, deliveryID uuid.UUID, input *RescheduleInput) (*Delivery, error)
    UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status DeliveryStatus) (*Delivery, error)
    GetQR(ctx context.Context, userID, deliveryID uuid.UUID) (*QRTokenResponse, error)
    VerifyQR(ctx context.Context, token string) (*Delivery, error)
}

type OrderClient interface {
    GetOrderByID(ctx context.Context, orderID uuid.UUID) (*OrderResponse, error)
    UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error
}

type Notifier interface {
    NotifyDeliveryLate(ctx context.Context, delivery *Delivery) error
    NotifyDeliveryStatusChanged(ctx context.Context, delivery *Delivery, oldStatus DeliveryStatus) error
}
```

---

## 5. Delivery State Machine

### 5.1 Диаграмма переходов

```
                    ┌───────────┐
         create     │           │  start transit
         ──────────►│  PENDING  │──────────────┐
                    │           │               │
                    └─────┬─────┘               ▼
                          │               ┌────────────┐
                          │ cancel        │            │
                          ▼               │ IN_TRANSIT │
                    ┌───────────┐        │            │
                    │           │        └─────┬──────┘
                    │ CANCELLED │              │
                    │           │◄─────────────┘ cancel
                    └───────────┘              │
                                               │ arrive at pickup
                                               ▼
                                        ┌─────────────────┐
                                ┌───────│                 │
                                │       │ READY_FOR_PICKUP │
                                │       │                 │
                          cancel│       └────┬────────────┘
                                │            │ verify QR
                                │            ▼
                          ┌───────────┐ ┌───────────┐
                          │           │ │           │
                          │ CANCELLED │ │ DELIVERED │
                          │           │ │           │
                          └───────────┘ └───────────┘
```

### 5.2 Таблица допустимых переходов

| Из                 | В                  | Условие                           |
| ------------------ | ------------------ | --------------------------------- |
| `PENDING`          | `IN_TRANSIT`       | Работник начинает доставку        |
| `PENDING`          | `CANCELLED`        | Работник отменяет                 |
| `IN_TRANSIT`       | `READY_FOR_PICKUP` | Доставка прибыла на пункт выдачи  |
| `IN_TRANSIT`       | `CANCELLED`        | Работник отменяет                 |
| `READY_FOR_PICKUP` | `DELIVERED`        | QR-код верифицирован работником   |
| `READY_FOR_PICKUP` | `CANCELLED`        | Работник отменяет (исключительно) |

**Терминальные статусы:** `DELIVERED`, `CANCELLED`. Переходы из них невозможны.

**Перенос (Reschedule):** допустим из `PENDING` и `IN_TRANSIT`. Не изменяет статус, только `estimated_delivery_date`.

### 5.3 Обновление статуса заказа

При переходах доставки статус заказа в order-service обновляется:

| Переход доставки                 | Статус заказа (order-service)        |
| -------------------------------- | ------------------------------------ |
| `PENDING` → `IN_TRANSIT`         | `PAID` → `DELIVERING`                |
| `READY_FOR_PICKUP` → `DELIVERED` | `DELIVERING` → `DELIVERED`           |
| Любой → `CANCELLED`              | `DELIVERING` → `PAID` (или остаётся) |

**Открытый вопрос:** точное поведение при отмене доставки требует бизнес-решения.

---

## 6. Authorization Matrix

### 6.1 Casbin-ресурсы и действия

```go
const (
    ResourceDelivery rbac.Resource = "delivery"

    ActionCreate    rbac.Action = "create"
    ActionView      rbac.Action = "view"
    ActionReschedule rbac.Action = "reschedule"
    ActionUpdateStatus rbac.Action = "update_status"
    ActionVerifyQR  rbac.Action = "verify_qr"
)
```

### 6.2 Матрица доступа

| Роль     | create | view | reschedule | update_status | verify_qr | get_qr |
| -------- | ------ | ---- | ---------- | ------------- | --------- | ------ |
| customer | ✗      | ✓\*  | ✗          | ✗             | ✗         | ✓\*    |
| employee | ✓      | ✓    | ✓          | ✓             | ✓         | ✗      |

\* _customer имеет доступ только к собственным доставкам_

### 6.3 Casbin-политика (дополнение к `common/rbac/policy.csv`)

```csv
p, customer, delivery, view
p, employee, delivery, create
p, employee, delivery, view
p, employee, delivery, reschedule
p, employee, delivery, update_status
p, employee, delivery, verify_qr
```

### 6.4 Дополнительная проверка принадлежности

Помимо Casbin, для `customer` выполняется проверка:

- `view`: `delivery.order_id → order.user_id == claims.UserID`
- `get_qr`: `delivery.order_id → order.user_id == claims.UserID`

Для `employee` принадлежность не проверяется — работник имеет доступ ко всем доставкам.

---

## 7. API Requirements

### 7.1 Endpoints

| Метод | Путь                        | Описание                   | Auth   | RBAC                    |
| ----- | --------------------------- | -------------------------- | ------ | ----------------------- |
| POST  | /deliveries                 | Создание доставки          | Bearer | delivery, create        |
| GET   | /deliveries/{id}            | Получение доставки по ID   | Bearer | delivery, view          |
| GET   | /deliverages                | Список доставок (employee) | Bearer | delivery, view          |
| PATCH | /deliveries/{id}/reschedule | Перенос доставки           | Bearer | delivery, reschedule    |
| PATCH | /deliveries/{id}/status     | Изменение статуса          | Bearer | delivery, update_status |
| GET   | /deliveries/{id}/qr         | Получение QR-кода          | Bearer | delivery, view\*        |
| POST  | /deliveries/verify-qr       | Верификация QR-кода        | Bearer | delivery, verify_qr     |

\* _GET /deliveries/{id}/qr требует RBAC `delivery, view` + проверку принадлежности для customer_

### 7.2 OpenAPI 3.1 — Схемы

```yaml
components:
  schemas:
    DeliveryStatus:
      type: string
      enum:
        - PENDING
        - IN_TRANSIT
        - READY_FOR_PICKUP
        - DELIVERED
        - CANCELLED

    DeliveryCreateRequest:
      type: object
      required:
        - order_id
        - pickup_address
        - estimated_delivery_date
      properties:
        order_id:
          type: string
          format: uuid
        pickup_address:
          type: string
          minLength: 1
          maxLength: 500
        estimated_delivery_date:
          type: string
          format: date-time

    DeliveryRescheduleRequest:
      type: object
      required:
        - new_date
      properties:
        new_date:
          type: string
          format: date-time
        reason:
          type: string
          maxLength: 1000

    DeliveryStatusUpdateRequest:
      type: object
      required:
        - status
      properties:
        status:
          $ref: "#/components/schemas/DeliveryStatus"

    VerifyQRRequest:
      type: object
      required:
        - token
      properties:
        token:
          type: string
          minLength: 1

    DeliveryResponse:
      type: object
      required:
        - id
        - order_id
        - status
        - pickup_address
        - estimated_delivery_date
        - is_late
        - created_at
        - updated_at
      properties:
        id:
          type: string
          format: uuid
        order_id:
          type: string
          format: uuid
        status:
          $ref: "#/components/schemas/DeliveryStatus"
        pickup_address:
          type: string
        estimated_delivery_date:
          type: string
          format: date-time
        is_late:
          type: boolean
        reschedules:
          type: array
          items:
            $ref: "#/components/schemas/DeliveryRescheduleResponse"
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    DeliveryRescheduleResponse:
      type: object
      required:
        - id
        - previous_date
        - new_date
        - created_at
      properties:
        id:
          type: string
          format: uuid
        previous_date:
          type: string
          format: date-time
        new_date:
          type: string
          format: date-time
        reason:
          type: string
        created_at:
          type: string
          format: date-time

    QRCodeResponse:
      type: object
      required:
        - token
        - expires_at
      properties:
        token:
          type: string
        expires_at:
          type: string
          format: date-time

    DeliveryListResponse:
      type: object
      required:
        - items
        - total
        - page
        - page_size
      properties:
        items:
          type: array
          items:
            $ref: "#/components/schemas/DeliveryResponse"
        total:
          type: integer
          format: int64
        page:
          type: integer
        page_size:
          type: integer

    VerifyQRResponse:
      type: object
      required:
        - delivery_id
        - order_id
        - verified
      properties:
        delivery_id:
          type: string
          format: uuid
        order_id:
          type: string
          format: uuid
        verified:
          type: boolean

    ErrorResponse:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: string
        message:
          type: string

    GetDeliveriesParams:
      type: object
      properties:
        status:
          $ref: "#/components/schemas/DeliveryStatus"
        is_late:
          type: boolean
        order_id:
          type: string
          format: uuid
        page:
          type: integer
          minimum: 1
          default: 1
        page_size:
          type: integer
          minimum: 1
          maximum: 100
          default: 20
```

### 7.3 HTTP-статусы ответов

| Endpoint                          | Успех | Ошибки                       |
| --------------------------------- | ----- | ---------------------------- |
| POST /deliveries                  | 201   | 400, 401, 403, 404, 409, 500 |
| GET /deliveries/{id}              | 200   | 401, 403, 404, 500           |
| GET /deliveries                   | 200   | 401, 403, 500                |
| PATCH /deliveries/{id}/reschedule | 200   | 400, 401, 403, 404, 409, 500 |
| PATCH /deliveries/{id}/status     | 200   | 400, 401, 403, 404, 409, 500 |
| GET /deliveries/{id}/qr           | 200   | 401, 403, 404, 409, 500      |
| POST /deliveries/verify-qr        | 200   | 400, 401, 403, 404, 500      |

---

## 8. Data Requirements

### 8.1 Миграция: 000003_create_deliveries.up.sql

```sql
CREATE TABLE IF NOT EXISTS deliveries (
    id UUID NOT NULL PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE,
    status VARCHAR(30) NOT NULL
        CHECK (status IN ('PENDING', 'IN_TRANSIT', 'READY_FOR_PICKUP', 'DELIVERED', 'CANCELLED')),
    pickup_address TEXT NOT NULL
        CHECK (length(pickup_address) > 0),
    estimated_delivery_date TIMESTAMP NOT NULL,
    is_late BOOLEAN NOT NULL DEFAULT FALSE,
    qr_nonce UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deliveries_order_id ON deliveries(order_id);
CREATE INDEX idx_deliveries_status ON deliveries(status);
CREATE INDEX idx_deliveries_is_late ON deliveries(is_late);

CREATE TABLE IF NOT EXISTS delivery_reschedules (
    id UUID NOT NULL PRIMARY KEY,
    delivery_id UUID NOT NULL
        REFERENCES deliveries(id) ON DELETE CASCADE,
    previous_date TIMESTAMP NOT NULL,
    new_date TIMESTAMP NOT NULL,
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_reschedules_delivery_id ON delivery_reschedules(delivery_id);
```

### 8.2 Миграция: 000003_create_deliveries.down.sql

```sql
DROP INDEX IF EXISTS idx_delivery_reschedules_delivery_id;
DROP INDEX IF EXISTS idx_deliveries_is_late;
DROP INDEX IF EXISTS idx_deliveries_status;
DROP INDEX IF EXISTS idx_deliveries_order_id;
DROP TABLE IF EXISTS delivery_reschedules;
DROP TABLE IF EXISTS deliveries;
```

### 8.3 Данные, не дублируемые из order-service

Delivery-service **не** хранит:

- `user_id` — извлекается из заказа через order-service client
- Информацию о товарах заказа
- Информацию о клиенте

Delivery-service **хранит**:

- `order_id` — ссылка на заказ
- `status` — собственный статус доставки
- `pickup_address` — адрес пункта выдачи
- `estimated_delivery_date` — плановая дата доставки
- `is_late` — флаг опоздания
- `qr_nonce` — nonce для инвалидации QR-кода
- Историю переносов в `delivery_reschedules`

---

## 9. QR Security Requirements

### 9.1 Формат payload

```json
{
  "delivery_id": "uuid",
  "order_id": "uuid",
  "user_id": "uuid",
  "nonce": "uuid",
  "iat": 1700000000,
  "exp": 1700086400
}
```

### 9.2 Подпись

- **Алгоритм:** HMAC-SHA256.
- **Ключ:** Серверный секретный ключ, настраиваемый через `QR_SECRET_KEY` в конфигурации.
- **Длина ключа:** минимум 32 байта.
- **Вычисление:** `signature = HMAC-SHA256(base64url(JSON(payload)), secret_key)`.

### 9.3 Формат токена

```
base64url(JSON(payload)) + "." + base64url(signature)
```

### 9.4 Срок действия

- Настраиваемый: `QR_TOKEN_TTL` в конфигурации, по умолчанию 24 часа.
- Поле `exp` в payload содержит unix timestamp истечения.

### 9.5 Отзыв/инвалидация

- При каждой генерации QR-кода в `deliveries.qr_nonce` записывается новый UUID.
- При верификации проверяется совпадение `nonce` из токена с `qr_nonce` в БД.
- При инвалидации: генерация нового QR автоматически инвалидирует старый (nonce не совпадает).
- При отмене доставки: `qr_nonce` обнуляется (NULL), все токены невалидны.

### 9.6 Защита от подделки

- HMAC-SHA256 обеспечивает криптографическую целостность. Без секретного ключа невозможно создать валидную подпись.
- Секретный ключ хранится на сервере и никогда не включается в QR-код.

### 9.7 Защита от повторного использования

- После успешной верификации статус доставки переходит в `DELIVERED`.
- Повторная верификация невозможна, т.к. статус `DELIVERED` — терминальный.
- Переход `READY_FOR_PICKUP` → `DELIVERED` допустим только один раз.

### 9.8 Серверная валидация (порядок проверок)

1. Декодировать токен: разделить на payload и signature.
2. Верифицировать HMAC-SHA256-подпись. При несовпадении → `INVALID_TOKEN`.
3. Декодировать JSON payload. При ошибке парсинга → `INVALID_TOKEN`.
4. Проверить `exp > now`. При истечении → `TOKEN_EXPIRED`.
5. Найти доставку по `delivery_id`. При отсутствии → `DELIVERY_NOT_FOUND`.
6. Проверить `delivery.status == READY_FOR_PICKUP`. При несовпадении → `INVALID_DELIVERY_STATUS`.
7. Проверить `payload.user_id == claims.UserID`. При несовпадении → `FORBIDDEN`.
8. Проверить `payload.nonce == delivery.qr_nonce`. При несовпадении → `TOKEN_REVOKED`.
9. Атомарно перевести доставку в статус `DELIVERED`.
10. Вернуть успешный ответ.

---

## 10. Error Cases

### 10.1 Создание доставки

| Код | Ошибка                    | Условие                                    |
| --- | ------------------------- | ------------------------------------------ |
| 400 | `validation_error`        | Отсутствуют обязательные поля              |
| 400 | `invalid_order_id`        | `order_id` равен UUID.Nil                  |
| 400 | `invalid_pickup_address`  | Пустой адрес пункта выдачи                 |
| 400 | `invalid_estimated_date`  | Дата в прошлом или нулевая                 |
| 401 | `unauthorized`            | Отсутствуют JWT-claims                     |
| 403 | `forbidden`               | Роль не `employee`                         |
| 404 | `order_not_found`         | Заказ не найден в order-service            |
| 409 | `delivery_already_exists` | Доставка для данного заказа уже существует |
| 409 | `order_not_paid`          | Статус заказа не `PAID`                    |
| 500 | `internal_error`          | Внутренняя ошибка сервера                  |

### 10.2 Просмотр доставки

| Код | Ошибка               | Условие                             |
| --- | -------------------- | ----------------------------------- |
| 401 | `unauthorized`       | Отсутствуют JWT-claims              |
| 403 | `forbidden`          | Customer запрашивает чужую доставку |
| 404 | `delivery_not_found` | Доставка не найдена                 |
| 500 | `internal_error`     | Внутренняя ошибка сервера           |

### 10.3 Список доставок

| Код | Ошибка           | Условие                   |
| --- | ---------------- | ------------------------- |
| 401 | `unauthorized`   | Отсутствуют JWT-claims    |
| 403 | `forbidden`      | Роль не `employee`        |
| 500 | `internal_error` | Внутренняя ошибка сервера |

### 10.4 Перенос доставки

| Код | Ошибка                    | Условие                         |
| --- | ------------------------- | ------------------------------- |
| 400 | `validation_error`        | Отсутствует `new_date`          |
| 400 | `invalid_new_date`        | `new_date` в прошлом            |
| 400 | `invalid_delivery_status` | Доставка в терминальном статусе |
| 401 | `unauthorized`            | Отсутствуют JWT-claims          |
| 403 | `forbidden`               | Роль не `employee`              |
| 404 | `delivery_not_found`      | Доставка не найдена             |
| 500 | `internal_error`          | Внутренняя ошибка сервера       |

### 10.5 Изменение статуса

| Код | Ошибка                       | Условие                             |
| --- | ---------------------------- | ----------------------------------- |
| 400 | `validation_error`           | Отсутствует или невалидный `status` |
| 400 | `invalid_status_transition`  | Переход недопустим по State Machine |
| 400 | `delivery_already_delivered` | Доставка уже завершена              |
| 400 | `delivery_already_cancelled` | Доставка уже отменена               |
| 401 | `unauthorized`               | Отсутствуют JWT-claims              |
| 403 | `forbidden`                  | Роль не `employee`                  |
| 404 | `delivery_not_found`         | Доставка не найдена                 |
| 500 | `internal_error`             | Внутренняя ошибка сервера           |

### 10.6 Получение QR-кода

| Код | Ошибка                    | Условие                                |
| --- | ------------------------- | -------------------------------------- |
| 401 | `unauthorized`            | Отсутствуют JWT-claims                 |
| 403 | `forbidden`               | Customer запрашивает QR чужой доставки |
| 404 | `delivery_not_found`      | Доставка не найдена                    |
| 409 | `invalid_delivery_status` | Статус не `READY_FOR_PICKUP`           |
| 500 | `internal_error`          | Внутренняя ошибка сервера              |

### 10.7 Верификация QR-кода

| Код | Ошибка                       | Условие                                          |
| --- | ---------------------------- | ------------------------------------------------ |
| 400 | `invalid_token`              | Невалидная подпись или формат                    |
| 400 | `token_expired`              | Срок действия QR истёк                           |
| 400 | `token_revoked`              | Nonce не совпадает (QR отозван или заменён)      |
| 400 | `invalid_delivery_status`    | Доставка не в статусе `READY_FOR_PICKUP`         |
| 400 | `delivery_already_delivered` | Доставка уже завершена (повторное использование) |
| 401 | `unauthorized`               | Отсутствуют JWT-claims                           |
| 403 | `forbidden`                  | Роль не `employee`                               |
| 404 | `delivery_not_found`         | Доставка по `delivery_id` из токена не найдена   |
| 500 | `internal_error`             | Внутренняя ошибка сервера                        |

---

## 11. Integration Requirements

### 11.1 Зависимость от order-service

Delivery-service зависит от order-service:

1. **GET /orders/{id}** — проверка существования заказа и получение `user_id` и статуса заказа при создании доставки.
2. **PATCH /orders/{id}/status** (требуется новый endpoint) — обновление статуса заказа при изменении статуса доставки.

**Предварительное условие:** order-service должен предоставить endpoint для обновления статуса заказа. В текущей реализации такого endpoint нет. Это требует дополнения order-service.

### 11.2 HTTP-клиент для order-service

Создать `common/client/order/` по аналогии с `common/client/product/`:

- `client.go` — интерфейс `OrderClient` и реализация `OrderClient`
- `errors.go` — `ErrOrderNotFound`, `ErrOrderNotPaid`
- `generated/api.gen.go` — сгенерированный клиент из OpenAPI order-service

### 11.3 Обновление Casbin-политики

Добавить в `common/rbac/policy.csv`:

```csv
p, customer, delivery, view
p, employee, delivery, create
p, employee, delivery, view
p, employee, delivery, reschedule
p, employee, delivery, update_status
p, employee, delivery, verify_qr
```

Добавить в `common/rbac/authorizer.go`:

```go
const (
    ResourceDelivery  Resource = "delivery"
    ActionReschedule  Action = "reschedule"
    ActionUpdateStatus Action = "update_status"
    ActionVerifyQR   Action = "verify_qr"
)
```

### 11.4 Docker Compose

Добавить сервис `delivery-service` в `docker-compose.yml`:

```yaml
delivery-service:
  build:
    context: .
    dockerfile: services/delivery-service/Dockerfile
  env_file:
    - .env
  depends_on:
    migrate:
      condition: service_completed_successfully
  ports:
    - "8082:8082"
```

Добавить в `.env` и `.env.example`:

```
DELIVERY_SERVICE_URL=http://delivery-service:8082
QR_SECRET_KEY=<secret>
QR_TOKEN_TTL=24h
```

### 11.5 Makefile

Добавить цель:

```makefile
generate-d:
	oapi-codegen \
		-config ./services/delivery-service/api/oapi-codegen.yaml \
		./services/delivery-service/api/delivery-service.yaml
```

### 11.6 Конфигурация

```yaml
http:
  host: 0.0.0.0
  port: 8082

postgres:
  host: ${POSTGRES_HOST}
  port: ${POSTGRES_PORT}
  database: ${POSTGRES_DB}
  user: ${POSTGRES_USER}
  password: ${POSTGRES_PASSWORD}
  sslmode: ${POSTGRES_SSLMODE}

log:
  level: ${LOG_LEVEL}

jwt:
  issuer: ${JWT_ISSUER}
  public_key_path: ${JWT_PUBLIC_KEY_PATH}

order_service_url: ${ORDER_SERVICE_URL}

qr:
  secret_key: ${QR_SECRET_KEY}
  token_ttl: ${QR_TOKEN_TTL}
```

### 11.7 Обновление статуса заказа

При переходах доставки:

| Переход доставки                 | Вызов order-service                            |
| -------------------------------- | ---------------------------------------------- |
| `PENDING` → `IN_TRANSIT`         | `PATCH /orders/{id}/status` → `DELIVERING`     |
| `READY_FOR_PICKUP` → `DELIVERED` | `PATCH /orders/{id}/status` → `DELIVERED`      |
| `→ CANCELLED`                    | `PATCH /orders/{id}/status` → `PAID` (возврат) |

**Открытый вопрос:** поведение при ошибке обновления статуса в order-service. Варианты: (A) откатить транзакцию доставки, (B) записать в лог и продолжить. Рекомендуется вариант A для консистентности.

---

## 12. Test Requirements

### 12.1 Unit-тесты (service-слой)

Следуя конвенции проекта (table-driven tests с `t.Parallel()`, ручные моки):

| Тест                       | Сценарии                                                                                                  |
| -------------------------- | --------------------------------------------------------------------------------------------------------- |
| `TestService_Create`       | Успех, nil input, невалидный order_id, заказ не найден, заказ не PAID, доставка уже существует, ошибка БД |
| `TestService_GetByID`      | Успех (employee), успех (customer, своя), 403 (чужая), не найдена, невалидный ID                          |
| `TestService_List`         | Успех (employee), фильтрация по статусу, фильтрация по is_late, пагинация                                 |
| `TestService_Reschedule`   | Успех, невалидная дата, терминальный статус, не найдена                                                   |
| `TestService_UpdateStatus` | Успех (каждый допустимый переход), недопустимый переход, терминальный статус, не найдена                  |
| `TestService_GetQR`        | Успех, не READY_FOR_PICKUP, чужая доставка, не найдена                                                    |
| `TestService_VerifyQR`     | Успех, невалидная подпись, истёкший токен, отозванный nonce, не READY_FOR_PICKUP, уже DELIVERED           |

### 12.2 Unit-тесты (transport-слой)

| Тест                         | Сценарии                                                                                               |
| ---------------------------- | ------------------------------------------------------------------------------------------------------ |
| `TestHandler_CreateDelivery` | Успех, 400 валидация, 401 неавторизован, 403 не employee, 404 заказ не найден, 409 доставка существует |
| `TestHandler_GetDelivery`    | Успех (customer), 403 чужая, 404 не найдена, 401 неавторизован                                         |
| `TestHandler_ListDeliveries` | Успех (employee), 403 не employee                                                                      |
| `TestHandler_Reschedule`     | Успех, 400 невалидная дата, 400 терминальный статус, 403 не employee                                   |
| `TestHandler_UpdateStatus`   | Успех, 400 недопустимый переход, 403 не employee                                                       |
| `TestHandler_GetQR`          | Успех, 409 неверный статус, 403 чужая доставка                                                         |
| `TestHandler_VerifyQR`       | Успех, 400 невалидный токен, 400 истёкший, 400 отозванный                                              |

### 12.3 Интеграционные тесты (repository-слой)

Следуя конвенции проекта (`testcontainers-go`, `testtools.NewTestDB`):

| Тест                              | Сценарии                                |
| --------------------------------- | --------------------------------------- |
| `TestRepository_Create`           | Успех, нарушение UNIQUE(order_id)       |
| `TestRepository_GetByID`          | Успех, не найдена                       |
| `TestRepository_GetByOrderID`     | Успех, не найдена                       |
| `TestRepository_List`             | Успех, фильтрация по статусу, пагинация |
| `TestRepository_Update`           | Успех, не найдена                       |
| `TestRepository_CreateReschedule` | Успех                                   |
| `TestRepository_ListReschedules`  | Успех, пустой список                    |

### 12.4 Моки

Ручные моки по конвенции проекта (`mocks/repository.go`, `mocks/service.go`, `mocks/order_client.go`):

- `MockDeliveryRepository` с функциональными полями и счётчиками вызовов
- `MockDeliveryService` с функциональными полями и счётчиками вызовов
- `MockOrderClient` с функциональными полями и счётчиками вызовов

### 12.5 Тестирование QR-механизма

| Тест                         | Сценарии                                             |
| ---------------------------- | ---------------------------------------------------- |
| `TestQR_GenerateAndVerify`   | Успешная генерация → верификация                     |
| `TestQR_InvalidSignature`    | Подмена payload → ошибка подписи                     |
| `TestQR_ExpiredToken`        | Истёкший exp → ошибка                                |
| `TestQR_RevokedNonce`        | Генерация нового QR → старый невалиден               |
| `TestQR_WrongDeliveryStatus` | Верификация при статусе PENDING → ошибка             |
| `TestQR_AlreadyDelivered`    | Повторная верификация → ошибка                       |
| `TestQR_DifferentUser`       | user_id из токена не совпадает с владельцем → ошибка |

---

## 13. Структура сервиса (список файлов)

```
services/delivery-service/
├── api/
│   ├── delivery-service.yaml
│   └── oapi-codegen.yaml
├── cmd/delivery-service/
│   └── main.go
├── config.yaml
├── Dockerfile
├── internal/
│   ├── app/
│   │   ├── module.go
│   │   ├── config.go
│   │   ├── logger.go
│   │   ├── database.go
│   │   ├── auth.go
│   │   ├── rbac.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   ├── handler.go
│   │   ├── server.go
│   │   ├── client.go
│   │   └── validate.go
│   ├── client/
│   │   └── order.go
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── delivery.go
│   │   ├── reschedule.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   ├── errors.go
│   │   └── auth.go
│   ├── generated/openapi/
│   │   └── api.gen.go
│   ├── mocks/
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── order_client.go
│   ├── qr/
│   │   ├── generator.go
│   │   └── verifier.go
│   ├── repository/
│   │   └── repository.go
│   ├── service/
│   │   ├── service.go
│   │   └── errors.go
│   └── transport/
│       ├── handler.go
│       ├── router.go
│       ├── errors.go
│       ├── mapper.go
│       ├── create.go
│       ├── get.go
│       ├── list.go
│       ├── reschedule.go
│       ├── update_status.go
│       ├── get_qr.go
│       ├── verify_qr.go
│       └── middlwr/
│           ├── auth.go
│           └── rbac.go
```

### Дополнения в common

```
services/common/
├── client/
│   ├── order/                        # Новый клиент
│   │   ├── client.go
│   │   ├── errors.go
│   │   └── generated/
│   │       └── api.gen.go
│   └── product/                      # Существующий
├── rbac/
│   ├── authorizer.go                 # Добавить ResourceDelivery, ActionReschedule, ActionUpdateStatus, ActionVerifyQR
│   ├── model.conf                    # Без изменений
│   └── policy.csv                    # Добавить строки для delivery
└── test-tools/
    └── helper.go                     # Без изменений
```

### Дополнения в корне

```
migrations/
├── 000003_create_deliveries.up.sql     # Новый
└── 000003_create_deliveries.down.sql   # Новый
```

---

## 14. Конфигурация сервиса

```go
type Config struct {
    HTTP           HTTPConfig     `yaml:"http" env:"HTTP_HOST" env-default:"0.0.0.0"`
    Postgres       PostgresConfig `yaml:"postgres"`
    Log            LogConfig      `yaml:"log" env:"LOG_LEVEL" env-default:"info"`
    JWT            JWT            `yaml:"jwt"`
    OrderService   OrderServiceConfig `yaml:"order_service"`
    QR             QRConfig       `yaml:"qr"`
}

type QRConfig struct {
    SecretKey string        `yaml:"secret_key" env:"QR_SECRET_KEY" env-required:"true"`
    TokenTTL  time.Duration `yaml:"token_ttl" env:"QR_TOKEN_TTL" env-default:"24h"`
}

type OrderServiceConfig struct {
    URL string `yaml:"url" env:"ORDER_SERVICE_URL" env-required:"true"`
}
```
