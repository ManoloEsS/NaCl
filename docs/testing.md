# Testing

## Backend (Go)
- Run: `make test`
- Location: `*_test.go` files alongside source code
- Framework: Go's built-in `testing` package

## Frontend (React)
- Run: `npm test`
- Location: `nacl_frontend/e2e/*.spec.ts`
- Framework: Playwright (E2E)
- Requires: backend running on `localhost:3333` and frontend dev server on `localhost:5173`

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

# Frontend Test Overview

We use **Playwright** for end-to-end testing. Instead of testing components in isolation with mocked dependencies, E2E tests launch a real browser, interact with the app as a user would, and assert against the actual rendered page — including real API calls.

### Why E2E?

- Test real user flows (register → login → create credential → view vault)
- No mocking of API calls, auth context, or routing
- Catches integration bugs between frontend and backend
- Less code to write and maintain compared to unit + integration tests
- Backend engineers feel at home: you test the **whole stack**, not isolated React components

### Trade-offs

- Slower than unit tests (browser launch, full stack running)
- Requires the backend to be running
- Harder to test edge-case UI states (loading spinners, error boundaries)
- Flakier (network, timing, database state)

---

# Setup

## 1. Install Playwright

Already in `devDependencies`. If starting from scratch:

```bash
npm install -D @playwright/test
npx playwright install chromium
```

## 2. Configuration

`nacl_frontend/playwright.config.ts`:

```ts
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
    timeout: 30000,
  },
})
```

**What each part does:**
- `testDir` — Playwright looks for tests in `e2e/`
- `workers: 1` — tests run sequentially (important when they share database state)
- `baseURL` — all `page.goto('/login')` relative paths resolve against this
- `webServer` — Playwright auto-starts `npm run dev` before tests and kills it after
- `reuseExistingServer` — in dev, if the Vite dev server is already running, Playwright reuses it instead of starting a new one

## 3. Prerequisites

The **Go backend must be running** on port `3333`. Start it in a separate terminal:

```bash
cd nacl_backend
make run
```

Playwright will auto-start the Vite frontend dev server, but the backend is your responsibility.

---

# How to Write a Test

## Test File Structure

Tests live in `nacl_frontend/e2e/` and use the `.spec.ts` extension:

```
nacl_frontend/
  e2e/
    auth.spec.ts
    services.spec.ts
```

Each file follows this shape:

```ts
import { test, expect } from '@playwright/test'

test.describe('Feature Name', () => {
  test('does something specific', async ({ page }) => {
    // 1. Navigate
    await page.goto('/login')

    // 2. Interact (fill, click, select)
    await page.getByLabel('Username').fill('alice')
    await page.getByRole('button', { name: 'Login' }).click()

    // 3. Assert
    await expect(page).toHaveURL(/\/dash/)
    await expect(page.getByText('Dashboard')).toBeVisible()
  })
})
```

## Key Concepts

### Locators — How You Find Elements

Playwright locators are **auto-waiting**: they automatically wait for the element to exist and be visible before acting or asserting.

| Locator | When to use |
|---|---|
| `page.getByRole('button', { name: /login/i })` | Accessible roles (preferred) |
| `page.getByLabelText('Username')` | Form inputs with `<label>` |
| `page.getByText('Dashboard')` | Arbitrary text content |
| `page.getByPlaceholder('Password')` | Inputs with a placeholder |
| `page.locator('.css-class')` | CSS selectors (last resort) |

### Actions

```ts
// Typing
await page.getByLabel('Username').fill('alice')

// Clicking
await page.getByRole('button', { name: 'Login' }).click()

// Selecting
await page.getByLabel('Algorithm').selectOption('aes-gcm')
```

### Assertions

Playwright assertions also auto-wait (they retry until the condition is met or timeout):

```ts
await expect(page).toHaveURL('/dash')
await expect(page.getByText('Dashboard')).toBeVisible()
await expect(page.getByLabel('Username')).toBeEmpty()
await expect(page.getByRole('button')).toBeDisabled()
await expect(page.locator('.toast')).toHaveText(/success/i)
```

The default timeout is 5 seconds. Use `toBeVisible()` instead of `toBeInTheDocument()` — in Playwright you nearly always want visible.

---

