# Goose Migration Workflow

Quick reference for managing database migrations with goose.

## Quick Commands

```bash
# Create new migration
make db-new name=00003_your_migration_name

# Apply all pending migrations
make db-migrate

# Rollback last migration
make db-down

# Check migration status
make db-status

# Full database reset (WARNING: destroys all data!)
make db-reset
```

---

## Adding a New Migration

### Step 1: Create Migration File

```bash
make db-new name=00003_add_sessions_table
```

Creates: `sql/migrations/20260527123456_00003_add_sessions_table.sql`

### Step 2: Edit the Migration

```sql
-- +goose Up
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);

-- +goose Down
DROP TABLE sessions;
```

### Step 3: Apply Migration

```bash
# Apply to development database
make db-migrate

# Generate Go code with sqlc
sqlc generate
```

### Step 4: Verify

```bash
# Check migration was applied
make db-status

# Test rollback works
make db-down
make db-migrate
```

---

## Modifying Migrations

### Scenario 1: Migration NOT Yet Committed

**Safe to edit!**

```bash
# 1. Rollback the migration
make db-down

# 2. Edit the migration file
# Edit sql/migrations/XXXXX_your_migration.sql

# 3. Re-apply
make db-migrate

# 4. Regenerate code if schema changed
sqlc generate
```

### Scenario 2: Migration Already Committed/Merged

**DO NOT EDIT!** Create a new migration instead:

```bash
# Create new migration that modifies the old one
make db-new name=00004_add_username_to_users
```

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN username TEXT UNIQUE;
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- +goose Down
ALTER TABLE users DROP COLUMN username;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
```

**Why?** Other developers or production may have already applied the old migration.

---

## Deleting/Rolling Back Migrations

### Rollback Last Migration

```bash
make db-down
```

### Rollback to Specific Version

```bash
# Find the version you want to rollback to
make db-status

# Rollback to that version
goose postgres "$DATABASE_URL" -dir sql/migrations down-to 20260527004445
```

### Reset Everything (Development Only!)

```bash
# WARNING: This destroys all data!
make db-reset
```

---

## Migration File Naming

### Convention

```
TIMESTAMP_description.sql
```

**Examples:**
- `20260527004445_00001_initial_schema.sql`
- `20260527005638_00002_add_passwords_table.sql`
- `20260527123456_00003_add_sessions_table.sql`

### Sequential vs Timestamped

Your project uses **timestamped** migrations (goose default).

**Benefits:**
- Avoids merge conflicts in team environments
- Automatically ordered by creation time
- No need to manage sequence numbers

---

## Best Practices

### DO ✅

- Write both Up and Down migrations
- Test rollback before committing
- Use descriptive names
- Keep migrations small and focused
- Commit migration files to git
- Run `sqlc generate` after schema changes

### DON'T ❌

- Edit committed migrations
- Delete migration files
- Skip the Down migration
- Make multiple schema changes in one migration
- Forget to regenerate sqlc code
- Use `make db-reset` on shared databases

---

## Troubleshooting

### Migration Fails with "Relation Already Exists"

```bash
# Migration was already applied
make db-status

# If it shows as pending, there's a mismatch
# Reset dev database (safe for local dev)
make db-reset
```

### Migration Fails with Syntax Error

```bash
# Rollback and fix
make db-down

# Edit the migration file
# Fix the SQL syntax

# Re-apply
make db-migrate
```

### sqlc Generate Fails

```bash
# Ensure migrations are up to date
make db-migrate

# Check SQL syntax in queries
# Regenerate
sqlc generate
```

### Goose Command Not Found

```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest
```

---

## Environment-Specific Commands

### Development

```bash
make db-migrate  # Uses DATABASE_URL from .env
```

### Test

```bash
# Handled by CI automatically
# Or manually:
DATABASE_URL="postgresql://postgres:postgres@localhost:5433/nacl_test" \
  goose postgres "$DATABASE_URL" -dir sql/migrations up
```

### Production (Future)

```bash
# Connect to production database
goose postgres "$PROD_DATABASE_URL" -dir sql/migrations up
```

---

## Migration Examples

### Example 1: Create Table

```sql
-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- +goose Down
DROP TABLE users;
```

### Example 2: Add Column

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN username TEXT UNIQUE;

-- +goose Down
ALTER TABLE users DROP COLUMN username;
```

### Example 3: Create Enum Type

```sql
-- +goose Up
CREATE TYPE encryption_strategy AS ENUM ('aes-256-gcm', 'chacha20-poly1305');

ALTER TABLE passwords ADD COLUMN strategy encryption_strategy DEFAULT 'aes-256-gcm';

-- +goose Down
ALTER TABLE passwords DROP COLUMN strategy;
DROP TYPE encryption_strategy;
```

### Example 4: Add Foreign Key

```sql
-- +goose Up
ALTER TABLE passwords 
ADD CONSTRAINT fk_passwords_user 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE passwords DROP CONSTRAINT fk_passwords_user;
```

---

## Version History

| Date | Changes |
|------|---------|
| 2026-05-26 | Initial migration workflow documentation |

---

## Related Documentation

- [Development Environment Setup](./dev_env_config.md)
- [CI/CD Pipeline](./ci_cd.md)
