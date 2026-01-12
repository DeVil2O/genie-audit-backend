# Automation Templates — Solutioning Doc

## Objectives
- Persist user-created React Flow graphs as cheap `jsonb` in Postgres.
- Provide production-style HTTP APIs (Gin) with request tracing, structured logs, validation, and graceful shutdown.
- Keep infra minimal (Postgres), allow search/filter by tags/visibility, and support versioning.

## Data Model (Postgres)
- Table: `automation_templates`
  - `id uuid PK DEFAULT gen_random_uuid()`
  - `name text NOT NULL`
  - `description text`
  - `tags text[] NOT NULL DEFAULT '{}'::text[]`
  - `graph_json jsonb NOT NULL`
  - `created_by_project_id uuid NULL` (for scoping/ownership; nullable for built-ins)
  - `visibility text NOT NULL DEFAULT 'private'` (`private|public|builtin`)
  - `version int NOT NULL DEFAULT 1`
  - `created_at timestamptz NOT NULL DEFAULT now()`
  - `updated_at timestamptz NOT NULL DEFAULT now()`
- Indexes:
  - GIN on `tags` for fast tag filter.
  - B-tree on `(visibility, created_by_project_id)` for listing.
  - Updated `updated_at` trigger handled in code (no triggers needed for now).

## API Surface (Gin)
- Base path: `/v1/templates`
- Endpoints (all JSON, auth can be added via middleware later):
  - `POST /v1/templates` — create template.
  - `GET /v1/templates` — list with filters `visibility`, `project_id`, `tag`.
  - `GET /v1/templates/:id` — fetch one.
  - `PATCH /v1/templates/:id` — partial update; bumps `version` automatically.
- Request validation: required `name`, `graph_json` object, `visibility` in allowed set, tags length bounded, `project_id` UUID format when provided.
- Responses: template payload; errors as JSON with `message` + trace id.

## Application Structure
- `cmd/api/main.go`: entrypoint; loads env, initializes logger, pgx pool, Gin router, starts HTTP with graceful shutdown and context cancelation.
- `http/router.go`: `NewRouter(db)` wires middlewares (request ID/trace, recovery, JSON logging) and routes.
- `http/routes.go`: registers template routes with handlers.
- `middleware/request_context.go`: injects per-request `traceID` (UUID), logs structured fields, sets `X-Trace-ID`.
- `stores/templates.go`: pgx queries with context timeouts; no long-lived goroutines; connections closed via pool shutdown.
- `models/template.go`: shared DTOs for DB + JSON.
- `logger/`: reuse existing logrus wrapper; include traceID in all handler logs.

## Observability & Ops
- Structured logs with trace id, method, path, status, latency.
- Request-scoped context passed to DB to avoid leaks; per-call `context.WithTimeout` guards.
- Graceful shutdown: OS signals (SIGINT/SIGTERM) shut down HTTP server with timeout, then close pgx pool.

## Versioning
- `version` starts at 1 on create.
- `PATCH` increments `version` whenever persisted changes occur.
- Future: move to temporal table/history if needed.

## Security/Validation
- Input sanitized/validated; `visibility` whitelist.
- `graph_json` stored as-is in `jsonb`; size is small (few KB).
- Hooks available to plug API-key auth middleware later.

## Testing/Manual Checks
- `go test ./...` (unit tests can be added incrementally).
- Manual curl examples:
  - `POST /v1/templates` with sample graph.
  - `GET /v1/templates?visibility=public&tag=foo`.

## Out of Scope (not built yet)
- Worker/runner integration with templates.
- Full auth/ACL; only scaffolding for project scoping.
