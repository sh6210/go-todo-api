# Installation & Setup Guide

Complete steps to get `go-todo-api` running locally, from a clean machine to a fully working API + background worker.

## Prerequisites

| Tool | Minimum version | Check with |
|---|---|---|
| Go | 1.22+ (required for stdlib routing features used) | `go version` |
| Docker | any recent version | `docker --version` |
| curl (or Postman/similar) | any | `curl --version` |

Docker is used to run PostgreSQL and Redis locally, and is also required to run the integration test suite (via `testcontainers-go`).

## 1. Clone / set up the project

```bash
mkdir go-todo-api
cd go-todo-api
go mod init github.com/yourusername/go-todo-api
```

> Replace `yourusername` with your actual module path. If you rename it, every internal import across the codebase (`internal/config`, `internal/repository`, etc.) must be updated to match.

Create the folder structure:

```bash
mkdir -p cmd/api cmd/worker
mkdir -p internal/config internal/database internal/logger internal/models
mkdir -p internal/queue internal/repository internal/tasks
mkdir -p internal/validation internal/handlers
```

## 2. Install dependencies

```bash
go get -u github.com/go-chi/chi/v5
go get -u github.com/jackc/pgx/v5
go get -u github.com/jackc/pgx/v5/pgxpool
go get -u github.com/joho/godotenv
go get -u github.com/go-playground/validator/v10
go get -u github.com/google/uuid
go get -u github.com/hibiken/asynq
go get -u github.com/testcontainers/testcontainers-go
go get -u github.com/testcontainers/testcontainers-go/modules/postgres
```

Verify everything resolved correctly:

```bash
go build ./...
```

(No output = success — Go's tooling is silent on success by convention.)

## 3. Start PostgreSQL

```bash
docker run --name todo-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=todo_db \
  -p 5432:5432 \
  -d postgres:16
```

Confirm it's running:
```bash
docker ps
```

## 4. Start Redis

```bash
docker run --name todo-redis -p 6379:6379 -d redis:7-alpine
```

## 5. Create the database schema

Connect to Postgres:
```bash
docker exec -it todo-postgres psql -U postgres -d todo_db
```

Run:
```sql
CREATE TABLE todos (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_todos_created_at ON todos (created_at DESC);

CREATE TABLE archived_todos (
    id INTEGER PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    archived_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Verify:
```sql
\d todos
\d archived_todos
```

Exit:
```sql
\q
```

## 6. Configure environment variables

Create a `.env` file in the project root:

```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable
PORT=8080
REDIS_ADDR=localhost:6379
```

Create a `.gitignore` so secrets never get committed:

```
.env
```

> In production, don't use a `.env` file at all — set these as real environment variables on your server/container/orchestration platform. `.env` is a local-development convenience only.

## 7. Run the application

You need **two terminals** running simultaneously (plus Postgres and Redis already running in the background via Docker from steps 3–4):

**Terminal 1 — the background worker:**
```bash
go run cmd/worker/main.go
```
Expected output: `{"level":"INFO","msg":"worker starting"}`

**Terminal 2 — the API server:**
```bash
go run cmd/api/main.go
```
Expected output: `{"level":"INFO","msg":"server starting","port":"8080"}`

Both processes will keep running until stopped — this is expected (they're servers/workers, not one-shot commands). Use `Ctrl+C` in each terminal to stop them; the API server will run its graceful shutdown sequence before exiting.

## 8. Verify it works

In a third terminal:

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Test setup", "description": "Confirming install works"}'
```

You should get back a `201 Created` response with the new todo, and shortly after, Terminal 1 (the worker) should log a `"notification sent"` line.

```bash
curl http://localhost:8080/todos
```

Should return the paginated list including the todo you just created.

## 9. Run the test suite

Unit tests (fast, no dependencies beyond Go itself):
```bash
go test ./internal/handlers/... -v
```

Integration tests (spins up real, throwaway Postgres containers — requires Docker running):
```bash
go test ./internal/repository/... -v
```

Everything at once:
```bash
go test ./...
```

## Troubleshooting

| Symptom | Likely cause / fix |
|---|---|
| `go build ./...` shows an "undefined" error | An import path doesn't match your actual module name from `go mod init` — check every `github.com/yourusername/go-todo-api/...` import matches your real module path exactly |
| Server fails to start with a database connection error | Postgres container isn't running (`docker ps` to check) or `.env`'s `DATABASE_URL` is wrong |
| `go run cmd/api/main.go` seems to "hang" | This is expected — it's a server, and blocks intentionally. Use a second terminal for `curl` commands |
| Integration tests fail immediately | Docker isn't running, or isn't accessible to the test process |
| Enqueued tasks never get processed | The worker process (`cmd/worker/main.go`) isn't running — it's a separate process from the API server and must be started independently |
| `"undefined: validation.Validate"` or similar | Check `internal/validation/validator.go` exists and the import path in the file using it matches your module name |

## Stopping everything (cleanup)

```bash
docker stop todo-postgres todo-redis
docker rm todo-postgres todo-redis
```