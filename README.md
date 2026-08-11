# go-todo-api

A production-style Todo REST API built in Go from scratch — no framework, just the standard library plus a small set of well-respected packages (chi, pgx, asynq). Built step-by-step as a learning project covering the full range of skills expected of a senior Go backend developer: clean layering, validation, testing (unit + integration), structured logging, graceful shutdown, pagination, bulk operations, transactions, indexing/N+1 awareness, and background job processing.

## Tech Stack

| Concern | Choice |
|---|---|
| Language | Go 1.22+ |
| HTTP routing | [chi](https://github.com/go-chi/chi) (thin router on top of `net/http`) |
| Database | PostgreSQL |
| DB driver | [pgx/v5](https://github.com/jackc/pgx) + `pgxpool` |
| Validation | [go-playground/validator](https://github.com/go-playground/validator) |
| Logging | `log/slog` (standard library, JSON structured) |
| Background jobs | [asynq](https://github.com/hibiken/asynq) (Redis-backed queue) |
| Testing | standard `testing` package + [testcontainers-go](https://github.com/testcontainers/testcontainers-go) for integration tests |

## Project Structure

```
go-todo-api/
├── cmd/
│   ├── api/
│   │   └── main.go              # API server entry point
│   └── worker/
│       └── main.go              # Background job worker entry point
├── internal/
│   ├── config/                  # Environment variable loading
│   ├── database/                # Postgres connection pool setup
│   ├── logger/                  # Structured (slog) logger setup
│   ├── models/                  # Todo struct + request/response DTOs
│   ├── queue/                   # asynq client (producer side)
│   ├── repository/              # All SQL — CRUD, pagination, bulk ops, transactions
│   ├── tasks/                   # Background task type + payload definitions
│   ├── validation/               # Shared validator instance
│   └── handlers/                # HTTP handlers, router, middleware
├── go.mod
├── go.sum
├── .env                         # Local environment variables (not committed)
├── README.md
└── INSTALL.md
```

## Architecture / Request Flow

```
Client
  │
  ▼
Router (chi) — method/path matching
  │
  ▼
Middleware — Recoverer → RequestID → Logging
  │
  ▼
Handler — parse request, validate, call repository, write response
  │
  ▼
Repository — raw parameterized SQL via pgx
  │
  ▼
PostgreSQL
```

Background work (e.g. notifications) is enqueued from a handler into Redis via `asynq`, and processed asynchronously by the separate `cmd/worker` process — decoupled entirely from the request/response cycle.

## API Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/todos` | Create a new todo |
| `GET` | `/todos?page=1&limit=10` | List todos (paginated) |
| `GET` | `/todos/{id}` | Get a single todo by ID |
| `PUT` | `/todos/{id}` | Update a todo |
| `DELETE` | `/todos/{id}` | Delete a todo |
| `POST` | `/todos/{id}/archive` | Archive a todo (transactional move to `archived_todos`) |

### Example: Create a todo

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn Go", "description": "Build a CRUD API from scratch"}'
```

Response (`201 Created`):
```json
{
  "id": 1,
  "title": "Learn Go",
  "description": "Build a CRUD API from scratch",
  "completed": false,
  "created_at": "2026-08-11T10:00:00Z",
  "updated_at": "2026-08-11T10:00:00Z"
}
```

### Example: Validation error response

```json
{
  "error": "validation failed",
  "fields": {
    "Title": "this field is required"
  }
}
```

See [INSTALL.md](./INSTALL.md) for full setup and running instructions.

## Key Design Decisions

- **No ORM** — raw SQL via `pgx`, for full control and no hidden query behavior. Keeps SQL visible and easy to reason about.
- **Layered architecture** — handlers never touch SQL; repositories never touch HTTP. Each layer can be tested, swapped, or reasoned about independently.
- **Interface-based repository** (`TodoRepositoryInterface`) — allows handlers to be unit tested with a mock, without a real database.
- **Sentinel errors** (e.g. `repository.ErrTodoNotFound`) — domain-level errors translated into precise HTTP status codes at the handler layer via `errors.Is`.
- **Structured logging with request IDs** — every request gets a unique ID, attached to every log line produced while handling it, for traceable debugging.
- **Graceful shutdown** — in-flight requests are given time to finish before the process exits, instead of being abruptly killed.
- **Cursor pagination available alongside offset pagination** — offset pagination for simple listing with page numbers; cursor-based for scale (avoids the `OFFSET` slowdown and skip/duplicate bugs on large, frequently-changing tables).
- **Background jobs decoupled via Redis/asynq** — slow or non-critical side effects (e.g. notifications) never block the API response.

## Testing

Two distinct layers of tests exist:

- **Unit tests** (`internal/handlers/*_test.go`) — fast, no external dependencies, use a hand-written mock repository.
- **Integration tests** (`internal/repository/*_test.go`) — spin up a real, throwaway PostgreSQL container via `testcontainers-go` and run actual SQL against it.

```bash
go test ./...                          # run everything
go test ./internal/handlers/... -v     # fast unit tests only
go test ./internal/repository/... -v   # slower integration tests only (requires Docker)
```

## Roadmap / Not Yet Covered

- Worker pool / goroutine concurrency patterns (`sync.WaitGroup`, `errgroup`)
- Caching layer (Redis, cache-aside pattern)
- Authentication (JWT) + protected routes
- Rate limiting
- Health check / readiness endpoints
- Dockerfile + docker-compose (API + Postgres + Redis together)
- Observability (Prometheus metrics, OpenTelemetry tracing)

## License

Personal learning project — no license applied.