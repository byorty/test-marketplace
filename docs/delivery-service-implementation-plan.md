# Delivery Service — Implementation Plan

## Notation

- **Task ID:** `T<phase>.<number>` — фаза и номер задачи внутри фазы.
- **Depends on:** список task IDs, которые должны быть завершены до начала данной задачи.
- **Files created:** новые файлы, которые нужно создать.
- **Files modified:** существующие файлы, которые нужно изменить.
- **Tests:** какие тесты нужно написать/запустить.
- **Done when:** критерии готовности задачи.

Все команды выполняются из каталога `services/` (где находится `go.mod`).

---

## Phase 1: Foundation & Skeleton

### T1.1 — Создать каркас каталогов delivery-service

**Goal:** Создать минимальную структуру каталогов и пустые файлы, позволяющую `go build` пройти.

**Depends on:** nothing

**Files created:**

```
services/delivery-service/
├── cmd/delivery-service/main.go
├── config.yaml
├── Dockerfile
├── internal/
│   ├── app/module.go
│   ├── config/config.go
│   └── domain/delivery.go
```

**`cmd/delivery-service/main.go`:** точка входа `fx.New(app.Module).Run()`, но Module пока пустой `fx.Options()`.

**`internal/app/module.go`:** `var Module = fx.Options()` — пустой модуль, компилируется.

**`internal/domain/delivery.go`:** минимальный пакет `domain` с комментарием `// TODO: implement`.

**`internal/config/config.go`:** минимальная структура `Config` с `HTTP` и `Postgres` секциями, функция `Load()` через `yaml.Unmarshal` + `os.ExpandEnv`. Без QR и OrderService секций пока.

**`config.yaml`:** минимальный конфиг с `http`, `postgres`, `log`, `jwt` секциями.

**`Dockerfile`:** по шаблону product-service/order-service.

**Tests:** `go build -o /tmp/delivery-service ./delivery-service/cmd/delivery-service/`

**Done when:**
- `go build -o /tmp/delivery-service ./delivery-service/cmd/delivery-service/` из `services/` проходит без ошибок;
- `go vet ./delivery-service/...` проходит.

---

### T1.2 — Добавить Fx-модули: Config, Logger, Database, Auth, Validate

**Goal:** Подключить инфраструктурные модули, позволяющие сервису стартовать и подключиться к PostgreSQL.

**Depends on:** T1.1

**Files created:**

```
services/delivery-service/internal/
├── app/logger.go
├── app/database.go
├── auth.go
└── validate.go
```

**Files modified:**

- `internal/app/module.go` — добавить `ConfigModule`, `LoggerModule`, `DatabaseModule`, `AuthModule`, `ValidateModule`
- `internal/config/config.go` — добавить `JWT` секцию

**Описание модулей (по конвенции проекта):**

- `LoggerModule = fx.Provide(NewLogger)` — `zap.NewProduction()`
- `DatabaseModule = fx.Provide(NewDB)` — `database.New(database.PostgresConfig(cfg.Postgres))`
- `AuthModule = fx.Provide(NewJWTValidator)` — `auth.NewValidator(publicKey, cfg.JWT.Issuer)`
- `ValidateModule = fx.Provide(NewValidator)` — `validator.New()`

Все функции — копии из product-service/order-service с заменой пакета на `delivery-service`.

**Tests:** `go build`, `go vet`

**Done when:**
- Сервис компилируется;
- `go vet` проходит;
- Модули подключены в правильном порядке (Config → Logger → Database → Auth → Validate).

---

### T1.3 — Добавить RBAC-модуль и расширить common/rbac

**Goal:** Добавить ресурсы и действия для delivery в `common/rbac/authorizer.go` и `common/rbac/policy.csv`.

**Depends on:** T1.2

**Files created:** none

**Files modified:**

- `services/common/rbac/authorizer.go` — добавить константы:
  ```go
  ResourceDelivery  Resource = "delivery"
  ActionReschedule  Action = "reschedule"
  ActionUpdateStatus Action = "update_status"
  ActionVerifyQR    Action = "verify_qr"
  ```
- `services/common/rbac/policy.csv` — добавить строки:
  ```csv
  p, customer, delivery, view
  p, employee, delivery, create
  p, employee, delivery, view
  p, employee, delivery, reschedule
  p, employee, delivery, update_status
  p, employee, delivery, verify_qr
  ```

**Files created:**

- `services/delivery-service/internal/app/rbac.go` — `RBACModule` с `rbac.NewEnforcer`, `rbac.New`, `NewDomainAuthorizer` (как в order-service)
- `services/delivery-service/internal/domain/auth.go` — `Authorizer` interface (как в order-service)

**Files modified:**

- `services/delivery-service/internal/app/module.go` — добавить `RBACModule`

**Tests:**
- `go vet ./common/rbac/...`
- `go vet ./delivery-service/...`
- Существующие тесты product-service и order-service проходят.

**Done when:**
- `go vet ./...` из `services/` проходит;
- Существующие тесты не сломаны;
- Новые константы доступны для импорта.

---

### T1.4 — Добавить ServerModule и запустить HTTP-сервер

**Goal:** Сервис стартует, слушает порт, отвечает 404 на любые запросы (маршрутов ещё нет).

**Depends on:** T1.3

**Files created:**

- `services/delivery-service/internal/app/server.go`

**Files modified:**

- `services/delivery-service/internal/app/module.go` — добавить `ServerModule`

**`server.go`:** `RunServer(lifecycle, handler, cfg, log, jwt, authorizer)` — создаёт `chi.NewRouter()`, подключает middleware, запускает `http.Server`. Handler пока не передаётся в роутер, просто пустой роутер.

**Tests:** `go build`, `go vet`

**Done when:**
- `go build -o /tmp/delivery-service ./delivery-service/cmd/delivery-service/` проходит;
- Сервис можно запустить (порт 8082) и получить 404.

