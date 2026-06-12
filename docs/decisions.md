# Architecture Decisions

## DEC-001: Validator Interface for Request Struct Validation

**Date:** 2026-06-01
**Status:** Accepted

### Context

The NaCl backend receives JSON payloads from the frontend (React SPA) via REST API endpoints. Each endpoint needs to validate incoming data before processing. Currently, validation is done inline in handlers (e.g., `handlers_users.go` checks `strings.TrimSpace(userRequest.Username) == ""`), which duplicates logic and makes handlers harder to read as more endpoints are added.

Upcoming endpoints (login, cipher CRUD, password change, operations) will each require their own request structs and validation rules. Without a consistent pattern, validation logic would be scattered across handlers with increasing inconsistency.

### Decision

Use a `Validator` interface defined on each request DTO struct to encapsulate validation logic:

```go
type Validator interface {
    Validate() error
}
```

Every request struct (e.g., `CreateUserRequest`, `LoginRequest`, `CreateCipherRequest`) implements `Validate() error`, returning a descriptive error if validation fails, or `nil` if valid.

Handlers call `Validate()` on the decoded struct before processing, replacing inline checks:

```go
var req CreateUserRequest
// ... decode JSON ...
if err := req.Validate(); err != nil {
    s.RespondWithError(w, http.StatusBadRequest, err.Error(), nil)
    return
}
```

### Consequences

- **Consistency**: All request validation follows the same pattern across every endpoint.
- **Testability**: Validation logic can be unit-tested independently of the handler by calling `req.Validate()` directly.
- **Encapsulation**: Validation rules live next to the struct definition in `structs.go`, not buried in handler code.
- **Extensibility**: Adding new request structs automatically enforces the `Validator` contract.
- **No external dependencies**: This is a plain Go interface, no third-party validation library required.

### Alternatives Considered

1. **Struct tags + `go-playground/validator`**: Tag-based validation (`binding:"required"`) is concise but hides rules in tags, requires a dependency, and is harder to unit-test or add cross-field logic to.
2. **Validation middleware**: A middleware that calls `Validate()` on decoded bodies. Adds complexity for minimal gain since handlers already need to decode the body and handle decode errors before validating.
3. **Keep inline validation in handlers**: Simple for one endpoint but doesn't scale. Validation rules would be duplicated and inconsistent across handlers.

---

## DEC-002: Base64 Encoding for Cryptographic Fields

**Date:** 2026-06-01
**Status:** Accepted

### Context

The NaCl encryption model stores two cryptographic byte fields in the `users` table:

- `master_key_salt` — 32 random bytes generated during registration, used as input to Argon2id key derivation
- `encrypted_master_key` — AES-256-GCM ciphertext (plaintext + 12-byte nonce + 16-byte auth tag), encrypted with the derived key

PostgreSQL does not have a native fixed-length byte type suitable for these values. They must be encoded as strings for storage in `VARCHAR`/`TEXT` columns. The two common encodings are hex and base64.

### Decision

Encode both `master_key_salt` and `encrypted_master_key` as **base64 strings** (standard encoding, `base64.StdEncoding`) before storing in the database, and decode from base64 when reading back. No mixing of encodings — one rule for all cryptographic byte fields.

### Consequences

- **Consistency**: A single encoding rule applies to all binary-to-string conversions in the encryption layer. No risk of accidentally decoding a hex value as base64 or vice versa.
- **Compactness**: Base64 uses ~33% overhead vs hex's 100% overhead. A 32-byte salt is 44 chars in base64 vs 64 in hex; a 60-byte encrypted master key is ~80 chars vs 120.
- **DB column sizing**: `VARCHAR(255)` for `master_key_salt` and `TEXT` for `encrypted_master_key` are both sufficient. A 32-byte salt encodes to 44 base64 chars; service password ciphertexts will also fit comfortably.
- **Alignment with existing code**: The `Encrypt` function in `internal/encryption/crypto.go` already documents this convention: `"caller can get string with -> base64.StdEncoding.EncodeToString(returnedCiphertext)"`.

### Alternatives Considered

1. **Hex for salt, base64 for encrypted master key**: Matches the illustrative examples in `encryption_flow.md`, but mixing two encodings across similar fields creates cognitive overhead and increases the chance of a decode-with-wrong-encoding bug.
2. **BYTEA columns**: Store raw bytes directly. Eliminates encoding entirely but makes debugging with `psql` harder, complicates JSON serialization, and mismatches the sqlc-generated `string`/`pgtype.Text` types already in the codebase.
3. **Base64url encoding**: Slightly more URL-safe but unnecessary since these values are never placed in URLs. Standard base64 is simpler and matches `base64.StdEncoding` already referenced in code.

