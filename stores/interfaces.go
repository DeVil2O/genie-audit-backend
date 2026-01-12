package stores

import (
	"context"
	"genie-audit-backend/models"
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

// Storage interface aggregates all store operations.
type Storage interface {
	ProjectsStore
	InboxesStore
	AutomationsStore
	EventsStore
	RunsStore
	DeliveriesStore
}

type ProjectsStore interface {
	CreateProject(ctx context.Context, name, apiKey string) (*models.Project, error)
}

type InboxesStore interface {
	CreateInbox(ctx context.Context, projectID uuid.UUID, name, token string) (*models.Inbox, error)
	ListInboxes(ctx context.Context, projectID uuid.UUID) ([]models.Inbox, error)
	GetInboxByToken(ctx context.Context, token string) (*models.Inbox, error)
}

type AutomationsStore interface {
	CreateAutomation(ctx context.Context, projectID uuid.UUID, inboxID *uuid.UUID, name string, enabled bool, graph []byte) (*models.Automation, error)
	ListAutomations(ctx context.Context, projectID uuid.UUID) ([]models.Automation, error)
	GetAutomation(ctx context.Context, projectID, automationID uuid.UUID) (*models.Automation, error)
	UpdateAutomation(ctx context.Context, projectID, automationID uuid.UUID, name *string, enabled *bool, graph []byte) (*models.Automation, error)
}

type EventsStore interface {
	CreateEvent(ctx context.Context, inboxID uuid.UUID, source string, headers map[string][]string, body []byte) (*models.Event, error)
	ListEvents(ctx context.Context, inboxID uuid.UUID, limit int) ([]models.Event, error)
	GetEvent(ctx context.Context, eventID uuid.UUID) (*models.Event, error)
}

type RunsStore interface {
	CreateRun(ctx context.Context, automationID, eventID uuid.UUID, status string) (*models.Run, error)
	ListRuns(ctx context.Context, automationID uuid.UUID) ([]models.Run, error)
}

type DeliveriesStore interface {
	CreateDelivery(ctx context.Context, runID uuid.UUID, nodeID, channel, status string) (*models.Delivery, error)
	ListDeliveries(ctx context.Context, runID uuid.UUID) ([]models.Delivery, error)
}