---

## Phase 2: Configuration & Environment

### T2.1 — Полная конфигурация delivery-service

**Goal:** Добавить все секции конфигурации: QR, OrderService.

**Depends on:** T1.4

**Files modified:**

- `services/delivery-service/internal/config/config.go` — добавить `OrderServiceConfig`, `QRConfig`, `LogConfig`

**Files modified:**

- `services/delivery-service/config.yaml` — добавить `order_service_url`, `qr` секции

**Files modified:**

- `.env` — добавить `ORDER_SERVICE_URL=http://order-service:8081`, `QR_SECRET_KEY`, `QR_TOKEN_TTL=24h`
- `.env.example` — то же самое

**Полный Config:**

```go
type Config struct {
    HTTP          HTTPConfig        `yaml:"http"`
    Postgres      PostgresConfig    `yaml:"postgres"`
    Log           LogConfig         `yaml:"log"`
    JWT           JWT               `yaml:"jwt"`
    OrderService  OrderServiceConfig `yaml:"order_service"`
    QR            QRConfig          `yaml:"qr"`
}
```

**Tests:** `go build`, `go vet`

**Done when:**
- Конфигурация парсится из YAML с подстановкой env-переменных;
- Все секции описаны с тегами `yaml` и комментариями `env`.

---

## Phase 3: Database & Migrations

### T3.1 — Миграция: создание таблиц deliveries и delivery_reschedules

**Goal:** Создать SQL-миграции для таблиц доставки.

**Depends on:** T2.1

**Files created:**

- `migrations/000003_create_deliveries.up.sql`
- `migrations/000003_create_deliveries.down.sql`

**`up.sql`:**

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

**`down.sql`:**

```sql
DROP INDEX IF EXISTS idx_delivery_reschedules_delivery_id;
DROP INDEX IF EXISTS idx_deliveries_is_late;
DROP INDEX IF EXISTS idx_deliveries_status;
DROP INDEX IF EXISTS idx_deliveries_order_id;
DROP TABLE IF EXISTS delivery_reschedules;
DROP TABLE IF EXISTS deliveries;
```

**Tests:** `make migrate-up` из корня проекта; проверить через `psql` что таблицы созданы; `make migrate-down`; `make migrate-up` снова.

**Done when:**
- `make migrate-up` проходит без ошибок;
- Таблицы `deliveries` и `delivery_reschedules` созданы с правильными индексами;
- `make migrate-down` откатывает миграцию;
- Повторный `make migrate-up` проходит.

---

## Phase 4: Domain Model

### T4.1 — Domain: Delivery model и DeliveryStatus

**Goal:** Определить модели доставки и статусы в `domain/delivery.go`.

**Depends on:** T3.1

**Files created:**

- `services/delivery-service/internal/domain/delivery.go`
- `services/delivery-service/internal/domain/errors.go`

**`domain/delivery.go`:**

```go
type DeliveryStatus string

const (
    DeliveryStatusPending        DeliveryStatus = "PENDING"
    DeliveryStatusInTransit      DeliveryStatus = "IN_TRANSIT"
    DeliveryStatusReadyForPickup DeliveryStatus = "READY_FOR_PICKUP"
    DeliveryStatusDelivered      DeliveryStatus = "DELIVERED"
    DeliveryStatusCancelled      DeliveryStatus = "CANCELLED"
)

func (s DeliveryStatus) IsValid() bool { ... }
func (s DeliveryStatus) IsTerminal() bool { ... }

func CanTransition(from, to DeliveryStatus) bool { ... }

type Delivery struct {
    bun.BaseModel `bun:"table:deliveries"`
    ID                    uuid.UUID      `bun:"id,pk,type:uuid"`
    OrderID               uuid.UUID      `bun:"order_id,notnull,type:uuid"`
    Status                DeliveryStatus `bun:"status,notnull,type:varchar(30)"`
    PickupAddress         string         `bun:"pickup_address,notnull"`
    EstimatedDeliveryDate time.Time      `bun:"estimated_delivery_date,notnull"`
    IsLate                bool           `bun:"is_late,notnull,default:false"`
    QRNonce               uuid.NullUUID  `bun:"qr_nonce,type:uuid"`
    CreatedAt             time.Time      `bun:"created_at,notnull"`
    UpdatedAt             time.Time      `bun:"updated_at,notnull"`
}

type DeliveryList struct { ... }
type DeliveryListFilter struct { ... }
type DeliveryCreateInput struct { ... }
```

**`domain/errors.go`:**

```go
var (
    ErrDeliveryNotFound       = errors.New("delivery not found")
    ErrDeliveryAlreadyExists = errors.New("delivery already exists for this order")
    ErrInvalidDeliveryStatus = errors.New("invalid delivery status")
    ErrInvalidStatusTransition = errors.New("invalid status transition")
    ErrDeliveryAlreadyDelivered = errors.New("delivery already delivered")
    ErrDeliveryAlreadyCancelled = errors.New("delivery already cancelled")
    ErrRescheduleOnTerminal   = errors.New("cannot reschedule delivery in terminal status")
    ErrQRNotAvailable         = errors.New("QR code is not available for this delivery status")
    ErrOrderNotFound          = errors.New("order not found")
    ErrOrderNotPaid           = errors.New("order is not in PAID status")
)
```

**`CanTransition`** — реализует state machine из архитектурного документа:
- PENDING → IN_TRANSIT, CANCELLED
- IN_TRANSIT → READY_FOR_PICKUP, CANCELLED
- READY_FOR_PICKUP → DELIVERED, CANCELLED
- DELIVERED → none (terminal)
- CANCELLED → none (terminal)

**Tests:** Unit-тесты для `IsValid()`, `IsTerminal()`, `CanTransition()`.

