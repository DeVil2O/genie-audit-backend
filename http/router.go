package http

import (
	gin "github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"genie-audit-backend/http/handlers"
	"genie-audit-backend/logger"
	"genie-audit-backend/middleware"
	"genie-audit-backend/service"
	"genie-audit-backend/stores"
)

// NewRouter wires Gin engine, middlewares, and routes.
func NewRouter(pool *pgxpool.Pool) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestContext())
	router.Use(middleware.Timeout(middleware.DefaultTimeout))
	router.Use(middleware.RequestLogger())

	store := stores.NewDBStore(pool)
	svc := service.NewService(store)
	templateStore := stores.NewTemplateStore(pool)
	templateHandler := handlers.NewTemplateHandler(templateStore)
	coreHandler := handlers.NewCoreHandler(svc)
	logger.Log.Info("HTTP router initialized with middlewares and routes")
	RegisterCoreRoutes(router, coreHandler, templateHandler)
	return router
}
