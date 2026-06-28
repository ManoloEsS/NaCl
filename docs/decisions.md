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

Handler methods on the `Server` struct used inconsistent naming: some were unexported (`handlerCreateUser`), some used unclear actions (`handlerGetAllCredentialsUser`), and some abbreviated names (`handlerUpdateServicePass`). The `handler` prefix was redundant on methods since the receiver (`Server`) already provides context.

### Decision

Use the `Handle<Action><Resource>` pattern for all handler methods:

```go
func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request)
func (s *Server) HandleListCredentials(w http.ResponseWriter, r *http.Request)
func (s *Server) HandleDecryptCredentialByID(w http.ResponseWriter, r *http.Request)
func (s *Server) HandleUpdateCredentialPassword(w http.ResponseWriter, r *http.Request)
```

Rules:
- Methods are exported (`Handle` prefix) for consistency with Go's `http.Handler` conventions
- Action verbs are clear: `Create`, `List`, `Decrypt`, `Update`, `Login`
- Collection endpoints use `List` (REST convention) not `GetAll`
- No abbreviations: `Password` not `Pass`, `Service` not `Svc`
- Resource names are explicit: `CredentialPassword` not `CredentialPass`, `CredentialByID` not `ByID`

### Consequences

- Handler names are self-documenting and searchable
- Route registration in `RegisterRoutes` reads clearly: `r.Post("/api/credentials", s.HandleCreateCredential)`
- Test names follow naturally: `TestHandleCreateUser`, `TestHandleListCredentials`

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
    ErrCredentialNotFound    = errors.New("credential not found")
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
| `NewServiceRequest` | `CreateCredentialRequest` | No redundant "New" prefix |
| `CredentialsRequest` | `DecryptCredentialRequest` | Describes the action, not a vague noun |
| `UpdateServiceRequest` | `UpdateCredentialRequest` | Renamed to match resource |

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
- `internal/service/errors.go` — service-level errors (`ErrInvalidCredentials`, `ErrUserNotFound`, `ErrCredentialNotFound`)

### Consequences

- Shared errors are discoverable in one place per layer
- Handler files are cleaner — no `var` blocks mixed with handler logic
- Service errors are exported (`ErrXxx`) so handlers can compare with `errors.Is()` across package boundaries

---

## DEC-011: Envelope Encryption (Two-Key Design)

**Date:** 2026-06-28
**Status:** Accepted

### Context

A password manager must encrypt many stored credentials per user and must support changing the user's login password. The naive approach derives the encryption key directly from the login password: changing the password requires re-encrypting every credential row. Two undesirable consequences follow: the cost of rotation is `O(credentials)`, not `O(1)`, and the rotation flow is exposed to interruption leaving rows half-migrated.

### Decision

Use a two-key envelope scheme:

1. Generate a random 32-byte **master key** per user at registration.
2. Encrypt every credential (service username and service password) with that master key using AES-256-GCM, with a fresh 12-byte nonce per encryption.
3. Derive a **wrapping key** from the user's login password and a 32-byte random salt using Argon2id.
4. Encrypt the master key with the wrapping key using AES-256-GCM, and persist only the encrypted master key (`users.encrypted_master_key`) and the salt (`users.master_key_salt`).
5. On password rotation, decrypt the master key with the old wrapping key, derive a new wrapping key from the new password, and re-encrypt only the master key. Credential rows are untouched.

The master key never goes to the database in plaintext. The login password never encrypts credentials directly.

### Consequences

- Password rotation is `O(1)`: one row update, atomic.
- Master key is cryptographically random, stronger than any user-chosen password used as a key.
- Authentication (Argon2id password hash) and confidentiality (master key) use separate algorithms with separate parameters, so a weakness in one does not break the other.
- Server-side encryption: the master key exists in process memory for the duration of one operation, so a compromised server can exfiltrate keys. This is an accepted trade-off; a full zero-knowledge posture would require client-side encryption.
- Industry precedent: 1Password and Bitwarden use this envelope pattern.

### Alternatives Considered

1. **Direct encryption from password-derived key.** Simpler (one key) but rotation is `O(credentials)` and risks partial migration on interruption. Rejected.
2. **Per-credential random keys wrapped by the login password.** Fine-grained but raises per-credential overhead and complicates key management. Not needed for the threat model of this MVP.

Full design reference: [`nacl_backend/docs/encryption_flow.md`](../nacl_backend/docs/encryption_flow.md).

