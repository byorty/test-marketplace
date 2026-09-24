# Анализ проекта для интеграции delivery-service

## 1. Существующая архитектура

### 1.1 Общая структура проекта

```
test-marketplace/
├── .env
├── .env.example
├── .gitignore
├── Makefile
├── docker-compose.yml
├── migrations/
│   ├── 000001_create_products.up.sql
│   ├── 000001_create_products.down.sql
│   ├── 000002_create_orders.up.sql
│   └── 000002_create_orders.down.sql
├── services/
│   ├── go.mod / go.sum          # Единый Go-модуль для всех сервисов
│   ├── keys/
│   │   ├── private.pem
│   │   └── public.pem
│   ├── common/
│   │   ├── auth/                 # JWT-валидация, claims, context
│   │   ├── client/product/       # HTTP-клиент для product-service
│   │   ├── database/             # Подключение к PostgreSQL (Bun)
│   │   ├── rbac/                 # Casbin-авторизатор
│   │   └── test-tools/           # Testcontainers-хелпер для интеграционных тестов
│   ├── product-service/
│   └── order-service/
└── opencode.json
```

**Ключевой факт:** все сервисы находятся в едином Go-модуле (`github.com/byorty/test-marketplace/services`), что позволяет импортировать `common`-пакеты напрямую без отдельных go.mod.

### 1.2 Стек технологий

| Компонент           | Решение                                      |
|---------------------|----------------------------------------------|
| Язык                | Go 1.25.5                                    |
| HTTP-маршрутизатор  | chi v5                                       |
| ORM                 | Bun (uptrace/bun) с pgdialect               |
| БД                  | PostgreSQL 17                                 |
| DI                  | Uber Fx                                      |
| Логирование          | Uber zap                                     |
| RBAC                | Casbin v2 (модель + CSV-политика)            |
| JWT                 | golang-jwt/jwt/v5, RS256                     |
| Валидация            | go-playground/validator/v10                  |
| Конфигурация         | YAML + env через gopkg.in/yaml.v3 + os.ExpandEnv |
| OpenAPI             | OpenAPI 3.1 → oapi-codegen (strict-server, chi) |
| Тестирование         | testify, testcontainers-go (postgres module)  |
| Миграции             | golang-migrate/migrate v4                   |
| Docker              | docker-compose (postgres, migrate, services)  |

---

## 2. Структура сервисов (conventions)

### 2.1 Унифицированная внутренняя структура сервиса

Оба сервиса (`product-service` и `order-service`) следуют идентичной структуре:

```
<service>/
├── api/
│   ├── <service>.yaml          # OpenAPI-спецификация
│   └── oapi-codegen.yaml       # Конфигурация oapi-codegen
├── cmd/<service>/
│   └── main.go                 # Точка входа: fx.New(app.Module).Run()
├── config.yaml                 # YAML-конфиг сервиса
├── Dockerfile                   # Multi-stage Docker-сборка
├── internal/
│   ├── app/
│   │   ├── module.go            # Fx-модуль (композиция всех субмодулей)
│   │   ├── config.go             # Fx-провайдер конфигурации
│   │   ├── logger.go            # Fx-провайдер zap.Logger
│   │   ├── database.go          # Fx-провайдер *bun.DB
│   │   ├── auth.go              # Fx-провайдер *auth.Validator
│   │   ├── rbac.go              # Fx-провайдер RBAC (enforcer + authorizer)
│   │   ├── repository.go        # Fx-провайдер репозитория (с fx.As)
│   │   ├── service.go           # Fx-провайдер сервиса (с fx.As)
│   │   ├── handler.go           # Fx-провайдер HTTP-хендлера
│   │   ├── validate.go          # Fx-провайдер *validator.Validate (product-service)
│   │   ├── server.go            # Fx-invoke RunServer
│   │   └── client.go             # Fx-провайдер HTTP-клиента (order-service)
│   ├── config/
│   │   └── config.go            # Структура Config + Load()
│   ├── domain/
│   │   ├── <entity>.go          # Модели (bun.BaseModel)
│   │   ├── repository.go        # Интерфейс репозитория
│   │   ├── service.go           # Интерфейс сервиса
│   │   ├── errors.go            # Доменные ошибки (sentinel errors)
│   │   └── auth.go              # Интерфейс Authorizer (order-service)
│   ├── generated/openapi/
│   │   └── api.gen.go           # Сгенерированный oapi-codegen код
│   ├── mocks/
│   │   ├── repository.go        # Ручной mock-объект репозитория
│   │   └── service.go           # Ручной mock-объект сервиса
│   ├── repository/
│   │   └── repository.go        # Реализация репозитория (Bun)
│   ├── service/
│   │   ├── service.go           # Реализация бизнес-логики
│   │   └── errors.go            # Ошибки сервисного слоя
│   └── transport/
│       ├── handler.go           # Конструктор хендлера (StrictServerInterface)
│       ├── router.go            # NewRouter(handler, jwt, authorizer)
│       ├── errors.go            # Маппинг ошибок → HTTP-ответы
│       ├── mapper.go            # Маппинг domain → API-модели
│       ├── <operation>.go       # Один файл на endpoint
│       └── middlwr/
│           ├── auth.go          # JWT-авторизация (middleware)
│           └── rbac.go          # RBAC-авторизация (middleware)
```

