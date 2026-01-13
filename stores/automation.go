package stores

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"genie-audit-backend/models"
)

// DBStore implements Storage backed by Postgres.
type DBStore struct {
	pool *pgxpool.Pool
}

// NewDBStore constructs a Postgres-backed store.
func NewDBStore(pool *pgxpool.Pool) *DBStore {
	return &DBStore{pool: pool}
}

func (s *DBStore) CreateProject(ctx context.Context, name, apiKey string) (*models.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	apiKeyHash := sha256.Sum256([]byte(apiKey))

	var p models.Project
	row := s.pool.QueryRow(ctx, `
		INSERT INTO projects (name, api_key_hash)
		VALUES ($1, $2)
		RETURNING id, name, created_at
	`, name, apiKeyHash[:])
	if err := row.Scan(&p.ID, &p.Name, &p.CreatedAt); err != nil {
		return nil, err
	}
	p.APIKey = apiKey
	return &p, nil
}

func (s *DBStore) CreateInbox(ctx context.Context, projectID uuid.UUID, name, token string) (*models.Inbox, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var inbox models.Inbox
	row := s.pool.QueryRow(ctx, `
		INSERT INTO inboxes (project_id, token, name)
		VALUES ($1, $2, $3)
		RETURNING id, project_id, token, name, created_at
	`, projectID, token, name)
	if err := row.Scan(&inbox.ID, &inbox.ProjectID, &inbox.Token, &inbox.Name, &inbox.CreatedAt); err != nil {
		return nil, err
	}
	return &inbox, nil
}

