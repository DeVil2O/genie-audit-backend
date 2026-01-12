package models

import (
	"encoding/json"
)

// Requests
type ProjectRequest struct {
	Name string `json:"name"`
}

type InboxRequest struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}

type AutomationRequest struct {
	ProjectID string          `json:"project_id"`
	InboxID   *string         `json:"inbox_id"`
	Name      string          `json:"name"`
	Enabled   *bool           `json:"enabled"`
	GraphJSON json.RawMessage `json:"graph_json"`
}

type AutomationUpdateRequest struct {
	Name      *string          `json:"name"`
	Enabled   *bool            `json:"enabled"`
	GraphJSON *json.RawMessage `json:"graph_json"`
	ProjectID string           `json:"project_id"`
}

// Responses
type ProjectResponse struct {
	Project *Project `json:"project"`
	APIKey  string          `json:"api_key"`
}

type InboxResponse struct {
	Inbox     *Inbox `json:"inbox"`
	IngestURL string        `json:"ingest_url"`
}

type AutomationResponse struct {
	Automation *Automation `json:"automation"`
}

type AutomationsResponse struct {
	Automations []Automation `json:"automations"`
}	

type EventsResponse struct {
	Events []Event `json:"events"`
}

type EventResponse struct {
	Event *Event `json:"event"`
}

type RunsResponse struct {
	Runs []Run `json:"runs"`
}

type DeliveriesResponse struct {
	Deliveries []Delivery `json:"deliveries"`
}

type IngestResponse struct {
	Event      *Event     `json:"event"`
	Runs       []Run      `json:"runs"`
	Deliveries []Delivery `json:"deliveries"`
}