---

## DEC-012: Argon2id (not bcrypt) for Password Hashing

**Date:** 2026-06-28
**Status:** Accepted

### Context

Early design notes and the draft of `encryption_flow.md` described bcrypt as the login password hashing algorithm. Bcrypt is widely supported and battle-tested but is GPU-parallelizable; Argon2id (the PHC winner) is memory-hard and is the OWASP-recommended choice for new systems. The codebase was migrated to Argon2id before the registration endpoint shipped.

There are two distinct Argon2id call sites in NaCl, which is a common source of confusion:

- **Key derivation** (`internal/encryption/crypto.go`, via `golang.org/x/crypto/argon2`): `time=3`, `memory=32 MiB`, `threads=4`, `output=32 bytes`. Used to derive the wrapping key from the login password and salt.
- **Password hashing** (`internal/auth/password_hash.go`, via `alexedwards/argon2id`): uses the library's `DefaultParams`. Used to verify the login password for authentication.

They share an algorithm name but use different parameters because they answer different questions (a 32-byte symmetric key vs. a slow password verifier).

### Decision

Use Argon2id for both key derivation and password hashing, from different libraries with different parameters per call site. Bcrypt is no longer used anywhere.

### Consequences

- Memory-hard KDF/PHF across both call sites, matching modern OWASP guidance.
- Two separate parameter sets must be reasoned about independently. Documenting them side-by-side (this ADR and `encryption_flow.md`) prevents the conflation.
- Earlier documentation incorrectly described bcrypt as the hashing algorithm. Those references have been corrected.

### Alternatives Considered

1. **Bcrypt for password hashing.** Widely supported but GPU-parallelizable. Rejected in favor of Argon2id.
2. **scrypt.** Memory-hard but less favored by current OWASP guidance than Argon2id.
3. **Single Argon2id call site shared by both concerns.** Rejected because the key derivation parameter set (output length, parallelism, memory) is tuned for a 32-byte key, while the password hash parameter set is tuned for slow verification with a portable encoded output. Conflating them produces a worse fit for both.

---

## DEC-013: BYTEA for Credentials Columns, Base64 for Users Columns (Carve-Out from DEC-002)

**Date:** 2026-06-28
**Status:** Accepted

### Context

DEC-002 established base64 encoding for all cryptographic byte fields stored in PostgreSQL text columns. The credentials table stores ciphertext for the service username and service password. The users table stores the salt and encrypted master key. After sqlc generation, the two tables surfaced their encrypted fields differently, and a reader of the codebase would fairly ask whether one side is a bug.

### Decision

Different storage conventions apply to the two tables, and this is intentional:

- **`users.master_key_salt` and `users.encrypted_master_key`**: stored as base64 strings in `VARCHAR(255)` / `TEXT`, per DEC-002. Base64 keeps them inspectable via `psql`, matches the DTO round-trip shape, and survives JSON serialization cleanly.
- **`credentials.encrypted_service_username` and `credentials.encrypted_password`**: stored as raw `BYTEA`, surfaced by sqlc as `[]byte`. The credential ciphertext is opaque bulk data with no JSON surface (the `GET /api/credentials` response strips these fields entirely, per DEC-019), so BYTEA avoids encoding overhead and matches sqlc's native `[]byte` mapping for the values that flow directly `service.Encrypt` -> `db.CreateCredentialParams.EncryptedServiceUsername` with no base64 step on the path.

### Consequences

- DEC-002's "one rule for all cryptographic byte fields" claim is now scoped to the *users* table only. This ADR is the explicit carve-out for the *credentials* table.
- A new engineer will see the inconsistency and may suspect a bug; this ADR exists to make the choice intentional and explainable.
- No base64 overhead for per-credential ciphertext, and sqlc returns the type the service code uses directly.
- `psql` reads of the credentials table show bytes, not a readable encoding. The trade-off is acceptable because the values are never read in plaintext at the DB level anyway (the column content is ciphertext).

### Alternatives Considered

1. **Apply DEC-002 uniformly: base64 columns everywhere.** Cleaner rule but adds an encode/decode step on every credential read/write and a base64 column consumes more storage per row. No payoff for opaque ciphertext.
2. **Apply BYTEA uniformly: drop base64 from the users table too.** Loses `psql` inspectability of the salt and the encrypted master key, where inspectability is occasionally useful for support and debugging.

