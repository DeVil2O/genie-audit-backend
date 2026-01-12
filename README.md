# Genie Audit Backend (MVP)

Event ingestion API plus async delivery worker. Built with Gin (HTTP), Postgres (storage), Redis + Asynq (queue).

## Architecture
- API: Gin server, auth via `X-API-Key` (hashed in DB).
- Worker: Asynq consumer for delivery tasks with retries.
- Storage: Postgres (migrations in `migrations/`).
- Queue: Redis for job scheduling and backoff.
- Connectors: Slack webhook, generic webhook (HMAC signing), SMTP email(in further iterations).

## Prerequisites
- Go (module targets 1.25.5; use 1.21+ toolchain).
- Docker + docker-compose for Postgres and Redis.
- goose for migrations: `go install github.com/pressly/goose/v3/cmd/goose@latest`.

## Quick start
1) `cp .env.example .env` (fill values below).
2) `docker-compose up -d postgres redis` (when compose is added).
3) Run migrations: `goose -dir ./migrations postgres "$DATABASE_URL" up`.
4) Start API: `go run ./cmd/api`.
5) Start worker: `go run ./cmd/worker`.

## Environment variables
- `DATABASE_URL` (e.g. `postgres://user:pass@localhost:5432/genie?sslmode=disable`)
- `REDIS_URL` (e.g. `redis://localhost:6379/0`)
- `API_ADDR` (default `:8080`)
- `WORKER_CONCURRENCY` (e.g. `5`)
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM` (email)
- `HMAC_SECRET` (webhook signing), `LOG_LEVEL` (info/debug)

## Auth
- Private routes require `X-API-Key`.
- Keys are generated once and stored hashed (`api_key_hash`).
- All operations are scoped to the authenticated project.

## Endpoints (MVP)
- Public: `POST /i/:token` — ingest event for inbox token.
- Projects (auth): `POST /v1/projects`.
- Inboxes (auth): `POST /v1/inboxes`, `GET /v1/inboxes`.
- Automations (auth): `POST /v1/automations`, `GET /v1/automations`, `GET /v1/automations/:id`, `PATCH /v1/automations/:id`.
- Logs (auth): `GET /v1/events`, `GET /v1/events/:id`, `POST /v1/events/:id/replay`.
- Runs (auth): `GET /v1/runs`.
- Deliveries (auth): `GET /v1/deliveries`.

## Data model (Postgres)
- `projects`: id (uuid), name, api_key_hash (bytea), created_at.
- `inboxes`: id, project_id (fk), token unique, name, created_at.
- `automations`: id, project_id, inbox_id (nullable), name, enabled, trigger_type enum, graph_json jsonb, created_at/updated_at.
- `events`: id, inbox_id, source enum, content_type, headers jsonb, body_raw bytea, body_json jsonb nullable, dedupe_key nullable, created_at. Partial unique index on (inbox_id, dedupe_key) where dedupe_key is not null.
- `runs`: id, automation_id, event_id, status enum (queued/running/succeeded/failed), error nullable, started_at, finished_at, created_at.
- `deliveries`: id, run_id, node_id, channel enum (slack/email/webhook), destination, url nullable, secret nullable, payload jsonb, status enum, attempts, next_retry_at, last_error, created_at/updated_at.

## Workflow runner
- Graph stored as React Flow JSON (nodes + edges).
- Supported nodes: `trigger.webhook` (start), `logic.if` with `data.expr` (jq), `notify.slack`, `notify.email`, `notify.webhook`.
- IF evaluation uses gojq on `.payload` (body_json); non-JSON payload → treat as false.
- Template rendering via `text/template` + sprig with context:
  - `.payload` (parsed JSON)
  - `.headers` (request headers map)
  - `.meta` (eventId, runId, inboxId, receivedAt)
- Runner creates `runs` per automation and `deliveries` per notify node; enqueue delivery tasks.

## Queue + retries
- Asynq tasks per delivery.
- Exponential backoff, `attempts` recorded; on final failure mark `status=failed`.
- Webhook deliveries include header `X-Aux-Signature: sha256=<hmac(secret, raw_body)>`.

## Example automation JSON
```json
{
  "nodes": [
    {"id": "n1", "type": "trigger.webhook", "data": {}, "position": {"x":0,"y":0}},
    {"id": "n2", "type": "logic.if", "data": {"expr": ".payload.ok == true"}, "position": {"x":200,"y":0}},
    {"id": "n3", "type": "notify.slack", "data": {"destination": "https://hooks.slack.com/..." , "template": "Run {{.meta.eventId}} OK"} , "position": {"x":400,"y":-50}},
    {"id": "n4", "type": "notify.webhook", "data": {"url": "https://example.com/callback", "secret": "abc", "template": "{\"id\":\"{{.meta.eventId}}\"}"} , "position": {"x":400,"y":100}}
  ],
  "edges": [
    {"id": "e1", "source": "n1", "target": "n2"},
    {"id": "e2", "source": "n2", "target": "n3", "label": "true"},
    {"id": "e3", "source": "n2", "target": "n4", "label": "false"}
  ]
}
```

## Example curl flow
- Create project (get api_key once):
  - `curl -X POST http://localhost:8080/v1/projects`
- Create inbox:
  - `curl -H "X-API-Key: <api_key>" -X POST http://localhost:8080/v1/inboxes`
- Create automation:
  - `curl -H "X-API-Key: <api_key>" -H "Content-Type: application/json" -d @automation.json http://localhost:8080/v1/automations`
- Ingest event:
  - `curl -X POST http://localhost:8080/i/<token> -H "Content-Type: application/json" -H "Idempotency-Key: 123" -d '{"hello":"world"}'`
- Replay event:
  - `curl -H "X-API-Key: <api_key>" -X POST http://localhost:8080/v1/events/<event_id>/replay`

## Testing
- Run all tests: `go test ./...`
- Focused: add unit tests for workflow runner paths and webhook signature helper.

## Project layout
- `cmd/api`, `cmd/worker`: entrypoints.
- `internal/http`: router, middleware, handlers.
- `internal/store`: pgx implementations + models.
- `internal/workflow`: graph parsing, runner, templating, jq eval.
- `internal/queue`: task definitions and processor registration.
- `internal/connectors`: slack/webhook/email senders.
- `migrations`: goose SQL migrations.

## Operational notes
- Idempotency: `Idempotency-Key` header maps to `events.dedupe_key` (unique per inbox when set).
- Ownership: all lookups scoped by authenticated `project_id`.
- Logging: structured (zap) with request IDs.

## TODO / future
- Add docker-compose.yml with Postgres + Redis service definitions.
- Harden validation for automation graphs and templates.
- Expand tests for connectors and retry logic.
