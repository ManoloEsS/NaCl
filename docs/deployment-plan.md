# Deployment Plan: NaCl to Render

## Project Analysis — Key Findings

| Finding | Detail |
|---------|--------|
| Config needs adjustment | `NewConfig()` calls `godotenv.Load()` which returns an error if no `.env` — will crash on Render |
| Port mismatch | App reads `SALT_PORT`; Render provides `PORT` (default 10000) |
| Log file | Logger opens a file from `SALT_LOG_FILE` — Render filesystem is ephemeral per deploy, and the file path dir may not exist |
| `dist/` is gitignored | Blanket `dist` pattern (line 128) ignores `nacl_backend/dist/` |
| BrowserRouter | React Router v7 with path-based routing needs SPA fallback (`index.html` for non-API routes) |
| Frontend build output | Already configured to `../nacl_backend/dist/` (from `vite.config.ts`) |

---

## Step 1: Make Config Resilient for Render

**What**: Update `nacl_backend/internal/config/config.go` to handle the Render environment.

**Changes**:
- Make `godotenv.Load()` non-fatal (`_ = godotenv.Load()`)
- Add fallback from `SALT_PORT` to `PORT` env var when `SALT_PORT` is empty
- Validate `DATABASE_URL` is non-empty and exit with error if missing

**Why**: Render doesn't use a `.env` file — environment variables are set directly via the dashboard. The app will crash on startup otherwise. Render also expects services to bind to the port specified by the `PORT` env var (default 10000).

