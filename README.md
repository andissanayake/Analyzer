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

## Screenshot

Web Page Analyzer UI

## Project Overview

This project provides a backend API and frontend UI for analyzing a public web page URL.  
The application fetches page HTML and returns structured analysis data that can be consumed by the UI.

Core analysis outputs:

- HTML version
- Page title
- Heading counts (`h1`-`h6`)
- Internal and external link counts
- Inaccessible links count
- Login-page detection (`hasLoginForm`, `loginScore`, `loginReason`)

## Technologies Used

- Backend (BE): Go (`net/http`, `slog`), Prometheus client, pprof
- Frontend (FE): React + Vite
- DevOps: Docker, Docker Compose
- Field validation: backend request validation for input URL format and required fields

## Service URLs and API Docs

When running with Docker Compose:

- Frontend URL: `http://localhost:3000`
- Backend URL: `http://localhost:5000`
- Health endpoint: `http://localhost:5000/health`
- Analyze endpoint: `POST http://localhost:5000/analyze`
- Metrics endpoint: `http://localhost:5000/metrics`
- pprof endpoint: `http://127.0.0.1:6060/debug/pprof/`

API docs/spec:

- This project does not use Swagger/OpenAPI generation.
- API contract is documented manually in this README (request/response examples below).

## Main Build/Run/Deploy Steps

### 1) Prerequisites

- Docker Desktop (recommended for full stack)
- Optional local development:
  - Go (for `api`)
  - Node.js + npm (for `web`)

## External Dependencies

Backend Go module dependencies are defined in `api/go.mod` and installed automatically by Go tooling:

- `github.com/prometheus/client_golang` (Prometheus metrics)
- `golang.org/x/net/html` (HTML parsing)

Frontend dependencies are defined in `web/package.json` and installed via:

```bash
cd web
npm install
```

### 2) Build and run with Docker Compose (recommended)

From repository root:

```bash
docker-compose up --build
```

Expected endpoints:

- Web UI: `http://localhost:3000`
- API: `http://localhost:5000`
- Health check: `http://localhost:5000/health`
- Prometheus metrics: `http://localhost:5000/metrics`
- pprof (localhost-only): `http://127.0.0.1:6060/debug/pprof/`

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

### 3.1) Coverage

To view package coverage:

```bash
cd api
go test ./... -cover
```

Current test approach includes:

- Unit tests using mocks/fakes for deterministic behavior.
- Integration tests for live end-to-end behavior with real websites.

### 3.2) Metrics (Prometheus)

The API exposes Prometheus metrics on:

- `GET /metrics`

Minimal metric set:

- `analyze_requests_total{status}`: total analyze requests grouped by returned app status code.
- `analyze_duration_seconds`: analyze request duration histogram.
- `links_checked_total`: total number of unique HTTP(S) links checked during analysis.
- `links_inaccessible_total`: total number of links identified as inaccessible.

### 4) API Contract (Manual Spec)

`POST /analyze`

Request body:

```json
{
  "url": "https://example.com"
}
```

Success response (HTTP 200; app status in payload):

```json
{
  "statusCode": 200,
  "message": "Analysis complete.",
  "body": {
    "htmlVersion": "HTML5",
    "pageTitle": "Example Domain",
    "headings": { "h1": 1, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0 },
    "internalLinks": 0,
    "externalLinks": 1,
    "inaccessibleLinks": 0,
    "hasLoginForm": false,
    "loginScore": 0,
    "loginReason": "No strong login signals found."
  }
}
```

Validation error response example:

```json
{
  "statusCode": 400,
  "message": "The URL is not valid."
}
```

### 5) Local non-Docker run (optional)

API:

```bash
cd api
go run ./cmd/api
```

Debug/profiling defaults:

- `PPROF_ENABLED=true`
- `PPROF_ADDR=127.0.0.1:6060` (separate debug server, not on the public API mux)

Web:

```bash
cd web
npm install
npm run dev
```

## Application Usage and Main Functionalities

1. Open the frontend at `http://localhost:3000`.
2. Enter a URL and submit.
3. Review returned analysis data.

Main functionality roles in the application:

- URL validation: rejects missing/invalid URLs with clear error messages.
- Analysis engine: parses HTML and computes content/link/login insights.
- Logging: structured JSON logs via `slog` for observability and troubleshooting.
- Error handling: maps upstream and validation failures into user-friendly responses.
- Monitoring: Prometheus metrics and pprof profiling for runtime diagnostics.

## AI Usage Statement

No AI-generated code is intentionally included as a direct copy-paste artifact in this project.  
All submitted code is reviewed and adapted to align with project requirements and Go coding practices.

## Challenges and How They Were Addressed

1. **Balancing speed and correctness in link checks**
  - Approach: deduplicated link targets and used bounded worker concurrency with per-link timeout.
2. **Handling unstable external pages for integration tests**
  - Approach: separated unit tests (mocked and deterministic) from integration tests (live network), and documented integration-test behavior/limitations.
3. **Monitoring without overcomplicating the architecture**
  - Approach: integrated lightweight Prometheus counters/histograms and exposed pprof on a dedicated debug server.

## Assumptions and Decisions

These are the implementation assumptions/decisions made where requirements were unclear or unspecified:

1. **CORS policy**

- Default `CORS_ORIGIN` is `http://localhost:3000` for local development.
- This assumes single known frontend origin in dev.

1. **Server-rendered HTML dependency**

- Analysis is based on the HTML returned by the HTTP response body.
- Pages that rely heavily on SPA frameworks or JavaScript DOM manipulation/rendering may not be fully represented in results.

1. **Link accessibility counting approach**

- The analyzer parses the HTML document once, collects resolvable HTTP/HTTPS link targets during that pass, and deduplicates them before checking accessibility.
- Link checks are executed in parallel using a bounded worker pool (`LinkCheckWorkerCount`) with a per-link timeout (`LinkCheckTimeoutPerURL`).
- `InaccessibleLinks` includes both links that are invalid at parse/classification time (for example empty or malformed `href`) and links that fail runtime accessibility checks (request error or non-2xx/3xx response).

1. **Login page detection strategy**

- Detection is heuristic and score-based (`loginScore`, `loginReason`), not a strict classifier.
- A threshold-like interpretation is used by tests (high score expected for known login pages).

## Suggestions for Improvement

1. **Performance and scalability**

- Add concurrency limits, optional caching for repeated URL checks, and benchmark parsing/link-check stages.

1. **Render JavaScript-heavy pages**

- Use a headless browser (Chromium with Playwright) to render SPA pages and capture the generated DOM/HTML before analysis.
- Keep the current fast HTTP fetch path as a default and enable browser rendering as an optional mode for accuracy-sensitive use cases.

1. **Vision-assisted login detection**

- Capture one or more browser screenshots during page rendering and analyze visual login cues (for example sign-in forms/buttons) in addition to HTML signals.
- Optionally call an AI vision service (such as an OpenAI-backed server) to improve detection accuracy on pages where DOM-based heuristics are ambiguous.

1. **Test and CI improvements**

- Add package-level tests for currently low-coverage packages (`config`, `httpx`, `health`, `metrics`) and enforce a minimum total coverage threshold in CI.

1. **Developer workflow**

- Add a top-level `Makefile` with common commands (`test`, `coverage`, `docker-up`, `docker-down`, `lint`) for faster onboarding.