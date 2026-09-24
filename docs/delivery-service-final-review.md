# Delivery Service — Final Review

> **Note:** This review was conducted before the production-hardening pass. Many of the issues identified here have been resolved. See `docs/delivery-service-hardening-report.md` for the current status.

## Critical

### C1. GetByOrderID error mishandling allows duplicate deliveries on transient DB errors

- **File:** `internal/service/service.go:97-103`
- **Problem:** `Create` calls `GetByOrderID` and if ANY error is returned (including transient DB errors like connection timeouts), it falls through and attempts to create a new delivery. Only `domain.ErrDeliveryNotFound` should be treated as "doesn't exist."
- **Why it matters:** A transient DB error during the check would result in a duplicate delivery being created, violating the uniqueness constraint at the application level. The DB `UNIQUE` constraint would catch it, but the error returned to the client would be a raw 500, not a proper 409 Conflict.
- **Fix:** Check `if err != nil && !errors.Is(err, domain.ErrDeliveryNotFound)` and return error. Only proceed with create when `err` is `ErrDeliveryNotFound`.

### C2. Race condition in VerifyQR: check-then-act without row-level locking

- **File:** `internal/service/service.go:378-416`
- **Problem:** `VerifyQR` reads delivery, checks nonce and status, then updates status. Between read and update, another request could change the delivery (e.g., generate a new QR with a different nonce, cancel the delivery). The update uses full-row replacement with no optimistic concurrency control.
- **Why it matters:** Concurrent VerifyQR calls or a concurrent GetQR call could lead to inconsistent state. Example: two VerifyQR calls with the same token could both pass the nonce check and both set status to DELIVERED.
- **Fix:** Add `updated_at` optimistic concurrency check: `WHERE id = ? AND updated_at = ?`. If 0 rows affected, return an error indicating concurrent modification.

### C3. GetByID has write side effect (recalculateLate + Update) in a read path

- **File:** `internal/service/service.go:128-167`
- **Problem:** `GetByID` recalculates `is_late` and calls `repo.Update`. This creates a read-modify-write cycle without transactions. Between the read and update, another operation could change the delivery (e.g., status change to DELIVERED), and the Update would overwrite those changes with stale data.
- **Why it matters:** GET requests should be idempotent. A read operation that mutates state is a side channel that can cause data loss under concurrency.
- **Fix:** Use optimistic concurrency control on Update. If the update fails (stale data), either retry or return the stale data without the update. Long-term: move the `is_late` recalculation to a background job or a computed column.

## High

### H1. Silent error swallowing in GetByID

- **File:** `internal/service/service.go:147-155` and `:159-164`
- **Problem:** `ListReschedules` errors are logged but not returned. `Update` errors after `recalculateLate` are logged but not returned. The caller receives data that may be stale or incomplete.
- **Why it matters:** If the DB is under pressure, `ListReschedules` or `Update` could fail silently, and the client would receive incorrect data without any error indication.
- **Fix:** Return errors from both `ListReschedules` and `Update`. For the `is_late` update, if it fails, return the delivery as-is (the stale `is_late` flag is better than hiding the error).

### H2. Reschedule is non-atomic: CreateReschedule + Update without transaction

- **File:** `internal/service/service.go:238-251`
- **Problem:** `CreateReschedule` and `Update` are two separate DB operations without a transaction. If `Update` fails after `CreateReschedule` succeeds, there's an orphan reschedule record pointing to a delivery that still has the old date.
- **Why it matters:** Data integrity violation — the reschedule history says the date changed, but the delivery still has the old date.
- **Fix:** Wrap both operations in a database transaction. If the repository doesn't support transactions, at minimum, attempt to compensate (delete the reschedule record) if the Update fails.

### H3. Server ListenAndServe in goroutine without error propagation

- **File:** `internal/app/server.go:32-36`
- **Problem:** `go func() { ... server.ListenAndServe() }()` runs in a goroutine. If `ListenAndServe` fails immediately (e.g., port in use), the error is silently lost. The fx lifecycle `OnStart` hook returns nil, so fx considers the server started successfully.
- **Why it matters:** Application appears healthy but isn't serving traffic.
- **Fix:** Use a channel to propagate the error back to the fx lifecycle, or use `server.Shutdown` with context in `OnStop`.

### H4. RBAC middleware skips all GET requests

- **File:** `internal/transport/middlwr/rbac.go:23-26`
- **Problem:** The RBAC middleware short-circuits for all GET requests, returning `next.ServeHTTP(w, r)` without any RBAC check. Individual handlers perform their own RBAC checks, but this creates a defense-in-depth gap.
- **Why it matters:** If any GET handler forgets to check RBAC, the middleware won't catch it. The `permission()` function maps GET paths but is dead code since the middleware skips GETs.
- **Fix:** Remove the GET short-circuit. All requests should go through the RBAC middleware.

### H5. VerifyQR ignores UpdateOrderStatus failure