**References**:
- [Render docs: Environment Variables](https://render.com/docs/configure-environment-variables)
- [Render docs: Web Services — Port Binding](https://render.com/docs/web-services#port-binding)
- [Go `os.Getenv`](https://pkg.go.dev/os#Getenv)
- [godotenv](https://github.com/joho/godotenv)

---

## Step 2: Handle Log File on Render

**What**: Modify the logger to gracefully fall back to stderr when no writable log file is available.

**Approach**:
- When `os.OpenFile(logFile, ...)` fails, fall back to `os.Stderr` as the output destination
- When writing directly to stderr, use unbuffered writes (no `bufio.Writer`) to avoid data loss on crash
- When writing to stderr, use a single text handler at `LevelInfo` (no debug noise in production)
- When writing to a file, use a JSON handler at `LevelInfo` (buffered for performance) + text debug handler to stderr
- `closeLogger` only flushes and closes when writing to a real file, not stderr

**Why**: Render captures stdout/stderr as the service's log stream. Writing logs to a file means they won't appear in the Render dashboard logs and are lost on redeploy. Buffered stderr risks losing log lines on crash.

**Reference**: [Render docs: Logging](https://render.com/docs/logging)

---

## Step 3: Embed the Frontend Build in Go

**What**: Add the `//go:embed` directive in the Go backend to embed the `dist/` directory at compile time.

**Where**: Create or update a file in `nacl_backend/` (e.g., a new file in `internal/server/`):

```go
package server

import "embed"

//go:embed dist
var DistFS embed.FS
```

**Why**: `//go:embed` bakes the static files into the Go binary at compile time. This means the binary is self-contained — no need to ship the `dist/` directory alongside it on Render.

**References**:
- [Go `embed` package docs](https://pkg.go.dev/embed)
- [Go blog: Go 1.16 Embedding](https://go.dev/blog/go1.16)

---

## Step 4: Set Up Chi Routes with SPA Fallback

**What**: Create a file server handler for the embedded static files and a catch-all route that serves `index.html` for client-side routing.

**Where**: Add to `nacl_backend/internal/server/server.go` (or a dedicated file):

- Serve `/assets/*` directly from the embedded FS (Vite builds assets under `assets/`)
- For any other non-API path, serve `index.html` so React Router can handle client-side routing

**Why**: React Router v7 uses `BrowserRouter` (path-based, not hash-based). If a user navigates to `/vault` or `/dash` directly, the server receives a request for that path. Without the SPA fallback, those requests would 404 because those paths don't match API routes or static assets.

**References**:
- [Chi file server example](https://github.com/go-chi/chi/blob/master/_examples/fileserver/main.go)
- [`net/http.FileServer` with `embed.FS`](https://pkg.go.dev/net/http#FileServer)
- [Chi Router `NotFound`](https://pkg.go.dev/github.com/go-chi/chi/v5#Router)

---

## Step 5: Fix `.gitignore`

**What**: Adjust the blanket `dist` rule so `nacl_backend/dist/` can be tracked (or built during deploy).

**Options**:
- **(Recommended for simplicity)** Add `!/nacl_backend/dist` before the blanket `dist` rule to re-include backend's dist
- **(Better for build pipeline)** Remove the blanket `dist` rule entirely; add targeted ignores for `nacl_frontend/dist/` and `*/node_modules/`

**Why**: If you use `//go:embed`, the `dist/` directory must exist at compile time. You either need to commit it, or build the frontend before Go compilation during the Render deploy. The `.gitignore` currently prevents both.

**Reference**: [Git docs: .gitignore patterns](https://git-scm.com/docs/gitignore) — Pattern negation with `!`

---

## Step 6: Choose a Build Strategy

This is the most important architectural decision. Here are three options:

### Option A: Dockerfile (Recommended)

Create a `Dockerfile` at the repo root using multi-stage build:

```
Stage 1: Node image → install deps, build frontend
Stage 2: Go image → copy built `dist/`, compile embedded binary
Stage 3: Distroless or minimal image → run the binary
```

**Render Setup**: Set runtime to `Docker` (not native Go).

**Why**: Reliable, reproducible, full control over the build environment. No guessing about whether Node.js is available in Render's Go build environment. Single `Dockerfile` describes the entire pipeline.

**References**:
- [Render docs: Docker Services](https://render.com/docs/docker)
- [Docker multi-stage build docs](https://docs.docker.com/build/building/multi-stage/)
- [Render Blueprint: Docker fields](https://render.com/docs/blueprint-spec#docker)

### Option B: Build Script + Native Go Runtime

Create a `build.sh` script at the repo root or in `nacl_backend/`:

```bash
npm --prefix ../nacl_frontend ci
npm --prefix ../nacl_frontend run build
cd ../nacl_backend && go build -o app .
```

**Render Setup**:
- Runtime: Go
- Root directory: `nacl_backend`
- Build command: `bash build.sh`
- Start command: `./app`

**Why**: Uses Render's native Go runtime (potentially simpler), no Docker overhead.

**Risk**: Render's Go build environment may not include Node.js. If it doesn't, the script will fail on the `npm` step.

**References**:
- [Render docs: Monorepo Support](https://render.com/docs/monorepo-support)
- [Render docs: Deploy — Build Command](https://render.com/docs/deploys#build-command)

### Option C: Separate Services (Simplest Deploy)

- Frontend as a **Static Site** on Render (pointed at `nacl_frontend/`, build: `npm ci && npm run build`, publish dir: `dist`)
- Backend as a **Web Service** on Render (Go runtime, pointed at `nacl_backend/`)
- No embedding needed; frontend calls backend API via fetch (with proxy or CORS)

**Why**: No embedding complexity, no build pipeline issues. Each service uses its native runtime. But requires configuring CORS on the backend and a proxy on the frontend.

**Reference**: [Render docs: Static Sites](https://render.com/docs/static-sites)

---

## Step 7: Set Up Render PostgreSQL

**What**: Create a Render Postgres database through the dashboard or Blueprint.

**Notes**:
- Choose a plan that fits your budget (free tier available)
- The database connection string will be provided as the `DATABASE_URL` environment variable
- Render Postgres is private-network accessible to your web service in the same region

**Why**: The app needs PostgreSQL (uses pgx + pgxpool). Render manages the database, including backups and scaling.

**References**:
- [Render docs: PostgreSQL](https://render.com/docs/postgresql)
- [Render Blueprint: Database fields](https://render.com/docs/blueprint-spec#database-fields)

---

## Step 8: Configure Environment Variables on Render

| Variable | Value / Source | Why |
|----------|---------------|-----|
| `DATABASE_URL` | From Render Postgres (auto-provided) | Database connection |
| `JWT_SECRET` | Generated secret (use Render's `generateValue`) | JWT signing key |
| `SALT_PORT` | `$PORT` (Render's built-in PORT var) | Maps to Render's default port 10000 |
| `SALT_LOG_FILE` | `salt.access.log` (writes to `/tmp/` dir) | Log file path; create it if not present |

**Why**: These four variables are required by the config. `JWT_SECRET` should be a strong random value — Render can generate one. `SALT_PORT` maps to Render's `PORT`. `SALT_LOG_FILE` specifies the log destination.

**References**:
- [Render docs: Configuring Environment Variables](https://render.com/docs/configure-environment-variables)
- [Render Blueprint: Environment Variables](https://render.com/docs/blueprint-spec#setting-environment-variables)
- [Render docs: Default Environment Variables](https://render.com/docs/environment-variables)

---

## Step 9: Deploy and Verify

**What**: Trigger the first deploy and verify the app works.

**Checklist**:
1. Initial deploy completes without errors
2. Visit the `onrender.com` URL — should see the React app
3. Navigate through React Router routes (e.g., `/login`, `/register`, `/dash`)
4. Create an account (hits `/api/users`)
5. Log in (hits `/api/login`)
6. Create a credential (hits `/api/credentials`)
7. Check logs in Render Dashboard for any errors

**Why**: End-to-end validation ensures both the static file serving and API routing work correctly, and that the SPA fallback handles all client-side routes.

**References**:
- [Render docs: Deploys](https://render.com/docs/deploys)
- [Render docs: Logging](https://render.com/docs/logging)
- [Render docs: Health Checks](https://render.com/docs/health-checks)

---

## Summary of Code Changes Required Before Deploy

| File | What to Change | Priority |
|------|---------------|----------|
| `nacl_backend/internal/config/config.go` | Make `godotenv.Load()` non-fatal; add `PORT` fallback; require `DATABASE_URL` | **Required** |
| `nacl_backend/internal/logger/logger.go` | Fall back to stderr on file error; unbuffered stderr; skip close on stderr | **Required** |
| `nacl_backend/internal/server/` (new file) | Add `//go:embed dist` directive | **Required** |
| `nacl_backend/internal/server/server.go` | Add embedded FS routes with SPA fallback | **Required** |
| `.gitignore` | Allow `nacl_backend/dist/` to be tracked or built | **Required** |
| Project root | Create `Dockerfile` or `build.sh` (depending on chosen approach) | **Required** |
| `nacl_backend/internal/server/index_handler.go` | Update or remove the plain-text handler (replaced by SPA serving) | Optional |
| `nacl_backend/go.mod` / `go.sum` | No changes needed (Chi v5 already included) | — |
