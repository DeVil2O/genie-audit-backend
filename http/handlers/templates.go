package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"genie-audit-backend/logger"
	"genie-audit-backend/stores"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var allowedVisibility = map[string]struct{}{
	"private": {},
	"public":  {},
	"builtin": {},
}

// TemplateHandler handles automation template routes.
type TemplateHandler struct {
	store *stores.TemplateStore
}

// NewTemplateHandler constructs a new handler.
func NewTemplateHandler(store *stores.TemplateStore) *TemplateHandler {
	return &TemplateHandler{store: store}
}

type createTemplateRequest struct {
	Name               string          `json:"name"`
	Description        *string         `json:"description"`
	Tags               []string        `json:"tags"`
	GraphJSON          json.RawMessage `json:"graph_json"`
	CreatedByProjectID *string         `json:"created_by_project_id"`
	Visibility         string          `json:"visibility"`
	Version            *int            `json:"version"`
}

type updateTemplateRequest struct {
	Name               *string          `json:"name"`
	Description        *string          `json:"description"`
	Tags               *[]string        `json:"tags"`
	GraphJSON          *json.RawMessage `json:"graph_json"`
	Visibility         *string          `json:"visibility"`
	CreatedByProjectID *string          `json:"created_by_project_id"`
}

// CreateTemplate inserts a new template row.
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req createTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validateCreateTemplateRequest(req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	graph := []byte(req.GraphJSON)
	projectID, err := parseUUIDPtr(req.CreatedByProjectID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid created_by_project_id")
		return
	}

	version := 1
	if req.Version != nil && *req.Version > 0 {
		version = *req.Version
	}
	visibility := normalizeVisibility(req.Visibility)

	template, err := h.store.CreateTemplate(c.Request.Context(), stores.CreateTemplateParams{
		Name:               strings.TrimSpace(req.Name),
		Description:        req.Description,
		Tags:               normalizeTags(req.Tags),
		GraphJSON:          graph,
		CreatedByProjectID: projectID,
		Visibility:         visibility,
		Version:            version,
	})
	if err != nil {
		logger.Log.ErrorfWithContext(c.Request.Context(), "create template failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to create template")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"template": template})
}

// ListTemplates returns templates filtered by visibility/project/tag.
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	visibility := strings.TrimSpace(c.Query("visibility"))
	if visibility != "" && !isAllowedVisibility(visibility) {
		respondError(c, http.StatusBadRequest, "invalid visibility filter")
		return
	}
	projectID, err := parseUUIDPtrFromQuery(c.Query("project_id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid project_id filter")
		return
	}
	tag := strings.TrimSpace(c.Query("tag"))

	templates, err := h.store.ListTemplates(c.Request.Context(), stores.ListTemplatesParams{
		Visibility: visibility,
		ProjectID:  projectID,
		Tag:        tag,
	})
	if err != nil {
		logger.Log.ErrorfWithContext(c.Request.Context(), "list templates failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to list templates")
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// GetTemplate fetches one template.
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid template id")
		return
	}

	template, err := h.store.GetTemplate(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "template not found")
			return
		}
		logger.Log.ErrorfWithContext(c.Request.Context(), "get template failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to fetch template")
		return
	}
	c.JSON(http.StatusOK, gin.H{"template": template})
}

// UpdateTemplate partially updates template and bumps version.
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid template id")
		return
	}

	var req updateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	params := stores.UpdateTemplateParams{}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			respondError(c, http.StatusBadRequest, "name cannot be empty")
			return
		}
		params.Name = &name
	}
	if req.Description != nil {
		params.Description = req.Description
	}
	if req.Tags != nil {
		tags := normalizeTags(*req.Tags)
		if err := validateTags(tags); err != nil {
			respondError(c, http.StatusBadRequest, err.Error())
			return
		}
		params.Tags = &tags
	}
	if req.GraphJSON != nil {
		if len(*req.GraphJSON) == 0 || !json.Valid(*req.GraphJSON) {
			respondError(c, http.StatusBadRequest, "graph_json must be valid JSON")
			return
		}
		graph := []byte(*req.GraphJSON)
		params.GraphJSON = &graph
	}
	if req.Visibility != nil {
		visibility := normalizeVisibility(*req.Visibility)
		if !isAllowedVisibility(visibility) {
			respondError(c, http.StatusBadRequest, "invalid visibility")
			return
		}
		params.Visibility = &visibility
	}
	if req.CreatedByProjectID != nil {
		projectID, parseErr := parseUUIDPtr(req.CreatedByProjectID)
		if parseErr != nil {
			respondError(c, http.StatusBadRequest, "invalid created_by_project_id")
			return
		}
		params.CreatedByProjectID = projectID
	}

	template, err := h.store.UpdateTemplate(c.Request.Context(), id, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "template not found")
			return
		}
		logger.Log.ErrorfWithContext(c.Request.Context(), "update template failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to update template")
		return
	}
	c.JSON(http.StatusOK, gin.H{"template": template})
}

func validateCreateTemplateRequest(req createTemplateRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if len(req.GraphJSON) == 0 || !json.Valid(req.GraphJSON) {
		return errors.New("graph_json must be valid JSON")
	}
	if err := validateTags(req.Tags); err != nil {
		return err
	}
	if req.Visibility != "" && !isAllowedVisibility(req.Visibility) {
		return errors.New("invalid visibility")
	}
	return nil
}

func normalizeTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func validateTags(tags []string) error {
	if len(tags) > 30 {
		return errors.New("too many tags (max 30)")
	}
	for _, t := range tags {
		if len(t) > 64 {
			return errors.New("tag length exceeds 64 characters")
		}
	}
	return nil
}

func isAllowedVisibility(v string) bool {
	_, ok := allowedVisibility[v]
	return ok
}

func normalizeVisibility(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "private"
	}
	if isAllowedVisibility(v) {
		return v
	}
	return "private"
}

func parseUUIDPtr(value *string) (*uuid.UUID, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseUUIDPtrFromQuery(value string) (*uuid.UUID, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func respondError(c *gin.Context, status int, message string) {
	traceID := ""
	if t, ok := c.Get(string(logger.TraceIDKey)); ok {
		traceID, _ = t.(string)
	}
	c.JSON(status, gin.H{
		"message":  message,
		"trace_id": traceID,
	})
}