---

## DEC-014: JWT Auth Strategy (HS256, 30-Minute Expiry, No Refresh, No Revocation, No `/api/me`)

**Date:** 2026-06-28
**Status:** Accepted

### Context

The auth model needs to authenticate stateless HTTP requests without a per-request DB lookup. Stateful sessions would require a session store and a token blacklist for logout. The MVP scope (see `docs/user_stories.md`) explicitly excluded refresh tokens, token blacklisting, and account lockout. A known contract drift exists: the frontend `AuthContext` calls `GET /api/me` to refresh user data, but no such route is implemented.

### Decision

- Use symmetric HS256-signed JWTs with `sub = userID`, `iss = "nacl-access"`, and `exp = now + 30 minutes`.
- Validate tokens in the `TokenValidator` middleware by verifying the signature and claims; extract `userID` into the request context via `auth.WithUserID`.
- Logout is purely client-side (`localStorage.clear`); the server cannot revoke a token.
- Do not implement refresh tokens; clients re-authenticate with the login endpoint after expiry.
- Do not implement `GET /api/me`: the `POST /api/login` response is the single source of truth for the user shape, and clients are expected to persist it alongside the token.

### Consequences

- No DB lookup per request -> low auth overhead.
- A stolen token is valid for up to 30 minutes, with no server-side way to revoke it. This is the documented MVP risk; rate limiting would reduce but not eliminate the exposure.
- On password rotation, existing tokens remain valid until they expire. This is acceptable for the MVP threat model but is a known trade-off (see Roadmap).
- Frontend currently has a residual `GET /api/me` call (a 404 with no callable route). This is on the cleanup list and called out here so a future engineer treats it as drift, not a missing feature to implement. Either the call should be removed (preferred for MVP) or `GET /api/me` added explicitly with its own decision.

### Alternatives Considered

1. **Asymmetric (RS256) JWTs.** Allows key rotation and segmented trust, but the only signer is this backend. No payoff at MVP scale.
2. **Refresh tokens.** Required for long-lived sessions but adds storage, rotation, and revocation complexity explicitly excluded by MVP scope.
3. **Server-side session with blacklist.** Rejected for MVP; trade-off of stateless JWT accepted.

---

## DEC-015: Re-Authentication on Every Sensitive Operation (The `user_password`-in-body Pattern)

**Date:** 2026-06-28
**Status:** Accepted

### Context

JWT middleware proves the requester holds a *currently valid* token but does not prove the requester still knows the master password. A stale JWT (issued in the last 30 minutes, but the browser was left unlocked) is enough to enumerate credential metadata via `GET /api/credentials`. To actually expose plaintext, the server must prove the requester can still unlock the vault.

### Decision

Every credential-mutating endpoint, plus the password rotation endpoint, requires the login password in the request body as `user_password`:

- `POST /api/credentials` (create)
- `POST /api/credentials/{id}/decrypt` (decrypt/reveal)
- `PATCH /api/credentials/{id}` (update)
- `POST /api/credentials/{id}/delete` (delete)
- `PATCH /api/users` (rotate login password)

The service verifies `user_password` against the Argon2id hash, then derives the wrapping key, decrypts the master key, and uses it to encrypt or decrypt the targeted credential. The master key is never held in memory across requests; each sensitive op re-derives it from the user's password.

### Consequences

- Defense in depth: a stolen JWT alone is not sufficient to reveal plaintext or mutate credentials. The attacker must also know the login password.
- The master key has a tight lifetime: in-memory only, gc'd after the operation returns.
- UX cost: the frontend re-prompts for the login password on every save/reveal/update/delete/rotate. This is intentional and is the visible security posture of the app. Users who dislike it can be told the threat model.
- Every re-auth is also audited via the operations log (DEC-023).

### Alternatives Considered

1. **Cache the master key server-side per session.** Lower UX cost but increases the in-memory blast radius and requires session storage and invalidation logic. Rejected.
2. **Cache the master key client-side.** Stronger than option 1 but pushes key material into the browser, where it is harder to protect (XSS, devtools). Rejected.
3. **No re-auth (JWT-only).** Lowest UX cost but reduces a stolen-token attack to "can read plaintext for 30 minutes". Rejected as too permissive even for an MVP.

---

## DEC-016: Single Deployable Binary via `//go:embed` for the Embedded Frontend

**Date:** 2026-06-28
**Status:** Accepted

