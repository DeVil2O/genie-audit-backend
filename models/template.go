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
	GraphJSON          json.RawMessage `json:"graphJson"`
	CreatedByProjectID *uuid.UUID      `json:"createdByProjectId,omitempty"`
	Visibility         string          `json:"visibility"`
	Version            int             `json:"version"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
}