**Done when:**
- `go build` проходит;
- Unit-тесты для state machine покрывают все допустимые и запрещённые переходы;
- `go vet` проходит.

---

### T4.2 — Domain: DeliveryReschedule model

**Goal:** Определить модель переноса доставки.

**Depends on:** T4.1

**Files created:**

- `services/delivery-service/internal/domain/reschedule.go`

**`domain/reschedule.go`:**

```go
type DeliveryReschedule struct {
    bun.BaseModel `bun:"table:delivery_reschedules"`
    ID           uuid.UUID  `bun:"id,pk,type:uuid"`
    DeliveryID   uuid.UUID  `bun:"delivery_id,notnull,type:uuid"`
    PreviousDate time.Time  `bun:"previous_date,notnull"`
    NewDate      time.Time  `bun:"new_date,notnull"`
    Reason       string     `bun:"reason"`
    CreatedAt    time.Time  `bun:"created_at,notnull"`
}

type RescheduleInput struct {
    NewDate time.Time
    Reason  string
}
```

**Tests:** `go build`, `go vet`

**Done when:**
- Модель компилируется;
- Bun-теги соответствуют схеме БД.

---

### T4.3 — Domain: Repository и Service interfaces

**Goal:** Определить интерфейсы репозитория и сервиса.

**Depends on:** T4.2

**Files created:**

- `services/delivery-service/internal/domain/repository.go`
- `services/delivery-service/internal/domain/service.go`

**`domain/repository.go`:**

```go
type DeliveryRepository interface {
    Create(ctx context.Context, delivery *Delivery) error
    GetByID(ctx context.Context, id uuid.UUID) (*Delivery, error)
    GetByOrderID(ctx context.Context, orderID uuid.UUID) (*Delivery, error)
    List(ctx context.Context, filter DeliveryListFilter) (*DeliveryList, error)
    Update(ctx context.Context, delivery *Delivery) (*Delivery, error)
    CreateReschedule(ctx context.Context, reschedule *DeliveryReschedule) error
    ListReschedules(ctx context.Context, deliveryID uuid.UUID) ([]DeliveryReschedule, error)
}
```

**`domain/service.go`:**

```go
type DeliveryService interface {
    Create(ctx context.Context, input *DeliveryCreateInput) (*Delivery, error)
    GetByID(ctx context.Context, userID, id uuid.UUID) (*Delivery, error)
    List(ctx context.Context, userID uuid.UUID, role string, filter DeliveryListFilter) (*DeliveryList, error)
    Reschedule(ctx context.Context, deliveryID uuid.UUID, input *RescheduleInput) (*Delivery, error)
    UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status DeliveryStatus) (*Delivery, error)
    GetQR(ctx context.Context, userID, deliveryID uuid.UUID) (*qr.QRTokenResponse, error)
    VerifyQR(ctx context.Context, token string) (*Delivery, error)
}
```

**Tests:** `go build`, `go vet`

**Done when:**
- Интерфейсы компилируются;
- Импорты `qr.QRTokenResponse` корректны (qr-пакет пока пуст, но интерфейс определён).

---

## Phase 5: QR Package

### T5.1 — QR Token: структура и кодирование

**Goal:** Определить структуру QR-токена и функции кодирования/декодирования.

**Depends on:** T4.3

**Files created:**

- `services/delivery-service/internal/qr/token.go`

**`qr/token.go`:**

```go
type QRToken struct {
    DeliveryID uuid.UUID `json:"delivery_id"`
    OrderID    uuid.UUID `json:"order_id"`
    UserID     uuid.UUID `json:"user_id"`
    Nonce      uuid.UUID `json:"nonce"`
    IssuedAt   int64     `json:"iat"`
    ExpiresAt  int64     `json:"exp"`
}

type QRTokenResponse struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
}

func Encode(token *QRToken) (string, error) { ... }
func Decode(encoded string) (*QRToken, error) { ... }
```

**`Encode`:** `base64url(JSON(payload))`
**`Decode`:** `base64url_decode → JSON.Unmarshal`

**Tests:** Unit-тесты для `Encode`/`Decode`: round-trip, невалидный base64, невалидный JSON.

**Done when:**
- `go build` проходит;
- Unit-тесты для round-trip кодирования/декодирования проходят.

---

### T5.2 — QR Generator: HMAC-SHA256 подпись

**Goal:** Реализовать генерацию QR-токенов с HMAC-SHA256 подписью.

**Depends on:** T5.1

**Files created:**

- `services/delivery-service/internal/qr/generator.go`

**`qr/generator.go`:**

```go
type Generator struct {
    secretKey []byte
    tokenTTL  time.Duration
}

func NewGenerator(secretKey string, tokenTTL time.Duration) *Generator
func (g *Generator) Generate(deliveryID, orderID, userID, nonce uuid.UUID) (*QRTokenResponse, error)
```

**Логика Generate:**
1. Вычислить `iat = now.Unix()`, `exp = now.Add(g.tokenTTL).Unix()`.
2. Сформировать `QRToken` с полями.
3. Вызвать `Encode(token)` → payload string.
4. Вычислить `signature = HMAC-SHA256(payload, g.secretKey)`.
5. Вернуть `base64url(payload) + "." + base64url(signature)`.

**tests:** Unit-тесты:
- Успешная генерация: токен не пустой, содержит `.`-разделитель, expiresAt в будущем.
- TTL: expiresAt = now + tokenTTL (с допуском 1 секунда).
- Разные nonce → разные токены.

**Done when:**
- `go build` проходит;
- Unit-тесты для генерации проходят;
- `go vet` проходит.

---

### T5.3 — QR Verifier: проверка HMAC-SHA256 подписи

**Goal:** Реализовать верификацию QR-токенов.

**Depends on:** T5.2

**Files created:**