### Context

A React SPA needs to be served from somewhere. Two common options are (1) a separate static-file host (S3, Vercel, Netlify) fronting the API, or (2) embedding the built assets into the Go binary and serving them from the same process as the API. NaCl has one deployment target, no CDN requirement, and needs to work behind a single TLS endpoint with no CORS.

### Decision

Build the Vite frontend into `nacl_backend/static/` and embed it into the Go binary at compile time via `//go:embed static`. The Go process serves `index.html` from the embedded filesystem and serves `/assets/*` from `static/assets/` using `http.FileServer` + `http.StripPrefix`. No separate frontend host and no CORS.

### Consequences

- One artifact per release: a single binary that contains both the API and the SPA.
- Atomic deploys: the frontend and backend move together; version skew is impossible.
- No CORS configuration needed because the API and the SPA share the same origin.
- The build step must run Vite before `go build`; the Dockerfile handles this in a multi-stage build (`node` stage -> `golang` stage, see DEC-024).
- Dev workflow uses two processes (vite dev server + air hot-reload), with the Vite config proxying `/api/*` to the backend.

### Alternatives Considered

1. **Separate frontend host.** Adds operational surface (CDN, CORS, two deploy pipelines, two TLS endpoints). Rejected for an MVP with a simple deployment target.
2. **Embed via `go:generate` of compiled assets.** More indirection; `//go:embed` is one line and idiomatic.

---

## DEC-018: Embedded goose Migrations Run Automatically on Startup

**Date:** 2026-06-28
**Status:** Accepted

### Context

Database migrations can be applied (a) manually by an operator or CI step before deploy, or (b) automatically by the application on boot. Manual migrations require the deploy pipeline to own a `goose up` step, and version-skew between binary and schema becomes possible if the operator forgets. The MVP deploys one container at a time, so migrator contention is not a concern.

### Decision

Embed all `.sql` migration files via `//go:embed` and call `goose.Up(db, ...)` from `main.go` before the HTTP server starts. A failed migration is fatal: the process exits with a non-zero status and no requests are served against a mismatched schema.

### Consequences

- The binary is self-bootstrapping: no manual `migrate` step in the deploy pipeline.
- A bad migration bricks startup, which is the desired behavior (no serving against a torn schema).
- Multi-replica deploys would race on the migrator role; this is acceptable for the single-replica Render deploy and is a documented scaling concern for the future.
- Local dev still uses `make db-migrate` for incremental migration during development; the auto-migrate path is for production boot.

### Alternatives Considered

1. **CI-applied migrations.** Adds a deploy step; the deploy pipeline must own DB credentials and ordering. Rejected for MVP simplicity.
2. **Explicit CLI subcommand (`nacl migrate`).** Forces the operator to remember it on every release; error-prone and violates "the binary is self-bootstrapping".

---

## DEC-019: `GET /api/credentials` Returns Metadata Only; Plaintext Requires the Explicit Decrypt Endpoint

**Date:** 2026-06-28
**Status:** Accepted

### Context

A naive credential list endpoint would return the full row, encrypted or not, requiring the server to decrypt every credential on every list call. That maximizes the value of a stolen JWT: a single list call reveals every plaintext. It also forces the server to hold the master key in memory for the duration of a list, which is unnecessary.

### Decision

`GET /api/credentials` selects only `{id, service, description, encryption_algorithm}` from the database. The encrypted columns are not transmitted. To reveal plaintext, the client must call `POST /api/credentials/{id}/decrypt` with `user_password` in the body (re-authentication, DEC-015), which decrypts one credential and writes a `TypeDecrypt` audit row.

The distinction is enforced at the SQL layer: `GetAllCredentialsForUserId` explicitly selects the metadata columns, not the ciphertext columns. The `CredentialMetadataResponse` DTO therefore has no field for the service password.

### Consequences

- A stolen JWT cannot reveal plaintext; it can only enumerate metadata.
- Every plaintext reveal is a separate, auditable, re-authenticated HTTP call. Attack surface for "someone else's token reveals all my passwords" is limited to `O(1)` per request, not `O(vault)` per request.
- The list endpoint is cheap (no crypto, no master key in memory).
- The user re-prompts for the login password for each reveal, consistent with the re-auth model from DEC-015.

### Alternatives Considered

