package service

import (
	"context"
	"encoding/json"
	"genie-audit-backend/models"

	"github.com/google/uuid"
)

type ServiceInterface interface {
	CreateProject(ctx context.Context, name string) (*models.Project, string, error)
	CreateInbox(ctx context.Context, projectID uuid.UUID, name string) (*models.Inbox, error)
	ListInboxes(ctx context.Context, projectID uuid.UUID) ([]models.Inbox, error)
	CreateAutomation(ctx context.Context, projectID uuid.UUID, inboxID *uuid.UUID, name string, enabled bool, graph json.RawMessage) (*models.Automation, error)
	ListAutomations(ctx context.Context, projectID uuid.UUID) ([]models.Automation, error)
}
