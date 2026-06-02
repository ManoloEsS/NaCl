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