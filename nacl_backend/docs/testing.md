# Testing Guide - NACL Backend

This document describes the testing patterns and conventions used in the NACL backend.

---

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Integration Testing with ServeHTTP](#integration-testing-with-servehttp)
- [How ServeHTTP Testing Works](#how-servehttp-testing-works)
- [Setting Up a Test Server](#setting-up-a-test-server)
- [Writing Integration Tests](#writing-integration-tests)
- [Test Environment](#test-environment)
- [Cleaning Up Test Data](#cleaning-up-test-data)
- [Comparison: Direct Handler vs ServeHTTP](#comparison-direct-handler-vs-servehttp)

---

## Testing Philosophy

We use **integration-style tests** that exercise the full HTTP request lifecycle: middleware, routing, and handlers. This catches bugs that unit-testing handlers in isolation would miss, such as:

- Middleware not running (e.g., authentication bypassed)
- Missing `return` statements after error responses
- Route configuration issues
- Request context not being passed correctly

---

## Integration Testing with ServeHTTP

Instead of calling handler methods directly, we route requests through the full chi router using `ServeHTTP`. This sends the request through all middleware (logging, recovery, token validation) and into the handler, exactly as it would happen in production.

### Why ServeHTTP over httptest.NewServer

`httptest.NewServer` starts a real HTTP server on a random port, making actual network calls. `ServeHTTP` with `httptest.NewRecorder` is an **in-memory call** that exercises the same code path without network overhead, port binding, or startup cost.

| Approach | Network | Speed | Complexity |
|----------|---------|-------|------------|
| `ServeHTTP` + `httptest.NewRecorder` | None | Fast | Low |
| `httptest.NewServer` | Localhost | Slower | Medium |
| Direct handler call | None | Fast | Low |

We chose `ServeHTTP` because it gives us the full middleware chain without the overhead of starting a real HTTP server.

---

## How ServeHTTP Testing Works

### The Request Flow

```
Test Code
    │
    │  1. Build http.Request with method, path, body, headers
    │
    ▼
httptest.NewRequest(http.MethodPost, "/api/services", body)
    │
    │  2. Create ResponseRecorder to capture the response
    │
    ▼
rr := httptest.NewRecorder()
    │
    │  3. Send request through the full router + middleware chain
    │
    ▼
server.HTTPServer.Handler.ServeHTTP(rr, req)
    │
    │  ┌──────────────────────────────────────┐
    │  │         Chi Router + Middleware        │
    │  │                                       │
    │  │  RequestLogger ──► Recovery ──► ...   │
    │  │                                       │
    │  │  For /api/services:                  │
    │  │  TokenValidator ──► handlerCreateService │
    │  │                                       │
    │  │  For other routes:                    │
    │  │  handlerCreateUser / handlerLogin     │
    │  └──────────────────────────────────────┘
    │
    │  4. Assert on rr.Code, rr.Body, etc.
    │
    ▼
assert.Equal(t, http.StatusCreated, rr.Code)
```

### Step by Step

1. **Build the request** — `httptest.NewRequest` creates an `*http.Request` with the method, URL path, and body. You set headers (Content-Type, Authorization) directly on the request object.

2. **Create a response recorder** — `httptest.NewRecorder()` creates a `*httptest.ResponseRecorder` that implements `http.ResponseWriter`. It captures the status code, headers, and body that the handler writes — all in memory, no network.

3. **Send through the router** — `server.HTTPServer.Handler.ServeHTTP(rr, req)` calls the chi router's `ServeHTTP` method. Chi matches the request path to the correct route, runs all middleware in order (logging, recovery, token validation for protected routes), and then calls the handler.

4. **Assert on the response** — `rr.Code` contains the HTTP status code, `rr.Body` contains the response body. Use `assert.Equal`, `assert.Contains`, etc. to verify the response.

---

## Setting Up a Test Server

`newTestServer` in `test_utils.go` creates a Server struct with a fully wired chi router:

```go
func newTestServer(t *testing.T, database *db.Database) *Server {
    t.Helper()

    cfg := &config.Config{
        Port:      3334,
        JwtSecret: "test-secret",
        LogFile:   "/tmp/test.log",
    }

    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: slog.LevelError,
    }))

    s := &Server{
        Config: cfg,
        Db:     database,
        Logger: logger,
    }

    r := chi.NewRouter()
    s.RegisterRoutes(r)

    s.HTTPServer = &http.Server{Handler: r}

    return s
}
```

Key points:

- **`RegisterRoutes`** — The same method used by `NewServer` in production. This ensures tests use the exact same routing configuration as the real server: same middleware, same route groups, same protected routes.
- **`s.HTTPServer.Handler = r`** — The chi router is set as the handler so tests can call `server.HTTPServer.Handler.ServeHTTP(rr, req)`.
- **Test JWT secret** — Uses `"test-secret"` instead of the real secret. Tokens minted in tests use this secret, and the `TokenValidator` middleware uses it to verify them.
- **Error-level logging** — Reduces noise in test output by only logging errors.

---

## Writing Integration Tests

### Example: Protected Route Test

```go
func TestHandlerCreateService(t *testing.T) {
    testDB := newTestDB(t)
    defer testDB.Close()

    server := newTestServer(t, testDB)

    // Step 1: Create a user (routed through the full stack)
    body := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, testUser, testPass)
    req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    server.HTTPServer.Handler.ServeHTTP(rr, req)

    // Step 2: Log in to get a JWT token (routed through the full stack)
    body = fmt.Sprintf(`{"username": "%s", "password": "%s"}`, testUser, testPass)
    req = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rr = httptest.NewRecorder()
    server.HTTPServer.Handler.ServeHTTP(rr, req)

    var loginResp LoginResponse
    json.NewDecoder(rr.Body).Decode(&loginResp)
    token := loginResp.Token

    // Step 3: Test a protected route with table-driven tests
    tests := []struct {
        name       string
        authorized bool
        wantCode   int
    }{
        {"authorized request", true, http.StatusCreated},
        {"unauthorized - no token", false, http.StatusUnauthorized},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodPost, "/api/services", strings.NewReader(requestBody))
            req.Header.Set("Content-Type", "application/json")
            if tt.authorized {
                req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
            }
            rr := httptest.NewRecorder()

            // This goes through: RequestLogger → Recovery → TokenValidator → handler
            server.HTTPServer.Handler.ServeHTTP(rr, req)

            assert.Equal(t, tt.wantCode, rr.Code)
        })
    }
}
```

### Why the Authorization Header Instead of Injecting Context

Earlier versions of tests called handler methods directly and injected user IDs into the request context manually:

```go
// Old approach: bypasses middleware, doesn't test authentication
ctx := auth.WithUserID(r.Context(), userID)
r = r.WithContext(ctx)
server.handlerCreateService(rr, r)
```

The new approach sends a real JWT token through the `Authorization` header and lets the `TokenValidator` middleware parse and validate it:

```go
// New approach: tests the full middleware chain including authentication
req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
server.HTTPServer.Handler.ServeHTTP(rr, req)
```

This catches bugs that direct handler calls miss:
- Middleware returning 401 before the handler is reached
- Missing `return` after error responses in handlers
- Token validation logic errors
- Context not being set correctly by middleware

---

## Test Environment

Tests require a running PostgreSQL database. Set `DATABASE_URL_TEST` before running tests:

```bash
export DATABASE_URL_TEST="postgresql://postgres:postgres@localhost:5433/nacl_test?sslmode=disable"
go test ./internal/server/ -v
```

Or use the Makefile target which sets up the test database:

```bash
make test
```

---

## Cleaning Up Test Data

Use `cleanupTestDB` to truncate tables between test cases:

```go
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        cleanupTestDB(t, testDB, "services")

        // ... run test ...
    })
}
```

Truncate tables that the current test modifies, not all tables. For example, when testing service creation, truncate `"services"` but not `"users"` (the user persists across subtests).

---

## Comparison: Direct Handler vs ServeHTTP

| Aspect | Direct Handler Call | ServeHTTP Through Router |
|--------|---------------------|--------------------------|
| Middleware runs | No | Yes |
| Authentication tested | No (must inject context) | Yes (send real JWT) |
| Route matching tested | No (call handler directly) | Yes (request hits correct route) |
| Catches missing `return` | Only if you check response writes | Yes — second write causes visible error |
| Setup complexity | Low | Low-medium |
| Test fidelity | Low | High |

**Recommendation:** Use `ServeHTTP` integration tests for all endpoint tests. Use direct handler calls only for unit-testing internal helper functions that don't depend on middleware.

---

## Related Documentation

- [Encryption Flow](./encryption_flow.md)
- [Error Handling](./error_handling.md)
- [Development Environment Setup](./dev_env_config.md)