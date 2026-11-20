# Prompts & Development Process

This project was built with **Claude Code** from a specification I wrote.
I directed the work phase by phase and reviewed and tested each phase
before moving on.

## My role

**Specification design**
- Defined the architecture: Go REST API (standard library only) with a pure
  domain layer, and a React + TypeScript frontend where all calculations
  happen in the backend.
- Designed the API contract and the error model: `400` for invalid requests
  and `422` for valid requests that can't be computed (division by zero,
  square root of a negative, overflow).
- Chose the design patterns (Strategy, Factory/Registry, Facade, Adapter,
  Decorator, State machine, Command) and the quality bar: coverage
  thresholds, linters, accessibility and responsive design.
- Split the work into 8 phases with a commit per phase for easy review.

**Review at every phase**
- Read the generated code and had each design pattern explained with its
  location in the codebase.
- Ran unit tests, coverage reports and linters (`make test`, `make coverage`,
  `make lint`) and inspected uncovered code.
- Checked the commit history to keep each phase clean and the repo working.

**Manual testing**
- API: tested endpoints with `curl`, including error cases and status codes.
- UI: tested all operations, error messages, keyboard usage, mobile view,
  and behavior when the backend is down.
- Verified requests and responses in the browser DevTools.

## Main prompt




# Role & Context

You are a senior full-stack engineer completing a **technical assessment for Sezzle** (a fintech "Buy Now, Pay Later" company). The evaluators will judge **correctness, clarity, maintainability, test quality, and documentation**, not feature count. Act like a strong candidate: deliberate, clean, well-documented, no over-engineering, no dead code.

Work in **phases** (below). At the end of each phase, run the relevant tests/linters, fix everything, and make a **Conventional Commit** (`feat:`, `test:`, `docs:`, `chore:`, `refactor:`). Never leave the repo in a broken state. Do not ask me for confirmation between phases unless something is truly ambiguous; make a sensible decision and record it in the README's "Design Decisions & Assumptions" section.

---

# Assignment (must be satisfied 100%)

Build a full-stack calculator: **React + TypeScript frontend** consuming a **Go REST microservice**.

**Operations:** Addition, Subtraction, Multiplication, Division. Also implement the optional ones: Exponentiation, Square Root, Percentage.

**Frontend:** intuitive UI for input and results, input validation and error handling, responsive design (mobile support).

**Backend:** REST endpoints for the operations, input validation, edge cases (division by zero, invalid data, overflow), JSON responses.

**Non-functional:** clean, idiomatic code in both layers; unit tests for key functionality in both layers; coverage report; documentation (setup, API usage, design rationale); Dockerfile/compose to run everything together.

**Deliverables:** Git repo with frontend and backend, README (setup, how to run each part, API call examples, design decisions/assumptions), unit tests plus coverage report, Docker setup, and a `PROMPTS.md` file containing this prompt (the assessment requires sharing the prompts used).

---

# Repository Structure

```
sezzle-calculator/
├── backend/
│   ├── cmd/server/main.go           # composition root only (wiring, no logic)
│   ├── internal/
│   │   ├── config/                  # env-based config (Builder / functional options)
│   │   ├── domain/                  # Operation interface, errors, value types
│   │   ├── operations/              # one file per operation (Strategy)
│   │   ├── registry/                # operation registry/factory
│   │   ├── service/                 # CalculatorService (use cases)
│   │   └── transport/http/          # handlers, DTOs, middleware, router, error mapping
│   ├── go.mod
│   ├── Makefile
│   ├── .golangci.yml
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── api/                     # HTTP client + adapter (DTO <-> domain)
│   │   ├── domain/                  # types, operation metadata, validation
│   │   ├── features/calculator/     # components, hooks, state machine/reducer
│   │   ├── components/ui/           # reusable presentational components
│   │   ├── theme/                   # design tokens (CSS variables)
│   │   └── main.tsx / App.tsx
│   ├── Dockerfile + nginx.conf
│   └── package.json
├── docker-compose.yml
├── Makefile                         # root: make test, make coverage, make up
├── .github/workflows/ci.yml         # lint + test + coverage for both layers
├── README.md
└── PROMPTS.md
```