1. **Return all ciphertext on list, decrypt client-side.** Harder; needs key material on the client. Not in the MVP posture (DEC-015: server-side encryption).
2. **Return decrypted plaintext on list.** Maximum blast radius; rejected.

---

## DEC-020: Frontend Mirrors Backend DTOs via Zod with `.strict()`; Both Sides Validate

**Date:** 2026-06-28
**Status:** Accepted

### Context

The Go backend defines request and response structs in `internal/dto/dto.go`. The frontend has its own TypeScript types but could optionally assert the wire shape at runtime, which catches drift between the two sides earlier than the user noticing the page breaking.

### Decision

- Every request and response shape has a paired Zod schema in `nacl_frontend/src/lib/` (`requestValidation.ts`, `responseValidation.ts`).
- All Zod schemas use `.strict()` to reject unknown keys, so a backend change that adds a field will surface as a failing frontend parse rather than silent acceptance of an unexpected shape.
- Form-only fields (`confirm_service_password`, `confirm_new_password`, `credentialID`) are stripped from the request payload before the POST is sent; only the DTO-shaped fields go over the wire.
- ISO 8601 date strings from the backend are coerced to `Date` via `z.string().pipe(z.coerce.date())`.
- The crosswalk is documented in [`docs/dto-schema-map.md`](dto-schema-map.md), and that file also holds the API endpoint reference table.

### Consequences

- Contract drift between frontend and backend becomes a runtime-failing parse, not a silent bug.
- Both sides validate independently; the backend is authoritative.
- Frontend forms carry a few extra fields compared to the wire DTO (`confirm_*` for UX); the explicit strip keeps the wire shape clean.
- Schema updates incur a maintenance cost: a DTO change requires editing both `dto.go` and the corresponding Zod schema, plus `dto-schema-map.md`.

### Alternatives Considered

1. **No frontend validation (TypeScript types only, no runtime parse).** Faster to write, but backend shape changes surface only as visual bugs. Rejected.
2. **Generate the Zod schemas from the Go DTOs.** Removes the maintenance cost but requires a codegen step (no idiomatic bridge from Go structs to Zod schemas exists at MVP scale). Not worth the tooling for an app this size.

---

## DEC-021: `POST /api/credentials/{id}/delete` Instead of HTTP `DELETE`

**Date:** 2026-06-28
**Status:** Accepted

### Context

Deleting a credential requires re-authentication (DEC-015), which requires sending `user_password` in the request body. RFC 7231 allows a `DELETE` body but the semantics are discouraged and many proxies and clients strip it silently. A `DELETE` with a body works in tests and fails in some production paths.

### Decision

Use `POST /api/credentials/{id}/delete` with a JSON body containing `user_password`. Do not register an HTTP `DELETE` verb.

### Consequences

- The request body reliably reaches the handler, regardless of proxy chain.
- Slightly unconventional: a REST purist may flag it. The ADR is the answer to the question.
- Paired with the convention from DEC-015, the endpoint signature (`POST /xxx/delete` with `user_password`) reads as a re-authenticated action sub-resource, not a CRUD verb.

### Alternatives Considered

1. **`DELETE /api/credentials/{id}` with body.** Semantically discouraged; proxy stripping is a real risk.
2. **`DELETE` with `user_password` in a header.** Novel and unconstrained by spec; less idiomatic than a body, harder to document, and inconsistent with the rest of the API where sensitive fields go in the body.

---

## DEC-022: `slog` Dual-Handler with `apperr`-Aware `replaceAttr`

**Date:** 2026-06-28
**Status:** Accepted

### Context

The backend needed structured logging for observability and an `Apperr` package (`apperr.WithAttrs`) to attach key-value attributes to errors. The glue between the two is non-obvious: an `apperr` error with attributes needs to be able to flow through `slog` and have those attributes appear in the structured record.

### Decision

- Use stdlib `log/slog`.
- Run two handlers simultaneously in development: a debug-level text handler to stderr (readable in the terminal) and an info-level JSON handler to a file (`SALT_LOG_FILE`, structured-queryable).
- Buffer the JSON handler at 8 KB.
- In `replaceAttr`, detect `error`-typed values and unwrap `apperr.Attrs` into top-level slog fields, so attributes attached via `apperr.WithAttrs(err, "userID", x, ...)` survive into the JSON record.
- If the log file fails to open, fall back to a single unbuffered info-level text handler on stderr (Render uses an ephemeral filesystem).

### Consequences

