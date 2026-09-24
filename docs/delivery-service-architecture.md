# Delivery Service — Architecture Document

## 1. Структура delivery-service

```
services/delivery-service/
├── api/
│   ├── delivery-service.yaml          # OpenAPI 3.1 спецификация
│   └── oapi-codegen.yaml             # Конфигурация генератора
├── cmd/
│   └── delivery-service/
│       └── main.go                    # fx.New(app.Module).Run()
├── config.yaml                        # YAML-конфиг
├── Dockerfile                         # Multi-stage build
├── internal/
│   ├── app/
│   │   ├── module.go                  # Fx-модуль
│   │   ├── config.go                  # Fx: config.Load
│   │   ├── logger.go                  # Fx: zap.NewProduction
│   │   ├── database.go                # Fx: database.New
│   │   ├── auth.go                    # Fx: auth.NewValidator
│   │   ├── rbac.go                    # Fx: rbac.NewEnforcer + rbac.New + NewDomainAuthorizer
│   │   ├── repository.go             # Fx: fx.Annotate(repository.New, fx.As(domain.DeliveryRepository))
│   │   ├── service.go                # Fx: fx.Annotate(service.New, fx.As(domain.DeliveryService))
│   │   ├── handler.go                # Fx: httptransport.New
│   │   ├── server.go                 # Fx: RunServer
│   │   ├── client.go                 # Fx: fx.Annotate(order.NewOrderClient, fx.As(client.Client))
│   │   └── validate.go               # Fx: validator.New
│   ├── client/
│   │   └── order.go                  # Интерфейс OrderClient
│   ├── config/
│   │   └── config.go                 # Config + Load() через yaml.Unmarshal + os.ExpandEnv
│   ├── domain/
│   │   ├── delivery.go              # Delivery, DeliveryStatus, DeliveryList, DeliveryListFilter
│   │   ├── reschedule.go            # DeliveryReschedule
│   │   ├── repository.go            # DeliveryRepository interface
│   │   ├── service.go               # DeliveryService interface
│   │   ├── errors.go                # Sentinel errors
│   │   └── auth.go                  # Authorizer interface
│   ├── generated/
│   │   └── openapi/
│   │       └── api.gen.go           # oapi-codegen strict-server
│   ├── mocks/
│   │   ├── repository.go           # MockDeliveryRepository
│   │   ├── service.go              # MockDeliveryService
│   │   └── order_client.go         # MockOrderClient
│   ├── qr/
│   │   ├── token.go                # QRToken struct, Encode/Decode
│   │   ├── generator.go            # Generate(deliveryID, orderID, userID, nonce, ttl, secret)
│   │   └── verifier.go             # Verify(token, secret) → *QRToken, error
│   ├── repository/
│   │   └── repository.go           # DeliveryRepository implementation
│   ├── service/
│   │   ├── service.go              # DeliveryService implementation
│   │   └── errors.go              # Service-level errors
│   └── transport/
│       ├── handler.go              # DeliveryHandler struct
│       ├── router.go              # NewRouter(handler, validator, authorizer)
│       ├── errors.go              # mapXxxError functions
│       ├── mapper.go              # toResponse, toDeliveryList, toRescheduleList
│       ├── create.go              # CreateDelivery handler
│       ├── get.go                 # GetDeliveryByID handler
│       ├── list.go                # GetDeliveries handler
│       ├── reschedule.go          # RescheduleDelivery handler
│       ├── update_status.go       # UpdateDeliveryStatus handler
│       ├── get_qr.go             # GetQRCode handler
│       ├── verify_qr.go          # VerifyQR handler
│       └── middlwr/
│           ├── auth.go           # Auth middleware
│           └── rbac.go           # Authorization middleware
```

### Дополнения в common

```
services/common/
├── client/
│   ├── order/                          # Новый HTTP-клиент
│   │   ├── client.go                  # OrderClient struct + GetOrderByID + UpdateOrderStatus
│   │   ├── errors.go                  # ErrOrderNotFound, ErrOrderNotPaid
│   │   └── generated/
│   │       └── api.gen.go            # Сгенерированный из order-service.yaml
│   └── product/                       # Без изменений
├── rbac/
│   ├── authorizer.go                 # + ResourceDelivery, ActionReschedule, ActionUpdateStatus, ActionVerifyQR
│   ├── model.conf                    # Без изменений
│   └── policy.csv                    # + строки для delivery
├── auth/                             # Без изменений
├── database/                         # Без изменений
└── test-tools/                       # Без изменений
```

### Новые миграции

```
migrations/
├── 000003_create_deliveries.up.sql
└── 000003_create_deliveries.down.sql
```

---

## 2. Domain Entities

### 2.1 Delivery (delivery.go)

```go
package domain

import (
    "time"
    "github.com/google/uuid"
    "github.com/uptrace/bun"
)

type DeliveryStatus string

const (
    DeliveryStatusPending       DeliveryStatus = "PENDING"
    DeliveryStatusInTransit     DeliveryStatus = "IN_TRANSIT"
    DeliveryStatusReadyForPickup DeliveryStatus = "READY_FOR_PICKUP"
    DeliveryStatusDelivered     DeliveryStatus = "DELIVERED"
    DeliveryStatusCancelled     DeliveryStatus = "CANCELLED"
)

func (s DeliveryStatus) IsValid() bool {
    switch s {
    case DeliveryStatusPending,
        DeliveryStatusInTransit,
        DeliveryStatusReadyForPickup,
        DeliveryStatusDelivered,
        DeliveryStatusCancelled:
        return true
    }
    return false
}

func (s DeliveryStatus) IsTerminal() bool {
    return s == DeliveryStatusDelivered || s == DeliveryStatusCancelled
}

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

type DeliveryList struct {
    Items    []*Delivery
    Total    int64
    Page     int
    PageSize int
}

type DeliveryListFilter struct {
    Status     *DeliveryStatus
    IsLate     *bool
    OrderID    *uuid.UUID
    Page       int
    PageSize   int
}

type DeliveryCreateInput struct {
    OrderID               uuid.UUID
    PickupAddress         string
    EstimatedDeliveryDate time.Time
}
```

