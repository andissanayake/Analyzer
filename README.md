# Web Page Analyzer

This project analyzes a URL and returns structured information about the page:

- HTML version
- Page title
- Heading counts (`h1`-`h6`)
- Internal and external link counts
- Inaccessible links count
- Login-page detection (`hasLoginForm`, `loginScore`, `loginReason`)

The solution has two components:

- `api`: Go HTTP API (`/health`, `/analyze`)
- `web`: React + Vite frontend

## Main Build/Run/Deploy Steps

### 1) Prerequisites

- Docker Desktop (recommended for full stack)
- Optional local development:
  - Go (for `api`)
  - Node.js + npm (for `web`)

### 2) Build and run with Docker Compose (recommended)

From repository root:

```bash
docker-compose up --build
```

Expected endpoints:

- Web UI: `http://localhost:3000`
- API: `http://localhost:5000`
- Health check: `http://localhost:5000/health`

Stop:

```bash
docker-compose down
```

### 3) Run tests

API tests (from `api` directory):

```bash
go test ./...
```

Notes:

- Running `go test ./...` from repository root fails because the Go module is inside `api`.
- Integration tests include live network calls (for example, real external URLs), so they depend on internet availability and external-site stability.

### 4) Local non-Docker run (optional)

API:

```bash
cd api
go run ./cmd/api
```

Web:

```bash
cd web
npm install
npm run dev
```

## Assumptions and Decisions

These are the implementation assumptions/decisions made where requirements were unclear or unspecified:

1. **CORS policy**
  - Default `CORS_ORIGIN` is `http://localhost:3000` for local development.
  - This assumes single known frontend origin in dev.
2. **Network behavior**
  - API fetches target URLs live with timeout-based HTTP client behavior.
  - External site reachability/content is treated as dynamic and can affect analysis result quality.
3. **Server-rendered HTML dependency**
  - Analysis is based on the HTML returned by the HTTP response body.
  - Pages that rely heavily on SPA frameworks or JavaScript DOM manipulation/rendering may not be fully represented in results.
4. **Login page detection strategy**
  - Detection is heuristic and score-based (`loginScore`, `loginReason`), not a strict classifier.
  - A threshold-like interpretation is used by tests (high score expected for known login pages).
5. **Monorepo structure**
  - Go module is scoped to `api`, not root.
  - Web and API are built/deployed as separate services coordinated with `docker-compose.yml`.
6. **Development-first container setup**
  - Web container runs Vite dev server (`npm run dev`) on port `3000`.
  - This is optimized for local iteration, not final production serving.

## Suggestions for Improvement

1. **Performance and scalability**
  - Add concurrency limits, optional caching for repeated URL checks, and benchmark parsing/link-check stages.

