package stores

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"genie-audit-backend/models"
)

// MemoryStore is an in-memory implementation of Storage for MVP/testing.
// Not safe for multi-process; intended for single-node local dev.
type MemoryStore struct {
	mu          sync.RWMutex
	projects    map[uuid.UUID]models.Project
	inboxes     map[uuid.UUID]models.Inbox
	inboxesByTk map[string]uuid.UUID
	automations map[uuid.UUID]models.Automation
	events      map[uuid.UUID]models.Event
	runs        map[uuid.UUID]models.Run
	deliveries  map[uuid.UUID]models.Delivery
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		projects:    make(map[uuid.UUID]models.Project),
		inboxes:     make(map[uuid.UUID]models.Inbox),
		inboxesByTk: make(map[string]uuid.UUID),
		automations: make(map[uuid.UUID]models.Automation),
		events:      make(map[uuid.UUID]models.Event),
		runs:        make(map[uuid.UUID]models.Run),
		deliveries:  make(map[uuid.UUID]models.Delivery),
	}
}

func (s *MemoryStore) CreateProject(ctx context.Context, name, apiKey string) (*models.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New()
	now := time.Now().UTC()
	p := models.Project{ID: id, Name: name, APIKey: apiKey, CreatedAt: now}
	s.projects[id] = p
	return &p, nil
}

func (s *MemoryStore) CreateInbox(ctx context.Context, projectID uuid.UUID, name, token string) (*models.Inbox, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New()
	now := time.Now().UTC()
	i := models.Inbox{ID: id, ProjectID: projectID, Name: name, Token: token, CreatedAt: now}
	s.inboxes[id] = i
	s.inboxesByTk[token] = id
	return &i, nil
}

func (s *MemoryStore) ListInboxes(ctx context.Context, projectID uuid.UUID) ([]models.Inbox, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []models.Inbox
	for _, i := range s.inboxes {
		if i.ProjectID == projectID {
			out = append(out, i)
		}
	}
	return out, nil
}

func (s *MemoryStore) GetInboxByToken(ctx context.Context, token string) (*models.Inbox, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.inboxesByTk[token]
	if !ok {
		return nil, errors.New("not found")
	}
	inbox := s.inboxes[id]
	return &inbox, nil
}

func (s *MemoryStore) CreateAutomation(ctx context.Context, projectID uuid.UUID, inboxID *uuid.UUID, name string, enabled bool, graph []byte) (*models.Automation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New()
	now := time.Now().UTC()
	a := models.Automation{
		ID:        id,
		ProjectID: projectID,
		InboxID:   inboxID,
		Name:      name,
		Enabled:   enabled,
		GraphJSON: graph,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.automations[id] = a
	return &a, nil
}

func (s *MemoryStore) ListAutomations(ctx context.Context, projectID uuid.UUID) ([]models.Automation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []models.Automation
	for _, a := range s.automations {
		if a.ProjectID == projectID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (s *MemoryStore) GetAutomation(ctx context.Context, projectID, automationID uuid.UUID) (*models.Automation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.automations[automationID]
	if !ok || a.ProjectID != projectID {
		return nil, errors.New("not found")
	}
	return &a, nil
}

func (s *MemoryStore) UpdateAutomation(ctx context.Context, projectID, automationID uuid.UUID, name *string, enabled *bool, graph []byte) (*models.Automation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.automations[automationID]
	if !ok || a.ProjectID != projectID {
		return nil, errors.New("not found")
	}
	if name != nil {
		a.Name = *name
	}
	if enabled != nil {
		a.Enabled = *enabled
	}
	if graph != nil {
		a.GraphJSON = graph
	}
	a.UpdatedAt = time.Now().UTC()
	s.automations[automationID] = a
	return &a, nil
}

func (s *MemoryStore) CreateEvent(ctx context.Context, inboxID uuid.UUID, source string, headers map[string][]string, body []byte) (*models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New()
	now := time.Now().UTC()
	e := models.Event{
		ID:        id,
		InboxID:   inboxID,
		Source:    source,
		Headers:   headers,
		BodyRaw:   body,
		CreatedAt: now,
	}
	s.events[id] = e
	return &e, nil
}

func (s *MemoryStore) ListEvents(ctx context.Context, inboxID uuid.UUID, limit int) ([]models.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []models.Event
	for _, e := range s.events {
		if e.InboxID == inboxID {
			out = append(out, e)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *MemoryStore) GetEvent(ctx context.Context, eventID uuid.UUID) (*models.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.events[eventID]
	if !ok {
		return nil, errors.New("not found")
	}
	return &e, nil
}

func (s *MemoryStore) CreateRun(ctx context.Context, automationID, eventID uuid.UUID, status string) (*models.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New()
	now := time.Now().UTC()
	r := models.Run{ID: id, AutomationID: automationID, EventID: eventID, Status: status, CreatedAt: now}
	s.runs[id] = r
	return &r, nil
}

func (s *MemoryStore) ListRuns(ctx context.Context, automationID uuid.UUID) ([]models.Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []models.Run
	for _, r := range s.runs {
		if r.AutomationID == automationID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *MemoryStore) CreateDelivery(ctx context.Context, runID uuid.UUID, nodeID, channel, status string) (*models.Delivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New()
	now := time.Now().UTC()
	d := models.Delivery{ID: id, RunID: runID, NodeID: nodeID, Channel: channel, Status: status, CreatedAt: now}
	s.deliveries[id] = d
	return &d, nil
}

func (s *MemoryStore) ListDeliveries(ctx context.Context, runID uuid.UUID) ([]models.Delivery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []models.Delivery
	for _, d := range s.deliveries {
		if d.RunID == runID {
			out = append(out, d)
		}
	}
	return out, nil
}