### 2.2 Конвенции именования

| Аспект                    | Конвенция                                         | Пример                                      |
|---------------------------|---------------------------------------------------|----------------------------------------------|
| Имена файлов              | snake_case, один тип на файл                     | `product.go`, `repository.go`              |
| Имена пакетов             | snake_case, короткие                              | `domain`, `service`, `transport`            |
| Имена структур            | PascalCase                                        | `ProductService`, `OrderRepository`         |
| Имена интерфейсов         | PascalCase без I-префикса                        | `ProductRepository`, `OrderService`          |
| Имена методов интерфейсов | PascalCase, глагол + существительное            | `GetByID`, `Create`, `AddToCart`            |
| Имена констант/ошибок     | CamelCase с префиксом Err / SortBy               | `ErrProductNotFound`, `SortByPrice`          |
| Имена доменных ошибок     | `Err<Entity><Condition>`                         | `ErrProductNotFound`, `ErrCartEmpty`         |
| Имена сервисных ошибок    | `Err<Condition>`                                  | `ErrInvalidInput`, `ErrInvalidID`            |
| Имена Fx-модулей          | `<Component>Module` (var)                        | `LoggerModule`, `DatabaseModule`            |
| Имена Fx-провайдеров     | `New<Component>`                                 | `NewDB`, `NewLogger`, `NewJWTValidator`      |
| Имена конструкторов       | `New` (в пакете конкретного типа)                | `repository.New`, `service.New`, `httptransport.New` |
| Имена тестовых моков      | `Mock<Type>`                                     | `MockProductRepository`, `MockOrderService`  |
| Путь middleware           | `middlwr` (сокращённое)                          | Используется в обоих сервисах                |

### 2.3 Паттерн Dependency Injection (Uber Fx)

Каждый сервис следует единому паттерну:
- `cmd/<service>/main.go` — `fx.New(app.Module).Run()`
- `app.Module` — `fx.Options(...)` из субмодулей:
  - `ConfigModule`, `LoggerModule`, `DatabaseModule`, `AuthModule`, `RBACModule`, `RepositoryModule`, `ServiceModule`, `HandlerModule`, `ServerModule`
  - Опционально: `ValidateModule` (product-service), `ClientModule` (order-service)
- Привязка интерфейсов через `fx.Annotate(…, fx.As(new(domain.XxxRepository)))`
- Сервер запускается через `fx.Invoke(RunServer)` с `fx.Lifecycle`

---

## 3. Domain/Model-слой

### 3.1 product-service

```go
type Product struct {
    bun.BaseModel `bun:"table:products"`
    ID           uuid.UUID `bun:"id,pk,type:uuid"`
    Name         string    `bun:"name,notnull"`
    Description  string    `bun:"description"`
    Price        int64     `bun:"price,notnull"`       // Цена в копейках
    Category     string    `bun:"category,notnull"`
    Rating       float32   `bun:"rating,notnull"`
    DeliveryDays int       `bun:"delivery_days,notnull"` // Дни доставки
    CreatedAt    time.Time `bun:"created_at,notnull"`
    UpdatedAt    time.Time `bun:"updated_at,notnull"`
}
```

Дополнительные типы: `ListFilter`, `ProductList`, `SortBy`, `SortOrder`.

### 3.2 order-service

