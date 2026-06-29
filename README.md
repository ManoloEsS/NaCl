# NaCl

[![Backend CI](https://github.com/ManoloEsS/NaCl/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/ManoloEsS/NaCl/actions/workflows/backend-ci.yml)
[![Frontend CI](https://github.com/ManoloEsS/NaCl/actions/workflows/frontend-ci.yml/badge.svg)](https://github.com/ManoloEsS/NaCl/actions/workflows/frontend-ci.yml)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white)](https://react.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://nacl-g5mw.onrender.com)

> A single-binary password manager implementing envelope encryption (AES-256-GCM over an Argon2id-derived key) with a Go HTTP backend serving an embedded React SPA.

<!-- TODO screenshot: docs/readme_drafts/hero-shot.png -->

## What is it?

NaCl is a password manager. You create an account, you store the credentials you use to log into other services (your email, your social media, anything), and NaCl keeps them encrypted so they're unreadable to anyone who gets hold of the database.

The name comes from the chemical formula for salt, and a *cryptographic salt* is what NaCl uses to protect the key that encrypts your data.

One design choice is worth flagging up front: if you ever change your login password, none of your stored credentials need to be re-encrypted. The cost of a password rotation is constant, whether you've saved five credentials or five hundred. The rest of this document explains how that works.

## Overview

NaCl is a monolithic web application: a Go HTTP backend serving a React 19 single-page application, both compiled into one self-contained binary.

The architecture is recorded as a set of Architecture Decision Records (ADRs) in [`docs/decisions.md`](../decisions.md), and the encryption design is fully documented in [`nacl_backend/docs/encryption_flow.md`](../../nacl_backend/docs/encryption_flow.md). The repository follows strict conventional commits and is exercised by two path-scoped GitHub Actions workflows.

## Table of Contents

1. [What is it?](#what-is-it)
2. [Overview](#overview)
3. [Features](#features)
4. [How it works at a glance](#how-it-works-at-a-glance)
5. [The crypto in short](#the-crypto-in-short)
6. [Envelope encryption and key rotation](#envelope-encryption-and-key-rotation)
7. [DTO contracts: Go structs and Zod schemas](#dto-contracts-go-structs-and-zod-schemas)
8. [Tech stack](#tech-stack)
9. [Architecture](#architecture)
10. [Quickstart](#quickstart)
11. [Development](#development)
12. [CI / CD](#ci--cd)
13. [Disclaimers](#disclaimers)
14. [Roadmap](#roadmap)
15. [Further reading](#further-reading)
16. [Author & License](#author--license)

## Features

- **Account management**: register, log in, and rotate the login password.
- **Credential vault**: store service credentials with an optional description and encryption algorithm tag.
- **Re-authentication on access**: every encryption or decryption operation requires the user to re-enter the login password.
- **Atomic key rotation**: changing the login password re-encrypts only the master key; credential rows are untouched.
- **Audit log**: create, update, and delete operations are recorded per user and listed chronologically.
- **Single-binary deployment**: frontend and backend ship as one Go binary in an `alpine` runtime image.

## How it works at a glance

You register an account with a login password. Behind the scenes NaCl:

- Generates a 32-byte **salt** and a 32-byte random **master key**.
- Derives a wrapping key from your password + salt with **Argon2id**.
- Encrypts the master key with that wrapping key using **AES-256-GCM**.
- Stores the Argon2id password hash, salt, and encrypted master key, never your password.

When you save a credential (say, your Gmail password), NaCl:

- Asks you to re-enter your **login password** (re-authentication).
- Decrypts the master key in memory.
- Encrypts the service username *and* password with the master key.
- Writes the ciphertext to the database and logs the operation to the audit table.

When you change your login password, only the master key is re-encrypted; your stored credentials stay exactly as they are.

## The crypto in short

Two keys, not one:

- A random **master key** encrypts *your credentials*.
- Your login password derives a **wrapping key** (via Argon2id) that encrypts *only the master key*.

```mermaid
flowchart LR
    PW["Login password"] -->|Argon2id + salt| DK["Wrapping key"]
    MK["Random master key"] -->|AES-256-GCM| EMK["stored encrypted"]
    DK -->|encrypts| EMK
    MK -->|AES-256-GCM| Cred["All your credentials"]
```

So when you change your login password, you only re-encrypt the master key: `O(1)`, not `O(vault size)`. The full design, including the registration, decryption, and rotation flows in detail, is in [`nacl_backend/docs/encryption_flow.md`](../../nacl_backend/docs/encryption_flow.md).

## Envelope encryption and key rotation

The two-key design earns its keep on password rotation. With a naive scheme that derives the encryption key directly from the login password, rotating your password means re-encrypting every credential row. NaCl avoids that: the master key does the bulk encryption work, and the login password only ever wraps the master key.

### Registration

- Generate a 32-byte salt and a 32-byte random master key.
- Derive a wrapping key from the password + salt with Argon2id.
- Encrypt the master key with the wrapping key using AES-256-GCM.
- Store the Argon2id password hash, salt, and encrypted master key in the `users` table.

### Decryption (per operation)

- Verify the login password against the Argon2id hash.
- Derive the wrapping key from the password + stored salt.
- Decrypt the master key in memory.
- Decrypt the credential's service username and password with the master key.

### Rotation

- Verify the old login password.
- Decrypt the master key with the old wrapping key.
- Derive a new wrapping key from the new password.
- Re-encrypt the master key with the new wrapping key.
- Update the `password_hash` and `encrypted_master_key` atomically in one transaction.

Credential rows are never read or written during rotation. The cost of a rotation is independent of how many credentials you have stored.

Full design document and rationale: [`nacl_backend/docs/encryption_flow.md`](../../nacl_backend/docs/encryption_flow.md). ADRs covering crypto choices: [`docs/decisions.md`](../decisions.md).

## DTO contracts: Go structs and Zod schemas

The Go struct in [`nacl_backend/internal/dto/dto.go`](../../nacl_backend/internal/dto/dto.go) is the single source of truth for request and response shapes. The frontend mirrors those shapes with Zod schemas in [`nacl_frontend/src/lib/`](../../nacl_frontend/src/lib/): one set for requests, one for responses.

The crosswalk between the Go DTOs and the Zod schemas is documented in [`docs/dto-schema-map.md`](../dto-schema-map.md). If the two sides drift, the Zod schemas' `.strict()` mode rejects unknown keys at runtime, so a mismatch surfaces as a failing request rather than silent acceptance of an unexpected shape.

The same DTO types feed the generic `DecodeAndValidate[T]` helper used by every handler: the request is decoded into a typed struct and validated in one step, before the service layer is called. That keeps the handler layer doing only HTTP work, with no parsing or validation logic scattered across it.

Where to look:

- [`nacl_backend/internal/dto/dto.go`](../../nacl_backend/internal/dto/dto.go): request and response structs, the `Validator` interface, and `DecodeAndValidate[T]`.
- [`nacl_frontend/src/lib/requestValidation.ts`](../../nacl_frontend/src/lib/requestValidation.ts) and [`responseValidation.ts`](../../nacl_frontend/src/lib/responseValidation.ts): the Zod mirror, including `.strict()` and date coercion via `z.string().pipe(z.coerce.date())`.

For a full field-by-field mapping between the Go and TypeScript sides, see [`docs/dto-schema-map.md`](../dto-schema-map.md).

## Tech stack

### Backend

Written in **Go**. HTTP routing via `chi`, PostgreSQL access via `pgx` with `sqlc` query generation, migrations via `goose`, JWT auth (`HS256`), and Argon2id password hashing. AES-256-GCM is used for symmetric encryption of credentials and the master key.

### Frontend

Written in **TypeScript** on **React 19**, built with Vite. Forms via `react-hook-form`, runtime validation via `zod` (mirroring the backend DTOs), HTTP via `axios`. Plain CSS for styling.

### Infrastructure

- **PostgreSQL** for persistence.
- **Docker** for the runtime.
- Migrations run on application startup; no manual SQL step is needed.

## Architecture

```mermaid
flowchart LR
    Browser["Browser\nReact SPA\n(Vite dev → proxies /api to backend)"] -->|HTTP /api/*| Router["Router"]
    Router --> MW["Middleware\nlogging → recovery → JWT validation"]
    MW --> Handler["Handler\n1. decode + validate DTO\n2. call service\n3. map errors to HTTP"]
    Handler --> Service["Service\nbusiness logic + crypto orchestration"]
    Service --> Service layer
    Router --> Static["Embedded static FS\nSPA fallback"]
    Static --> Browser
```

Three properties worth pointing out in this diagram:

1. **The handler layer is pure HTTP.** Neither the database nor the encryption layer appears between the handler and the service. Handlers do only DTO decode + validate, call the service, and map errors to HTTP status codes.
2. **The embedded filesystem serves the SPA from the same binary.** Client-side routes resolve on direct navigation or refresh, not only on client-side clicks.
3. **The service is the crypto orchestrator.** Every operation that touches plaintext lives in the service, so the handler layer cannot accidentally leak plaintext or call crypto directly.

The layered design enforces separation of concerns by dependency direction. Each layer depends only on the one below it and exposes a narrow interface to the one above: the router dispatches to handlers, handlers speak only HTTP and DTOs, the service owns business logic and crypto orchestration, and the data layer is consumed by the service alone. Errors defined in the service are caught at the handler boundary and mapped to HTTP status codes, so transport concerns never leak into business logic and vice versa.

Architectural decisions are recorded as ADRs in [`docs/decisions.md`](../decisions.md).

## Quickstart

### Option A: Docker (recommended)

```bash
docker build -t nacl .
docker run -p 3333:3333 \
  -e DATABASE_URL="postgresql://postgres:postgres@host.docker.internal:5432/nacl_dev?sslmode=disable" \
  -e JWT_SECRET="your-secret" \
  -e SALT_PORT=3333 \
  nacl
```

The container ships the frontend embedded in the Go binary; no separate web server is required.

### Option B: Full local dev environment

Requires Go, Node, Docker, and the tools `goose`, `sqlc`, and `air` installed via `go install ...@latest`.

```bash
# From nacl_backend/: starts Postgres, applies migrations, hot-reloads Go
make dev-full
```

```bash
# From nacl_frontend/: Vite dev server on :5173, proxies /api to :3333
npm install
npm run dev
```

See [`nacl_backend/docs/dev_env_config.md`](../../nacl_backend/docs/dev_env_config.md) for the full prerequisites list.

### Option C: Manual build

```bash
# Build frontend into nacl_backend/static
cd nacl_frontend && npm ci && npm run build

# Build the Go binary (embeds static/)
cd ../nacl_backend && go build -o nacl .
./nacl
```

## Development

### Environment variables

| Variable | Required | Description |
| --- | --- | --- |
| `DATABASE_URL` | yes | PostgreSQL connection string for the dev database. |
| `DATABASE_URL_TEST` | tests | PostgreSQL connection string for the test database. |
| `JWT_SECRET` | yes (random fallback) | HS256 signing key. A random value is generated if missing. |
| `SALT_PORT` | no | Listen port. Falls back to `PORT`, then a platform default. |
| `PORT` | no | Alternate listen port (Render compatibility). |
| `SALT_LOG_FILE` | no | Path to the request log file. |
| `DB_SSL` | no | Enables SSL on the database connection. |

A working `.env` template is `include`d and exported by the Makefile at the start of every invocation.

### Backend (`nacl_backend/`)

| Target | Effect |
| --- | --- |
| `make dev` | Hot-reload via `air`. |
| `make dev-full` | Start Postgres, init databases, migrate, then `make dev`. |
| `make db-start` `db-stop` `db-restart` `db-reset` | Postgres container lifecycle. |
| `make db-migrate` `db-down` `db-status` | Goose migrations. |
| `make db-new name=add_users` | Create a new timestamped migration. |
| `make db-test-clean` | Stand up an isolated Postgres on port 5433 for tests. |
| `make build` `run` `clean` | Compile, run, or remove the binary. |
| `make test` | Reset test DB and run `go test ./...` through `tparse`. |
| `make lint` `fmt` `fmt-check` `ci` | Mirror the CI checks. |

### Frontend (`nacl_frontend/`)

| Script | Effect |
| --- | --- |
| `npm run dev` | Vite dev server. |
| `npm run build` | `tsc -b && vite build` into `../nacl_backend/static`. |
| `npm run lint` | ESLint. |
| `npm run format` | Prettier. |
| `npm test` / `npm run test:ui` | Playwright E2E. |

### API reference

For the full list of API endpoints with request and response shapes, see [`docs/dto-schema-map.md`](../dto-schema-map.md).

## CI / CD

Two workflows under [`.github/workflows/`](../../.github/workflows/):

- **`backend-ci.yml`** runs on PRs to `main` touching `nacl_backend/**`. It lints with `golangci-lint`, then spins up a PostgreSQL service container, runs migrations and `sqlc generate`, builds, and runs `go test -race ./...`.
- **`frontend-ci.yml`** runs on push and PR to `main` touching `nacl_frontend/**`. It runs ESLint and `tsc --noEmit`, then Playwright, then a production build.

Both workflows are path-scoped, so a frontend-only change does not trigger the backend pipeline and vice versa.

See [`docs/ci_cd.md`](../ci_cd.md).

## Disclaimers

This is a learning project and is explicitly not production-ready. Please do not store real credentials in it.

### Intentional MVP exclusions

The following are documented as out of scope for this MVP and recorded in [`docs/decisions.md`](../decisions.md) and [`docs/user_stories.md`](../user_stories.md):

- No rate limiting on authentication endpoints.
- No second-factor authentication.
- No token blacklist; logout is client-side only.
- No email verification or account lockout.
- No CAPTCHA.

### Accepted trade-offs

- **Server-side encryption.** The master key exists in backend memory for the duration of each operation, so NaCl does not satisfy the strict zero-knowledge threat model of client-side encryption schemes. A compromised server can exfiltrate keys. This is acknowledged openly in [`nacl_backend/docs/encryption_flow.md`](../../nacl_backend/docs/encryption_flow.md).
- **No ciphertext padding.** AES-GCM ciphertext length leaks plaintext length. Acceptable for short password fields and consistent with industry practice, but documented as such.

## Roadmap

Things we plan to work on in the future:

- 2FA / WebAuthn
- Server-side token blacklisting
- Rate limiting and account lockout
- Email verification and password reset
- True client-side encryption for a zero-knowledge posture
- Additional encryption algorithms to choose from

## Further reading

- [`docs/decisions.md`](../decisions.md) - the ADRs explaining why the architecture is the way it is.
- [`docs/user_stories.md`](../user_stories.md) - MVP scope and the explicit list of post-MVP features.
- [`nacl_backend/docs/encryption_flow.md`](../../nacl_backend/docs/encryption_flow.md) - the full crypto design with diagrams.
- [`docs/dto-schema-map.md`](../dto-schema-map.md) - how Go DTOs and Zod schemas stay in sync.
- [`nacl_backend/docs/dev_env_config.md`](../../nacl_backend/docs/dev_env_config.md) - local prerequisites and tool installs.

The remaining files under `docs/` and `nacl_backend/docs/` are bonus material on specific topics (serving React from Go, error handling, migrations, frontend setup, CSS patterns).

## Author & License

Built by **Manolo Estrada** ([@ManoloEsS](https://github.com/ManoloEsS)).

Licensed under the [MIT License](../../LICENSE).
