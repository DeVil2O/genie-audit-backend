package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"

	"genie-audit-backend/logger"
	"genie-audit-backend/models"
	"genie-audit-backend/stores"
)

// Service contains business logic sitting atop the storage layer.
type Service struct {
	store stores.Storage
}

func NewService(store stores.Storage) *Service {
	return &Service{store: store}
}

func (s *Service) CreateProject(ctx context.Context, name string) (*models.Project, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, "", errors.New("name is required")
	}
	apiKey := uuid.NewString()
	project, err := s.store.CreateProject(ctx, name, apiKey)
	if err != nil {
		return nil, "", err
	}
	return project, apiKey, nil
}

func (s *Service) CreateInbox(ctx context.Context, projectID uuid.UUID, name string) (*models.Inbox, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	token := uuid.NewString()
	inbox, err := s.store.CreateInbox(ctx, projectID, name, token)
	if err == nil {
		logger.Log.Infof("created inbox id=%s for project=%s", inbox.ID, projectID)
	}
	return inbox, err
}

func (s *Service) ListInboxes(ctx context.Context, projectID uuid.UUID) ([]models.Inbox, error) {
	return s.store.ListInboxes(ctx, projectID)
}

func (s *Service) CreateAutomation(ctx context.Context, projectID uuid.UUID, inboxID *uuid.UUID, name string, enabled bool, graph json.RawMessage) (*models.Automation, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	if len(graph) == 0 || !json.Valid(graph) {
		return nil, errors.New("graph_json must be valid JSON")
	}
	auto, err := s.store.CreateAutomation(ctx, projectID, inboxID, name, enabled, graph)
	if err == nil {
		logger.Log.Infof("created automation id=%s project=%s", auto.ID, projectID)
	}
	return auto, err
}

func (s *Service) ListAutomations(ctx context.Context, projectID uuid.UUID) ([]models.Automation, error) {
	return s.store.ListAutomations(ctx, projectID)
}

func (s *Service) GetAutomation(ctx context.Context, projectID, automationID uuid.UUID) (*models.Automation, error) {
	return s.store.GetAutomation(ctx, projectID, automationID)
}

func (s *Service) UpdateAutomation(ctx context.Context, projectID, automationID uuid.UUID, name *string, enabled *bool, graph *json.RawMessage) (*models.Automation, error) {
	var graphBytes []byte
	if graph != nil {
		if len(*graph) == 0 || !json.Valid(*graph) {
			return nil, errors.New("graph_json must be valid JSON")
		}
		graphBytes = *graph
	}
	return s.store.UpdateAutomation(ctx, projectID, automationID, name, enabled, graphBytes)
}

func (s *Service) IngestEvent(ctx context.Context, token string, headers map[string][]string, body []byte) (*models.Event, []models.Run, []models.Delivery, error) {
	inbox, err := s.store.GetInboxByToken(ctx, token)
	if err != nil {
		return nil, nil, nil, errors.New("inbox not found")
	}
	event, err := s.store.CreateEvent(ctx, inbox.ID, "web2", headers, body)
	if err != nil {
		return nil, nil, nil, err
	}

	autos, err := s.store.ListAutomations(ctx, inbox.ProjectID)
	if err != nil {
		return event, nil, nil, err
	}
	var runs []models.Run
	var deliveries []models.Delivery
	for _, a := range autos {
		if a.InboxID != nil && *a.InboxID != inbox.ID {
			continue
		}
		if !a.Enabled {
			continue
		}
		run, err := s.store.CreateRun(ctx, a.ID, event.ID, "queued")
		if err != nil {
			return event, runs, deliveries, err
		}
		runs = append(runs, *run)
		// For MVP, create a placeholder delivery record representing graph execution.
		d, err := s.store.CreateDelivery(ctx, run.ID, "node-1", "webhook", "pending")
		if err != nil {
			return event, runs, deliveries, err
		}
		deliveries = append(deliveries, *d)
	}
	logger.Log.Infof("ingested event id=%s inbox=%s runs=%d deliveries=%d", event.ID, inbox.ID, len(runs), len(deliveries))
	return event, runs, deliveries, nil
}

func (s *Service) ListEvents(ctx context.Context, inboxID uuid.UUID, limit int) ([]models.Event, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.store.ListEvents(ctx, inboxID, limit)
}

func (s *Service) GetEvent(ctx context.Context, eventID uuid.UUID) (*models.Event, error) {
	return s.store.GetEvent(ctx, eventID)
}

func (s *Service) ReplayEvent(ctx context.Context, eventID uuid.UUID) ([]models.Run, error) {
	// For in-memory MVP, simply return empty and rely on ingest path to enqueue runs.
	_, err := s.store.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return []models.Run{}, nil
}

func (s *Service) ListRuns(ctx context.Context, automationID uuid.UUID) ([]models.Run, error) {
	return s.store.ListRuns(ctx, automationID)
}

func (s *Service) ListDeliveries(ctx context.Context, runID uuid.UUID) ([]models.Delivery, error) {
	return s.store.ListDeliveries(ctx, runID)
}