- `services/delivery-service/internal/qr/verifier.go`
- `services/delivery-service/internal/qr/qr_test.go` (если не создан ранее)

**`qr/verifier.go`:**

```go
type Verifier struct {
    secretKey []byte
}

func NewVerifier(secretKey string) *Verifier
func (v *Verifier) Verify(tokenString string) (*QRToken, error)
```

**Логика Verify:**
1. Разделить токен на `payload` и `signature` по `.`.
2. Декодировать `base64url(payload)` → payload string.
3. Вычислить `expectedSig = HMAC-SHA256(payload, v.secretKey)`.
4. Сравнить `base64url_decode(signature)` с `expectedSig` через `hmac.Equal`. При несовпадении → `ErrInvalidSignature`.
5. Декодировать `QRToken` из JSON payload. При ошибке → `ErrInvalidToken`.
6. Проверить `token.ExpiresAt > now.Unix()`. При истечении → `ErrTokenExpired`.
7. Вернуть `*QRToken, nil`.

**Ошибки (qr package):**

```go
var (
    ErrInvalidToken     = errors.New("invalid QR token")
    ErrTokenExpired     = errors.New("QR token expired")
    ErrInvalidSignature = errors.New("invalid QR token signature")
)
```

**tests:** Unit-тесты:
- Успешная верификация: Generate → Verify → валидный QRToken.
- Невалидная подпись: подмена payload → `ErrInvalidSignature`.
- Истёкший токен: `exp` в прошлом → `ErrTokenExpired`.
- Невалидный формат: пустая строка, без `.` → `ErrInvalidToken`.
- Невалидный base64 → `ErrInvalidToken`.

**Done when:**
- `go build` проходит;
- Все unit-тесты QR-пакета проходят;
- `go vet` проходит.

---

## Phase 6: Repository

### T6.1 — DeliveryRepository: реализация

**Goal:** Реализовать все методы DeliveryRepository.

**Depends on:** T4.3, T5.3

**Files created:**

- `services/delivery-service/internal/repository/repository.go`

**Конструктор:** `New(db *bun.DB, log *zap.Logger) *DeliveryRepository`

**Методы:**
- `Create(ctx, *Delivery) error` — INSERT, логирование
- `GetByID(ctx, id) (*Delivery, error)` — SELECT, `sql.ErrNoRows` → `ErrDeliveryNotFound`
- `GetByOrderID(ctx, orderID) (*Delivery, error)` — SELECT WHERE order_id, `sql.ErrNoRows` → `ErrDeliveryNotFound`
- `List(ctx, filter) (*DeliveryList, error)` — SELECT с фильтрацией по status, is_late, order_id, пагинацией
- `Update(ctx, *Delivery) (*Delivery, error)` — UPDATE RETURNING *, `sql.ErrNoRows` → `ErrDeliveryNotFound`
- `CreateReschedule(ctx, *DeliveryReschedule) error` — INSERT
- `ListReschedules(ctx, deliveryID) ([]DeliveryReschedule, error)` — SELECT WHERE delivery_id

**Паттерн:** как в product-service/order-service repository. Bun query builder, zap.Named("delivery-repository"), `sql.ErrNoRows` → sentinel errors.

**Tests:** Пока без интеграционных (они будут в T6.2).

**Done when:**
- `go build` проходит;
- `go vet` проходит.

---

### T6.2 — Integration tests: DeliveryRepository

**Goal:** Написать интеграционные тесты для всех методов DeliveryRepository.

**Depends on:** T6.1

**Files created:**

- `services/delivery-service/internal/repository/repository_test.go`

**Паттерн:** как в product-service/order-service — `testtools.NewTestDB(t)`, `t.Parallel()`, `require`.

**Тесты:**

```
TestRepository_Create — успех, нарушение UNIQUE(order_id)
TestRepository_GetByID — успех, не найдена
TestRepository_GetByOrderID — успех, не найдена
TestRepository_List — фильтрация по status, по is_late, пагинация
TestRepository_Update — успех, не найдена
TestRepository_CreateReschedule — успех
TestRepository_ListReschedules — успех, пустой список
```

**Done when:**
- `go test ./delivery-service/internal/repository/... -tags=integration` проходит;
- Все тесты покрывают positive и negative cases.

---

## Phase 7: Order Client

### T7.1 — OrderClient: интерфейс и common/client/order

**Goal:** Создать HTTP-клиент для order-service в `common/client/order/`.

**Depends on:** T2.1

**Files created:**

- `services/common/client/order/errors.go`
- `services/common/client/order/client.go`
- `services/common/client/order/generated/api.gen.go` (сгенерированный из order-service OpenAPI)
- `services/common/client/order/oapi-codegen/order-client.yaml`

**`errors.go`:**

```go
var (
    ErrOrderNotFound = errors.New("order not found")
    ErrOrderNotPaid  = errors.New("order is not in PAID status")
)
```

**`client.go`:** `OrderClient` с методами `GetOrderByID` и `UpdateOrderStatus`. `UpdateOrderStatus` пока возвращает `ErrNotImplemented` (endpoint не существует в order-service).

**Примечание:** `generated/api.gen.go` — генерируется из существующего `order-service.yaml`. Цель `generate-cl-order` добавляется в Makefile в задаче T7.2.

**`internal/client/order.go`:**

```go
type Client interface {
    GetOrderByID(ctx context.Context, orderID uuid.UUID) (*OrderResponse, error)
    UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error
}
```

**Tests:** `go build`, `go vet`

**Done when:**
- `go build ./common/client/order/...` проходит;
- `go vet ./common/client/order/...` проходит;
- Существующие тесты product-service и order-service не сломаны.

---

### T7.2 — Makefile: добавить generate-d и generate-cl-order

**Goal:** Добавить цели для генерации кода delivery-service и order-client.

**Depends on:** T7.1

**Files modified:**

