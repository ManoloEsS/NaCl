# Development Environment Configuration

This document outlines all decisions, configurations, and setup steps for the NaCl backend development environment.

## Table of Contents

- [Overview](#overview)
- [Project Structure](#project-structure)
- [Tools & Versions](#tools--versions)
- [Configuration Files](#configuration-files)
- [Database Setup](#database-setup)
- [Makefile Commands](#makefile-commands)
- [Development Workflow](#development-workflow)
- [Troubleshooting](#troubleshooting)

---

## Overview

The NaCl backend is a Go-based REST API for a password management application. It provides endpoints for user authentication, password encryption/decryption, and secure storage using PostgreSQL.

### Key Technologies

- **Language**: Go 1.26.1
- **Database**: PostgreSQL 18
- **Query Generation**: sqlc v1.31.1
- **Migrations**: goose v3.27.1
- **HTTP Router**: chi/v5
- **Database Driver**: pgx/v5
- **Hot Reload**: air
- **Containerization**: Docker

---

## Project Structure

```
nacl_backend/
├── main.go                 # Application entry point
├── Makefile                # Development commands
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── .env                    # Development environment variables
├── .env.test              # Test environment variables
├── .air.toml              # Hot reload configuration
├── goose.yml              # Goose migration config
├── sqlc.yml               # sqlc codegen config
├── docs/                  # Documentation
├── server/                # HTTP server code
│   ├── server.go
│   └── handlers/
├── config/                # Configuration loading
├── sql/                   # SQL files
│   ├── migrations/        # Database migrations (goose)
│   └── queries/           # Query annotations (sqlc)
└── internal/              # Private application code
    ├── apperr/            # Error handling
    ├── auth/              # Authentication logic
    ├── crypto/            # Encryption strategies
    └── db/                # sqlc generated code
```

---

## Tools & Versions

### Required Tools

| Tool | Version | Installation |
|------|---------|--------------|
| Go | 1.26.1+ | `mise install go@1.26.1` or `brew install go@1.26` |
| Docker | Latest | `docker --version` |
| goose | v3.27.1+ | `go install github.com/pressly/goose/v3/cmd/goose@latest` |
| sqlc | v1.31.1+ | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| air | Latest | `go install github.com/air-verse/air@latest` |

### Verify Installation

```bash
go version
goose --version
sqlc version
air --version
docker --version
```

---

## Configuration Files

### `.env` (Development)

**Location**: Project root

**Purpose**: Development environment variables

**Contents**:
```env
SALT_LOG_FILE=salt.access.log
SALT_PORT=3333
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/nacl_dev?sslmode=disable
DATABASE_URL_TEST=postgresql://postgres:postgres@localhost:5433/nacl_test?sslmode=disable
DB_SSL=false
JWT_SECRET=hashandeggs
```

**Decisions**:
- No quotes around values (prevents parsing issues in Makefile)
- `sslmode=disable` in connection string (local development only)
- Separate `DATABASE_URL` and `DATABASE_URL_TEST` for dev/test isolation
- Simple JWT secret for development (change in production)

---

### `.env.test` (Testing)

**Location**: Project root

**Purpose**: Test environment variables

**Contents**:
```env
SALT_LOG_FILE=salt.test.log
SALT_PORT=3334
DATABASE_URL=postgresql://postgres:postgres@localhost:5433/nacl_test?sslmode=disable
DB_SSL=false
JWT_SECRET=test-secret-key
```

**Decisions**:
- Separate log file to avoid conflicts with dev logs
- Different port (3334) to avoid conflicts with dev server
- Test database on port 5433 (dev uses 5432)
- Isolated test database (`nacl_test`)

---

### `.air.toml` (Hot Reload)

**Location**: Project root

**Purpose**: Configure air for hot reload during development

**Contents**:
```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./..."
bin = "./tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "vendor", "sql"]
```

**Decisions**:
- Build output to `tmp/` directory (excluded from git)
- Exclude `sql/` directory (no need to rebuild on SQL changes)
- Only watch `.go` files
- Simple build command for fast reloads

---

### `goose.yml` (Migrations)

**Location**: Project root

**Purpose**: Configure goose database migrations

**Contents**:
```yaml
dialect: postgres
dir: sql/migrations
```

**Decisions**:
- Simple YAML format (no complex config needed)
- Points to `sql/migrations/` (shared with sqlc)
- Uses PostgreSQL dialect
- No additional options required

---

### `sqlc.yml` (Code Generation)

**Location**: Project root

**Purpose**: Configure sqlc for generating Go code from SQL

**Contents**:
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "sql/queries"
    schema: "sql/migrations"
    gen:
      go:
        package: "db"
        out: "internal/db"
        emit_json_tags: true
        emit_interface: true
        emit_empty_slices: true
        sql_package: "pgx/v5"
```

**Decisions**:

| Option | Value | Reason |
|--------|-------|--------|
| `schema` | `sql/migrations` | Share migration files with goose (single source of truth) |
| `package` | `db` | Conventional naming for database package |
| `out` | `internal/db` | Private generated code |
| `emit_json_tags` | `true` | Enable JSON serialization for API responses |
| `emit_interface` | `true` | Enable mocking for tests |
| `emit_empty_slices` | `true` | Return empty slices instead of nil (consistent behavior) |
| `sql_package` | `pgx/v5` | Use native pgx instead of `database/sql` |

---

## Database Setup

### PostgreSQL Version

**Decision**: PostgreSQL 18

**Rationale**:
- Latest stable version
- Improved performance
- Better type support
- Long-term support

### Docker Configuration

**Container Name**: `nacl_postgres` (dev), `nacl_postgres_test` (test)

**Volume Mount**: `/var/lib/postgresql` (PostgreSQL 18+ convention)

**Ports**:
- Dev: `5432:5432`
- Test: `5433:5432`

**Environment**:
```bash
POSTGRES_PASSWORD=postgres
```

### Databases

| Database | Purpose | Port | Connection String |
|----------|---------|------|-------------------|
| `nacl_dev` | Development | 5432 | `postgresql://postgres:postgres@localhost:5432/nacl_dev` |
| `nacl_test` | Testing | 5433 | `postgresql://postgres:postgres@localhost:5433/nacl_test` |

**Decision**: Separate databases for dev and test to avoid data contamination.

### Migration Strategy

**Tool**: goose

**Directory**: `sql/migrations/`

**Naming Convention**: Timestamped (default)

**Example**:
```
sql/migrations/
├── 20260527004445_00001_initial_schema.sql
├── 20260527005638_00002_add_passwords_table.sql
└── ...
```

**File Format**:
```sql
-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

-- +goose Down
DROP TABLE users;
```

**Decisions**:
- Timestamped migrations (avoids merge conflicts)
- Shared directory with sqlc (single source of truth)
- Explicit down migrations for rollback support
- Use `gen_random_uuid()` (PostgreSQL 13+ built-in)

---

## Makefile Commands

### Development

| Command | Description |
|---------|-------------|
| `make dev` | Run app with hot reload |
| `make dev-full` | Start DB + migrate + run dev |

### Database

| Command | Description |
|---------|-------------|
| `make db-start` | Start PostgreSQL container (auto-creates databases) |
| `make db-stop` | Stop PostgreSQL container |
| `make db-reset` | Full reset (drops volume, recreates, migrates) |
| `make db-migrate` | Run all pending migrations |
| `make db-down` | Rollback last migration |
| `make db-status` | Show migration status |
| `make db-new name=<name>` | Create new migration file |
| `make db-init` | Create databases if they don't exist |

### Testing

| Command | Description |
|---------|-------------|
| `make test` | Run tests with isolated test database |
| `make db-test-clean` | Reset test database and run migrations |

### Build

| Command | Description |
|---------|-------------|
| `make build` | Compile Go binary to `bin/nacl_backend` |
| `make run` | Build and run binary |
| `make clean` | Remove build artifacts (`bin/`, `tmp/`) |

---

## Development Workflow

### Initial Setup

```bash
# 1. Clone repository
cd nacl_backend

# 2. Install dependencies
go mod download

# 3. Start development environment
make dev-full
```

### Daily Development

```bash
# Start coding (hot reload enabled)
make dev

# Or if database needs to be started first
make dev-full
```

### Adding a Migration

```bash
# 1. Create migration file
make db-new name=00003_add_user_preferences

# 2. Edit the generated file in sql/migrations/

# 3. Run migration
make db-migrate

# 4. Generate Go code
sqlc generate
```

### Adding a Query

```bash
# 1. Add query to sql/queries/<table>.sql
# Example: sql/queries/users.sql
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

# 2. Generate Go code
sqlc generate

# 3. Use in code
import "github.com/.../nacl_backend/internal/db"
```

### Running Tests

```bash
# Run all tests (auto-creates test DB)
make test
```

### Database Reset

```bash
# Complete reset (WARNING: destroys all data)
make db-reset
```

---

## Troubleshooting

### PostgreSQL Container Won't Start

**Symptom**: `make db-start` fails

**Solution**:
```bash
# Remove old container and volume
docker rm -f nacl_postgres
docker volume rm nacl_dev_data

# Start fresh
make db-start
```

### Migration Errors

**Symptom**: `make db-migrate` fails with "relation already exists"

**Solution**:
```bash
# Check migration status
make db-status

# If migrations are out of sync, reset
make db-reset
```

### sqlc Generation Fails

**Symptom**: `sqlc generate` fails

**Common Causes**:
1. No queries in `sql/queries/`
2. Schema changes not migrated
3. Syntax errors in SQL files

**Solution**:
```bash
# Ensure migrations are up to date
make db-migrate

# Add at least one query to sql/queries/
# Then regenerate
sqlc generate
```

### Port Already in Use

**Symptom**: PostgreSQL can't bind to port 5432

**Solution**:
```bash
# Find and kill process using port
lsof -i :5432
kill -9 <PID>

# Or use different port in .env
```

### Test Database Issues

**Symptom**: Tests fail with connection errors

**Solution**:
```bash
# Clean test database
make db-test-clean

# Or full reset
docker rm -f nacl_postgres_test
docker volume rm nacl_test_data
make test
```

### air Hot Reload Not Working

**Symptom**: Changes don't trigger rebuild

**Solution**:
```bash
# Check .air.toml syntax
cat .air.toml

# Ensure tmp/ directory exists
mkdir -p tmp

# Restart air
make dev
```

---

## Environment Variables Reference

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `SALT_LOG_FILE` | Log file path | `salt.access.log` |
| `SALT_PORT` | HTTP server port | `3333` |
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://...` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_SSL` | Enable SSL for database | `false` |

---

## Security Notes

### Development vs Production

**Development**:
- `sslmode=disable` (acceptable for local dev)
- Simple passwords (`postgres`)
- Exposed ports (5432, 5433)

**Production Requirements**:
- Enable SSL (`sslmode=require`)
- Strong passwords
- Restricted network access
- Separate secrets management
- Environment-specific `.env.production` (not committed)

### Secrets Management

**Current**: `.env` files (not committed to git)

**Production Recommendations**:
- Use environment variables from CI/CD
- Secrets manager (AWS Secrets Manager, HashiCorp Vault)
- Kubernetes secrets
- Never commit `.env` files

---

## Version History

| Date | Version | Changes |
|------|---------|---------|
| 2026-05-26 | 1.0.0 | Initial dev environment setup |

---

## Contributing

When modifying dev environment configuration:

1. Update this document
2. Test all `make` commands
3. Ensure `.env` and `.env.test` are in `.gitignore`
4. Document any new required tools
5. Update troubleshooting section if applicable
