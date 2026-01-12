package http

import (
	gin "github.com/gin-gonic/gin"

	"genie-audit-backend/http/handlers"
	"genie-audit-backend/middleware"
)

// RegisterTemplateRoutes binds template handlers to routes.
func RegisterTemplateRoutes(router gin.IRouter, h *handlers.TemplateHandler) {
	templates := router.Group("/templates")
	templates.POST("", h.CreateTemplate)
	templates.GET("", h.ListTemplates)
	templates.GET("/:id", h.GetTemplate)
	templates.PATCH("/:id", h.UpdateTemplate)
}

// RegisterCoreRoutes wires ingest + core MVP routes.
func RegisterCoreRoutes(engine *gin.Engine, core *handlers.CoreHandler, templates *handlers.TemplateHandler) {
	// Public ingest
	engine.POST("/i/:token", core.IngestEvent)

	v1 := engine.Group("/v1")
	private := v1.Group("")
	private.Use(middleware.Auth())

	private.POST("/projects", core.CreateProject)

	private.POST("/inboxes", core.CreateInbox)
	private.GET("/inboxes", core.ListInboxes)

	private.POST("/automations", core.CreateAutomation)
	private.GET("/automations", core.ListAutomations)
	private.GET("/automations/:id", core.GetAutomation)
	private.PATCH("/automations/:id", core.UpdateAutomation)

	private.GET("/events", core.ListEvents)
	private.GET("/events/:id", core.GetEvent)
	private.POST("/events/:id/replay", core.ReplayEvent)

	private.GET("/runs", core.ListRuns)
	private.GET("/deliveries", core.ListDeliveries)

	// Templates (auth-protected for now)
	RegisterTemplateRoutes(private.Group(""), templates)
}