### 2.2 DeliveryReschedule (reschedule.go)

```go
package domain

import (
    "time"
    "github.com/google/uuid"
    "github.com/uptrace/bun"
)

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

### 2.3 QRToken (qr/token.go)

```go
package qr

import "github.com/google/uuid"

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
```

---

## 3. DTO

### 3.1 Маппинг domain → API

Конвенция проекта: функции-мапперы в `transport/mapper.go`.

```go
// mapper.go

func toDeliveryResponse(d *domain.Delivery, reschedules []domain.DeliveryReschedule) api.DeliveryResponse

func toRescheduleResponse(r *domain.DeliveryReschedule) api.DeliveryRescheduleResponse

func toDeliveryListResponse(list *domain.DeliveryList, rescheduleMap map[uuid.UUID][]domain.DeliveryReschedule) api.DeliveryListResponse

func toDeliveryListFilter(params api.GetDeliveriesParams) domain.DeliveryListFilter
```

### 3.2 API-модели (определяются через OpenAPI, генерируются oapi-codegen)

Ключевые типы из `generated/openapi/api.gen.go`:
- `DeliveryStatus` (enum)
- `DeliveryCreateRequest`
- `DeliveryRescheduleRequest`
- `DeliveryStatusUpdateRequest`
- `DeliveryResponse`
- `DeliveryRescheduleResponse`
- `DeliveryListResponse`
- `QRCodeResponse`
- `VerifyQRRequest`
- `VerifyQRResponse`
- `ErrorResponse`

### 3.3 Service Input DTO

```go
// domain/service.go — входные типы для сервисного слоя

type DeliveryCreateInput struct {
    OrderID               uuid.UUID
    PickupAddress         string
    EstimatedDeliveryDate time.Time
}

type RescheduleInput struct {
    NewDate time.Time
    Reason  string
}
```

Сервисный слой принимает domain-типы или простые параметры (uuid.UUID, DeliveryStatus), а не API-типы. Handler маппит API-типы → domain-типы.

---

## 4. Repository Interfaces

```go
// domain/repository.go

package domain

import (
    "context"
    "github.com/google/uuid"
)

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

**Конвенция:** как в product-service и order-service — интерфейс в `domain/`, реализация в `repository/`. Конструктор `New(*bun.DB, *zap.Logger) *DeliveryRepository`. Привязка через `fx.Annotate(repository.New, fx.As(new(domain.DeliveryRepository)))`.

---

## 5. Service/Use-case Interfaces

```go
// domain/service.go

package domain

import (
    "context"
    "github.com/google/uuid"
)

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

**Конвенция:** как в order-service — интерфейс в `domain/`, реализация в `service/`. Привязка через `fx.Annotate(service.New, fx.As(new(domain.DeliveryService)))`.

Зависимости `service.New`:
- `domain.DeliveryRepository`
- `*zap.Logger`
- `client.Client` (OrderClient)
- `*qr.Generator`
- `*qr.Verifier`

---

## 6. HTTP Handlers

### 6.1 Handler struct

```go
// transport/handler.go

type DeliveryHandler struct {
    service    domain.DeliveryService
    log        *zap.Logger
    authorizer domain.Authorizer
}

func New(service domain.DeliveryService, log *zap.Logger, authorizer domain.Authorizer) *DeliveryHandler
```

Следует конвенции order-service: handler хранит `domain.Authorizer` для программных RBAC-проверок.

### 6.2 Endpoints

| Файл               | Метод   | Путь                          | StrictServerInterface-метод      |
|---------------------|---------|-------------------------------|-----------------------------------|
| `create.go`         | POST    | /deliveries                   | `CreateDelivery`                  |
| `get.go`            | GET     | /deliveries/{id}              | `GetDeliveryByID`                 |
| `list.go`           | GET     | /deliveries                   | `GetDeliveries`                   |
| `reschedule.go`     | PATCH   | /deliveries/{id}/reschedule   | `RescheduleDelivery`              |
| `update_status.go`  | PATCH   | /deliveries/{id}/status       | `UpdateDeliveryStatus`            |
| `get_qr.go`         | GET     | /deliveries/{id}/qr           | `GetQRCode`                       |
| `verify_qr.go`      | POST    | /deliveries/verify-qr         | `VerifyQR`                        |

### 6.3 Handler-паттерн

Каждый handler-метод:
1. Извлекает claims из контекста через `auth.ClaimsFromContext(ctx)`.
2. Проверяет RBAC через `h.authorizer.Authorize(claims.Role, resource, action)` (для employee-операций).
3. Маппит входные данные в domain-типы через mapper-функции.
4. Вызывает сервисный метод.
5. Маппит ошибки через `mapXxxError(h.log, err)`.
6. Возвращает типизированный ответ.

---

## 7. Middleware

### 7.1 Auth middleware (`middlwr/auth.go`)

Полностью идентичен order-service: обязательная авторизация для **всех** запросов (все endpoints delivery-service требуют JWT).

```go
type Auth struct {
    validator *auth.Validator
}