---

## DEC-003: Handler Method Naming Convention

**Date:** 2026-06-11
**Status:** Accepted

### Context

Handler methods on the `Server` struct used inconsistent naming: some were unexported (`handlerCreateUser`), some used unclear actions (`handlerGetAllServicesUser`), and some abbreviated names (`handlerUpdateServicePass`). The `handler` prefix was redundant on methods since the receiver (`Server`) already provides context.

### Decision

Use the `Handle<Action><Resource>` pattern for all handler methods:

```go
func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request)
func (s *Server) HandleListServices(w http.ResponseWriter, r *http.Request)
func (s *Server) HandleDecryptServiceByID(w http.ResponseWriter, r *http.Request)
func (s *Server) HandleUpdateServicePassword(w http.ResponseWriter, r *http.Request)
```

Rules:
- Methods are exported (`Handle` prefix) for consistency with Go's `http.Handler` conventions
- Action verbs are clear: `Create`, `List`, `Decrypt`, `Update`, `Login`
- Collection endpoints use `List` (REST convention) not `GetAll`
- No abbreviations: `Password` not `Pass`, `Service` not `Svc`
- Resource names are explicit: `ServicePassword` not `ServicePass`, `ServiceByID` not `ByID`

### Consequences

- Handler names are self-documenting and searchable
- Route registration in `RegisterRoutes` reads clearly: `r.Post("/api/services", s.HandleCreateService)`
- Test names follow naturally: `TestHandleCreateUser`, `TestHandleListServices`

---

## DEC-004: Separate DTO Package for Request/Response Types

**Date:** 2026-06-11
**Status:** Accepted

### Context

Request and response structs were originally defined in the `server` package. As we extracted a service layer, these types needed to be shared between `server/` (handlers) and `service/` (business logic). Keeping them in `server/` would create a circular dependency or force the service to import the server package. Additionally, request validation logic (`Validate()` methods) was coupled to the same file as the generic `decodeAndValidate` function.

### Decision

Move all request/response types and their `Validate()` methods into `internal/dto/dto.go`. Move the `Validator` interface and `decodeAndValidate` generic function there as well. Service methods return DTO types directly (e.g., `dto.UserResponse`, `dto.LoginResponse`), eliminating double-mapping between "result" types and "response" types.

```go
// internal/dto/dto.go
type CreateUserRequest struct { ... }
type LoginRequest struct { ... }
type UserResponse struct { ... }
type LoginResponse struct { ... }

func DecodeAndValidate[T Validator](body io.Reader) (T, error) { ... }
```

### Consequences

- No circular dependencies: `dto` is imported by both `server/` and `service/`
- DTOs serve as the API contract between frontend and backend, centralized in one file
- Service methods skip intermediate "result" types — `CreateUser()` returns `dto.UserResponse` directly
- Handlers pass DTOs straight to `RespondWithJSON()` without mapping
- `LoginRequest` and `CreateUserRequest` are separate types despite identical fields — they represent different operations with potentially divergent validation rules

### Alternatives Considered

1. **Shared `internal/dto` package (chosen)**: Clean dependency flow, single source of truth for API shapes.
2. **Service returns internal types, handler maps to response types**: Requires double mapping (service result → response struct) for every endpoint. DRY violation.
3. **Service returns response types directly (in `server/` package)**: Leaks HTTP/json concerns into service layer. Service would depend on `server/`.

---

## DEC-005: Service Layer Extraction

**Date:** 2026-06-11
**Status:** Accepted

### Context

Handlers originally contained all business logic: request parsing, validation, database queries, encryption orchestration, and response formatting. This made handlers difficult to test, reuse, and maintain. Cryptographic key derivation logic (decode salt → derive key → decrypt master key) was duplicated across three handler files.

### Decision

Extract business logic into `internal/service/` with a `Service` struct holding `*db.Database` and `*config.Config` as dependencies. Each handler becomes: decode request → call service → map errors to HTTP status codes → encode response.

```go
// internal/service/service.go
type Service struct {
    Db     *db.Database
    Config *config.Config
}

// internal/service/errors.go
var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUserNotFound       = errors.New("user not found")
    ErrServiceNotFound    = errors.New("service not found")
)
```

The shared crypto helper `decryptMasterKey(password, user)` was extracted from three handler files into `service/service_credentials.go`.

### Consequences

- Handlers have no imports on `db`, `encryption`, `auth`, or `base64` — they only deal with HTTP concerns
- `Server` struct no longer has a `Db` field — all DB access goes through `Service`
- Business logic is testable independently by calling service methods directly
- Duplicate crypto code eliminated via `decryptMasterKey()` helper
- Sentinel errors (`ErrInvalidCredentials`, etc.) allow handlers to map domain failures to HTTP status codes without knowing implementation details

