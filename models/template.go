package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AutomationTemplate represents a stored workflow template.
type AutomationTemplate struct {
	ID                 uuid.UUID       `json:"id"`
	Name               string          `json:"name"`
	Description        *string         `json:"description,omitempty"`
	Tags               []string        `json:"tags"`
	GraphJSON          json.RawMessage `json:"graph_json"`
	CreatedByProjectID *uuid.UUID      `json:"created_by_project_id,omitempty"`
	Visibility         string          `json:"visibility"`
	Version            int             `json:"version"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}