func NewAuth(v *auth.Validator) *Auth
func (m *Auth) Handler(next http.Handler) http.Handler
```

Логика: извлечь `Authorization: Bearer <token>` → `validator.Parse(token)` → `auth.ContextWithClaims(ctx, claims)` → `next.ServeHTTP(w, r.WithContext(ctx))`.

### 7.2 RBAC middleware (`middlwr/rbac.go`)

Идентичен по структуре order-service. Маппит HTTP-метод + путь → (resource, action):

```go
func permission(r *http.Request) (rbac.Resource, rbac.Action) {
    switch {
    case r.Method == http.MethodPost && r.URL.Path == "/deliveries":
        return rbac.ResourceDelivery, rbac.ActionCreate

    case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/deliveries/") &&
         !strings.HasSuffix(r.URL.Path, "/reschedule") &&
         !strings.HasSuffix(r.URL.Path, "/status") &&
         !strings.HasSuffix(r.URL.Path, "/qr"):
        return rbac.ResourceDelivery, rbac.ActionView

    case r.Method == http.MethodGet && r.URL.Path == "/deliveries":
        return rbac.ResourceDelivery, rbac.ActionView

    case r.Method == http.MethodPatch &&
         strings.HasPrefix(r.URL.Path, "/deliveries/") &&
         strings.HasSuffix(r.URL.Path, "/reschedule"):
        return rbac.ResourceDelivery, rbac.ActionReschedule

    case r.Method == http.MethodPatch &&
         strings.HasPrefix(r.URL.Path, "/deliveries/") &&
         strings.HasSuffix(r.URL.Path, "/status"):
        return rbac.ResourceDelivery, rbac.ActionUpdateStatus

    case r.Method == http.MethodGet &&
         strings.HasPrefix(r.URL.Path, "/deliveries/") &&
         strings.HasSuffix(r.URL.Path, "/qr"):
        return rbac.ResourceDelivery, rbac.ActionView

    case r.Method == http.MethodPost && r.URL.Path == "/deliveries/verify-qr":
        return rbac.ResourceDelivery, rbac.ActionVerifyQR
    }
    return "", ""
}
```

---

## 8. Authorization

### 8.1 Casbin-политика (дополнение к `common/rbac/policy.csv`)

```csv
p, customer, delivery, view
p, employee, delivery, create
p, employee, delivery, view
p, employee, delivery, reschedule
p, employee, delivery, update_status
p, employee, delivery, verify_qr
```

### 8.2 Новые константы (дополнение к `common/rbac/authorizer.go`)

```go
const (
    ResourceDelivery  Resource = "delivery"

    ActionReschedule   Action = "reschedule"
    ActionUpdateStatus Action = "update_status"
    ActionVerifyQR     Action = "verify_qr"
)
```

### 8.3 Domain Authorizer interface (как в order-service)

```go
// domain/auth.go

package domain

import "github.com/byorty/test-marketplace/services/common/rbac"

type Authorizer interface {
    Authorize(role string, resource rbac.Resource, action rbac.Action) error
}
```

Привязка в `app/rbac.go`:

```go
func NewDomainAuthorizer(authorizer *rbac.Authorizer) domain.Authorizer {
    return authorizer
}

var RBACModule = fx.Provide(
    rbac.NewEnforcer,
    rbac.New,
    NewDomainAuthorizer,
)
```

### 8.4 Проверка принадлежности (дополнительная авторизация)

Для customer-доступа (`GetDeliveryByID`, `GetQRCode`) после Casbin-проверки `delivery.view` выполняется дополнительная проверка: `order.user_id == claims.UserID`. Эта проверка выполняется в handler-слое (как в `GetOrderByID` в order-service).

---

## 9. Database Schema

### 9.1 Миграция UP (000003_create_deliveries.up.sql)

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

### 9.2 Миграция DOWN (000003_create_deliveries.down.sql)

```sql
DROP INDEX IF EXISTS idx_delivery_reschedules_delivery_id;
DROP INDEX IF EXISTS idx_deliveries_is_late;
DROP INDEX IF EXISTS idx_deliveries_status;
DROP INDEX IF EXISTS idx_deliveries_order_id;
DROP TABLE IF EXISTS delivery_reschedules;
DROP TABLE IF EXISTS deliveries;
```

### 9.3 Обоснование schema-решений

- `UNIQUE` на `order_id` — одна активная доставка на заказ (бизнес-правило BR-1).
- `qr_nonce UUID NULLABLE` — NULL означает, что QR не генерировался; при генерации нового QR старый инвалидируется заменой nonce.
- `is_late BOOLEAN` — вычисляемое поле, обновляемое при каждом запросе и при переносе.
- `ON DELETE CASCADE` на `delivery_reschedules.delivery_id` — история переносов удаляется вместе с доставкой (если это потребуется).
- Индексы на `order_id`, `status`, `is_late` — для эффективной фильтрации в `List`.

---

## 10. Migrations

Следуют конвенции проекта: числовая нумерация `NNNNNN_create_<table>.{up,down}.sql` в корневом каталоге `migrations/`.

Применение через:
- `make migrate-up` (локально, через Docker)
- Docker Compose: контейнер `migrate` применяет все миграции при старте
- Интеграционные тесты: `testtools.NewTestDB(t)` в `common/test-tools/helper.go`

Новые файлы:
- `migrations/000003_create_deliveries.up.sql`
- `migrations/000003_create_deliveries.down.sql`

---

## 11. Dependency Graph

```
                          ┌──────────────┐
                          │  cmd/main.go │
                          └──────┬───────┘
                                 │ fx.New(app.Module)
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                       app.Module                                │
│                                                                 │
│  ConfigModule ──► LoggerModule ──► DatabaseModule               │
│       │               │                │                         │
│       ▼               ▼                ▼                         │
│  AuthModule ──► RBACModule ──► ClientModule                     │
│       │               │                │                         │
│       ▼               ▼                ▼                         │
│  RepositoryModule ──► ServiceModule ──► HandlerModule             │
│                           │                │                     │
│                           ▼                ▼                     │
│                      QRModule         ServerModule               │
│                                          │                      │
└──────────────────────────────────────────┼──────────────────────┘
                                           │
                                           ▼
                                    HTTP Server (chi)
