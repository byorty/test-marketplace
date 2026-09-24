# AI Development Protocol — test-marketplace

## Project Context

Go microservices marketplace project.

**Module:** `github.com/byorty/test-marketplace/services`

**Stack:** Go 1.25, Bun ORM, Chi v5, Uber Fx, Zap, Casbin v2, JWT RS256, go-playground/validator/v10, oapi-codegen, testcontainers, golang-migrate.

**Specs:** `docs/delivery-service-spec.md`, `docs/delivery-service-architecture.md`

**Working directory for all Go commands:** `services/` (where `go.mod` lives)

## Build & Run Commands

```bash
# Format
cd services && go fmt ./...

# Static analysis
cd services && go vet ./...

# Build (per service)
cd services && go build -o /tmp/product-service ./product-service/cmd/product-service/
cd services && go build -o /tmp/order-service ./order-service/cmd/order-service/

# Unit tests (per package)
cd services && go test ./product-service/internal/service/...
cd services && go test ./product-service/internal/transport/...
cd services && go test ./product-service/internal/repository/...
cd services && go test ./order-service/internal/service/...
cd services && go test ./order-service/internal/transport/...
cd services && go test ./order-service/internal/repository/...

# All unit tests
cd services && go test ./product-service/internal/... ./order-service/internal/...

# Integration tests (require Docker)
cd services && go test ./product-service/internal/repository/... -tags=integration
cd services && go test ./order-service/internal/repository/... -tags=integration

# OpenAPI code generation
make generate-p   # product-service
make generate-o   # order-service
make generate-cl  # product-service client

# Database migrations
make migrate-up
make migrate-down
```

## Step-by-Step Workflow

For every task:

1. **Read** relevant existing code before writing anything.
2. **Check** `docs/delivery-service-spec.md` and `docs/delivery-service-architecture.md`.
3. **Plan** a brief implementation outline (which files, which interfaces).
4. **Implement** one minimal complete step.
5. **Write/update** unit tests for that step.
6. **Run** `go fmt ./...` from `services/`.
7. **Run** `go vet ./...` from `services/`.
8. **Run** `go test ./delivery-service/internal/...` from `services/`.
9. **Run** `go build -o /tmp/delivery-service ./delivery-service/cmd/delivery-service/` from `services/`.
10. **Verify** all pass. Only then proceed to the next step.

## Prohibitions

- Do NOT rewrite existing services without explicit need.
- Do NOT change public APIs without updating the OpenAPI spec first.
- Do NOT add libraries without justification in the commit message.
- Do NOT create duplicate infrastructure (config, logging, auth, RBAC, DB connection) — reuse `common/`.
- Do NOT disable tests to make things pass.
- Do NOT hide errors with `_ = ...` or empty catch blocks.
- Do NOT replace integration tests with mock-only tests.
- Do NOT delete existing tests.
- Do NOT alter architecture for generation convenience.
- Do NOT invent APIs for neighboring services that do not exist. If order-service lacks an endpoint you need, record it as an integration contract task in `docs/delivery-service-architecture.md` section 20.2 and implement delivery-service without calling that endpoint until it exists.

## Handling Conflicts

If existing code contradicts the spec or architecture:

1. Stop.
2. Show the exact conflict with file paths and line numbers.
3. Propose options rooted in the existing project conventions.
4. Do NOT pick an option unless it clearly follows from the architecture doc.
5. Otherwise, ask for a decision.

## Testing Requirements

- Every business rule must have a unit test.
- Every HTTP endpoint must have handler-level tests.
- Every repository method must have integration tests using `testtools.NewTestDB(t)`.
- Authorization must be tested separately (RBAC allow/deny, ownership checks).
- Tests must cover positive and negative cases.
- Use table-driven tests with `t.Parallel()`.
- Use `testify/require` for assertions.
- Use `zap.NewNop()` for test loggers.
- Use hand-written mocks in `internal/mocks/` (no mockgen).
- Follow the naming convention: `TestRepository_Method`, `TestService_Method`, `TestHandler_Method`.

## Test File Conventions

```
internal/service/service_test.go      # package service
internal/transport/handler_test.go    # package transport
internal/repository/repository_test.go # package repository
internal/mocks/repository.go          # MockDeliveryRepository
internal/mocks/service.go             # MockDeliveryService
internal/mocks/order_client.go        # MockOrderClient
```

## Definition of Done

A feature is done only when ALL of the following are true:

- [ ] Implementation matches spec and architecture docs
- [ ] `go fmt ./...` passes
- [ ] `go vet ./...` passes
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Build succeeds (`go build ./delivery-service/cmd/delivery-service/`)
- [ ] OpenAPI spec (`api/delivery-service.yaml`) is up to date
- [ ] RBAC policy (`common/rbac/policy.csv`) is updated
- [ ] Error cases are covered in tests
- [ ] Docker integration verified (`docker-compose.yml` updated)
- [ ] Code follows existing conventions (see below)

## Conventions Checklist

Before submitting any file, verify:

- [ ] Package naming: `domain`, `service`, `repository`, `transport`, `config`, `app`, `mocks`
- [ ] Interface in `domain/`, implementation in respective package
- [ ] Constructor: `New(deps...) *Type`
- [ ] Fx wiring: `fx.Annotate(xxx.New, fx.As(new(domain.Interface)))`
- [ ] Logger: `zap.NewProduction()`, named via `zap.Named("xxx-service")`
- [ ] Errors: sentinel `var ErrXxx = errors.New(...)` in `domain/errors.go` and `service/errors.go`
- [ ] Error mapping: `mapXxxError(log, err)` in `transport/errors.go`
- [ ] Config: `yaml.Unmarshal` + `os.ExpandEnv`, no cleanenv
- [ ] Middleware directory: `middlwr/` (abbreviated, as in existing services)
- [ ] Handler: one file per endpoint (`create.go`, `get.go`, etc.)
- [ ] Handler struct holds `domain.Service` + `*zap.Logger` + `domain.Authorizer`
- [ ] Domain Authorizer interface in `domain/auth.go` (as in order-service)
- [ ] oapi-codegen: `strict-server: true`, `chi-server: true`, `models: true`
- [ ] Migrations: `NNNNNN_create_xxx.up.sql` / `.down.sql` in root `migrations/`
- [ ] Mocks: hand-written with `XxxFunc` fields and `XxxCalls` counters

## Implementation Order for delivery-service

Follow this sequence. Do not skip steps. Do not proceed until current step passes Definition of Done.

### Phase 1: Foundation
1. `common/rbac/authorizer.go` — add `ResourceDelivery`, `ActionReschedule`, `ActionUpdateStatus`, `ActionVerifyQR`
2. `common/rbac/policy.csv` — add delivery policy rows
3. `migrations/000003_create_deliveries.up.sql` and `.down.sql`
4. `api/delivery-service.yaml` — OpenAPI 3.1 spec
5. `api/oapi-codegen.yaml` — codegen config
6. Generate `internal/generated/openapi/api.gen.go`

### Phase 2: Domain Layer
7. `internal/domain/delivery.go` — Delivery model, DeliveryStatus, DeliveryList, DeliveryListFilter
8. `internal/domain/reschedule.go` — DeliveryReschedule
9. `internal/domain/errors.go` — sentinel errors
10. `internal/domain/repository.go` — DeliveryRepository interface
11. `internal/domain/service.go` — DeliveryService interface
12. `internal/domain/auth.go` — Authorizer interface

### Phase 3: QR Package
13. `internal/qr/token.go` — QRToken struct
14. `internal/qr/generator.go` — HMAC-SHA256 generation
15. `internal/qr/verifier.go` — HMAC-SHA256 verification
16. Unit tests for QR package

### Phase 4: Repository
17. `internal/repository/repository.go` — DeliveryRepository implementation
18. Integration tests for repository

### Phase 5: Client
19. `common/client/order/errors.go`
20. `common/client/order/client.go` — OrderClient (GetOrderByID only, UpdateOrderStatus stub that returns ErrNotImplemented until order-service adds the endpoint)
21. `internal/client/order.go` — Client interface

### Phase 6: Service
22. `internal/service/errors.go`
23. `internal/service/service.go` — DeliveryService implementation
24. `internal/mocks/repository.go` — MockDeliveryRepository
25. `internal/mocks/order_client.go` — MockOrderClient
26. Unit tests for service

### Phase 7: Transport
27. `internal/transport/handler.go` — DeliveryHandler struct
28. `internal/transport/router.go` — NewRouter
29. `internal/transport/errors.go` — error mapping functions
30. `internal/transport/mapper.go` — toResponse, toDeliveryList, etc.
31. `internal/transport/create.go`
32. `internal/transport/get.go`
33. `internal/transport/list.go`
34. `internal/transport/reschedule.go`
35. `internal/transport/update_status.go`
36. `internal/transport/get_qr.go`
37. `internal/transport/verify_qr.go`
38. `internal/transport/middlwr/auth.go`
39. `internal/transport/middlwr/rbac.go`
40. `internal/mocks/service.go` — MockDeliveryService
41. Handler unit tests

### Phase 8: Configuration & Wiring
42. `internal/config/config.go`
43. `config.yaml`
44. `internal/app/module.go`
45. `internal/app/config.go`
46. `internal/app/logger.go`
47. `internal/app/database.go`
48. `internal/app/auth.go`
49. `internal/app/rbac.go`
50. `internal/app/repository.go`
51. `internal/app/service.go`
52. `internal/app/handler.go`
53. `internal/app/client.go`
54. `internal/app/validate.go`
55. `internal/app/qr.go`
56. `internal/app/server.go`
57. `cmd/delivery-service/main.go`

### Phase 9: Integration
58. `Dockerfile`
59. Update `docker-compose.yml` — add delivery-service
60. Update `.env` and `.env.example` — add ORDER_SERVICE_URL, QR_SECRET_KEY, QR_TOKEN_TTL
61. Update `Makefile` — add `generate-d` target
62. End-to-end verification: build, run, test against PostgreSQL

## Key File Paths

```
# Specification & Architecture
docs/delivery-service-spec.md
docs/delivery-service-architecture.md

# Common packages (modify)
services/common/rbac/authorizer.go        # Add Delivery resource and actions
services/common/rbac/policy.csv           # Add delivery policy rows
services/common/client/order/client.go     # New OrderClient
services/common/client/order/errors.go     # New order errors
services/common/client/order/generated/    # Generated from order-service OpenAPI

# Migrations (add)
migrations/000003_create_deliveries.up.sql
migrations/000003_create_deliveries.down.sql

# Delivery service (create)
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
│   │   ├── validate.go
│   │   └── qr.go
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
│   │   ├── token.go
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

# Project-level (modify)
docker-compose.yml
.env
.env.example
Makefile
```