---

# Backend (Go)

**Stack:** latest stable Go, **standard library only** for HTTP (`net/http` with Go 1.22+ method/path routing patterns), `log/slog` for structured logging. No web frameworks. Testing with `testing` + `net/http/httptest`; `testify` is optional, keep dependencies minimal.

## API Design

- `GET  /api/v1/health` → `{"status":"ok"}`
- `GET  /api/v1/operations` → list of supported operations with their metadata (name, symbol, arity)
- `POST /api/v1/calculate`
  - Request: `{"operation":"divide","operands":[10, 4]}`
  - Success 200: `{"operation":"divide","operands":[10,4],"result":2.5}`
  - Error 4xx: `{"error":{"code":"DIVISION_BY_ZERO","message":"division by zero is not allowed"}}`

Operation names: `add`, `subtract`, `multiply`, `divide`, `power`, `sqrt`, `percentage`.

- `sqrt` is unary (1 operand); all others are binary (2 operands).
- `percentage`: `a` percent of `b` → `(a / 100) * b`. Document this assumption.

## Validation & Edge Cases (each one needs a test)

- Malformed JSON, unknown fields (`DisallowUnknownFields`), empty body, body size limit (`http.MaxBytesReader`)
- Unknown operation → `400 UNSUPPORTED_OPERATION`
- Wrong number of operands → `400 INVALID_OPERAND_COUNT`
- Division by zero → `422 DIVISION_BY_ZERO`
- Square root of a negative number → `422 NEGATIVE_SQUARE_ROOT`
- Results that are `NaN` or `±Inf` (overflow, e.g. `power(10, 400)`, or `power(0, -1)`) → `422 RESULT_OVERFLOW` / `UNDEFINED_RESULT`
- Wrong HTTP method → 405; unknown route → 404 with JSON body
- Use **typed domain errors** (sentinel errors + `errors.Is/As`) and a **single error-mapping function** in the transport layer that converts domain errors to HTTP status + code. The domain layer must know nothing about HTTP.

## Design Patterns (apply naturally; name them in code comments and README)

- **Strategy (behavioral):** `Operation` interface (`Name()`, `Arity()`, `Apply(operands []float64) (float64, error)`); one implementation per operation.
- **Factory + Registry (creational):** `registry.New(ops ...domain.Operation)` builds the registry; `Get(name)` resolves the strategy. Adding a new operation = add one file + register it in one place (Open/Closed Principle).
- **Builder / Functional Options (creational):** server and config construction (`NewServer(opts ...Option)`).
- **Decorator / Chain of Responsibility (structural/behavioral):** HTTP middleware chain: request ID, structured logging, panic recovery, CORS, timeouts.
- **Facade (structural):** `CalculatorService` exposes a single `Calculate(ctx, op, operands)` and hides registry, validation, and result checking.
- **Adapter (structural):** DTO ↔ domain mapping in the transport layer.
- **Dependency Injection** via interfaces and constructors, wired only in `main.go`. No globals and no `init()` magic.

## Quality

- Graceful shutdown (`signal.NotifyContext` + `server.Shutdown`), read/write/idle timeouts, port from `PORT` env (default 8080), CORS origin configurable via env.
- **Table-driven tests** for every operation, the service, the registry, the error mapper, middleware, and handlers (via `httptest`). Target **≥ 90% coverage** of `internal/`.
- `make test`, `make coverage` (generates `coverage.out` + `coverage.html`), `make lint` (golangci-lint with `govet`, `errcheck`, `staticcheck`, `revive`, `gofmt`, `goimports`).
- Go doc comments on all exported identifiers.

---

# Frontend (React + TypeScript)