```

### Fx-провайдеры и invoke

```go
var Module = fx.Options(
    ConfigModule,      // fx.Provide(config.Load)
    LoggerModule,      // fx.Provide(NewLogger)
    DatabaseModule,    // fx.Provide(NewDB)
    AuthModule,        // fx.Provide(NewJWTValidator)
    RBACModule,        // fx.Provide(rbac.NewEnforcer, rbac.New, NewDomainAuthorizer)
    ClientModule,     // fx.Provide(fx.Annotate(order.NewOrderClient, fx.As(new(client.Client))))
    RepositoryModule, // fx.Provide(fx.Annotate(repository.New, fx.As(new(domain.DeliveryRepository))))
    ServiceModule,    // fx.Provide(fx.Annotate(service.New, fx.As(new(domain.DeliveryService))))
    QRModule,         // fx.Provide(qr.NewGenerator) + fx.Provide(qr.NewVerifier)
    HandlerModule,    // fx.Provide(httptransport.New)
    ValidateModule,   // fx.Provide(NewValidator)
    ServerModule,     // fx.Invoke(RunServer)
)
```

---

## 12. Fx Modules

### 12.1 app/config.go

```go
var ConfigModule = fx.Provide(config.Load)
```

### 12.2 app/logger.go

```go
func NewLogger() *zap.Logger {
    logger, err := zap.NewProduction()
    if err != nil {
        panic(err)
    }
    return logger
}

var LoggerModule = fx.Provide(NewLogger)
```

Идентично product-service и order-service.

### 12.3 app/database.go

```go
func NewDB(cfg *config.Config) (*bun.DB, error) {
    return database.New(database.PostgresConfig(cfg.Postgres))
}

var DatabaseModule = fx.Provide(NewDB)
```

Идентично обоим сервисам.

### 12.4 app/auth.go

```go
func NewJWTValidator(cfg *config.Config) (*auth.Validator, error) {
    publicKey, err := auth.LoadPublicKey(cfg.JWT.PublicKeyPath)
    if err != nil {
        return nil, err
    }
    return auth.NewValidator(publicKey, cfg.JWT.Issuer), nil
}

var AuthModule = fx.Provide(NewJWTValidator)
```

Идентично обоим сервисам.

### 12.5 app/rbac.go

```go
func NewDomainAuthorizer(authorizer *rbac.Authorizer) domain.Authorizer {
    return authorizer
}

var RBACModule = fx.Provide(
    rbac.NewEnforcer,
    rbac.New,
    NewDomainAuthorizer,
)
```

Идентично order-service.

### 12.6 app/client.go

```go
var ClientModule = fx.Provide(
    fx.Annotate(
        order.NewOrderClient,
        fx.As(new(client.Client)),
    ),
)
```

Аналогично order-service `app/client.go`, но для order-service вместо product-service.

### 12.7 app/validate.go

```go
func NewValidator() *validator.Validate {
    return validator.New()
}

var ValidateModule = fx.Provide(NewValidator)
```

Идентично product-service.

### 12.8 app/qr.go (новый модуль)

```go
func NewGenerator(cfg *config.Config) *qr.Generator {
    return qr.NewGenerator(cfg.QR.SecretKey, cfg.QR.TokenTTL)
}

func NewVerifier(cfg *config.Config) *qr.Verifier {
    return qr.NewVerifier(cfg.QR.SecretKey)
}

var QRModule = fx.Provide(NewGenerator, NewVerifier)
```

---

## 13. Configuration

### 13.1 config/config.go

Следует конвенции проекта: `yaml.Unmarshal` + `os.ExpandEnv`. **Без cleanenv.**

```go
package config

import (
    "fmt"
    "os"
    "time"

    "gopkg.in/yaml.v3"
)

type Config struct {
    HTTP          HTTPConfig        `yaml:"http"`
    Postgres      PostgresConfig    `yaml:"postgres"`
    Log           LogConfig         `yaml:"log"`
    JWT           JWT               `yaml:"jwt"`
    OrderService  OrderServiceConfig `yaml:"order_service"`
    QR            QRConfig          `yaml:"qr"`
}

type HTTPConfig struct {
    Host string `yaml:"host" env:"HTTP_HOST" env-default:"0.0.0.0"`
    Port int    `yaml:"port" env:"HTTP_PORT" env-default:"8082"`
}

