# Error Handling and Logging Guide

This document explains error handling patterns, when to use different error wrapping strategies, and how the logger handles errors.

---

## Table of Contents

- [Error Types](#error-types)
- [Error Wrapping Tools](#error-wrapping-tools)
- [Decision Tree](#decision-tree)
- [Logger Behavior](#logger-behavior)
- [Examples by Layer](#examples-by-layer)
- [Best Practices](#best-practices)

---

## Error Types

### 1. **Validation Errors** ❌

User input errors that are expected and self-explanatory.

**Examples:**
- Empty required fields
- Invalid email format
- Password too short
- Missing required parameters

**Handling:**
```go
if email == "" {
    return errors.New("email is required")
}

if len(password) < 8 {
    return errors.New("password must be at least 8 characters")
}
```

**Why no wrapping:**
- Error message is clear enough
- No stack trace needed
- No additional context needed

---

### 2. **Operational Errors** ⚠️

Expected errors from external systems that should be handled gracefully.

**Examples:**
- Database connection failures
- File not found
- Network timeouts
- External API errors

**Handling:**
```go
// Database operations
user, err := queries.GetUserByEmail(ctx, email)
if err != nil {
    return apperr.WithAttrs(
        pkgerrors.Wrap(err, "get user failed"),
        "email", email,
        "operation", "get_user",
    )
}

// File operations
file, err := os.Open("config.json")
if err != nil {
    return pkgerrors.Wrap(err, "failed to open config file")
}
```

**Why wrap:**
- Stack trace helps debugging
- Context helps trace user impact
- Should be logged but not crash the app

---

### 3. **Programming Errors** 💥

Bugs in the code that should never happen.

**Examples:**
- Nil pointer dereference
- Index out of bounds
- Type assertion failures
- Invariant violations

**Handling:**
```go
// Let it panic - Recovery middleware will catch it
user := users[0]  // Panics if slice is empty

// Or explicitly panic
if config == nil {
    panic("config must not be nil")
}
```

**Why panic:**
- Indicates a bug that needs fixing
- Should be caught by Recovery middleware
- Will be logged with stack trace

---

## Error Wrapping Tools

### 1. **Plain `errors` Package**

```go
import "errors"

errors.New("simple error message")
```

**Use for:**
- ✅ Validation errors
- ✅ Simple, self-explanatory errors
- ✅ Internal helper functions

**Does NOT provide:**
- ❌ Stack traces
- ❌ Error wrapping
- ❌ Context/attributes

---

### 2. **`fmt.Errorf` with `%w`**

```go
import "fmt"

fmt.Errorf("context: %w", underlyingError)
```

**Use for:**
- ✅ Simple error wrapping
- ✅ Preserving error chain
- ✅ When you don't need stack traces

**Provides:**
- ✅ Error chain (works with `errors.Unwrap()`)
- ✅ Custom context message
- ❌ No stack traces
- ❌ No structured attributes

---

### 3. **`pkg/errors` Package**

```go
import pkgerrors "github.com/pkg/errors"

pkgerrors.Wrap(err, "context")
pkgerrors.WithStack(err)
pkgerrors.Errorf("formatted: %v", err)
```

**Use for:**
- ✅ Database errors
- ✅ External service errors
- ✅ Any operational error needing stack trace

**Provides:**
- ✅ Stack traces (captured at wrap point)
- ✅ Error chain (works with `errors.Unwrap()`)
- ✅ Custom context message
- ❌ No structured attributes

---

### 4. **`apperr` Package (Custom)**

```go
import "github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"

apperr.WithAttrs(err, "key1", value1, "key2", value2)
apperr.Attrs(err)  // Extract all attributes from error chain
```

**Use for:**
- ✅ HTTP handler errors
- ✅ Service layer errors
- ✅ User-facing operations

**Provides:**
- ✅ Structured attributes (key-value pairs)
- ✅ Error chain preservation
- ✅ Works with logger to extract all context
- ❌ No stack traces (combine with pkg/errors)

---

## Decision Tree

```
Did an error occur?
    │
    ├─→ Is it a validation error? (user input, simple checks)
    │       └─→ Return plain error: errors.New("message")
    │
    ├─→ Is it an internal helper error? (not user-facing)
    │       └─→ Use fmt.Errorf: fmt.Errorf("context: %w", err)
    │
    ├─→ Is it a database/external service error?
    │       └─→ Wrap with stack trace: pkgerrors.Wrap(err, "context")
    │
    └─→ Is it a user-facing operational error?
            └─→ Wrap with stack trace + context:
                apperr.WithAttrs(pkgerrors.Wrap(err, "msg"), "key", value)
```

---

## Logger Behavior

### Your Logger Handles ALL Cases ✅

Current implementation in `internal/logger/logger.go`:

```go
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
    if a.Key == "error" {
        err, ok := a.Value.Any().(error)
        if !ok {
            return a
        }
        
        // 1. Extract error message
        errorAttrs := []slog.Attr{
            slog.String("message", err.Error()),
        }
        
        // 2. Extract stack trace (from pkg/errors)
        if stackErr, ok := errors.AsType[stackTracer](err); ok {
            errorAttrs = append(errorAttrs, slog.Any("stack", stackErr.StackTrace()))
        }
        
        // 3. Extract custom attributes (from apperr)
        errorAttrs = append(errorAttrs, apperr.Attrs(err)...)
        
        // 4. Group all under "error" key
        return slog.GroupAttrs("error", errorAttrs...)
    }
    return a
}
```

### What Gets Extracted

| Error Type | Message | Stack Trace | Attributes |
|------------|---------|-------------|------------|
| Plain `errors.New()` | ✅ Yes | ❌ No | ❌ No |
| `fmt.Errorf()` | ✅ Yes | ❌ No | ❌ No |
| `pkgerrors.Wrap()` | ✅ Yes | ✅ Yes | ❌ No |
| `apperr.WithAttrs()` | ✅ Yes | ❌ No | ✅ Yes |
| Combined (pkgerrors + apperr) | ✅ Yes | ✅ Yes | ✅ Yes |

### Error Chain Walking

The logger's `apperr.Attrs()` function walks the **entire error chain**:

```go
// Create error chain
err := errors.New("base error")
err = apperr.WithAttrs(err, "step", "1")
err = pkgerrors.Wrap(err, "wrapped")
err = apperr.WithAttrs(err, "step", "2")

// Logger extracts ALL attributes from ALL levels
logger.Error("failed", "error", err)

// Output includes:
// - message: "wrapped: base error"
// - stack: (from pkgerrors)
// - step: "1" (from first apperr)
// - step: "2" (from second apperr)
```

---

## Examples by Layer

### Configuration Layer

```go
// ❌ Don't use apperr - validation error
if cfg.JWTSecret == "" {
    return errors.New("JWT_SECRET is required")
}

// ❌ Don't use apperr - self-explanatory
file, err := os.Open("config.json")
if err != nil {
    return fmt.Errorf("failed to open config: %w", err)
}
```

---

### Database Layer

```go
// ✅ Use pkgerrors for stack trace
func (d *Database) NewDatabase(ctx context.Context, connString string) (*Database, error) {
    pool, err := pgxpool.New(ctx, connString)
    if err != nil {
        return nil, pkgerrors.Wrap(err, "failed to create connection pool")
    }
    
    if err := pool.Ping(ctx); err != nil {
        return nil, pkgerrors.Wrap(err, "database ping failed")
    }
    
    return &Database{Pool: pool}, nil
}
```

---

### Service Layer

```go
// ✅ Combine pkgerrors + apperr
func (s *Service) CreateUser(email, password string) error {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return apperr.WithAttrs(
            pkgerrors.Wrap(err, "failed to hash password"),
            "email", email,
            "operation", "create_user",
        )
    }
    
    user, err := s.queries.CreateUser(s.ctx, db.CreateUserParams{
        Email: email,
        PasswordHash: string(hash),
    })
    if err != nil {
        return apperr.WithAttrs(
            pkgerrors.Wrap(err, "failed to create user record"),
            "email", email,
            "operation", "create_user_db",
        )
    }
    
    return nil
}
```

---

### HTTP Handler Layer

```go
// ✅ Combine pkgerrors + apperr with request context
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // Validation error - plain error is fine
        s.RespondWithError(w, 400, "invalid request", err)
        return
    }
    
    // Validate required fields
    if req.Email == "" {
        s.RespondWithError(w, 400, "email is required", nil)
        return
    }
    
    // Operational error - use apperr
    user, err := queries.GetUserByEmail(r.Context(), req.Email)
    if err != nil {
        err = apperr.WithAttrs(
            pkgerrors.Wrap(err, "failed to get user"),
            "email", req.Email,
            "endpoint", "/api/login",
            "operation", "login",
        )
        s.Logger.Error("login failed", "error", err)
        s.RespondWithError(w, 500, "login failed", err)
        return
    }
}
```

---

### Middleware Layer

```go
// ✅ Recovery middleware uses logger automatically
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil && err != http.ErrAbortHandler {
                    // Logger extracts stack trace from panic
                    logger.Error("panic recovered",
                        "error", fmt.Errorf("%v", err),
                        "path", r.URL.Path,
                        "method", r.Method,
                        "stack", string(debug.Stack()),
                    )
                    w.WriteHeader(500)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## Best Practices

### ✅ DO

1. **Wrap errors at layer boundaries**
   ```go
   // Database → Service → Handler
   // Wrap at each layer with layer-specific context
   ```

2. **Use pkgerrors for stack traces**
   ```go
   pkgerrors.Wrap(err, "operation failed")
   ```

3. **Use apperr for user-facing context**
   ```go
   apperr.WithAttrs(err, "user_id", userID, "operation", "login")
   ```

4. **Combine both for critical errors**
   ```go
   apperr.WithAttrs(pkgerrors.Wrap(err, "msg"), "key", value)
   ```

5. **Let validation errors be simple**
   ```go
   errors.New("field is required")
   ```

6. **Log errors with context**
   ```go
   logger.Error("operation failed", "error", err)
   ```

---

### ❌ DON'T

1. **Don't wrap validation errors**
   ```go
   // Bad
   if email == "" {
       return apperr.WithAttrs(errors.New("email required"), "field", "email")
   }
   
   // Good
   if email == "" {
       return errors.New("email is required")
   }
   ```

2. **Don't wrap errors that already have context**
   ```go
   // Bad - SQL error already includes query
   err := db.Query("SELECT * FROM users")
   if err != nil {
       return apperr.WithAttrs(err, "query", "SELECT * FROM users")
   }
   ```

3. **Don't swallow errors**
   ```go
   // Bad
   err := doSomething()
   // ... forgot to handle error
   
   // Good
   err := doSomething()
   if err != nil {
       return err
   }
   ```

4. **Don't use panic for expected errors**
   ```go
   // Bad
   if user == nil {
       panic("user not found")  // This is expected, not a bug
   }
   
   // Good
   if user == nil {
       return errors.New("user not found")
   }
   ```

---

## Quick Reference

| Scenario | Use This | Example |
|----------|----------|---------|
| Validation | `errors.New()` | `errors.New("email required")` |
| Simple wrapping | `fmt.Errorf()` | `fmt.Errorf("context: %w", err)` |
| DB errors | `pkgerrors.Wrap()` | `pkgerrors.Wrap(err, "query failed")` |
| Handler errors | `apperr.WithAttrs()` | `apperr.WithAttrs(err, "user_id", id)` |
| Critical errors | Both combined | `apperr.WithAttrs(pkgerrors.Wrap(err, "msg"), "key", val)` |
| Programming bugs | `panic()` | `panic("invariant violated")` |

---

## Testing Error Handling

### Unit Tests

```go
func TestHandleLogin(t *testing.T) {
    // Test validation error
    req := &http.Request{Body: io.NopCloser(strings.NewReader(`{}`))}
    rr := httptest.NewRecorder()
    
    server.HandleLogin(rr, req)
    
    if rr.Code != 400 {
        t.Errorf("expected 400, got %d", rr.Code)
    }
}
```

### Integration Tests

```go
func TestDatabaseError(t *testing.T) {
    err := db.GetUser(ctx, "invalid-uuid")
    
    // Check error is wrapped correctly
    var pgErr *pgconn.PgError
    if !errors.As(err, &pgErr) {
        t.Error("expected PostgreSQL error")
    }
    
    // Check attributes are present
    attrs := apperr.Attrs(err)
    if len(attrs) == 0 {
        t.Error("expected error attributes")
    }
}
```

---

## Version History

| Date | Version | Changes |
|------|---------|---------|
| 2026-05-27 | 1.0.0 | Initial error handling documentation |

---

## Related Documentation

- [Development Environment Setup](./dev_env_config.md)
- [Testing Guide](./testing.md)
- [Logger Configuration](./dev_env_config.md#logger-configuration)
