# Testing

## Backend (Go)
- Run: `make test`
- Location: `*_test.go` files alongside source code
- Framework: Go's built-in `testing` package

## Frontend (React)
- Run: `npm test -- --run`
- Location: `src/test/*.test.tsx`
- Framework: Vitest + React Testing Library

---

# Backend Test Example

```go
package main

import (
	"testing"
)

func TestSomething(t *testing.T) {
	// Your test here
	t.Log("Test is working")
}
```

Run tests:
```bash
make test
```

---

# Frontend Test Example

```tsx
import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import App from '../App'

describe('App', () => {
  it('renders hello world', () => {
    render(<App />)
    expect(screen.getByText('Hello World')).toBeInTheDocument()
  })
})
```

Run tests:
```bash
npm test -- --run
```

---

# Test Files Location

**Backend:**
```
nacl_backend/
├── main_test.go
├── server/
│   └── server_test.go
└── internal/
    └── auth/
        └── auth_test.go
```

**Frontend:**
```
nacl_frontend/
└── src/
    └── test/
        ├── setup.ts
        ├── App.test.tsx
        └── components/
            └── Button.test.tsx
```

---

# Coverage

## Backend
```bash
# Run with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

## Frontend
```bash
# Run with coverage (add to vitest.config.ts)
npm test -- --run --coverage
```

---

# CI Integration

Tests run automatically on every push/PR via GitHub Actions:
- Backend: Lint → Build → Test (with PostgreSQL)
- Frontend: Lint → Type-check → Test → Build

See `.github/workflows/` for CI configuration.

---

# Best Practices

## Backend
- Name test files: `*_test.go`
- Use table-driven tests for multiple cases
- Use `t.Parallel()` for independent tests
- Mock database with interfaces

## Frontend
- Test user behavior, not implementation
- Use `screen` queries (Testing Library)
- Mock API calls with MSW or vitest mocks
- Keep tests focused on one behavior

---

# Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Vitest Documentation](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