```go
type Status string
const (
    StatusCreated   Status = "CREATED"
    StatusPaid      Status = "PAID"
    StatusDelivering Status = "DELIVERING"
    StatusDelivered  Status = "DELIVERED"
)

type Order struct {
    bun.BaseModel `bun:"table:orders"`
    ID           uuid.UUID    `bun:"id,pk,type:uuid"`
    UserID       uuid.UUID    `bun:"user_id,notnull,type:uuid"`
    Status       Status       `bun:"status,notnull,type:varchar(20)"`
    TotalPrice   int64        `bun:"total_price,notnull"`
    CreatedAt    time.Time    `bun:"created_at,notnull"`
    DeliveryDate time.Time    `bun:"delivery_date"`
    Items        []OrderItem   `bun:"rel:has-many,join:id=order_id"`
}

type OrderItem struct {
    bun.BaseModel `bun:"table:order_items"`
    ID           uuid.UUID `bun:"id,pk,type:uuid"`
    OrderID      uuid.UUID `bun:"order_id,notnull,type:uuid"`
    ProductID    uuid.UUID `bun:"product_id,notnull,type:uuid"`
    ProductPrice int64     `bun:"product_price,notnull"`
    Quantity     int       `bun:"quantity,notnull"`
}

type CartItem struct {
    bun.BaseModel `bun:"table:cart_items"`
    ID        uuid.UUID `bun:"id,pk,type:uuid"`
    UserID    uuid.UUID `bun:"user_id,notnull,type:uuid"`
    ProductID uuid.UUID `bun:"product_id,notnull,type:uuid"`
    Quantity  int       `bun:"quantity,notnull"`
}

type Cart struct {
    Items      []CartItem
    TotalPrice int64
}
```

**Ключевое наблюдение для delivery-service:** Order содержит `Status` с `StatusDelivering` и `StatusDelivered`, а также `DeliveryDate`. Это создаёт точку интеграции для delivery-service.

---

## 4. Repository-слой

### Паттерн

1. Интерфейс определяется в `domain/repository.go`.
2. Реализация — в `repository/repository.go`, структура с `*bun.DB` и `*zap.Logger`.
3. Конструктор `New(*bun.DB, *zap.Logger) *XxxRepository`.
4. Репозиторий использует Bun query builder.
5. Ошибки: `sql.ErrNoRows` → доменные sentinel-ошибки, прочие — оборачиваются через `fmt.Errorf`.
6. Логирование: информационные логи при успехе, ошибки при неудаче, с `zap.Named("xxx-repository")`.
7. В order-service репозиторий принимает `bun.IDB` (интерфейс) вместо `*bun.DB` для поддержки транзакций.
8. Транзакции: `Transaction(ctx, func(repo OrderRepository) error)` — создаёт `txRepo` с `bun.Tx`.

---

## 5. Service/Use-case-слой

### Паттерн

1. Интерфейс определяется в `domain/service.go`.
2. Реализация — в `service/service.go`, структура с зависимостями (repo, logger, validator/client).
3. Конструктор `New(deps...) *XxxService`.
4. Валидация: `validator.Validate.Struct()` (product-service), ручная (order-service).
5. Сервисные ошибки — в `service/errors.go` (sentinel `errors.New`).
6. Бизнес-логика: проверка входных данных → вызов репозитория → возврат результата/ошибки.
7. Логирование: `zap.Named("xxx-service")`, `time.Since(start)` для замеров длительности.
8. Order-service делает HTTP-вызов к product-service через `client.Client`.

---

## 6. HTTP Handlers / Transport-слой

### Паттерн

1. Handler реализует `api.StrictServerInterface` (сгенерированный oapi-codegen).
2. Один файл на endpoint: `create.go`, `get.go`, `list.go`, `update.go`, `delete.go`.
3. Маппинг domain → API-модели через функции в `mapper.go` (`toResponse`, `toXxxList`).
4. Маппинг ошибок через функции в `errors.go` (`mapXxxError`) → типизированные ответы по `errors.Is`.
5. Router: `chi.NewRouter()` + middleware + `api.HandlerFromMux(strictHandler, router)`.
6. В order-service дополнительно: `chi.middleware.RequestID`, `chi.middleware.Logger`, `chi.middleware.Recoverer`.

---

## 7. Middleware

### 7.1 Auth middleware (`middlwr/auth.go`)

**product-service:** пропускает GET-запросы без авторизации; для прочих методов парсит `Authorization: Bearer <token>`, валидирует JWT, кладёт claims в контекст через `auth.ContextWithClaims`.