- Dev logs are readable; production logs are structured-queryable.
- The `apperr` -> `slog` bridge requires the `replaceAttr` to know about `apperr`; this indirection is the source of truth for the bridge and an engineer reading `logger.go` should start here.
- Stdlib only, no third-party logging dependency.
- Render-friendly: the stderr fallback survives an ephemeral filesystem.

### Alternatives Considered

1. **`logrus` or `zap`.** Mature but heavier; stdlib `slog` (introduced in Go 1.21) is sufficient and adds no dependency.
2. **Single handler to stderr only.** Fine for dev but loses structured queryability for production debugging.
3. **No `replaceAttr` bridge.** `apperr` attributes would not surface in log records; the indirection between errors and logs would be invisible.

---

## DEC-023: Best-Effort Operation Log Audit Trail

**Date:** 2026-06-28
**Status:** Accepted

### Context

Credential mutations and logins should be auditable per user ("when was this credential decrypted, updated, or deleted?"). Failing to insert a `TypeDecrypt` / `TypeCreate` / `TypeUpdate` / `TypeDelete` row should not roll back the user's successful operation, but neither should silent audit failure be invisible.

### Decision

- Every credential-mutating handler and the login handler call `svc.SaveOperation` with a typed `OperationType` and the affected service name.
- `operations` is a PostgreSQL table with `ON DELETE CASCADE` on `users.id`, so deleting a user atomically purges their audit trail.
- `SaveOperation` failures are logged via `slog` but do not fail the user's request. Audit logging is best-effort.
- Operations are surfaced to the user via `GET /api/operations`, newest first.
- For operations without a specific service (login, password change), the `service` column is set to the literal string `"nil"`, not NULL, because the column is `TEXT NOT NULL` and a placeholder is more honest than a NULL.

### Consequences

- Audit rows can be silently missing if `SaveOperation` errors; the trade-off is that user-facing operations do not fail because an audit couldn't be written.
- The "nil" sentinel in the `service` column is a known string magic value; documented here so it is not mistaken for a real service name.
- `ON DELETE CASCADE` means deleting a user deletes their audit trail too. Acceptable for GDPR-style deletion; worth knowing if audit retention is ever regulated.

### Alternatives Considered

1. **Fail the request if `SaveOperation` errors.** User operations become coupled to audit availability; rejected as too brittle for an MVP.
2. **Out-of-band audit pipeline (queue + worker).** Out of scope for the MVP; the synchronous best-effort insert in the same transaction context is sufficient.

---

## DEC-024: Multi-Stage Dockerfile: `node` -> `golang` -> `alpine:3.20`

**Date:** 2026-06-28
**Status:** Accepted

### Context

The release artifact is a single `alpine:3.20` image containing the Go binary with the embedded frontend (DEC-016). Producing that artifact requires Node for the Vite build and Go for the binary build; neither belongs in the runtime image.

### Decision

Three-stage Dockerfile at the repo root:

1. **`node:25.9.0-slim` (frontend builder)**: `npm ci`, copy `nacl_frontend/`, run `npm run build`. Outputs `nacl_backend/static/` (the Vite build target is `../nacl_backend/static`, see `vite.config.ts`).
2. **`golang:1.26-alpine` (backend builder)**: `go mod download`, copy `nacl_backend/`, `COPY --from=frontend-builder` the `static/` directory into the Go build context, run `go build -o nacl`. The binary now embeds the frontend via `//go:embed static`.
3. **`alpine:3.20` (runtime)**: install `ca-certificates`, `COPY --from=backend-builder /app/nacl_backend/nacl .`, set `CMD ["./nacl"]`. No Node, no Go toolchain, no source.

### Consequences

- The runtime image contains only the binary and CA certificates; it is small and has minimal attack surface.
- Reproducible builds: the frontend and backend are produced in the same Docker invocation, so the binary is guaranteed to embed the matching frontend commit.
- A single `docker build` command produces the deployable artifact.
- Image selection rationale: `alpine:3.20` over `distroless` keeps a shell available for debugging on a Render host; the trade-off is a slightly larger image and the presence of a package manager in the runtime image.

### Alternatives Considered

1. **Two stages (slim runtime with Node).** Smaller build, but leaks Node into the runtime image.
2. **`distroless` runtime.** Smaller and no shell; harder to debug in a Render-only environment without exec access.
3. **Alpine variants other than 3.20.** Pinned for reproducibility across rebuilds.