func (h HTTPConfig) Address() string {
    return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

type PostgresConfig struct {
    Host            string        `yaml:"host" env:"POSTGRES_HOST" env-required:"true"`
    Port            int           `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
    User            string        `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
    Password        string        `yaml:"password" env:"POSTGRES_PASSWORD"`
    Database        string        `yaml:"database" env:"POSTGRES_DB"`
    SSLMode         string        `yaml:"sslmode" env:"POSTGRES_SSLMODE" env-default:"disable"`
    MaxOpenConns    int           `yaml:"max_open_conns" env-default:"20"`
    MaxIdleConns    int           `yaml:"max_idle_conns" env-default:"10"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env-default:"30m"`
    ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env-default:"15m"`
}

type LogConfig struct {
    Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

type JWT struct {
    Issuer        string `yaml:"issuer" env:"JWT_ISSUER" env-required:"true"`
    PublicKeyPath string `yaml:"public_key_path" env:"JWT_PUBLIC_KEY_PATH" env-required:"true"`
}

type OrderServiceConfig struct {
    URL string `yaml:"url"`
}

type QRConfig struct {
    SecretKey string        `yaml:"secret_key" env:"QR_SECRET_KEY"`
    TokenTTL  time.Duration `yaml:"token_ttl" env:"QR_TOKEN_TTL" env-default:"24h"`
}

func Load() (*Config, error) {
    configPath := os.Getenv("CONFIG_PATH")
    if configPath == "" {
        configPath = "config/config.yaml"
    }

    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("read config %q: %w", configPath, err)
    }

    data = []byte(os.ExpandEnv(string(data)))

    var cfg Config

    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config %q: %w", configPath, err)
    }

    return &cfg, nil
}
```

### 13.2 config.yaml

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

### 13.3 .env дополнения

```
ORDER_SERVICE_URL=http://order-service:8081
QR_SECRET_KEY=<generate-32-byte-secret>
QR_TOKEN_TTL=24h
```

---

## 14. Logging

Следует конвенции проекта: `zap.NewProduction()`, именованные логгеры через `zap.Named()`.

```go
// repository
log.Named("delivery-repository")

// service
log.Named("delivery-service")

// handler
log.Named("delivery-handler")

// client
log.Named("order-client")
```

Паттерн логирования: Info при успехе, Error при неудаче, Warn при некритичных ситуациях. Ключевые поля: `delivery_id`, `order_id`, `user_id`, `duration`.

---

## 15. Error Model

### 15.1 Domain errors (domain/errors.go)

```go
package domain

import "errors"

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

### 15.2 Service errors (service/errors.go)

```go
package service

import "errors"

var (
    ErrInvalidID             = errors.New("invalid id")
    ErrInvalidInput          = errors.New("invalid input")
    ErrInvalidOrderID        = errors.New("invalid order id")
    ErrInvalidPickupAddress  = errors.New("invalid pickup address")
    ErrInvalidEstimatedDate  = errors.New("invalid estimated delivery date")
    ErrInvalidStatus         = errors.New("invalid status")
    ErrInvalidNewDate        = errors.New("invalid new date")
    ErrNilInput              = errors.New("input is nil")
    ErrForbidden             = errors.New("forbidden")
)
```

### 15.3 QR errors (qr/)

```go
package qr

import "errors"

var (
    ErrInvalidToken        = errors.New("invalid QR token")
    ErrTokenExpired         = errors.New("QR token expired")
    ErrTokenRevoked         = errors.New("QR token has been revoked")
    ErrInvalidSignature    = errors.New("invalid QR token signature")
)
```

### 15.4 Client errors (common/client/order/errors.go)

```go
package order

import "errors"

var (
    ErrOrderNotFound = errors.New("order not found")
    ErrOrderNotPaid  = errors.New("order is not in PAID status")
)
```

### 15.5 Error mapping в transport/errors.go

```go
func errorResponse(code, message string) api.ErrorResponse {
    return api.ErrorResponse{Code: code, Message: message}
}

func mapCreateDeliveryError(log *zap.Logger, err error) api.CreateDeliveryResponseObject
func mapGetDeliveryError(log *zap.Logger, err error) api.GetDeliveryByIDResponseObject
func mapListDeliveriesError(log *zap.Logger, err error) api.GetDeliveriesResponseObject
func mapRescheduleError(log *zap.Logger, err error) api.RescheduleDeliveryResponseObject
func mapUpdateStatusError(log *zap.Logger, err error) api.UpdateDeliveryStatusResponseObject
func mapGetQRError(log *zap.Logger, err error) api.GetQRCodeResponseObject
func mapVerifyQRError(log *zap.Logger, err error) api.VerifyQRResponseObject
```

Каждая функция маппит `errors.Is` → соответствующий HTTP-ответ (400/401/403/404/409/500), по конвенции проекта.

---

## 16. Validation

Используется `go-playground/validator/v10`, как в product-service.

OpenAPI-схема содержит `x-oapi-codegen-extra-tags` для правил валидации:

```yaml
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
      x-oapi-codegen-extra-tags:
        validate: "required,min=1,max=500"
    estimated_delivery_date:
      type: string
      format: date-time
      x-oapi-codegen-extra-tags:
        validate: "required"
```

Валидация вызывается в сервисном слое: `s.validate.Struct(input)`, аналогично product-service.

---

## 17. OpenAPI Structure

### 17.1 delivery-service.yaml (структура)

```yaml
openapi: 3.1.0

info:
  title: Delivery service API
  description: "Микросервис доставки для маркетплейса"
  version: 1.0.0
  contact:
    name: Полина Соломонова
    email: snoomew@mail.ru

tags:
  - name: Deliveries
  - name: QR

paths:
  /deliveries:
    post: ...
    get: ...

  /deliveries/{id}:
    get: ...

  /deliveries/{id}/reschedule:
    patch: ...

  /deliveries/{id}/status:
    patch: ...

  /deliveries/{id}/qr:
    get: ...

  /deliveries/verify-qr:
    post: ...

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    DeliveryStatus: ...
    DeliveryCreateRequest: ...
    DeliveryRescheduleRequest: ...
    DeliveryStatusUpdateRequest: ...
    VerifyQRRequest: ...
    DeliveryResponse: ...
    DeliveryRescheduleResponse: ...
    DeliveryListResponse: ...
    QRCodeResponse: ...
    VerifyQRResponse: ...
    ErrorResponse: ...
```

### 17.2 oapi-codegen.yaml

```yaml
package: api

generate:
  models: true
  chi-server: true
  strict-server: true

output: ./services/delivery-service/internal/generated/openapi/api.gen.go
```

---

## 18. Delivery State Machine

### 18.1 Состояния

| Статус               | Смысл                                                   |
|-----------------------|---------------------------------------------------------|
| `PENDING`             | Доставка создана, ожидает назначения курьера/логистики  |
| `IN_TRANSIT`          | Доставка в пути к пункту выдачи                         |
| `READY_FOR_PICKUP`    | Доставка прибыла на пункт выдачи, клиент может получить QR |
| `DELIVERED`           | Заказ выдан клиенту (QR верифицирован) — **терминальный** |
| `CANCELLED`           | Доставка отменена — **терминальный**                   |

### 18.2 Допустимые переходы

| Из                    | В                      | Кто инициирует | Условие                     |
|-----------------------|------------------------|----------------|-----------------------------|
| `PENDING`              | `IN_TRANSIT`           | employee       | Курьер забрал доставку       |
| `PENDING`              | `CANCELLED`            | employee       | Отмена до отправки          |
| `IN_TRANSIT`           | `READY_FOR_PICKUP`     | employee       | Прибытие на пункт выдачи    |
| `IN_TRANSIT`           | `CANCELLED`            | employee       | Отмена в пути               |
| `READY_FOR_PICKUP`    | `DELIVERED`            | employee       | Верификация QR (автоматически) |
| `READY_FOR_PICKUP`    | `CANCELLED`            | employee       | Исключительная отмена        |

### 18.3 Запрещённые переходы

| Переход                                   | Причина                                     |
|-------------------------------------------|---------------------------------------------|
| `DELIVERED` → любой статус                | Терминальный, неизменяемый                  |
| `CANCELLED` → любой статус                | Терминальный, неизменяемый                  |
| `PENDING` → `READY_FOR_PICKUP`           | Пропуск IN_TRANSIT                          |
| `PENDING` → `DELIVERED`                  | Пропуск промежуточных статусов              |
| `IN_TRANSIT` → `PENDING`                 | Обратный переход запрещён                   |
| `IN_TRANSIT` → `DELIVERED`               | Пропуск READY_FOR_PICKUP                    |
| `READY_FOR_PICKUP` → `PENDING`           | Обратный переход запрещён                   |
| `READY_FOR_PICKUP` → `IN_TRANSIT`        | Обратный переход запрещён                   |
| Любой → `PENDING`                         | Создание — только через POST /deliveries    |

### 18.4 Поведение после получения заказа (DELIVERED)

- Статус `DELIVERED` — терминальный. Переходы из него невозможны.
- При переходе `READY_FOR_PICKUP` → `DELIVERED`:
  - delivery-service обновляет статус доставки.
  - delivery-service вызывает `PATCH /orders/{id}/status` в order-service со статусом `DELIVERED`.
  - QR-код, связанный с доставкой, больше не может быть сгенерирован (статус не `READY_FOR_PICKUP`).
  - QR-код с тем же nonce не может быть использован повторно (доставка уже `DELIVERED`).

### 18.5 Перенос доставки (Reschedule)

Допустим из статусов: `PENDING`, `IN_TRANSIT`.
Недопустим из: `READY_FOR_PICKUP`, `DELIVERED`, `CANCELLED`.

При переносе:
1. Проверяется, что статус не терминальный и не `READY_FOR_PICKUP`.
2. `estimated_delivery_date` обновляется на `new_date`.
3. Создаётся запись в `delivery_reschedules` с `previous_date`, `new_date`, `reason`, `created_at`.
4. Пересчитывается `is_late`: если `new_date` в будущем → `false`, иначе `true`.

---

## 19. QR Security

### 19.1 Сравнение подходов

| Критерий                    | HMAC-SHA256                                       | RSA/ECDSA цифровая подпись                    |
|-----------------------------|---------------------------------------------------|-----------------------------------------------|
| Ключ                        | Один общий секретный ключ                         | Пара: приватный ключ для подписи, публичный для проверки |
| Длина подписи                | 32 байта                                         | RSA-2048: 256 байт; ECDSA P-256: 64 байта    |
| Производительность           | ~0.5 μs/op                                        | RSA: ~100 μs/sign; ECDSA: ~50 μs/sign        |
| Распределение ключей         | Один ключ на все инстансы delivery-service        | Приватный ключ только для генерации, публичный для проверки |
| Риск компрометации          | Компрометация одного ключа = компрометация всех токенов | Компрометация приватного ключа = компрометация генерации, но не позволяет подделать уже выпущенные токены |
| Сложность реализации         | Простая                                           | Умеренная                                     |
| Согласованность с проектом  | Проект использует RSA для JWT, но QR — внутренний механизм | Не добавляет значимой выгоды для внутреннего сервиса |

**Выбор: HMAC-SHA256.**

Обоснование:
1. **Производительность:** HMAC-SHA256 на порядок быстрее RSA. QR-коды генерируются и верифицируются синхронно в HTTP-запросах.
2. **Простота:** Один симметричный ключ, единая логика генерации и верификации. Не требуется управление ключевыми парами.
3. **Достаточность:** delivery-service — единственный генератор и верификатор QR-кодов. Нет необходимости в асимметричном разделении ролей.
4. **Размер payload:** HMAC-SHA256 подпись — 32 байта (base64url ≈ 43 символа), что значительно компактнее RSA-2048 (256 байт ≈ 344 символа). Меньший QR-код легче сканировать.
5. **Согласованность:** Проект уже использует RS256 для JWT, но JWT — это внешний токен, верифицируемый публичным ключом. QR — внутренний механизм одного сервиса.

### 19.2 QR-механизм

#### Payload

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

Поля:
- `delivery_id` — связь с конкретной доставкой.
- `order_id` — связь с заказом (для проверки принадлежности).
- `user_id` — владелец QR-кода.
- `nonce` — совпадает с `delivery.qr_nonce` в БД, используется для инвалидации.
- `iat` — unix timestamp выдачи (необязательно, но полезно для аудита).
- `exp` — unix timestamp истечения.

#### Подпись

```
signature = HMAC-SHA256(base64url(JSON(payload)), secret_key)
```

Секретный ключ: минимум 32 байта, конфигурируемый через `QR_SECRET_KEY`.

#### Формат токена

```
base64url(JSON(payload)) + "." + base64url(signature)
```

#### Генерация (qr/generator.go)

```go
type Generator struct {
    secretKey []byte
    tokenTTL  time.Duration
}

func NewGenerator(secretKey string, tokenTTL time.Duration) *Generator

func (g *Generator) Generate(deliveryID, orderID, userID uuid.UUID, nonce uuid.UUID) (*QRTokenResponse, error)
```

Логика:
1. Вычислить `iat = now`, `exp = now + tokenTTL`.
2. Сформировать payload.
3. Вычислить HMAC-SHA256-подпись.
4. Вернуть `QRTokenResponse{Token: encoded, ExpiresAt: exp}`.

#### Верификация (qr/verifier.go)

```go
type Verifier struct {
    secretKey []byte
}

func NewVerifier(secretKey string) *Verifier

func (v *Verifier) Verify(tokenString string) (*QRToken, error)
```

Логика:
1. Разделить токен на `payload` и `signature`.
2. Вычислить HMAC-SHA256 от payload и сравнить с декодированной signature. При несовпадении → `ErrInvalidSignature`.
3. Декодировать JSON payload. При ошибке → `ErrInvalidToken`.
4. Проверить `exp > now`. При истечении → `ErrTokenExpired`.

**Примечание:** Проверка `nonce` и `delivery.qr_nonce` выполняется в сервисном слое, а не в `Verifier`, т.к. требует доступа к БД.

### 19.3 Защита от подделки

HMAC-SHA256 обеспечивает криптографическую целостность. Без `secret_key` невозможно создать валидную подпись. Секрет хранится только на сервере и никогда не включается в QR-код.

### 19.4 Защита от повторного использования

1. После верификации статус доставки атомарно переходит в `DELIVERED`.
2. Статус `DELIVERED` — терминальный. Повторная верификация невозможна: статус не `READY_FOR_PICKUP`, что проверяется в сервисном слое.
3. Дополнительная защита: `nonce` в токене совпадает с `qr_nonce` в БД. При генерации нового QR старый nonce заменяется, что делает старый токен невалидным (`ErrTokenRevoked`).

### 19.5 Отзыв/инвалидация

- Генерация нового QR: `qr_nonce` в БД обновляется на новый UUID. Старый токен с прежним nonce больше не проходит проверку `nonce == qr_nonce`.
- Отмена доставки: `qr_nonce` устанавливается в NULL. Любой токен с любым nonce не проходит проверку.
- Истечение TTL: проверка `exp > now` в `Verifier`.

### 19.6 Серверная валидация (полный порядок проверок в service.VerifyQR)

1. Извлечь claims из контекста (employee).
2. Верифицировать JWT-токен и подпись через `qr.Verifier.Verify(token)` → получить `QRToken`.
3. Найти доставку по `token.DeliveryID`. Если не найдена → `domain.ErrDeliveryNotFound`.
4. Проверить `delivery.Status == READY_FOR_PICKUP`. Иначе → `service.ErrQRNotAvailable`.
5. Проверить `token.Nonce == delivery.QRNonce`. Иначе → `qr.ErrTokenRevoked`.
6. Проверить `token.UserID` совпадает с владельцем заказа (через order-client). Иначе → `service.ErrForbidden`.
7. Атомарно обновить статус: `READY_FOR_PICKUP → DELIVERED`.
8. Вызвать order-service: `UpdateOrderStatus(delivery.OrderID, "DELIVERED")`.
9. Вернуть `VerifyQRResponse{DeliveryID, OrderID, Verified: true}`.

---

## 20. Integration Boundary с order-service

### 20.1 Существующие контракты order-service

Order-service предоставляет:
- `POST /cart/items` — добавление в корзину
- `GET /cart` — просмотр корзины
- `DELETE /cart/items/{id}` — удаление из корзины
- `POST /orders` — создание заказа
- `GET /orders/{id}` — просмотр заказа

### 20.2 Требуемые, но отсутствующие контракты

#### INTEGRATION-1: GET /orders/{id} (существует, но недостаточно)

Существующий endpoint возвращает заказ с `user_id` в ответе (см. `CreateOrderResponse` и `Order`). Это **достаточно** для:
- Проверки существования заказа при создании доставки.
- Получения `user_id` для проверки принадлежности доставки.

#### INTEGRATION-2: PATCH /orders/{id}/status (ОТСУТСТВУЕТ)

Delivery-service должен обновлять статус заказа при изменении статуса доставки:

| Переход доставки | Вызов к order-service |
|------------------|------------------------|
| `PENDING` → `IN_TRANSIT` | `PATCH /orders/{id}/status` с `status: "DELIVERING"` |
| `READY_FOR_PICKUP` → `DELIVERED` | `PATCH /orders/{id}/status` с `status: "DELIVERED"` |
| Любой → `CANCELLED` | `PATCH /orders/{id}/status` с `status: "PAID"` |

**Задача изменения integration contract:** Необходимо добавить в order-service endpoint `PATCH /orders/{id}/status`, который:
- Принимает `{ "status": "DELIVERING" | "DELIVERED" | "PAID" }`.
- Допустим только для переходов: `PAID → DELIVERING`, `DELIVERING → DELIVERED`, `DELIVERING → PAID`.
- Доступен только для внутренних вызовов между сервисами (или для employee с RBAC).

**Открытый вопрос:** Способ авторизации межсервисного вызова. Варианты:
- (A) Добавить service-to-service токен (не реализовано в проекте).
- (B) Использовать employee JWT-claims, прокидывая их от запроса сотрудника.
- (C) Добавить отдельный внутренний endpoint без авторизации (только внутри Docker-сети).

Рекомендация: вариант (B) — прокидывать claims от employee. Это согласуется с текущей моделью авторизации и не требует новой инфраструктуры.

### 20.3 OrderClient interface

```go
// internal/client/order.go

package client

import (
    "context"
    "github.com/google/uuid"
)

type Client interface {
    GetOrderByID(ctx context.Context, orderID uuid.UUID) (*OrderResponse, error)
    UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error
}

type OrderResponse struct {
    ID     uuid.UUID
    UserID uuid.UUID
    Status string
}
```

### 20.4 common/client/order/

```
common/client/order/
├── client.go          # OrderClient implementation
├── errors.go          # ErrOrderNotFound, ErrOrderNotPaid
└── generated/
    └── api.gen.go     # Сгенерированный из order-service.yaml
```

Реализация `OrderClient` по аналогии с `common/client/product/client.go`, но с двумя методами: `GetOrderByID` и `UpdateOrderStatus`.

**Примечание:** `UpdateOrderStatus` будет возвращать ошибку, пока order-service не реализует `PATCH /orders/{id}/status`. До этого момента delivery-service должен корректно обрабатывать ошибку обновления статуса заказа.

### 20.5 Обработка ошибки при обновлении статуса заказа

При ошибке вызова `UpdateOrderStatus`:
- В лог записывается ошибка (Warn-уровень).
- Доставка не откатывается — её статус обновлён локально.
- В будущем можно добавить механизм ретри или outbox-паттерн.

Это прагматичное решение для pet-project. В продакшене потребовался бы саговый паттерн или outbox.

---

## 21. Consistency Check

### 21.1 Соответствие спецификации

| Требование спецификации          | Реализовано в архитектуре            | Статус |
|-----------------------------------|--------------------------------------|--------|
| US-1: Создание доставки           | `POST /deliveries`, `DeliveryService.Create` | ✓ |
| US-2: Просмотр доставки           | `GET /deliveries/{id}`, `DeliveryService.GetByID` | ✓ |
| US-3: Получение QR-кода           | `GET /deliveries/{id}/qr`, `DeliveryService.GetQR` | ✓ |
| US-4: Верификация QR              | `POST /deliveries/verify-qr`, `DeliveryService.VerifyQR` | ✓ |
| US-5: Перенос доставки            | `PATCH /deliveries/{id}/reschedule`, `DeliveryService.Reschedule` | ✓ |
| US-6: Изменение статуса           | `PATCH /deliveries/{id}/status`, `DeliveryService.UpdateStatus` | ✓ |
| US-7: Информация об опоздании     | `is_late` в Delivery, пересчёт при запросе/переносе | ✓ |
| BR-1: Одна доставка на заказ      | `UNIQUE` на `order_id` в БД | ✓ |
| BR-2: Заказ должен быть PAID      | Проверка через OrderClient | ✓ |
| BR-3: Принадлежность клиенту      | Проверка `user_id` из JWT-claims через OrderClient | ✓ |
| BR-4: QR только в READY_FOR_PICKUP | Проверка статуса в `GetQR` | ✓ |
| BR-5: Терминальные статусы        | `DeliveryStatus.IsTerminal()` | ✓ |
| BR-6: История переносов immutable | Отдельная таблица `delivery_reschedules` | ✓ |
| BR-7: is_late вычисляется         | Вычисление в сервисном слое | ✓ |
| BR-8: Перенос обновляет дату      | Обновление `estimated_delivery_date` + создание записи | ✓ |
| BR-9: Перенос невозможен для терминальных | Проверка `IsTerminal()` | ✓ |
| State Machine                     | 5 статусов, таблица переходов | ✓ |
| QR HMAC-SHA256                    | `qr.Generator` + `qr.Verifier` | ✓ |
| QR nonce-инвалидация              | `qr_nonce` в БД + сравнение при верификации | ✓ |
| QR одноразовость                  | Терминальный статус DELIVERED | ✓ |

### 21.2 Соответствие существующему test-marketplace

| Конвенция проекта                | Применена в delivery-service       | Статус |
|-----------------------------------|--------------------------------------|--------|
| Единый Go-модуль                  | `github.com/byorty/test-marketplace/services` | ✓ |
| Структура `cmd/<service>/main.go` | `cmd/delivery-service/main.go` | ✓ |
| Структура `internal/app/`        | Все Fx-модули | ✓ |
| Структура `internal/domain/`     | Entities, interfaces, errors | ✓ |
| Структура `internal/repository/` | DeliveryRepository | ✓ |
| Структура `internal/service/`    | DeliveryService + errors.go | ✓ |
| Структура `internal/transport/`  | Handler, Router, errors, mapper, операции | ✓ |
| Структура `internal/transport/middlwr/` | auth.go, rbac.go | ✓ |
| Структура `internal/config/`     | Config + Load() | ✓ |
| Структура `internal/mocks/`      | MockDeliveryRepository, MockDeliveryService, MockOrderClient | ✓ |
| `fx.Annotate(..., fx.As(...))`  | Для repository и service | ✓ |
| `domain.Authorizer` interface    | Как в order-service | ✓ |
| `middlwr` (сокращение middleware) | `middlwr/auth.go`, `middlwr/rbac.go` | ✓ |
| chi router + oapi-codegen strict | `chi.NewRouter()` + `api.NewStrictHandler` | ✓ |
| `common/auth`                    | Переиспользуется | ✓ |
| `common/rbac`                    | Переиспользуется + расширена | ✓ |
| `common/database`                | Переиспользуется | ✓ |
| `common/test-tools`              | Переиспользуется в интеграционных тестах | ✓ |
| yaml.Unmarshal + os.ExpandEnv    | В config.Load() | ✓ |
| zap.NewProduction + zap.Named    | Во всех слоях | ✓ |
| errors.Is + sentinel errors      | domain/errors.go, service/errors.go, qr errors | ✓ |
| mapXxxError pattern              | transport/errors.go | ✓ |
| go-playground/validator          | В сервисном слое | ✓ |
| Table-driven tests               | mocks, t.Parallel() | ✓ |
| testcontainers                   | Для интеграционных тестов | ✓ |
| Docker multi-stage build         | В Dockerfile | ✓ |
| Конфигурация YAML + env          | config.yaml + .env | ✓ |

### 21.3 Нарушения conventions

Нарушений не обнаружено. Архитектура полностью следует существующим конвенциям проекта.

### 21.4 Необоснованные усложнения

| Потенциальное усложнение          | Решение                            |
|-------------------------------------|--------------------------------------|
| Отдельный пакет `qr/`              | Обосновано: HMAC-генерация и верификация — отдельная ответственность, не вписывается в domain/service/repository слои |
| Механизм нотификаций (Notifier)   | Не реализуется в текущей итерации. Флаг `is_late` вычисляется синхронно. Расширение — через интерфейс в domain, без реализации |
| Саги/outbox для обновления order   | Не реализуется. Ошибка обновления статуса логируется как Warning. Доставка не откатывается |
| Очереди сообщений (Kafka/RabbitMQ)| Не требуется архитектурой проекта. Синхронные HTTP-вызовы между сервисами |

### 21.5 Итог

Архитектура delivery-service полностью соответствует:
1. **Спецификации** — все user stories, business rules, QR-требования покрываются.
2. **Существующему проекту** — конвенции структуры, именования, DI, логирования, тестирования, моков соблюдены.
3. **Не содержит необоснованных усложнений** — QR-механизм минимален (HMAC), нотификации отложены, синхронная интеграция.
4. **Чётко фиксирует integration contract** — `PATCH /orders/{id}/status` в order-service — задача, которую нужно реализовать до или параллельно с delivery-service.