- `Makefile` — добавить:
  ```makefile
  generate-d:
  	oapi-codegen \
  		-config ./services/delivery-service/api/oapi-codegen.yaml \
  		./services/delivery-service/api/delivery-service.yaml

  generate-cl-order:
  	oapi-codegen \
  		-config ./services/common/client/order/oapi-codegen/order-client.yaml \
  		./services/order-service/api/order-service.yaml
  ```

**Done when:**
- `make generate-d` генерирует `api.gen.go` (пока без спецификации — это задача T17.1);
- `make generate-cl-order` генерирует клиент для order-service.

---

## Phase 8: Service Layer

### T8.1 — Service: errors и базовая структура

**Goal:** Создать файл ошибок сервисного слоя и структуру сервиса.

**Depends on:** T6.1, T7.1

**Files created:**

- `services/delivery-service/internal/service/errors.go`
- `services/delivery-service/internal/service/service.go` (скелет)

**`service/errors.go`:**

```go
var (
    ErrInvalidID            = errors.New("invalid id")
    ErrInvalidInput         = errors.New("invalid input")
    ErrInvalidOrderID       = errors.New("invalid order id")
    ErrInvalidPickupAddress = errors.New("invalid pickup address")
    ErrInvalidEstimatedDate = errors.New("invalid estimated delivery date")
    ErrInvalidStatus        = errors.New("invalid status")
    ErrInvalidNewDate       = errors.New("invalid new date")
    ErrNilInput             = errors.New("input is nil")
    ErrForbidden            = errors.New("forbidden")
    ErrOrderNotAccessible   = errors.New("order is not accessible")
)
```

**`service/service.go`:** структура `DeliveryService` с зависимостями: `repo domain.DeliveryRepository`, `log *zap.Logger`, `orderClient client.Client`, `qrGen *qr.Generator`, `qrVer *qr.Verifier`, `validate *validator.Validate`. Конструктор `New(...)`.

Методы: заглушки, возвращающие `nil, errors.New("not implemented")`.

**Done when:**
- `go build` проходит;
- `go vet` проходит.

---

### T8.2 — Service: Create (создание доставки)

**Goal:** Реализовать `DeliveryService.Create`.

**Depends on:** T8.1

**Files modified:**

- `services/delivery-service/internal/service/service.go`

**Логика Create:**
1. Проверить input (non-nil, order_id != uuid.Nil, pickup_address не пустой, estimated_delivery_date не нулевая, не в прошлом).
2. Вызвать `orderClient.GetOrderByID` — если не найден → `ErrOrderNotFound`, если не PAID → `ErrOrderNotPaid`.
3. Проверить `repo.GetByOrderID` — если найдена → `ErrDeliveryAlreadyExists`.
4. Установить `Status = DeliveryStatusPending`, `IsLate = false`, `CreatedAt`, `UpdatedAt`.
5. Вызвать `repo.Create`.
6. Вернуть созданную доставку.

**Done when:**
- `go build` проходит;
- Метод реализован.

---

### T8.3 — Service: GetByID (просмотр доставки)

**Goal:** Реализовать `DeliveryService.GetByID` с проверкой принадлежности.

**Depends on:** T8.2

**Логика GetByID(userID, id):**
1. Проверить id != uuid.Nil.
2. Получить доставку из repo. Если не найдена → `ErrDeliveryNotFound`.
3. Вызвать `orderClient.GetOrderByID(delivery.OrderID)` — получить `UserID` заказа.
4. Проверить принадлежность: `order.UserID == userID` (для customer). Если не совпадает → `ErrForbidden`.
5. Пересчитать `is_late` если `now.After(d.EstimatedDeliveryDate)` и статус не терминальный — обновить через repo.Update.
6. Вернуть доставку.

**Примечание:** Проверка принадлежности выполняется в сервисном слое, а не в handler, т.к. требует вызова order-client.

**Done when:**
- `go build` проходит;
- Метод реализован.

---

### T8.4 — Service: List (список доставок)

**Goal:** Реализовать `DeliveryService.List` с фильтрацией.

**Depends on:** T8.3

**Логика List(userID, role, filter):**
1. Если role == "customer" — фильтровать по заказам, принадлежащим userID (вызвать orderClient для каждого заказа или передать userID в фильтр).
   **Упрощение для pet-project:** для customer возвращаем только доставки, где order_id принадлежит пользователю. Это требует дополнительного вызова order-client. Для простоты: customer не может использовать List, только GetByID. List доступен только employee.
2. Если role == "employee" — вернуть все доставки с фильтрацией.
3. Установить defaults для Page и PageSize (1 и 20).
4. Вызвать repo.List.
5. Для каждой доставки пересчитать is_late (batch update или вычислить на лету).

**Done when:**
- `go build` проходит.

---

### T8.5 — Service: UpdateStatus (state machine)

**Goal:** Реализовать `DeliveryService.UpdateStatus` с проверкой переходов.

**Depends on:** T8.4

**Логика UpdateStatus(deliveryID, newStatus):**
1. Проверить deliveryID != uuid.Nil.
2. Проверить `newStatus.IsValid()`.
3. Получить доставку из repo. Если не найдена → `ErrDeliveryNotFound`.
4. Проверить `CanTransition(delivery.Status, newStatus)`. Если нет → `ErrInvalidStatusTransition`.
5. Обновить статус: `delivery.Status = newStatus`, `delivery.UpdatedAt = now`.
6. Если `newStatus == DeliveryStatusInTransit` → вызвать `orderClient.UpdateOrderStatus(orderID, "DELIVERING")`.
7. Если `newStatus == DeliveryStatusDelivered` → обнулить `qr_nonce` (NULL).
8. Если `newStatus == DeliveryStatusCancelled` → обнулить `qr_nonce` (NULL).
9. Вызвать `repo.Update`.
10. Вернуть обновлённую доставку.