**order-service:** обязательная авторизация для всех запросов (нет исключения для GET).

### 7.2 RBAC middleware (`middlwr/rbac.go`)

Общий паттерн: извлечь claims из контекста → определить (resource, action) по HTTP-методу и пути → вызвать `authorizer.Authorize(role, resource, action)`. При отказе → 403.

Ресурсы и действия определены в `common/rbac/authorizer.go`:
- Resources: `product`, `cart`, `order`
- Actions: `create`, `view`, `update`, `delete`, `add`, `remove`

---

## 8. Authentication (JWT)

### Конфигурация

```yaml
jwt:
  issuer: ${JWT_ISSUER}
  public_key_path: ${JWT_PUBLIC_KEY_PATH}
```

### Реализация (`common/auth/`)

- `Claims`: `UserID uuid.UUID`, `Role string`, `jwt.RegisteredClaims`.
- `Validator.Parse(tokenString)`: парсит и валидирует JWT с RS256, проверяет issuer.
- `LoadPublicKey(path)`: загружает RSA public key из PEM-файла.
- `ContextWithClaims(ctx, claims)`, `ClaimsFromContext(ctx)` — работа с контекстом.

Ключи хранятся в `services/keys/` (private.pem, public.pem).

---

## 9. Authorization (RBAC / Casbin)

### Модель (`common/rbac/model.conf`)

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

### Политика (`common/rbac/policy.csv`)

```csv
p, customer, product, view

p, employee, product, create
p, employee, product, view
p, employee, product, update
p, employee, product, delete

p, customer, cart, add
p, customer, cart, remove
p, customer, cart, view

p, customer, order, create
p, customer, order, view

p, employee, order, view
```

**Роли:** `customer`, `employee`.

### Паттерн использования в order-service

В `domain/auth.go` определён интерфейс:
```go
type Authorizer interface {
    Authorize(role string, resource rbac.Resource, action rbac.Action) error
}
```

Этот интерфейс используется в handler для программной проверки доступа (в дополнение к middleware).

---

## 10. Конфигурация

### Паттерн

- Структура `config.Config` в `internal/config/config.go`.
- YAML-файл + подстановка env-переменных через `os.ExpandEnv`.
- Общие поля: `HTTP`, `Postgres`, `Log`, `JWT`.
- Специфичные поля: `ProductService` (в order-service — URL продукта).

### Общий формат Config

```go
type Config struct {
    HTTP           HTTPConfig     `yaml:"http"`
    Postgres       PostgresConfig `yaml:"postgres"`
    Log            LogConfig      `yaml:"log"`
    JWT            JWT            `yaml:"jwt"`
    // + специфичные поля сервиса
}

type HTTPConfig struct {
    Host string `yaml:"host" env:"HTTP_HOST" env-default:"0.0.0.0"`
    Port int    `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
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
```

**Примечание:** теги `env-required` и `env-default` определены, но парсинг осуществляется через `yaml.Unmarshal` + `os.ExpandEnv`, а не через библиотеку `cleanenv`. Библиотека `cleanenv` не используется, хотя упомянута в описании проекта.

---

## 11. Logging

- Логгер: `zap.NewProduction()`, именованные: `zap.Named("xxx-service")`, `zap.Named("xxx-repository")`, `zap.Named("xxx-handler")`.
- Тестовый логгер: `zap.NewNop()`.

---

## 12. Error Handling

### Доменный слой (`domain/errors.go`)

Sentinel errors через `errors.New`: `ErrProductNotFound`, `ErrOrderNotFound`, `ErrCartEmpty`, и т.д.

### Сервисный слой (`service/errors.go`)

Sentinel errors: `ErrInvalidInput`, `ErrInvalidID`, `ErrNilInput`, `ErrEmptyUpdate`, `ErrForbidden`, и т.д.

### Transport-слой (`transport/errors.go`)

Функции `mapXxxError(log, err)` с `switch { case errors.Is(...) }` → типизированные HTTP-ответы (400/401/403/404/500). Общий паттерн:
- Доменные ошибки → 400/404
- Неизвестные ошибки → 500 с `http.StatusText(http.StatusInternalServerError)`

---

## 13. Validation

- **product-service:** `go-playground/validator/v10` — `validate.Struct(input)`.
- **order-service:** ручная валидация (`if id == uuid.Nil`, `if item == nil`, и т.д.).

---

## 14. Database Migrations

- Используется `golang-migrate/migrate/v4`.
- SQL-файлы в `migrations/` с нумерацией `NNNNNN_create_<table>.{up,down}.sql`.
- Запуск через Makefile (`migrate-up`, `migrate-down`) или через Docker Compose (контейнер `migrate`).
- Миграции общие для всех сервисов (единая БД `marketplace`).

---

## 15. Docker Compose

```yaml
services:
  postgres:     # PostgreSQL 17, порт 5432
  migrate:      # golang-migrate, depends_on postgres
  product-service:  # depends_on migrate, порт 8008:8080
  order-service:    # depends_on migrate, порт 8081:8081