func (s *DBStore) ListInboxes(ctx context.Context, projectID uuid.UUID) ([]models.Inbox, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
		SELECT id, project_id, token, name, created_at
		FROM inboxes
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Inbox
	for rows.Next() {
		var i models.Inbox
		if err := rows.Scan(&i.ID, &i.ProjectID, &i.Token, &i.Name, &i.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (s *DBStore) GetInboxByToken(ctx context.Context, token string) (*models.Inbox, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := s.pool.QueryRow(ctx, `
		SELECT id, project_id, token, name, created_at
		FROM inboxes
		WHERE token = $1
	`, token)

	var inbox models.Inbox
	if err := row.Scan(&inbox.ID, &inbox.ProjectID, &inbox.Token, &inbox.Name, &inbox.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &inbox, nil
}

func (s *DBStore) CreateAutomation(ctx context.Context, projectID uuid.UUID, inboxID *uuid.UUID, name string, enabled bool, graph []byte) (*models.Automation, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := s.pool.QueryRow(ctx, `
		INSERT INTO automations (project_id, name, enabled, graph_json, trigger_type, inbox_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, project_id, inbox_id, name, enabled, graph_json, created_at, updated_at
	`, projectID, name, enabled, graph, "webhook", inboxID)

	return scanAutomation(row)
}

func (s *DBStore) ListAutomations(ctx context.Context, projectID uuid.UUID) ([]models.Automation, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
		SELECT id, project_id, inbox_id, name, enabled, graph_json, created_at, updated_at
		FROM automations
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var autos []models.Automation
	for rows.Next() {
		a, err := scanAutomation(rows)
		if err != nil {
			return nil, err
		}
		autos = append(autos, *a)
	}
	return autos, rows.Err()
}

func (s *DBStore) GetAutomation(ctx context.Context, projectID, automationID uuid.UUID) (*models.Automation, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := s.pool.QueryRow(ctx, `
		SELECT id, project_id, inbox_id, name, enabled, graph_json, created_at, updated_at
		FROM automations
		WHERE project_id = $1 AND id = $2
	`, projectID, automationID)

	a, err := scanAutomation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return a, nil
}

func (s *DBStore) UpdateAutomation(ctx context.Context, projectID, automationID uuid.UUID, name *string, enabled *bool, graph []byte) (*models.Automation, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	if name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *name)
		argPos++
	}
	if enabled != nil {
		setClauses = append(setClauses, fmt.Sprintf("enabled = $%d", argPos))
		args = append(args, *enabled)
		argPos++
	}
	if graph != nil {
		setClauses = append(setClauses, fmt.Sprintf("graph_json = $%d", argPos))
		args = append(args, graph)
		argPos++
	}

	if len(setClauses) == 0 {
		return s.GetAutomation(ctx, projectID, automationID)
	}

	setClauses = append(setClauses, "updated_at = now()")

	projectPlaceholder := fmt.Sprintf("$%d", argPos)
	args = append(args, projectID)
	argPos++
	automationPlaceholder := fmt.Sprintf("$%d", argPos)
	args = append(args, automationID)

	query := fmt.Sprintf(`
		UPDATE automations
		SET %s
		WHERE project_id = %s AND id = %s
		RETURNING id, project_id, inbox_id, name, enabled, graph_json, created_at, updated_at
	`, strings.Join(setClauses, ", "), projectPlaceholder, automationPlaceholder)

	row := s.pool.QueryRow(ctx, query, args...)
	auto, err := scanAutomation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return auto, nil
}

func (s *DBStore) CreateEvent(ctx context.Context, inboxID uuid.UUID, source string, headers map[string][]string, body []byte) (*models.Event, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	headersJSON, err := json.Marshal(headers)
	if err != nil {
		return nil, err
	}

	var bodyJSON interface{}
	if json.Valid(body) {
		bodyJSON = body
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO events (inbox_id, source, headers, body_raw, body_json)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, inbox_id, source, headers, body_raw, body_json, created_at
	`, inboxID, source, headersJSON, body, bodyJSON)

	return scanEvent(row)
}

func (s *DBStore) ListEvents(ctx context.Context, inboxID uuid.UUID, limit int) ([]models.Event, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
		SELECT id, inbox_id, source, headers, body_raw, body_json, created_at
		FROM events
		WHERE inbox_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, inboxID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *e)
	}
	return events, rows.Err()
}

func (s *DBStore) GetEvent(ctx context.Context, eventID uuid.UUID) (*models.Event, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := s.pool.QueryRow(ctx, `
		SELECT id, inbox_id, source, headers, body_raw, body_json, created_at
		FROM events
		WHERE id = $1
	`, eventID)

	e, err := scanEvent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return e, nil
}

func (s *DBStore) CreateRun(ctx context.Context, automationID, eventID uuid.UUID, status string) (*models.Run, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var run models.Run
	row := s.pool.QueryRow(ctx, `
		INSERT INTO runs (automation_id, event_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, automation_id, event_id, status, created_at
	`, automationID, eventID, status)
	if err := row.Scan(&run.ID, &run.AutomationID, &run.EventID, &run.Status, &run.CreatedAt); err != nil {
		return nil, err
	}
	return &run, nil
}

func (s *DBStore) ListRuns(ctx context.Context, automationID uuid.UUID) ([]models.Run, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
		SELECT id, automation_id, event_id, status, created_at
		FROM runs
		WHERE automation_id = $1
		ORDER BY created_at DESC
	`, automationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []models.Run
	for rows.Next() {
		var r models.Run
		if err := rows.Scan(&r.ID, &r.AutomationID, &r.EventID, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

func (s *DBStore) CreateDelivery(ctx context.Context, runID uuid.UUID, nodeID, channel, status string) (*models.Delivery, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var d models.Delivery
	row := s.pool.QueryRow(ctx, `
		INSERT INTO deliveries (run_id, node_id, channel, destination, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, run_id, node_id, channel, status, created_at
	`, runID, nodeID, channel, nodeID, status)
	if err := row.Scan(&d.ID, &d.RunID, &d.NodeID, &d.Channel, &d.Status, &d.CreatedAt); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *DBStore) ListDeliveries(ctx context.Context, runID uuid.UUID) ([]models.Delivery, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
		SELECT id, run_id, node_id, channel, status, created_at
		FROM deliveries
		WHERE run_id = $1
		ORDER BY created_at DESC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ds []models.Delivery
	for rows.Next() {
		var d models.Delivery
		if err := rows.Scan(&d.ID, &d.RunID, &d.NodeID, &d.Channel, &d.Status, &d.CreatedAt); err != nil {
			return nil, err
		}
		ds = append(ds, d)
	}
	return ds, rows.Err()
}

func scanAutomation(row pgx.Row) (*models.Automation, error) {
	var a models.Automation
	var inboxID *uuid.UUID
	if err := row.Scan(&a.ID, &a.ProjectID, &inboxID, &a.Name, &a.Enabled, &a.GraphJSON, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, err
	}
	a.InboxID = inboxID
	return &a, nil
}

func scanEvent(row pgx.Row) (*models.Event, error) {
	var e models.Event
	var headersBytes []byte
	var bodyJSONBytes []byte
	if err := row.Scan(&e.ID, &e.InboxID, &e.Source, &headersBytes, &e.BodyRaw, &bodyJSONBytes, &e.CreatedAt); err != nil {
		return nil, err
	}
	if len(headersBytes) > 0 {
		if err := json.Unmarshal(headersBytes, &e.Headers); err != nil {
			return nil, err
		}
	}
	if len(bodyJSONBytes) > 0 {
		tmp := json.RawMessage(bodyJSONBytes)
		e.BodyJSON = &tmp
	}
	return &e, nil
}
