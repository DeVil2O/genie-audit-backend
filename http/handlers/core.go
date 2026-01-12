package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"genie-audit-backend/logger"
	"genie-audit-backend/models"
	"genie-audit-backend/service"
)

const (
	defaultEventsListLimit = 50
)

// CoreHandler holds delivery-layer handlers with validation.
type CoreHandler struct {
	svc *service.Service
}

func NewCoreHandler(svc *service.Service) *CoreHandler {
	return &CoreHandler{svc: svc}
}

// Public ingest endpoint.
func (h *CoreHandler) IngestEvent(c *gin.Context) {
	token := c.Param("token")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondError(c, http.StatusBadRequest, "failed to read body")
		return
	}
	logger.Log.Infof("ingest request token=%s", token)
	event, runs, deliveries, err := h.svc.IngestEvent(c.Request.Context(), token, c.Request.Header, body)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, models.IngestResponse{
		Event:      event,
		Runs:       runs,
		Deliveries: deliveries,
	})
}

// Projects
func (h *CoreHandler) CreateProject(c *gin.Context) {
	var req models.ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid body")
		return
	}
	project, apiKey, err := h.svc.CreateProject(c.Request.Context(), req.Name)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	logger.Log.Infof("created project id=%s", project.ID)
	c.JSON(http.StatusCreated, models.ProjectResponse{Project: project, APIKey: apiKey})
}

func (h *CoreHandler) CreateInbox(c *gin.Context) {
	var req models.InboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid body")
		return
	}
	pid, err := uuid.Parse(req.ProjectID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid project_id")
		return
	}
	inbox, err := h.svc.CreateInbox(c.Request.Context(), pid, req.Name)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	logger.Log.Infof("created inbox id=%s project=%s", inbox.ID, req.ProjectID)
	c.JSON(http.StatusCreated, models.InboxResponse{
		Inbox:     inbox,
		IngestURL: "/i/" + inbox.Token,
	})
}

func (h *CoreHandler) ListInboxes(c *gin.Context) {
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		respondError(c, http.StatusBadRequest, "project_id is required")
		return
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid project_id")
		return
	}
	inboxes, err := h.svc.ListInboxes(c.Request.Context(), pid)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"inboxes": inboxes})
}

// Automations
func (h *CoreHandler) CreateAutomation(c *gin.Context) {
	var req models.AutomationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid body")
		return
	}
	pid, err := uuid.Parse(req.ProjectID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid project_id")
		return
	}
	var inboxID *uuid.UUID
	if req.InboxID != nil && strings.TrimSpace(*req.InboxID) != "" {
		parsed, err := uuid.Parse(*req.InboxID)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid inbox_id")
			return
		}
		inboxID = &parsed
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	auto, err := h.svc.CreateAutomation(c.Request.Context(), pid, inboxID, req.Name, enabled, req.GraphJSON)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	logger.Log.Infof("created automation id=%s project=%s", auto.ID, req.ProjectID)
	c.JSON(http.StatusCreated, models.AutomationResponse{Automation: auto})
}

func (h *CoreHandler) ListAutomations(c *gin.Context) {
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		respondError(c, http.StatusBadRequest, "project_id is required")
		return
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid project_id")
		return
	}
	autos, err := h.svc.ListAutomations(c.Request.Context(), pid)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, models.AutomationsResponse{Automations: autos})
}

func (h *CoreHandler) GetAutomation(c *gin.Context) {
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		respondError(c, http.StatusBadRequest, "project_id is required")
		return
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid project_id")
		return
	}
	aid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid automation id")
		return
	}
	auto, err := h.svc.GetAutomation(c.Request.Context(), pid, aid)
	if err != nil {
		respondError(c, http.StatusNotFound, "not found")
		return
	}
	c.JSON(http.StatusOK, models.AutomationResponse{Automation: auto})
}

func (h *CoreHandler) UpdateAutomation(c *gin.Context) {
	var req models.AutomationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid body")
		return
	}
	pid, err := uuid.Parse(req.ProjectID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid project_id")
		return
	}
	aid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid automation id")
		return
	}
	auto, err := h.svc.UpdateAutomation(c.Request.Context(), pid, aid, req.Name, req.Enabled, req.GraphJSON)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, models.AutomationResponse{Automation: auto})
}

// Logs / Events
func (h *CoreHandler) ListEvents(c *gin.Context) {
	inboxID := strings.TrimSpace(c.Query("inboxId"))
	if inboxID == "" {
		respondError(c, http.StatusBadRequest, "inboxId is required")
		return
	}
	iid, err := uuid.Parse(inboxID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid inboxId")
		return
	}
	limit := 50
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	events, err := h.svc.ListEvents(c.Request.Context(), iid, limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, models.EventsResponse{Events: events})
}

func (h *CoreHandler) GetEvent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid event id")
		return
	}
	e, err := h.svc.GetEvent(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusNotFound, "event not found")
		return
	}
	c.JSON(http.StatusOK, models.EventResponse{Event: e})
}

func (h *CoreHandler) ReplayEvent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid event id")
		return
	}
	runs, err := h.svc.ReplayEvent(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusAccepted, models.RunsResponse{Runs: runs})
}

// Runs
func (h *CoreHandler) ListRuns(c *gin.Context) {
	automationID := strings.TrimSpace(c.Query("automationId"))
	if automationID == "" {
		respondError(c, http.StatusBadRequest, "automationId is required")
		return
	}
	aid, err := uuid.Parse(automationID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid automationId")
		return
	}
	runs, err := h.svc.ListRuns(c.Request.Context(), aid)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, models.RunsResponse{Runs: runs})
}

// Deliveries
func (h *CoreHandler) ListDeliveries(c *gin.Context) {
	runID := strings.TrimSpace(c.Query("runId"))
	if runID == "" {
		respondError(c, http.StatusBadRequest, "runId is required")
		return
	}
	rid, err := uuid.Parse(runID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid runId")
		return
	}
	deliveries, err := h.svc.ListDeliveries(c.Request.Context(), rid)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, models.DeliveriesResponse{Deliveries: deliveries})
}