# Walkthrough: Writing Tests for Login

This walkthrough tests the login flow against the real backend. We'll write tests that exercise the same behaviors as the original component-unit-test approach, but as real E2E tests.

## Step 1: Create the Test File

```
nacl_frontend/e2e/auth.spec.ts
```

## Step 2: Unique Test Users

Since tests hit a real database, use a unique username per run to avoid collisions:

```ts
import { test, expect } from '@playwright/test'

const TEST_USER = {
  username: `testuser_${Date.now()}`,
  user_password: 'TestPass123!',
}
```

Using `Date.now()` means every test run creates a fresh user. If you need to reuse the same user across multiple tests in a file, declare it at the top and run tests serially (which we already configured with `workers: 1`).

## Step 3: Registration as Setup

Before you can test login, you need a user. Register one first — this is both a test and test setup:

```ts
test('register a new user', async ({ page }) => {
  await page.goto('/register')

  await page.getByLabel('Username').fill(TEST_USER.username)
  await page.getByLabel('Password').fill(TEST_USER.user_password)
  await page.getByRole('button', { name: /register/i }).click()

  // Verify success
  await expect(page.getByText(/registration successful/i)).toBeVisible()

  // Should redirect to login page
  await expect(page).toHaveURL(/\/login/)
})
```

**What this teaches:**
- `page.goto()` navigates to a relative URL (resolved against `baseURL`)
- `getByLabel()` finds `<label for="...">` or `aria-labelledby` elements
- `fill()` types into an input (clearing existing content first)
- `toBeVisible()` waits for the element to be in the DOM and visible (not `display: none` or `visibility: hidden`)

## Step 4: Test the Login Page

### Test: Form renders with all expected elements

```ts
test('login page has heading, form fields, and link to register', async ({
  page,
}) => {
  await page.goto('/login')

  await expect(page.getByRole('heading', { name: /nacl/i })).toBeVisible()
  await expect(page.getByLabel(/username/i)).toBeVisible()
  await expect(page.getByLabel(/password/i)).toBeVisible()
  await expect(page.getByRole('button', { name: /login/i })).toBeVisible()
  await expect(
    page.getByRole('link', { name: /register new user/i })
  ).toBeVisible()
})
```

### Test: Validation errors on empty submit

```ts
test('login shows validation errors when submitting empty fields', async ({
  page,
}) => {
  await page.goto('/login')

  // Ensure errors are NOT visible before submit
  await expect(page.getByText(/username is required/i)).not.toBeVisible()

  // Submit empty form
  await page.getByRole('button', { name: /login/i }).click()

  // Assert errors appear
  await expect(page.getByText(/username is required/i)).toBeVisible()
  await expect(page.getByText(/password is required/i)).toBeVisible()
})
```

**Key difference from Vitest/RTL:** Playwright uses `.not.toBeVisible()` instead of `.not.toBeInTheDocument()`. In Playwright you almost always check visibility, not just DOM presence.

### Test: Wrong password shows error

```ts
test('login with wrong password shows error message', async ({ page }) => {
  await page.goto('/login')

  await page.getByLabel(/username/i).fill(TEST_USER.username)
  await page.getByLabel(/password/i).fill('WrongPassword!')
  await page.getByRole('button', { name: /login/i }).click()

  await expect(
    page.getByText(/invalid email or password/i)
  ).toBeVisible()
})
```

**Note:** This is testing against the real backend. The user must already exist (from the registration test). That's why serial execution matters — run tests in order within a file.

### Test: Successful login

```ts
test('login with valid credentials redirects to dashboard', async ({
  page,
}) => {
  await page.goto('/login')

  await page.getByLabel(/username/i).fill(TEST_USER.username)
  await page.getByLabel(/password/i).fill(TEST_USER.user_password)
  await page.getByRole('button', { name: /login/i }).click()

  await expect(page).toHaveURL(/\/dash/)
  await expect(page.getByText(/dashboard/i)).toBeVisible()
})
```

**What this teaches:**
- `toHaveURL()` asserts current page URL (supports regex or string)
- After a form redirect, Playwright automatically waits for navigation to complete
- No need to `waitFor` or `await` navigation — Playwright actions auto-wait

## Step 5: Running the Tests

First, make sure the backend is running:

```bash
# Terminal 1
cd nacl_backend && make run
```

Then run the tests:

```bash
# Terminal 2
cd nacl_frontend && npm test
```

Playwright will:
1. Start the Vite dev server (or reuse an existing one)
2. Launch Chromium in headless mode
3. Run each test
4. Generate an HTML report

To see tests visually with the Playwright UI:

```bash
npm run test:ui
```

This opens a browser UI where you can watch tests run, inspect each step's DOM snapshot, and debug failures.

---

# Patterns for Common User Flows

## Register → Login → Create Credential → Verify Vault

This is the core user journey. Write it as a serial test block:

```ts
test.describe('Full user flow', () => {
  const user = {
    username: `flowuser_${Date.now()}`,
    user_password: 'Pass123!',
  }

  const service = {
    name: 'example.com',
    username: 'bob@example.com',
    password: 'svc_pass_456',
    description: 'My email account',
  }

  test('register', async ({ page }) => { /* ... */ })
  test('login', async ({ page }) => { /* ... */ })
  test('create credential', async ({ page }) => { /* ... */ })
  test('view credential in vault', async ({ page }) => { /* ... */ })
})
```

## Testing Vault (Authenticated Pages)

For pages that require auth, either:
1. Run the login test first in the same file (serial)
2. Re-login in each test

Approach 1 is more efficient. Approach 2 is more isolated. Pick based on your tolerance for shared state.

## Testing Loading and Empty States

The Vault page shows "loading" while fetching services, then either a list or nothing. To test the loading state specifically, you'd need to slow the network — which is harder in E2E. For most backend engineers, the pragmatic choice is to **test the happy path** and **test the error path**, and not worry about intermediate states like spinner visibility. If those break, they're usually obvious in manual testing.

---

# Debugging Failed Tests

Playwright has excellent debugging tools:

```bash
# Run with visible browser (no headless)
npm test -- --headed

# Slow down each action by 100ms
npm test -- --slowmo 100

# Pause after each action for debugging
await page.pause()

# View the HTML report after a run
npx playwright show-report
```

The HTML report shows each test step with before/after DOM snapshots, console logs, network requests, and trace viewer.

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
└── e2e/
    ├── auth.spec.ts
    └── services.spec.ts
```

Tests are grouped by feature area. Each file covers one user-facing feature (auth, services, account).

---

# Writing Tests for New Features

1. Create `nacl_frontend/e2e/<feature>.spec.ts`
2. Import `{ test, expect }` from `@playwright/test`
3. Decide: standalone test or serial block?
4. Use `page.goto()` to navigate, locators to interact, and `expect().toBeVisible()` / `toHaveURL()` to assert
5. If the page requires authentication, register+login in the same block or via a `beforeEach`
6. Run with `npm test` to verify

---

# CI Integration

The CI pipeline (`.github/workflows/frontend-ci.yml`) currently runs lint → type-check → test → build. We'll need to update the test step to:
- Start the backend (e.g., with a PostgreSQL service container)
- Run Playwright tests
- Upload the Playwright report as an artifact

For now, E2E tests run locally. CI integration is tracked as a separate task.

---

# Best Practices

## Backend
- Name test files: `*_test.go`
- Use table-driven tests for multiple cases
- Use `t.Parallel()` for independent tests
- Mock database with interfaces

## Frontend (Playwright E2E)
- **Test user journeys, not components.** Each test should reflect a real user goal (register, add a service, view vault).
- **Use unique test data** per run (`Date.now()` in usernames) to avoid collisions with the database.
- **Keep tests independent** at the file level, or explicitly serial within a file.
- **Prefer `getByRole` and `getByLabel`** over CSS selectors. They match how users interact with the page.
- **Don't over-assert.** Test what matters: did the user end up on the right page? Is the expected data visible? Skip asserting intermediate CSS classes or internal state.
- **When debug is needed** use `--headed` or `await page.pause()` instead of sprinkling `console.log` everywhere.
- **Write tests as you build features,** not after. It's faster and you'll catch bugs sooner.

---

# Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Playwright Documentation](https://playwright.dev/)
- [Playwright Locators Guide](https://playwright.dev/docs/locators)
- [Playwright Assertions](https://playwright.dev/docs/test-assertions)