**Stack:** Vite + React 18+ + TypeScript in **strict mode**, Vitest + React Testing Library + `@testing-library/user-event`, **MSW** to mock the API in tests, ESLint + Prettier. Plain CSS Modules + CSS variables (no heavy UI library), so the styling stays clearly hand-crafted.

## UX

- Calculator layout: a display showing the expression and the result, a numeric keypad, and operation buttons (+, −, ×, ÷, xʸ, √, %), plus clear (C), backspace, decimal point, and equals.
- **Full keyboard support** (digits, operators, Enter = equals, Backspace, Escape = clear).
- Unary operations (√) work on the current value immediately; binary operations follow the classic `a op b =` flow.
- Calculations are performed by the **backend**, never computed locally (the frontend only validates input).
- Loading state while the request is in flight (disable equals, show a subtle indicator); prevent double submits.
- Errors are shown inline in the display area with friendly messages mapped from backend error codes (e.g. `DIVISION_BY_ZERO` → "You can't divide by zero"). Network failure → "Can't reach the calculator service. Please try again."
- Client-side validation: one decimal point max, max input length, no operator without operands, no NaN being sent.
- Optional, only if time allows: small history of the last 5 operations (session only).
- Accessibility: semantic buttons, `aria-label`s on symbol buttons, `aria-live="polite"` on the result, visible focus states, WCAG AA contrast.
- Responsive: mobile-first; the keypad fills the screen width on phones and is centered as a card on desktop.

## Visual Style: inspired by Sezzle's brand

Fintech, friendly, trustworthy, modern:
- Primary: Sezzle-like purple (around `#8333D4`), a deep purple/near-black for text and the display (around `#2A1A3E`), white/very light lavender surfaces (`#F7F3FC`), and a soft mint/green accent for success states. Define everything as **design tokens** in `theme/tokens.css` (colors, spacing, radius, shadows, typography) so the palette can be adjusted in one place. These hex values are approximate; keep them consistent.
- Rounded corners (12–20px), generous whitespace, soft shadows, pill-shaped operator buttons, a modern sans-serif (e.g. Inter or system UI stack).
- Subtle hover and press micro-interactions (respect `prefers-reduced-motion`).
- A simple header with the text "Calculator" and a small "Built for Sezzle technical assessment" caption. **Do not use Sezzle's logo, wordmark, or any copied brand assets**: only an inspired color palette and tone.

## Design Patterns (frontend)

- **State pattern / State machine (behavioral):** calculator logic in a pure `useReducer` reducer with explicit states (`idle`, `enteringFirstOperand`, `operatorSelected`, `enteringSecondOperand`, `loading`, `result`, `error`). The reducer is a pure function and is unit-tested exhaustively.
- **Command (behavioral):** every key/button press is dispatched as a typed action (`{ type: 'DIGIT', value: '7' }`, etc.). The keyboard and the on-screen keypad share the same commands.
- **Adapter (structural):** `api/calculatorClient.ts` maps backend DTOs and error payloads to frontend domain types and a typed `CalculatorError`.
- **Facade (structural):** a `useCalculator()` hook exposes a simple API (`state`, `press(command)`) and hides the reducer, API client, and effects.
- **Factory (creational):** operation buttons generated from a single operation metadata config (`domain/operations.ts`), and an API client factory `createCalculatorClient({ baseUrl, fetch })` for injection and testing.
- **Dependency Injection:** the client is provided via React Context, so tests can inject a fake.
- Presentational vs. container components separation, small single-responsibility components, no `any`.

## Quality

- Tests: reducer (all transitions), validation utilities, API adapter (success, each error code, network failure), hooks, and component/integration tests with MSW (full flows: `12 ÷ 4 = 3`, division by zero shows an error, keyboard input works, loading state).
- Target **≥ 85% coverage** (`vitest --coverage` with v8, thresholds enforced in config).
- API base URL via `VITE_API_BASE_URL` (default `/api/v1`, using the Vite dev proxy to `localhost:8080`).
- Scripts: `dev`, `build`, `test`, `test:coverage`, `lint`, `format`, `typecheck`.

