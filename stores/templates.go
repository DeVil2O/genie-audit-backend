package stores

import (
	"context"
	"fmt"
	"genie-audit-backend/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TemplateStore provides CRUD operations for automation templates.
type TemplateStore struct {
	pool *pgxpool.Pool
}

// NewTemplateStore constructs a TemplateStore.
func NewTemplateStore(pool *pgxpool.Pool) *TemplateStore {
	return &TemplateStore{pool: pool}
}

type CreateTemplateParams struct {
	Name               string
	Description        *string
	Tags               []string
	GraphJSON          []byte
	CreatedByProjectID *uuid.UUID
	Visibility         string
	Version            int
}

type ListTemplatesParams struct {
	Visibility string
	ProjectID  *uuid.UUID
	Tag        string
}

type UpdateTemplateParams struct {
	Name               *string
	Description        *string
	Tags               *[]string
	GraphJSON          *[]byte
	CreatedByProjectID *uuid.UUID
	Visibility         *string
}

func (s *TemplateStore) CreateTemplate(ctx context.Context, params CreateTemplateParams) (*models.AutomationTemplate, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := s.pool.QueryRow(ctx, `
		INSERT INTO automation_templates
		(name, description, tags, graph_json, created_by_project_id, visibility, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, description, tags, graph_json, created_by_project_id, visibility, version, created_at, updated_at
	`, params.Name, params.Description, params.Tags, params.GraphJSON, params.CreatedByProjectID, params.Visibility, params.Version)

	return scanTemplate(row)
}

func (s *TemplateStore) ListTemplates(ctx context.Context, params ListTemplatesParams) ([]models.AutomationTemplate, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	base := `
		SELECT id, name, description, tags, graph_json, created_by_project_id, visibility, version, created_at, updated_at
		FROM automation_templates
	`
	conds := []string{}
	args := []interface{}{}

	if params.Visibility != "" {
		args = append(args, params.Visibility)
		conds = append(conds, fmt.Sprintf("visibility = $%d", len(args)))
	}
	if params.ProjectID != nil {
		args = append(args, *params.ProjectID)
		conds = append(conds, fmt.Sprintf("created_by_project_id = $%d", len(args)))
	}
	if params.Tag != "" {
		args = append(args, params.Tag)
		conds = append(conds, fmt.Sprintf("tags @> ARRAY[$%d]::text[]", len(args)))
	}

	query := base
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.AutomationTemplate
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *t)
	}
	return results, rows.Err()
}

func (s *TemplateStore) GetTemplate(ctx context.Context, id uuid.UUID) (*models.AutomationTemplate, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := s.pool.QueryRow(ctx, `
		SELECT id, name, description, tags, graph_json, created_by_project_id, visibility, version, created_at, updated_at
		FROM automation_templates
		WHERE id = $1
	`, id)
	return scanTemplate(row)
}

func (s *TemplateStore) UpdateTemplate(ctx context.Context, id uuid.UUID, params UpdateTemplateParams) (*models.AutomationTemplate, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	setClauses := []string{}
	args := []interface{}{}

	if params.Name != nil {
		args = append(args, *params.Name)
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", len(args)))
	}
	if params.Description != nil {
		args = append(args, *params.Description)
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", len(args)))
	}
	if params.Tags != nil {
		args = append(args, *params.Tags)
		setClauses = append(setClauses, fmt.Sprintf("tags = $%d", len(args)))
	}
	if params.GraphJSON != nil {
		args = append(args, *params.GraphJSON)
		setClauses = append(setClauses, fmt.Sprintf("graph_json = $%d", len(args)))
	}
	if params.CreatedByProjectID != nil {
		args = append(args, *params.CreatedByProjectID)
		setClauses = append(setClauses, fmt.Sprintf("created_by_project_id = $%d", len(args)))
	}
	if params.Visibility != nil {
		args = append(args, *params.Visibility)
		setClauses = append(setClauses, fmt.Sprintf("visibility = $%d", len(args)))
	}

	if len(setClauses) == 0 {
		// No-op update; return current value.
		return s.GetTemplate(ctx, id)
	}

	// Always bump version and updated_at.
	setClauses = append(setClauses, "version = version + 1", "updated_at = now()")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE automation_templates
		SET %s
		WHERE id = $%d
		RETURNING id, name, description, tags, graph_json, created_by_project_id, visibility, version, created_at, updated_at
	`, strings.Join(setClauses, ", "), len(args))

	row := s.pool.QueryRow(ctx, query, args...)
	return scanTemplate(row)
}

func scanTemplate(row pgx.Row) (*models.AutomationTemplate, error) {
	var t models.AutomationTemplate
	err := row.Scan(
		&t.ID,
		&t.Name,
		&t.Description,
		&t.Tags,
		&t.GraphJSON,
		&t.CreatedByProjectID,
		&t.Visibility,
		&t.Version,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
