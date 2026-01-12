-- +goose Up
-- Automation templates storage (React Flow graph JSON in jsonb).

CREATE TABLE automation_templates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    description text,
    tags text[] NOT NULL DEFAULT '{}'::text[],
    graph_json jsonb NOT NULL,
    created_by_project_id uuid,
    visibility text NOT NULL DEFAULT 'private' CHECK (visibility IN ('private', 'public', 'builtin')),
    version int NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_automation_templates_tags ON automation_templates USING GIN (tags);
CREATE INDEX idx_automation_templates_visibility_project ON automation_templates (visibility, created_by_project_id);

-- +goose Down
DROP TABLE IF EXISTS automation_templates;