### Alternatives Considered

1. **Keep logic in handlers**: Simple but doesn't scale. Handlers become 150+ lines of mixed concerns.
2. **Repository pattern (separate DB layer)**: Overkill for sqlc-generated queries. The service layer wraps queries when needed; no abstraction gap to fill.

---

## DEC-006: Request Struct Naming Convention

**Date:** 2026-06-11
**Status:** Accepted

### Context

Request structs used inconsistent naming: `UserRequest` was overloaded for both user creation and login, `NewServiceRequest` had a redundant prefix, and `CredentialsRequest` used a noun instead of describing the action.

### Decision

Name request structs with the pattern `<Action><Resource>Request`, matching the handler action:

| Old | New | Rationale |
|---|---|---|
| `UserRequest` (for create) | `CreateUserRequest` | Explicit action |
| `UserRequest` (for login) | `LoginRequest` | Separate type for separate operation |
| `NewServiceRequest` | `CreateServiceRequest` | No redundant "New" prefix |
| `CredentialsRequest` | `DecryptServiceRequest` | Describes the action, not a vague noun |
| `UpdateServiceRequest` | `UpdateServiceRequest` | Already correct |

### Consequences

- Request type names mirror their handler: `HandleCreateUser` ↔ `CreateUserRequest`
- No single struct serves two different operations
- Validation rules are scoped to the specific operation

---

## DEC-007: GenerateRandomBytes Returns Error Instead of Panic

**Date:** 2026-06-11
**Status:** Accepted

### Context

`encryption.GenerateRandomBytes(n int) []byte` called `panic(err)` if `crypto/rand.Read` failed. While rare, this is an operational error (e.g., file descriptor exhaustion) that should be handled gracefully, not crash the entire server.

### Decision

Changed signature to return an error:

```go
func GenerateRandomBytes(n int) ([]byte, error)
```

All callers in production code (`service/user.go`) now handle the error. Test callers use a `mustGenerateRandomBytes` helper that calls `t.Fatalf` on failure.

### Consequences

- No unexpected panics from the crypto package under resource pressure
- Consistent with the error-returning pattern of `DeriveKey`, `Encrypt`, and `Decrypt`
- Callers must handle the error, which is appropriate for cryptographic operations

### Alternatives Considered

1. **Keep panicking**: Simpler caller code, but crashes the server on a recoverable condition.
2. **Return error (chosen)**: Standard Go practice for library code. Forces callers to acknowledge the failure mode.

---

## DEC-008: Consistent HTTP Status Code Constants

**Date:** 2026-06-11
**Status:** Accepted

### Context

Some handlers used raw integers (`200`, `201`) while others used `http.StatusXxx` constants (`http.StatusOK`, `http.StatusCreated`). Mixed styles reduce readability and make it easier to use wrong codes.

### Decision

All status codes use `http.StatusXxx` constants. No raw integers.

```go
// Do:
s.RespondWithJSON(w, http.StatusCreated, result)
s.RespondWithError(w, http.StatusUnauthorized, ...)

// Don't:
s.RespondWithJSON(w, 201, result)
s.RespondWithError(w, 401, ...)
```

---

## DEC-009: Test Helpers Excluded from Production Binary

**Date:** 2026-06-11
**Status:** Accepted

### Context

`test_utils.go` contained test helper functions (`newTestDB`, `newTestServer`, `cleanupTestDB`) but used a non-`_test.go` filename, meaning it was compiled into the production binary. These functions all require `*testing.T` and are only callable from test code.

### Decision

Renamed `test_utils.go` → `test_utils_test.go`. Go only compiles `_test.go` files during `go test`, so these helpers are excluded from production builds.

### Consequences

- Smaller production binary
- No dead code in production
- Functions are still accessible to all test files in the same package

---

## DEC-010: Centralized Package-Level Sentinel Errors

**Date:** 2026-06-11
**Status:** Accepted

### Context

The sentinel error `errInvalidUserID` was declared in `services_create_handler.go` but referenced across four handler files. This created an implicit dependency and made it unclear where shared errors should live.

### Decision

Package-level sentinel errors are consolidated into dedicated files:
- `internal/server/errors.go` — handler-level errors (`errInvalidUserID`)
- `internal/service/errors.go` — service-level errors (`ErrInvalidCredentials`, `ErrUserNotFound`, `ErrServiceNotFound`)

### Consequences

- Shared errors are discoverable in one place per layer
- Handler files are cleaner — no `var` blocks mixed with handler logic
- Service errors are exported (`ErrXxx`) so handlers can compare with `errors.Is()` across package boundaries