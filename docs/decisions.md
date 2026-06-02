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