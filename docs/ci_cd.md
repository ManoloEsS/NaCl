# CI/CD Pipeline Configuration

This document describes the CI/CD pipeline setup for the NaCl project.

## Overview

The project uses GitHub Actions for continuous integration and deployment. The pipeline is split into frontend and backend workflows, with the backend workflow tailored for Go.

## Workflow Files

### Backend CI (`.github/workflows/backend-ci.yml`)

**Triggers:**
- Push to `main` branch (nacl_backend changes only)
- Pull requests to `main` branch (nacl_backend changes only)

**Jobs:**

#### 1. Lint
Runs Go code quality checks:
- `gofmt` - Check code formatting
- `go vet` - Static analysis for suspicious constructs
- `staticcheck` - Advanced static analysis

**Requirements:**
- Go 1.26
- Dependencies installed via `go mod download`

#### 2. Type Check
Verifies the code compiles correctly:
- Builds all packages with `go build -v ./...`
- Catches compilation errors early

#### 3. Test
Runs the test suite with PostgreSQL:

**Services:**
- PostgreSQL 18 container
  - Database: `nacl_test`
  - Port: 5432
  - Health check enabled

**Steps:**
1. Install goose (migrations)
2. Install sqlc (code generation)
3. Run database migrations
4. Generate Go code from SQL
5. Run tests with race detector
6. Upload coverage to Codecov

**Environment Variables:**
```yaml
DATABASE_URL: postgresql://postgres:postgres@localhost:5432/nacl_test?sslmode=disable
DATABASE_URL_TEST: postgresql://postgres:postgres@localhost:5432/nacl_test?sslmode=disable
SALT_LOG_FILE: salt.test.log
SALT_PORT: 3334
DB_SSL: false
JWT_SECRET: test-secret-key-ci
```

#### 4. Build
Creates production binary:
- Builds Go binary for Linux
- Uploads as artifact (7-day retention)

**Output:** `bin/nacl_backend`

---

## Local Testing

### Run CI Steps Locally

```bash
# Lint
cd nacl_backend
gofmt -d .
go vet ./...
staticcheck ./...

# Type check
go build -v ./...

# Test (requires PostgreSQL)
make db-start
make db-migrate
sqlc generate
go test -v -race ./...
```

### Simulate CI Environment

```bash
# Set CI environment variables
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/nacl_test?sslmode=disable"
export JWT_SECRET="test-secret-key-ci"

# Run full test suite
cd nacl_backend
make db-test-clean
go test -v -race -coverprofile=coverage.out ./...
```

---

## Adding New Checks

### Add New Linter

```yaml
- name: Install <linter>
  run: go install <linter-path>@latest

- name: Run <linter>
  run: |
    cd nacl_backend
    <linter-command> ./...
```

### Add Integration Tests

```yaml
- name: Run integration tests
  run: |
    cd nacl_backend
    go test -v -tags=integration ./...
```

### Add Multi-Platform Build

```yaml
- name: Build for multiple platforms
  run: |
    cd nacl_backend
    GOOS=linux GOARCH=amd64 go build -o bin/nacl_backend-linux
    GOOS=darwin GOARCH=amd64 go build -o bin/nacl_backend-macos
    GOOS=windows GOARCH=amd64 go build -o bin/nacl_backend-windows.exe
```

---

## Deployment Workflow (Future)

### Staging Deployment

```yaml
name: Deploy to Staging

on:
  push:
    branches: [main]

jobs:
  deploy-staging:
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4
      - name: Deploy to staging server
        run: |
          # SSH and deploy
          ssh user@staging.example.com "cd /app && ./deploy.sh"
```

### Production Deployment

```yaml
name: Deploy to Production

on:
  release:
    types: [published]

jobs:
  deploy-production:
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - name: Build and deploy
        run: |
          # Build and deploy
```

---

## Environment Configuration

### GitHub Secrets Required

| Secret | Description | Required For |
|--------|-------------|--------------|
| `CODECOV_TOKEN` | Codecov upload token | Test job (optional) |
| `DEPLOY_KEY` | SSH key for deployment | Deploy jobs |
| `PROD_DB_URL` | Production database URL | Production deploy |

### GitHub Environments

Configure in Repository Settings → Environments:

**staging:**
- Required reviewers: (optional)
- Deployment branches: `main`
- Environment variables: `STAGING_URL`, `STAGING_DB_URL`

**production:**
- Required reviewers: 1
- Deployment branches: `main`
- Environment variables: `PROD_URL`, `PROD_DB_URL`

---

## Troubleshooting CI Failures

### Go Module Cache Issues

```yaml
- name: Clear Go cache
  run: |
    go clean -cache -modcache
    go mod download
```

### PostgreSQL Connection Failures

```yaml
# Increase health check retries
services:
  postgres:
    options: >-
      --health-cmd "pg_isready -U postgres"
      --health-interval 10s
      --health-timeout 5s
      --health-retries 10  # Increase from 5
```

### Migration Failures

```yaml
- name: Debug migrations
  run: |
    cd nacl_backend
    goose -dir sql/migrations postgres "$DATABASE_URL" status
    goose -dir sql/migrations postgres "$DATABASE_URL" up -v
```

### Test Timeouts

```yaml
- name: Run tests
  run: |
    cd nacl_backend
    go test -v -race -timeout=5m ./...
```

---

## Performance Optimization

### Cache Dependencies

```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

### Parallel Jobs

Jobs run in parallel when possible:
- `lint` and `type-check` run simultaneously
- `test` waits for both
- `build` waits for `test`

### Selective Path Triggers

```yaml
on:
  push:
    paths:
      - 'nacl_backend/**'  # Only run on backend changes
```

---

## Coverage Reports

### View Coverage

After CI runs, coverage reports are uploaded to Codecov:
- Visit: `https://app.codecov.io/gh/ManoloEsS/NaCl`
- Compare coverage between commits
- View line-by-line coverage

### Coverage Thresholds (Future)

```yaml
- name: Check coverage threshold
  run: |
    cd nacl_backend
    go tool cover -func=coverage.out | grep total | awk '{if ($3 < 80.0) exit 1}'
```

---

## Version History

| Date | Version | Changes |
|------|---------|---------|
| 2026-05-26 | 1.0.0 | Initial Go CI pipeline |

---

## Related Documentation

- [Development Environment Setup](./dev_env_config.md)
- [Testing Guide](./testing.md) (TODO)
- [Deployment Guide](./deployment.md) (TODO)