**Примечание:** Вызов `orderClient.UpdateOrderStatus` может вернуть ошибку (endpoint не существует). В этом случае логировать Warning и продолжать. Это фиксируется как integration contract task в архитектурном документе.

**Done when:**
- `go build` проходит;
- State machine корректно отклоняет недопустимые переходы.

---

### T8.6 — Service: Reschedule (перенос доставки)

**Goal:** Реализовать `DeliveryService.Reschedule`.

**Depends on:** T8.5

**Логика Reschedule(deliveryID, input):**
1. Проверить deliveryID != uuid.Nil.
2. Проверить input.NewDate не нулевая, не в прошлом.
3. Получить доставку из repo.
4. Проверить `!delivery.Status.IsTerminal()` и `delivery.Status != DeliveryStatusReadyForPickup`. Если нарушение → `ErrRescheduleOnTerminal`.
5. Создать `DeliveryReschedule` с `PreviousDate = delivery.EstimatedDeliveryDate`, `NewDate = input.NewDate`, `Reason = input.Reason`.
6. Обновить `delivery.EstimatedDeliveryDate = input.NewDate`.
7. Пересчитать `is_late` на основе `input.NewDate`.
8. Вызвать `repo.Update(delivery)`.
9. Вызвать `repo.CreateReschedule(reschedule)`.
10. Вернуть обновлённую доставку.

**Done when:**
- `go build` проходит.

---

### T8.7 — Service: GetQR (генерация QR-кода)

**Goal:** Реализовать `DeliveryService.GetQR`.

**Depends on:** T8.6

**Логика GetQR(userID, deliveryID):**
1. Проверить deliveryID != uuid.Nil.
2. Получить доставку из repo.
3. Получить заказ через orderClient → проверить `order.UserID == userID`. Иначе → `ErrForbidden`.
4. Проверить `delivery.Status == DeliveryStatusReadyForPickup`. Иначе → `ErrQRNotAvailable`.
5. Сгенерировать новый nonce: `uuid.New()`.
6. Обновить `delivery.QRNonce = nonce` через `repo.Update`.
7. Вызвать `qrGen.Generate(delivery.ID, delivery.OrderID, userID, nonce)`.
8. Вернуть `QRTokenResponse`.

**Done when:**
- `go build` проходит.

---

### T8.8 — Service: VerifyQR (верификация QR-кода)

**Goal:** Реализовать `DeliveryService.VerifyQR`.

**Depends on:** T8.7

**Логика VerifyQR(token):**
1. Вызвать `qrVer.Verify(token)` → получить `QRToken`. При ошибке → вернуть ошибку верификатора.
2. Получить доставку по `token.DeliveryID` из repo. Если не найдена → `ErrDeliveryNotFound`.
3. Проверить `delivery.Status == DeliveryStatusReadyForPickup`. Иначе → `ErrQRNotAvailable`.
4. Проверить `token.Nonce == delivery.QRNonce`. Иначе → `ErrTokenRevoked` (qr package).
5. Атомарно перевести статус в `DELIVERED`: `delivery.Status = DeliveryStatusDelivered`, `delivery.QRNonce = uuid.NullUUID{}`.
6. Вызвать `repo.Update(delivery)`.
7. Вызвать `orderClient.UpdateOrderStatus(delivery.OrderID, "DELIVERED")`.
8. Вернуть обновлённую доставку.

**Done when:**
- `go build` проходит;
- Все методы сервиса реализованы.

---

### T8.9 — Service: unit tests

**Goal:** Написать unit-тесты для всех методов DeliveryService.

**Depends on:** T8.8

**Files created:**

- `services/delivery-service/internal/service/service_test.go`
- `services/delivery-service/internal/mocks/repository.go`
- `services/delivery-service/internal/mocks/order_client.go`

**Моки:** hand-written, по конвенции проекта (`XxxFunc` + `XxxCalls`).

**Тесты (table-driven, `t.Parallel()`):**

```
TestService_Create — успех, nil input, невалидный order_id, заказ не найден, заказ не PAID, доставка уже существует, ошибка БД
TestService_GetByID — успех (employee), успех (customer, своя), 403 (чужая), не найдена, невалидный ID
TestService_List — успех (employee), фильтрация по status, фильтрация по is_late, пагинация
TestService_Reschedule — успех, невалидная дата, терминальный статус, READY_FOR_PICKUP, не найдена
TestService_UpdateStatus — каждый допустимый переход, недопустимый переход, терминальный статус, не найдена
TestService_GetQR — успех, не READY_FOR_PICKUP, чужая доставка, не найдена
TestService_VerifyQR — успех, невалидная подпись, истёкший токен, отозванный nonce, не READY_FOR_PICKUP, уже DELIVERED
```

**Done when:**
- `go test ./delivery-service/internal/service/...` проходит;
- `go vet` проходит.

---

## Phase 9: Transport Layer

### T9.1 — OpenAPI спецификация и генерация

**Goal:** Создать OpenAPI 3.1 спецификацию и сгенерировать код.

**Depends on:** T8.9

**Files created:**

- `services/delivery-service/api/delivery-service.yaml`
- `services/delivery-service/api/oapi-codegen.yaml`

**Files generated:**

- `services/delivery-service/internal/generated/openapi/api.gen.go`

**`oapi-codegen.yaml`:**

```yaml
package: api
generate:
  models: true
  chi-server: true
  strict-server: true
output: ./services/delivery-service/internal/generated/openapi/api.gen.go
```

**Запуск:** `make generate-d`

**Done when:**
- `make generate-d` проходит без ошибок;
- `api.gen.go` создан;
- `go build` компилирует сгенерированный код.

---

### T9.2 — Handler, Router, Mapper, Errors

**Goal:** Создать транспортный слой: handler, router, mapper, error mapping.

**Depends on:** T9.1

