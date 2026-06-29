# Serving a React SPA from a Go Chi Server

## Table of Contents

- [Why embed the frontend in the backend?](#why-embed-the-frontend-in-the-backend)
- [Overview of the full pipeline](#overview-of-the-full-pipeline)
- [Step 1: Configure Vite to build to the Go project](#step-1-configure-vite-to-build-to-the-go-project)
- [Step 2: Embed the build output with `//go:embed`](#step-2-embed-the-build-output-with-goembed)
- [Step 3: Pass the embedded filesystem to Chi](#step-3-pass-the-embedded-filesystem-to-chi)
- [Step 4: Understand `fs.FS`, `http.FS`, and `http.FileServer`](#step-4-understand-fsfs-httpfs-and-https-fileserver)
- [Step 5: Set up Chi routes for static files](#step-5-set-up-chi-routes-for-static-files)
- [Step 6: Handle SPA client-side routing (the fallback)](#step-6-handle-spa-client-side-routing-the-fallback)
- [Step 7: Fix `.gitignore` so the build directory is available](#step-7-fix-gitignore)
- [Full request trace: what happens when a browser visits](#full-request-trace-what-happens-when-a-browser-visits)
- [Summary of key types and functions](#summary-of-key-types-and-functions)
- [Troubleshooting](#troubleshooting)

---

## Why embed the frontend in the backend?

When you deploy a web app, you have two programs that need to reach the user: the Go API server and the React frontend. There are three common deployment strategies:

| Strategy | How it works | Pros | Cons |
|----------|-------------|------|------|
| **Separate services** | React on a static host (Vercel/Netlify), Go on a server | Independent scaling, clear separation | CORS config, two URLs, two deploy pipelines |
| **Reverse proxy** | Go serves API, nginx/Caddy serves static files | Standard pattern | Extra infrastructure to manage |
| **Embedding (this guide)** | React build is compiled into the Go binary | Single binary, single deploy, no CORS | Slightly larger binary, frontend rebuild requires Go recompile |

Embedding is ideal when you want a self-contained deployment — one binary, one process, one URL.

---

## Overview of the full pipeline

```
┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│  npm run build      │     │  go build            │     │  ./app (the binary) │
│                     │     │                      │     │                     │
│  Vite produces:     │ ──► │  Go reads dist/      │ ──► │  Embedded FS        │
│  nacl_backend/      │     │  at compile time     │     │  lives in memory    │
│  └── dist/          │     │  via //go:embed      │     │                     │
│      ├── index.html │     │                      │     │  Chi routes:        │
│      └── assets/    │     │  Compiles everything │     │  /assets/* → FS     │
│          ├── main.js│     │  into one binary     │     │  /* → index.html    │
│          └── ...    │     │                      │     │  /api/* → handlers  │
└─────────────────────┘     └─────────────────────┘     └─────────────────────┘
         build time                 compile time                  runtime
```

The key insight: the React build output (`dist/`) is treated as a **compile-time resource**, not a runtime one. It gets baked into the binary so that at runtime, the Go server can serve it from memory without needing the original files on disk.

---

## Step 1: Configure Vite to build to the Go project

In your React project's `vite.config.ts`, set the output directory to a location inside your Go project. This makes the build output available for embedding.

```ts
// nacl_frontend/vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: './',    // relative paths so it works behind any proxy
  build: {
    outDir: '../nacl_backend/dist', // output into the Go project
  },
})
```

**What `base: './'` does**: Vite normally generates paths like `/assets/main.js` (absolute). Setting `base: './'` makes them relative (`./assets/main.js`), which matters when the frontend and backend are served from the same origin. With `base: './'`, the built `index.html` references assets with relative paths, so they resolve correctly regardless of the deployment URL.

**What `outDir` does**: Tells Vite where to write the build output. After running `npm run build` in the frontend directory, the compiled files end up in `nacl_backend/dist/`.

**The output structure**:
```
nacl_backend/dist/
├── index.html
└── assets/
    ├── main-abc123.js
    ├── react-xyz456.js
    └── ...
```

The `index.html` is the entry point. The `assets/` directory contains all the JavaScript, CSS, and other static resources that `index.html` references.

---

## Step 2: Embed the build output with `//go:embed`

Go's `embed` package (introduced in Go 1.16) lets you include files from disk into the binary at compile time.

Create a file at the root of your Go project (the same directory that contains `dist/`):

```go
// nacl_backend/embed.go
package main

import "embed"

//go:embed dist
var DistFS embed.FS
```

**How `//go:embed` works**:

- The directive is processed at **compile time**, not runtime.
- `//go:embed dist` embeds the entire `dist/` directory tree recursively.
- The `embed.FS` variable is populated at compile time with the contents of all files under `dist/`.
- Files are stored in the binary's read-only data section — they use memory but can't be modified.
- The embedded paths preserve the `dist/` prefix: inside `DistFS`, paths look like `dist/index.html`, `dist/assets/main.js`, etc.

**Why the variable is in the `main` package**: The embed path is relative to the source file's directory. `//go:embed dist` in a file at `nacl_backend/embed.go` resolves to `nacl_backend/dist/`. If we placed the embed file deeper (like in `internal/server/`), the relative path would need to be different, and `..` is not allowed in embed patterns.

**Rules for `//go:embed` paths**:
- Paths are relative to the source file's directory
- Patterns may not contain `.` or `..` (no parent directory traversal)
- Patterns may not begin or end with `/`
- A directory name embeds the entire subtree recursively
- Hidden files (starting with `.` or `_`) are excluded by default

References:
- [Go embed package documentation](https://pkg.go.dev/embed)
- [Go blog: Go 1.16 Embedding](https://go.dev/blog/go1.16)

---

## Step 3: Pass the embedded filesystem to Chi

The embedded filesystem lives in the `main` package as `DistFS`, but our route registration happens in a `server` package. We need to pass it in.

### 3a: Update the server constructor to accept an `fs.FS`

```go
// nacl_backend/internal/server/server.go
import "io/fs"

type Server struct {
    Config     *config.Config
    HTTPServer *http.Server
    Svc        *service.Service
    Logger     *slog.Logger
    StaticFS   fs.FS           // <-- new field
}

func NewServer(
    db db.Querier,
    logger *slog.Logger,
    config *config.Config,
    static fs.FS,              // <-- new parameter
) *Server {
    // ... existing code ...
    s := &Server{
        // ... existing fields ...
        StaticFS: static,       // <-- store it
    }
    // ...
}
```

**Why `fs.FS` (interface) instead of `embed.FS` (concrete type)**: `fs.FS` is an interface from the standard `io/fs` package that has a single method: `Open(name string) (fs.File, error)`. `embed.FS` implements this interface. By accepting `fs.FS`, the server package doesn't need to import `embed` — it only depends on the standard `io/fs` interface. This also makes testing easier: you can pass a `testing/fstest.MapFS` in unit tests without needing real embedded files.

### 3b: Pass `DistFS` from main

```go
// nacl_backend/main.go
s := server.NewServer(queries, log, cfg, DistFS)
```

`DistFS` is in the `main` package (declared in `embed.go`), so it's directly accessible in `main.go`.

---

## Step 4: Understand `fs.FS`, `http.FS`, and `http.FileServer`

Three key types that work together:

### `fs.FS` (from `io/fs`)

The standard filesystem interface introduced in Go 1.16:

```go
type FS interface {
    Open(name string) (File, error)
}
```

It represents a read-only hierarchy of files. `embed.FS` implements it, but so does `os.DirFS` (for real directories), `testing/fstest.MapFS` (for tests), and `zip.Reader` (for zip archives). This abstraction means your server code never needs to know where files actually come from.

### `http.FS()` (from `net/http`)

A bridge function that converts `fs.FS` to `http.FileSystem`:

```go
func FS(fsys fs.FS) FileSystem
```

`http.FileSystem` is an older, `net/http`-specific interface:

```go
type FileSystem interface {
    Open(name string) (File, error)
}
```

It looks similar to `fs.FS`, but `http.File` is different from `fs.File` — it adds `Readdir` and `Stat` methods that `http.FileServer` needs for directory listings and content-type detection.

`http.FS()` wraps an `fs.FS` and returns an `http.FileSystem` that the old `net/http` serving machinery can use.

### `http.FileServer()` (from `net/http`)

Takes an `http.FileSystem` and returns an `http.Handler`:

```go
func FileServer(root FileSystem) Handler
```

This handler:
1. Receives an HTTP request with a URL path (e.g., `/assets/main.js`)
2. Opens the file at that path from the provided `http.FileSystem`
3. Detects the Content-Type from the file extension
4. Sets caching headers (Last-Modified, ETag)
5. Supports conditional requests (304 Not Modified)
6. Writes the file contents to the response

**The chain in code**:
```go
http.FileServer(http.FS(embeddedFs))
//              ▲       ▲
//              │       └── Your embed.FS (implements fs.FS)
//              │
//              └── Converts fs.FS → http.FileSystem
```

---

## Step 5: Set up Chi routes for static files

A Vite build produces two kinds of files: `index.html` and hashed assets in `assets/`. They need different serving strategies.

### 5a: Create sub-filesystems

The embedded FS paths look like `dist/index.html` and `dist/assets/main.js`. We need to strip the `dist/` prefix.

```go
// In RegisterRoutes:
assetsRoot, _ := fs.Sub(s.StaticFS, "dist/assets")
```

**What `fs.Sub` does**: Creates a new `fs.FS` rooted at the given path within the original FS.

```
Before fs.Sub:                       After fs.Sub(s.StaticFS, "dist/assets"):
  s.StaticFS paths:                    assetsRoot paths:
  ┌─────────────────────┐             ┌─────────────────────┐
  │ dist/index.html     │             │ main-abc123.js      │
  │ dist/assets/        │   fs.Sub    │ vendor-xyz.js       │
  │   ├── main-abc.js   │  ───────►  │ styles-123.css      │
  │   ├── vendor-xyz.js │             └─────────────────────┘
  │   └── styles.css    │
  └─────────────────────┘
```

It's a purely virtual operation — no disk I/O, just path rewriting.

### 5b: Register the assets route

```go
import "net/http"

r.Handle("/assets/*", http.StripPrefix("/assets", http.FileServer(http.FS(assetsRoot))))
```

Breaking this down from the inside out:

1. **`http.FS(assetsRoot)`** — wraps our `fs.FS` (the sub-FS rooted at `dist/assets`) into an `http.FileSystem`
2. **`http.FileServer(...)`** — creates an HTTP handler that serves files from that filesystem
3. **`http.StripPrefix("/assets", ...)`** — decorates the handler to remove `/assets` from the request path before passing it to the file server

**Why `StripPrefix` is needed**:

| Step | Path |
|------|------|
| Browser requests | `/assets/main-abc123.js` |
| After `StripPrefix("/assets", ...)` | `/main-abc123.js` |
| FileServer looks up in `assetsRoot` | `main-abc123.js`  |

Without `StripPrefix`, the file server would look for `/assets/main-abc123.js` in `assetsRoot`, which doesn't exist (the file is at `main-abc123.js` at the root of `assetsRoot`).

**Why `r.Handle` and not `r.Get`**: `Handle` matches any HTTP method (GET, POST, etc.), while `Get` only matches GET. For static files this usually doesn't matter, but `Handle` is more correct — browsers only GET these files, but using `Handle` is idiomatic for static file serving.

**Why the `*` in the pattern**: Chi's `/*` is a wildcard that matches the path and all sub-paths. `/assets/*` matches `/assets/main.js`, `/assets/react.js`, and so on. Without the `*`, Chi would only match the exact path `/assets`.

---

## Step 6: Handle SPA client-side routing (the fallback)

### The problem

React apps using `BrowserRouter` (instead of `HashRouter`) use real URL paths for routing. When a user navigates directly to a URL (or refreshes the page), the browser sends a request for that path to your server:

| User visits | Browser requests | What React expects |
|-------------|-----------------|--------------------|
| `/` | `GET /` | `index.html` → show Dashboard |
| `/login` | `GET /login` | `index.html` → show Login |
| `/dash` | `GET /dash` | `index.html` → show Dashboard |
| `/vault` | `GET /vault` | `index.html` → show Vault |

The server receives `GET /login`, but there's no route for `GET /login` in Chi — the API route is `POST /api/login`. If we don't handle this, the server returns 404.

**The solution**: For any request that doesn't match an API route or a static asset, serve `index.html`. React Router, loaded in the browser, reads the current URL and renders the correct page.

### 6a: Implement the fallback handler

```go
// nacl_backend/internal/server/index_handler.go
package server

import (
    "io"
    "io/fs"
    "net/http"
)

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
    root, _ := fs.Sub(s.StaticFS, "dist")
    index, err := root.Open("index.html")
    if err != nil {
        http.Error(w, "Not Found", http.StatusNotFound)
        s.Logger.Error("could not serve index.html")
        return
    }
    defer index.Close()
    w.Header().Set("Content-Type", "text/html")
    io.Copy(w, index)
}
```

Key details:
- `fs.Sub(s.StaticFS, "dist")` creates a filesystem rooted at `dist/`, where `index.html` lives
- `root.Open("index.html")` opens the embedded `dist/index.html` file — returns an `fs.File` which implements `io.ReadCloser`
- `defer index.Close()` ensures the file handle is released
- `w.Header().Set("Content-Type", "text/html")` tells the browser to interpret the response as HTML (without this, some browsers may display it as plain text)
- `io.Copy(w, index)` streams the file contents into the HTTP response — it reads from the file and writes to the response writer in chunks

### 6b: Register the fallback

```go
func (s *Server) RegisterRoutes(r chi.Router) {
    // 1. Static assets
    assetsRoot, _ := fs.Sub(s.StaticFS, "dist/assets")
    r.Handle("/assets/*", http.StripPrefix("/assets", http.FileServer(http.FS(assetsRoot))))

    // 2. SPA fallback — any unmatched route gets index.html
    r.NotFound(s.HandleIndex)

    // 3. Root also serves index.html
    r.Get("/", s.HandleIndex)

    // 4. API routes
    r.Post("/api/users", s.HandleCreateUser)
    r.Post("/api/login", s.HandleLogin)
    // ... etc
}
```

**Route priority matters**: Chi evaluates routes in registration order. By putting `/assets/*` before `NotFound` and API routes, we ensure that:
1. Asset requests are served first (most specific)
2. API requests are matched next
3. Everything else falls through to `NotFound` for the SPA fallback

**Why both `r.Get("/", ...)` and `r.NotFound(...)`**: The root path `/` is a specific route that Chi would match before `NotFound`. Both handlers happen to do the same thing (serve `index.html`). You could skip `r.Get("/", ...)` and let `NotFound` handle it, but having it explicit makes the intent clearer.

### How `NotFound` works in Chi

Chi's `Router` interface provides a `NotFound` method:

```go
type Router interface {
    NotFound(h http.HandlerFunc)
    // ...
}
```

When Chi receives a request, it walks the routing tree. If no registered pattern matches the request path, Chi calls the `NotFound` handler instead of returning a hardcoded 404. This is the catch-all for the SPA pattern.

**What `NotFound` does NOT catch**: Route pattern matches where the HTTP method doesn't match. For example, `GET /api/users` matches the pattern `/api/users` but isn't a `POST`. Chi handles this separately with the `MethodNotAllowed` handler (which defaults to returning 405 Method Not Allowed).

This behavior is actually desirable — it means a `GET` to `/api/users` returns a proper 405 instead of getting `index.html` back.

---

## Step 7: Fix `.gitignore`

For `//go:embed dist` to work at compile time, the `dist/` directory must exist when `go build` runs. If `.gitignore` ignores `dist/`, then either:
- The directory won't be in git (and won't exist when cloning/building on CI)
- The embed directive will fail at compile time because there's nothing to embed

**Option A: Allow `nacl_backend/dist/` while still ignoring other `dist/` directories**

```gitignore
# Keep ignoring most dist/ directories
nacl_frontend/dist/
**/node_modules/

# But allow the backend's dist/ to be tracked
!/nacl_backend/dist/
```

The `!` prefix negates the ignore pattern. This is placed **before** any blanket `dist` rule.

**Option B: Build the frontend during CI before Go compilation**

If you don't want to commit build artifacts, ensure your CI/deploy pipeline runs `npm run build` before `go build`. On Render, this means either:
- A Dockerfile with a multi-stage build (Node stage + Go stage)
- A build script that installs Node.js, builds the frontend, then builds Go

---

## Full request trace: what happens when a browser visits

### Scenario 1: User visits the app for the first time

```
Browser: GET https://yourapp.onrender.com/
                │
                ▼
Chi router receives the request
                │
                ▼
RequestLogger middleware logs: "GET / 200"
                │
                ▼
Recovery middleware (no panic, passes through)
                │
                ▼
Chi checks registered routes in order:
  1. "/assets/*" → path doesn't start with /assets → skip
  2. "/api/*" → path doesn't start with /api → skip
                │
                ▼
Chi matches "/" → calls HandleIndex
                │
                ▼
HandleIndex:
  1. fs.Sub(s.StaticFS, "dist") → get filesystem rooted at dist/
  2. root.Open("index.html") → read embedded dist/index.html
  3. Set Content-Type: text/html
  4. io.Copy(w, index) → stream HTML to browser
                │
                ▼
Browser receives index.html
                │
                ▼
Browser discovers <script src="/assets/main-abc123.js">
                │
                ▼
Browser: GET /assets/main-abc123.js
                │
                ▼
Chi matches "/assets/*"
                │
                ▼
StripPrefix("/assets", fileServer) strips prefix → "/main-abc123.js"
                │
                ▼
FileServer looks up "main-abc123.js" in assetsRoot
                │
                ▼
assetsRoot (fs.Sub of dist/assets) → file found 
                │
                ▼
FileServer responds with:
  - Content-Type: text/javascript (detected from .js extension)
  - Content-Length: 48293
  - Last-Modified: [embed time]
  - [file contents]
                │
                ▼
Browser receives and executes the JavaScript
                │
                ▼
React Router reads the current URL (/) and renders the Dashboard
```

### Scenario 2: User navigates directly to /login

```
Browser: GET https://yourapp.onrender.com/login
                │
                ▼
Chi checks all routes: no match for "/login"
                │
                ▼
Chi fires NotFound handler → calls HandleIndex
                │
                ▼
Same as above: index.html is served
                │
                ▼
Browser receives index.html, React starts
                │
                ▼
React Router reads the URL (/login) → navigates to LoginPage
```

### Scenario 3: User submits the login form

```
Browser: POST https://yourapp.onrender.com/api/login
{
  "email": "...",
  "password": "..."
}
                │
                ▼
Chi matches "POST /api/login"
                │
                ▼
TokenValidator middleware (no token needed for login, passes through)
                │
                ▼
HandleLogin:
  - Validates request body
  - Queries database
  - Returns JWT token in JSON
```

---

## Summary of key types and functions

| Type/Function | Package | Purpose |
|---------------|---------|---------|
| `embed.FS` | `embed` | Compile-time read-only filesystem from `//go:embed` |
| `fs.FS` | `io/fs` | Interface for any read-only filesystem (embed.FS, os.DirFS, etc.) |
| `fs.Sub(fsys, dir)` | `io/fs` | Returns a new FS rooted at a subdirectory of the input FS |
| `http.FS(fsys)` | `net/http` | Converts `fs.FS` → `http.FileSystem` (bridge between old/new APIs) |
| `http.FileServer(root)` | `net/http` | Returns an `http.Handler` that serves files from an `http.FileSystem` |
| `http.StripPrefix(prefix, h)` | `net/http` | Returns a handler that removes a URL prefix before delegating to `h` |
| `chi.Router.NotFound(h)` | `chi/v5` | Registers a handler for routes that don't match any pattern |

**The chain visualized**:

```
embed.FS ──► fs.FS ──► http.FS() ──► http.FileSystem ──► http.FileServer() ──► http.Handler ──► Chi route
```

---

## Troubleshooting

### "404 Not Found" for static assets

**Check**: Does your Vite `base` config match your route prefix? If Vite generates `/assets/main.js` paths and your Chi route is `/static/*`, the browser will request `/assets/main.js` which won't match.

**Fix**: Either set `base: '/static/'` in Vite, or register the route at the path Vite expects.

### "no matching files found" error when building Go

**Check**: Does `nacl_backend/dist/` actually exist? Run `ls nacl_backend/dist/` to verify. If not, build the frontend first: `npm run build` in the frontend directory.

**Fix**: Ensure `dist/` is either committed to git or built before `go build`.

### "index.html not served for some routes"

**Check**: Are you using `r.NotFound()` correctly? The `NotFound` handler only fires when no route pattern matches. If you have a pattern like `r.Get("/{param}", ...)`, it might catch `GET /login` before `NotFound` gets a chance.

**Fix**: Be deliberate about which routes are registered. API routes under `/api/*` are safe because they won't match `/login`. Avoid catch-all patterns that might intercept SPA routes before `NotFound`.

### Content-Type is "application/octet-stream" for JS files

**Check**: Does `http.FileServer` have access to the OS MIME type database? On minimal Docker images, MIME types might not be configured.

**Fix**: Import `mime` and either run `mime.AddExtensionType(".js", "text/javascript")` at startup, or register the MIME types your app needs.

### Binary is too large

**Check**: The embedded `dist/` directory size. Run `du -sh nacl_backend/dist/`. A typical React build is 100-500 KB after compression.

**Fix**: The binary size increase equals the uncompressed size of the `dist/` directory. This is normal. If size is a concern, consider gzip compression at the web server level or moving to a separate static file host.