- **File:** `internal/service/service.go:416`
- **Problem:** `_ = s.orderClient.UpdateOrderStatus(ctx, updated.OrderID, "DELIVERED")` discards the error. If the order service is unavailable, the delivery is marked DELIVERED but the order remains in its previous state.
- **Why it matters:** Cross-service data inconsistency.
- **Fix:** At minimum, log the error with sufficient detail for manual reconciliation. Consider a retry mechanism or outbox pattern for production.

### H6. Duplicate RBAC checks in transport and service layers

- **Files:** `internal/transport/list.go:18`, `internal/service/service.go:177-184`; and similar patterns in get_qr.go/service.go
- **Problem:** RBAC is checked at both the transport handler level and the service level. The service uses hardcoded string `"customer"` for role checks instead of the RBAC system.
- **Why it matters:** DRY violation. If a new role is added, both layers need updating. The hardcoded string comparison bypasses the RBAC system.
- **Fix:** Remove RBAC checks from the service layer. The service should only enforce ownership checks (user_id matching). Role-based access should be solely at the transport/middleware layer.

## Medium

### M1. Verifier uses `time.Now()` directly, not injectable Clock

- **File:** `internal/qr/verifier.go:37-39`
- **Problem:** `nowUnix()` calls `time.Now()` directly, while `Generator` uses an injectable `Clock`. This inconsistency makes it harder to test token expiry and creates a gap where production behavior is fine but test behavior may differ.
- **Fix:** Add `Clock` interface to `Verifier` (same as Generator) with `NewVerifierWithClock`.

### M2. No maximum page size validation

- **File:** `internal/service/service.go:170-175`
- **Problem:** `List` defaults to `pageSize=20` but doesn't enforce a maximum. A client could request `pageSize=1000000`, causing a huge DB query.
- **Fix:** Cap `pageSize` to a reasonable maximum (e.g., 100).

### M3. Token-related errors return 400 instead of 401

- **File:** `internal/transport/errors.go:113-120`
- **Problem:** `ErrInvalidToken`, `ErrTokenExpired`, `ErrTokenRevoked` all map to 400. An invalid or expired token is more accurately a 401 (authentication failure), not 400 (bad request).
- **Fix:** Map these to 401 for `ErrInvalidToken` and `ErrInvalidSignature`, and keep `ErrTokenRevoked` and `ErrQRNotAvailable` as 400.

### M4. List endpoint doesn't load reschedules

- **File:** `internal/service/service.go:169-193`
- **Problem:** `List` returns deliveries without reschedules, but `GetByID` loads them. API consumers using the list endpoint never see reschedule history.
- **Fix:** Either load reschedules for each delivery in the list, or document that list responses don't include reschedules.

### M5. OrderClient.GetOrderByID doesn't pass authentication context

- **File:** `services/common/client/order/client.go`
- **Problem:** The HTTP request to the order service doesn't include an Authorization header. If the order service requires authentication, the request will fail.
- **Fix:** Add service-to-service authentication (API key, JWT, or mTLS).

### M6. Service layer hardcodes role string "customer"

- **Files:** `internal/service/service.go:178, 340`
- **Problem:** Hardcoded `"customer"` string comparisons bypass the RBAC system and are fragile.
- **Fix:** Use constants or the RBAC authorizer for role checks.

### M7. Error messages may leak internal details

- **File:** `internal/transport/errors.go`
- **Problem:** `err.Error()` strings from the service layer are passed directly to API error responses. Internal error messages could reveal architecture details.
- **Fix:** Map known errors to sanitized messages; use a generic "internal error" for unknown errors.

## Low

### L1. Role field in DeliveryCreateInput domain model

- **File:** `internal/domain/delivery.go:93-98`
- **Problem:** `Role` is an auth concern that leaked into the domain model.
- **Fix:** Move role handling to the service or transport layer; domain should not know about roles.

### L2. Logger not flushed on shutdown

- **File:** `internal/app/logger.go`
- **Problem:** `zap.NewProduction()` is created but `Sync()` is not called on shutdown, potentially losing buffered log entries.
- **Fix:** Add `logger.Sync()` to the fx lifecycle `OnStop` hook.

### L3. RBAC permission() URL matching uses HasPrefix/HasSuffix

- **File:** `internal/transport/middlwr/rbac.go:49-81`
- **Problem:** `strings.HasPrefix` and `strings.HasSuffix` could match unintended paths.
- **Fix:** Use more precise path matching or chi route patterns.

### L4. Delivery Reschedules field is `bun:"-"`

- **File:** `internal/domain/delivery.go:72`
- **Problem:** `Reschedules` is not persisted with the delivery, only loaded separately. This is by design but creates inconsistency between endpoints that load reschedules and those that don't.
- **Fix:** Document this behavior in the code or API spec.

### L5. Config doesn't validate QR_SECRET_KEY length at startup

- **File:** `internal/config/config.go`, `internal/app/qr.go`
- **Problem:** If `QR_SECRET_KEY` is empty or too short, the application will fail at fx.Provide time with a cryptic error, not at config validation time.
- **Fix:** Add config validation for `QR_SECRET_KEY` minimum length in `config.Load()`.