---

# Docker & Tooling

- `backend/Dockerfile`: multi-stage, static binary (`CGO_ENABLED=0`), final image `gcr.io/distroless/static` (or `scratch`), non-root user.
- `frontend/Dockerfile`: multi-stage, Node build → `nginx:alpine` serving static files, with `nginx.conf` proxying `/api/` to the backend service and SPA fallback.
- `docker-compose.yml`: `docker compose up --build` → app at `http://localhost:3000`, backend healthcheck, frontend depends on a healthy backend.
- Root `Makefile`: `make test`, `make coverage`, `make lint`, `make up`, `make down`.
- `.github/workflows/ci.yml`: lint, test, and coverage for both layers on push/PR.
- Proper `.gitignore` files and `.editorconfig`. No committed `node_modules`, binaries, or coverage HTML (coverage summaries go in the README instead).

---

# README.md (professional, scannable)

1. Title + one-paragraph overview + screenshot placeholder
2. Tech stack
3. Quick start with Docker (one command)
4. Local setup without Docker (prerequisites with versions, how to run the backend, how to run the frontend)
5. Running tests and generating coverage (both layers) + a **coverage summary table** with the real numbers you obtained
6. **API reference**: every endpoint, request/response schemas, an error code table, and `curl` examples for each operation plus each error case
7. Architecture: an ASCII or Mermaid diagram of the layers (frontend → API client → REST → handler → service → registry → strategy)
8. **Design Decisions & Assumptions**: the patterns used and *why*, in both layers (a short justification for each, no buzzword padding); why stdlib over a framework; why calculations happen in the backend; float64 precision trade-off (and why a decimal library was not used, or was); percentage semantics; error-code design; anything you decided on your own
9. Project structure tree
10. Possible future improvements (short list)
11. AI usage note: mention that Claude Code was used and that the prompt is in `PROMPTS.md`

---

# Execution Phases

1. **Scaffold:** repo structure, `go.mod`, Vite project, linters, `.gitignore`, Makefiles. Commit.
2. **Backend domain + operations + registry + service** with tests. Commit.
3. **Backend HTTP layer** (handlers, DTOs, error mapping, middleware, server, graceful shutdown) with tests. Run `go vet`, lint, coverage. Commit.
4. **Frontend domain, reducer, API client** with tests. Commit.
5. **Frontend UI** (theme tokens, components, keyboard support, responsive, accessibility) with component tests. Commit.
6. **Docker + compose + CI.** Verify `docker compose up --build` works and the frontend calls the backend end-to-end. Commit.
7. **Docs:** README, `PROMPTS.md` (copy this exact prompt into it). Commit.
8. **Final review pass:** act as a strict Sezzle reviewer. Remove dead code, unused deps, and TODOs; check naming consistency, that every requirement in "Assignment" is met, that all tests and linters pass, and that coverage thresholds are met. Then print a checklist mapping each assignment requirement to where it is satisfied in the repo.

**Rules:** keep functions small, prefer clarity over cleverness, no premature abstractions beyond the patterns listed, no commented-out code, meaningful names, consistent English throughout.


## Follow-up prompts

- "Read PROMPTS.md carefully... Start with Phase 1. Stop after each phase
  and wait for my confirmation."
- "Phase N looks good. Continue with Phase N+1." (after reviewing each phase)
---

# Follow-up prompts used during the session

The specification above was given to Claude Code unchanged. Work was then driven by these short messages, in order:

1. "Read PROMPTS.md carefully. It contains the full specification for this technical assessment. Execute it phase by phase exactly as described, committing at the end of each phase. Start with Phase 1. Stop after finishing each phase and wait for my confirmation before starting the next one."
2. "Start phase 2"
3. "Phase 2 looks good. Continue with Phase 3 and stop when finished."
4. "START PHASE 4"
5. "Start phase 5"
6. "Start phase 6" (plus answering "I started Docker Desktop" when asked how to verify the Docker setup)
7. "Start phase 7"