```

Один `.env`-файл на все сервисы. Каждый сервис имеет собственный Dockerfile с multi-stage build.

---

## 16. OpenAPI / Swagger / Code Generation

- OpenAPI 3.1 YAML-спецификации: `api/<service>.yaml`.
- Конфигурация oapi-codegen: `api/oapi-codegen.yaml` — генерация `strict-server: true`, `chi-server: true`, `models: true`.
- Сгенерированный код: `internal/generated/openapi/api.gen.go`.
- Клиент для product-service: `common/client/product/generated/api.gen.go` (генерация `client: true`).
- Makefile-цели: `generate-p`, `generate-o`, `generate-cl`.

---

## 17. Тестирование

### Unit-тесты

- **Расположение:** тот же пакет (`package service`, `package transport`, `package repository`).
- **Моки:** ручные (`mocks/repository.go`, `mocks/service.go`) — не используют mockgen.
- **Фреймворк:** `testing` + `testify/require`.
- **Паттерн:** table-driven tests с `t.Parallel()`.

### Интеграционные тесты

- **Расположение:** `repository/*_test.go`.
- **Инфраструктура:** `testcontainers-go/modules/postgres` — поднимает PostgreSQL 17-alpine в Docker.
- **Хелпер:** `common/test-tools/helper.go` → `NewTestDB(t)` — создаёт контейнер, применяет миграции, возвращает `*bun.DB`.
- **Миграции в тестах:** `golang-migrate` с путём `../../../../migrations` (относительный путь от тестового файла).

---

## 18. Существующие зависимости между сервисами

### order-service → product-service

Order-service вызывает product-service через HTTP-клиент для:
1. **Добавление в корзину** (`AddToCart`): проверяет существование товара через `client.GetProduct`.
2. **Создание заказа** (`CreateOrder`): получает цену и `DeliveryDays` товара для расчёта итоговой суммы и даты доставки.

Клиентская абстракция:
```go
// internal/client/product.go
type Client interface {
    GetProduct(ctx context.Context, id uuid.UUID) (*client.ProductResponse, error)
}
```

Реализация: `common/client/product/client.go` → `ProductClient` с захардкоженным URL `http://product-service:8080`.

Fx-провайдер в `app/client.go`:
```go
var ClientModule = fx.Provide(
    fx.Annotate(
        product.NewProductClient,
        fx.As(new(client.Client)),
    ),
)
```

---

## 19. API-контракты

### product-service (порт 8080)

| Метод  | Путь          | Описание         | Auth  |
|--------|---------------|------------------|-------|
| GET    | /products     | Список товаров   | Нет   |
| POST   | /products     | Создание товара  | Bearer|
| GET    | /products/{id}| Получение товара | Нет   |
| PATCH  | /products/{id}| Обновление товара| Bearer|
| DELETE | /products/{id}| Удаление товара  | Bearer|

### order-service (порт 8081)

| Метод  | Путь            | Описание              | Auth  |
|--------|-----------------|-----------------------|-------|
| POST   | /cart/items     | Добавить в корзину   | Bearer|
| GET    | /cart           | Просмотр корзины     | Bearer|
| DELETE | /cart/items/{id}| Удалить из корзины   | Bearer|
| POST   | /orders         | Оформить заказ       | Bearer|
| GET    | /orders/{id}   | Просмотр заказа      | Bearer|

---

## 20. Как delivery-service должен вписаться в проект

### 20.1 Структура каталога

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
│   │   └── client.go           # Клиент для order-service
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── delivery.go          # Модель Delivery
│   │   ├── repository.go        # Интерфейс DeliveryRepository
│   │   ├── service.go           # Интерфейс DeliveryService
│   │   ├── errors.go            # Доменные ошибки
│   │   └── auth.go              # Интерфейс Authorizer
│   ├── generated/openapi/
│   │   └── api.gen.go
│   ├── mocks/
│   │   ├── repository.go
│   │   └── service.go
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
│       ├── <operation>.go
│       └── middlwr/
│           ├── auth.go
│           └── rbac.go
```

### 20.2 Что нужно переиспользовать из common

| Пакет              | Что переиспользовать              | Как                            |
|--------------------|------------------------------------|--------------------------------|
| `common/auth`     | `Validator`, `Claims`, `ContextWithClaims`, `ClaimsFromContext`, `LoadPublicKey` | Импорт напрямую                |
| `common/rbac`     | `Authorizer`, `NewEnforcer`, `New`, `Resource`, `Action`, `ErrAccessDenied` | Импорт напрямую                |
| `common/database` | `PostgresConfig`, `New()`          | Импорт напрямую                |
| `common/test-tools`| `NewTestDB()`                     | В интеграционных тестах        |

**Нельзя дублировать:**
- JWT-валидацию (RS256, Claims-структура)
- RBAC-модель и политику
- Подключение к БД
- HTTP-клиент для product-service (если нужен)

### 20.3 Новые ресурсы и действия для RBAC

В `common/rbac/authorizer.go` нужно добавить:
```go
const (
    ResourceDelivery Resource = "delivery"
)
```

В `common/rbac/policy.csv` нужно добавить строки для delivery-ресурса:
```csv
p, employee, delivery, create
p, employee, delivery, view
p, employee, delivery, update
p, customer, delivery, view
```

**Открытый вопрос:** какие именно роли и действия нужны для delivery-service — требует архитектурного решения.

### 20.4 Новые миграции

В `migrations/` добавить:
- `000003_create_deliveries.up.sql`
- `000003_create_deliveries.down.sql`

### 20.5 Docker Compose

Добавить сервис `delivery-service` с собственным портом и зависимостью от `migrate`.

### 20.6 Makefile

Добавить цель `generate-d` для oapi-codegen delivery-service.

### 20.7 Конфигурация

Добавить в `.env` и `.env.example` переменные для delivery-service (порт, URL order-service).

---

## 21. Открытые архитектурные вопросы

### 21.1 Интеграция с order-service

**Вопрос:** Как delivery-service взаимодействует с order-service?

Варианты:
- **A) delivery-service вызывает order-service через HTTP-клиент** (как order-service вызывает product-service) — для получения данных заказа, обновления статуса доставки.
- **B) order-service вызывает delivery-service** — при создании заказа создаётся доставка.
- **C) Оба сервиса подписываются на события** (необходим брокер сообщений, которого сейчас нет в проекте).
- **D) Общую БД** — delivery-service пишет в таблицу deliveries, order-service читает (нарушает микросервисную изоляцию).

**Текущая конвенция:** используется синхронный HTTP-вызов между сервисами (вариант A). Это наиболее вероятный путь, но требуется решение.

### 21.2 Обновление статуса заказа

**Вопрос:** Кто обновляет `Order.Status` с `CREATED` → `PAID` → `DELIVERING` → `DELIVERED`?

Текущий код order-service создаёт заказ со статусом `CREATED` и `DeliveryDate`. Нет API для обновления статуса. Delivery-service должен как-то менять статус заказа.

Варианты:
- **A) delivery-service вызывает PATCH /orders/{id}/status на order-service** — но такого endpoint нет. Нужно ли его добавить?
- **B) order-service предоставляет отдельный внутренний endpoint** для обновления статуса.
- **C) delivery-service пишет напрямую в БД order-service** — нарушает микросервисную изоляцию.

**Решение:** требуется архитектурное решение.

### 21.3 Обновление политики Casbin

**Вопрос:** Как delivery-service обновляет `policy.csv` и ресурсы?

Текущие ресурсы и политика находятся в `common/rbac/`. Delivery-service должен добавить `ResourceDelivery` и соответствующие политики. Это затрагивает общий код, который используют все сервисы.

**Решение:** либо добавить в `common/rbac/` до реализации, либо delivery-service поддерживает собственную модель/политику (нарушает консистентность).

### 21.4 URL-адрес order-service для HTTP-клиента

**Вопрос:** Как delivery-service узнаёт адрес order-service?

Текущий паттерн: `config.yaml` содержит `product_service_url`, а `ProductClient` в `common/client/product/client.go` захардкожен на `http://product-service:8080`.

**Решение:** добавить `ORDER_SERVICE_URL` в конфигурацию delivery-service, создать аналогичный клиент в `common/client/order/` или внутри `internal/client/`.

### 21.5 Данные о доставке в модели заказа

**Вопрос:** Должен ли order-service хранить `delivery_id` (ссылку на доставку) или delivery-service хранит `order_id` (ссылку на заказ)?

Текущая модель `Order` не содержит `delivery_id`. Если delivery-service ссылается на `order_id`, то это односторонняя связь. Если нужна двусторонняя, потребуется изменить схему `orders`.

**Решение:** требуется архитектурное решение.

### 21.6 Статус доставки vs статус заказа

**Вопрос:** Как соотносятся статусы доставки и заказа?

Текущие статусы заказа: `CREATED`, `PAID`, `DELIVERING`, `DELIVERED`. У доставки будут свои статусы. Требуется определить модель статусов доставки и их связь с `Order.Status`.

**Решение:** требуется архитектурное решение.

### 21.7 Кто вычисляет DeliveryDate?

**Вопрос:** `Order.DeliveryDate` сейчас вычисляется в order-service на основе `Product.DeliveryDays`. Если delivery-service управляет доставкой, должен ли он переопределять/обновлять эту дату?

**Решение:** требуется архитектурное решение.

### 21.8 Единая БД vs отдельные БД

**Вопрос:** Сейчас все сервисы используют одну PostgreSQL-БД `marketplace`. Delivery-service тоже будет использовать эту же БД или свою?

Текущий паттерн — единая БД. Это упрощает разработку, но может стать проблемой при масштабировании.

**Решение:** следовать текущему паттерну (единая БД), если нет требования об изоляции.

### 21.9 Использование cleanenv

**Вопрос:** В описании проекта указана библиотека `cleanenv` для конфигурации, но фактический код использует `yaml.Unmarshal` + `os.ExpandEnv`. Какой подход следует использовать для delivery-service?

**Рекомендация:** следовать фактическому паттерну (`yaml.Unmarshal` + `os.ExpandEnv`), чтобы сохранить консистентность.

### 21.10 Валидация: validator vs ручная

**Вопрос:** Product-service использует `go-playground/validator`, order-service — ручную валидацию. Какой подход выбрать для delivery-service?

**Рекомендация:** использовать `go-playground/validator`, если OpenAPI-схема содержит `x-oapi-codegen-extra-tags` с правилами валидации (как в product-service). Иначе — ручная валидация.

### 21.11 Auth middleware: пропускать GET или нет?

**Вопрос:** Product-service пропускает авторизацию для GET-запросов, order-service — нет. Для delivery-service требуется определить политику.

**Решение:** зависит от бизнес-требований. Если доставка может содержать чувствительные данные — требовать авторизацию для всех запросов.

---

## 22. Резюме найденного

**Проект test-marketplace** — это монорепозиторий с двумя микросервисами (product-service, order-service), общей библиотекой (common) и единым Go-модулем.

**Архитектурный стиль:**
- Clean Architecture (domain → repository → service → transport)
- Строгая типизация через oapi-codegen (strict-server)
- Dependency Injection через Uber Fx
- Sentinel errors + errors.Is для маппинга ошибок
- Ручные моки (без mockgen/mockery)
- Table-driven тесты с testify
- Testcontainers для интеграционных тестов
- Единая БД PostgreSQL с общими миграциями

**Для delivery-service нужно переиспользовать:**
- `common/auth` — JWT-валидация
- `common/rbac` — авторизация (с добавлением новых ресурсов/политик)
- `common/database` — подключение к PostgreSQL
- `common/test-tools` — тестовый хелпер
- Паттерны Fx-модулей, конфигурации, моков, тестирования
- OpenAPI → oapi-codegen → strict-server конвейер

**Критические архитектурные решения перед реализацией:**
1. Способ интеграции delivery-service ↔ order-service (HTTP vs общая БД vs события)
2. Обновление статуса заказа (кто и как меняет Order.Status)
3. Расширение Casbin-политики (новые ресурсы/действия)
4. Модель данных доставки и её связь с заказом
5. HTTP-клиент для order-service (в common/client/order или внутренний)