**Files created:**

- `services/delivery-service/internal/transport/handler.go`
- `services/delivery-service/internal/transport/router.go`
- `services/delivery-service/internal/transport/errors.go`
- `services/delivery-service/internal/transport/mapper.go`

**`handler.go`:**

```go
type DeliveryHandler struct {
    service    domain.DeliveryService
    log        *zap.Logger
    authorizer domain.Authorizer
}

func New(service domain.DeliveryService, log *zap.Logger, authorizer domain.Authorizer) *DeliveryHandler
```

**`router.go`:** `NewRouter(handler, jwt, authorizer)` — chi router + auth + rbac middleware + `api.HandlerFromMux`.

**`errors.go`:** функции `mapCreateDeliveryError`, `mapGetDeliveryError`, `mapListDeliveriesError`, `mapRescheduleError`, `mapUpdateStatusError`, `mapGetQRError`, `mapVerifyQRError`.

**`mapper.go`:** `toDeliveryResponse`, `toDeliveryListResponse`, `toRescheduleResponse`, `toDeliveryListFilter`.

**Done when:**
- `go build` проходит.

---

### T9.3 — Handler: CreateDelivery

**Goal:** Реализовать handler для `POST /deliveries`.

**Depends on:** T9.2

**Files created:**

- `services/delivery-service/internal/transport/create.go`

**Логика:**
1. Извлечь claims из контекста.
2. Проверить RBAC: `authorizer.Authorize(claims.Role, rbac.ResourceDelivery, rbac.ActionCreate)`.
3. Маппить request body → `DeliveryCreateInput`.
4. Вызвать `service.Create`.
5. Маппить ошибки через `mapCreateDeliveryError`.

**Done when:**
- `go build` проходит.

---

### T9.4 — Handler: GetDeliveryByID, GetDeliveries

**Goal:** Реализовать handlers для `GET /deliveries/{id}` и `GET /deliveries`.

**Depends on:** T9.3

**Files created:**

- `services/delivery-service/internal/transport/get.go`
- `services/delivery-service/internal/transport/list.go`

**GetDeliveryByID:** извлечь claims, маппить, вызвать service, маппить ошибки. Для customer — проверка принадлежности выполняется в сервисе.

**GetDeliveries:** извлечь claims, проверить RBAC (employee only), маппить параметры, вызвать service.

**Done when:**
- `go build` проходит.

---

### T9.5 — Handler: RescheduleDelivery, UpdateDeliveryStatus

**Goal:** Реализовать handlers для `PATCH /deliveries/{id}/reschedule` и `PATCH /deliveries/{id}/status`.

**Depends on:** T9.4

**Files created:**

- `services/delivery-service/internal/transport/reschedule.go`
- `services/delivery-service/internal/transport/update_status.go`

**Done when:**
- `go build` проходит.

---

### T9.6 — Handler: GetQRCode, VerifyQR

**Goal:** Реализовать handlers для QR.

**Depends on:** T9.5

**Files created:**

- `services/delivery-service/internal/transport/get_qr.go`
- `services/delivery-service/internal/transport/verify_qr.go`

**GetQRCode:** извлечь claims, вызвать service.GetQR (проверка принадлежности внутри сервиса), маппить ошибки.

**VerifyQR:** извлечь claims, проверить RBAC (`rbac.ActionVerifyQR`), вызвать service.VerifyQR, маппить результат и ошибки.

**Done when:**
- `go build` проходит.

---

### T9.7 — Middleware: auth и rbac

**Goal:** Создать middleware для аутентификации и авторизации.

**Depends on:** T9.6

**Files created:**

- `services/delivery-service/internal/transport/middlwr/auth.go`
- `services/delivery-service/internal/transport/middlwr/rbac.go`

**auth.go:** Идентичен order-service — обязательная авторизация для всех запросов.

**rbac.go:** Маппинг HTTP-методов и путей на `(ResourceDelivery, Action)`:
- POST /deliveries → (delivery, create)
- GET /deliveries/{id} → (delivery, view)
- GET /deliveries → (delivery, view)
- PATCH /deliveries/{id}/reschedule → (delivery, reschedule)
- PATCH /deliveries/{id}/status → (delivery, update_status)
- GET /deliveries/{id}/qr → (delivery, view)
- POST /deliveries/verify-qr → (delivery, verify_qr)

**Done when:**
- `go build` проходит;
- `go vet` проходит.

---

### T9.8 — Handler unit tests

**Goal:** Написать unit-тесты для всех handlers.

**Depends on:** T9.7

**Files created:**

- `services/delivery-service/internal/transport/handler_test.go`
- `services/delivery-service/internal/mocks/service.go`

**Мок:** `MockDeliveryService` с `XxxFunc` и `XxxCalls` по конвенции.

**Тесты (table-driven, `t.Parallel()`):**

```
TestHandler_CreateDelivery — успех, 400 валидация, 401 неавторизован, 403 не employee, 404 заказ не найден, 409 доставка существует, 500
TestHandler_GetDeliveryByID — успех (employee), успех (customer, своя), 403 чужая, 404 не найдена, 401
TestHandler_ListDeliveries — успех (employee), 403 customer
TestHandler_Reschedule — успех, 400 невалидная дата, 400 терминальный статус, 403 не employee, 404
TestHandler_UpdateStatus — успех (каждый переход), 400 недопустимый переход, 403 не employee, 404
TestHandler_GetQRCode — успех, 409 неверный статус, 403 чужая, 404
TestHandler_VerifyQR — успех, 400 невалидный токен, 400 истёкший, 400 отозванный, 403 не employee
```

**Done when:**
- `go test ./delivery-service/internal/transport/...` проходит;
- Все positive и negative cases покрыты.

---

## Phase 10: Wiring & Configuration

### T10.1 — Fx wiring: все модули

