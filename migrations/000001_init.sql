-- +goose Up
-- Initial schema for projects, inboxes, automations, events, runs, and deliveries.
-- Postgres-specific (uses gen_random_uuid and jsonb).

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE automation_trigger_type AS ENUM ('webhook', 'manual', 'cron', 'evm');
CREATE TYPE event_source AS ENUM ('web2', 'manual', 'cron', 'web3');
CREATE TYPE run_status AS ENUM ('queued', 'running', 'succeeded', 'failed');
CREATE TYPE delivery_channel AS ENUM ('slack', 'email', 'webhook');
CREATE TYPE delivery_status AS ENUM ('pending', 'sent', 'failed');

CREATE TABLE projects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    api_key_hash bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE inboxes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    token text NOT NULL UNIQUE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_inboxes_project_id ON inboxes(project_id);

CREATE TABLE automations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name text NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    graph_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    trigger_type automation_trigger_type NOT NULL,
    inbox_id uuid REFERENCES inboxes(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_automations_project_id ON automations(project_id);
CREATE INDEX idx_automations_inbox_id ON automations(inbox_id);

CREATE TABLE events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    inbox_id uuid NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
    source event_source NOT NULL,
    content_type text,
    headers jsonb NOT NULL DEFAULT '{}'::jsonb,
    body_raw bytea NOT NULL,
    body_json jsonb,
    dedupe_key text,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_events_inbox_id ON events(inbox_id);
CREATE INDEX idx_events_dedupe ON events(inbox_id, dedupe_key) WHERE dedupe_key IS NOT NULL;

CREATE TABLE runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id uuid NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
    event_id uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    status run_status NOT NULL DEFAULT 'queued',
    started_at timestamptz,
    finished_at timestamptz,
    error text,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_runs_automation_id ON runs(automation_id);
CREATE INDEX idx_runs_event_id ON runs(event_id);
CREATE INDEX idx_runs_status ON runs(status);

CREATE TABLE deliveries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    node_id text NOT NULL,
    channel delivery_channel NOT NULL,
    destination text NOT NULL,
    url text,
    secret text,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    status delivery_status NOT NULL DEFAULT 'pending',
    attempts integer NOT NULL DEFAULT 0,
    next_retry_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_deliveries_run_id ON deliveries(run_id);
CREATE INDEX idx_deliveries_status ON deliveries(status);
CREATE INDEX idx_deliveries_next_retry_at ON deliveries(next_retry_at);

-- +goose Down
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS automations;
DROP TABLE IF EXISTS inboxes;
DROP TABLE IF EXISTS projects;

DROP TYPE IF EXISTS delivery_status;
DROP TYPE IF EXISTS delivery_channel;
DROP TYPE IF EXISTS run_status;
DROP TYPE IF EXISTS event_source;
DROP TYPE IF EXISTS automation_trigger_type;
