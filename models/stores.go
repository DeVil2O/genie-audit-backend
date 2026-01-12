package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID        uuid.UUID
	Name      string
	APIKey    string
	CreatedAt time.Time
}

type Inbox struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Token     string
	Name      string
	CreatedAt time.Time
}

type Automation struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	InboxID   *uuid.UUID
	Name      string
	Enabled   bool
	GraphJSON []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Event struct {
	ID        uuid.UUID
	InboxID   uuid.UUID
	Source    string
	Headers   map[string][]string
	BodyRaw   []byte
	BodyJSON  *json.RawMessage
	CreatedAt time.Time
}

type Run struct {
	ID           uuid.UUID
	AutomationID uuid.UUID
	EventID      uuid.UUID
	Status       string
	CreatedAt    time.Time
}

type Delivery struct {
	ID        uuid.UUID
	RunID     uuid.UUID
	NodeID    string
	Channel   string
	Status    string
	CreatedAt time.Time
}