**Goal:** Подключить все модули в app.Module.

**Depends on:** T9.8

**Files created:**

- `services/delivery-service/internal/app/repository.go`
- `services/delivery-service/internal/app/service.go`
- `services/delivery-service/internal/app/handler.go`
- `services/delivery-service/internal/app/client.go`
- `services/delivery-service/internal/app/qr.go`

**Files modified:**

- `services/delivery-service/internal/app/module.go` — добавить все модули

**Все модули:**
- `ConfigModule`, `LoggerModule`, `DatabaseModule`, `AuthModule`, `RBACModule`, `ValidateModule`
- `ClientModule` (OrderClient)
- `RepositoryModule` (`fx.Annotate(repository.New, fx.As(new(domain.DeliveryRepository)))`)
- `ServiceModule` (`fx.Annotate(service.New, fx.As(new(domain.DeliveryService)))`)
- `QRModule` (qr.NewGenerator, qr.NewVerifier)
- `HandlerModule` (httptransport.New)
- `ServerModule` (RunServer)

**Done when:**
- `go build -o /tmp/delivery-service ./delivery-service/cmd/delivery-service/` проходит;
- Все Fx-провайдеры корректно связаны.

---

## Phase 11: Docker & Infrastructure

### T11.1 — Dockerfile

**Goal:** Создать Dockerfile по шаблону product-service/order-service.

**Depends on:** T10.1

**Files created:**

- `services/delivery-service/Dockerfile`

**Шаблон:** multi-stage build, как в product-service.

**Done when:**
- `docker build` проходит.

---

### T11.2 — docker-compose.yml, .env, .env.example

**Goal:** Добавить delivery-service в Docker Compose.

**Depends on:** T11.1

**Files modified:**

- `docker-compose.yml` — добавить сервис `delivery-service`
- `.env` — добавить `ORDER_SERVICE_URL`, `QR_SECRET_KEY`, `QR_TOKEN_TTL`
- `.env.example` — то же самое

**Done when:**
- `docker-compose up` запускает все сервисы;
- delivery-service доступен на порту 8082.

---

## Phase 12: Final Verification

### T12.1 — Полный build и vet

**Goal:** Убедиться, что весь проект компилируется и проходит статический анализ.

**Depends on:** T11.2

**Commands:**

```bash
cd services && go fmt ./...
cd services && go vet ./...
cd services && go build -o /tmp/product-service ./product-service/cmd/product-service/
cd services && go build -o /tmp/order-service ./order-service/cmd/order-service/
cd services && go build -o /tmp/delivery-service ./delivery-service/cmd/delivery-service/
```

**Done when:**
- Все команды проходят без ошибок;
- Существующие сервисы не сломаны.

---

### T12.2 — Все unit tests

**Goal:** Запустить все unit-тесты проекта.

**Depends on:** T12.1

**Commands:**

```bash
cd services && go test ./product-service/internal/... ./order-service/internal/... ./delivery-service/internal/...
```

**Done when:**
- Все тесты проходят;
- Ни один тест не пропущен.

---

### T12.3 — Integration tests

**Goal:** Запустить интеграционные тесты (требуют Docker).

**Depends on:** T12.2

**Commands:**

```bash
cd services && go test ./delivery-service/internal/repository/... -tags=integration
```

**Done when:**
- Интеграционные тесты проходят;
- Миграции применяются корректно.

---

### T12.4 — Consistency check

**Goal:** Финальная проверка соответствия спецификации, архитектуре и конвенциям.

**Depends on:** T12.3

**Checklist:**

- [ ] Все файлы следуют конвенциям именования (packages, constructors, Fx wiring)
- [ ] Все sentinel errors определены в domain/errors.go и service/errors.go
- [ ] Error mapping в transport/errors.go покрывает все случаи
- [ ] RBAC-политика обновлена в common/rbac/policy.csv
- [ ] OpenAPI-спецификация актуальна
- [ ] Docker Compose включает delivery-service
- [ ] .env и .env.example обновлены
- [ ] Makefile включает generate-d
- [ ] Все интеграционные контракты задокументированы в архитектурном документе
- [ ] QR-механизм реализован корректно (HMAC-SHA256, nonce, TTL)
- [ ] State machine реализована корректно (CanTransition)
- [ ] Все handler-тесты покрывают positive и negative cases

**Done when:**
- Все чеклист-пункты проверены и подтверждены.

---

## Summary: Task Dependency Graph

```
T1.1 ──► T1.2 ──► T1.3 ──► T1.4
                                    │
                                    ▼
                               T2.1 ──► T3.1
                                                │
                                                ▼
                                          T4.1 ──► T4.2 ──► T4.3
                                                                    │
                                                                    ▼
                                                              T5.1 ──► T5.2 ──► T5.3
                                                                                                │
                                                                                                ▼
                                                                                          T6.1 ──► T6.2
                                                                                                    │
                                                                                                    ▼
T7.1 ──► T7.2                                                                                         │
  │                                                                                                     │
  ▼                                                                                                     ▼
T8.1 ──► T8.2 ──► T8.3 ──► T8.4 ──► T8.5 ──► T8.6 ──► T8.7 ──► T8.8 ──► T8.9
                                                                                          │
                                                                                          ▼
                                                                                    T9.1 ──► T9.2 ──► T9.3 ──► T9.4 ──► T9.5 ──► T9.6 ──► T9.7 ──► T9.8
                                                                                                                                                                    │
                                                                                                                                                                    ▼
                                                                                                                                                              T10.1
                                                                                                                                                                │
                                                                                                                                                                ▼
                                                                                                                                                          T11.1 ──► T11.2
                                                                                                                                                                        │
                                                                                                                                                                        ▼
                                                                                                                                                                  T12.1 ──► T12.2 ──► T12.3 ──► T12.4
```

Total: 34 tasks across 12 phases.