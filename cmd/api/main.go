package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gin "github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"genie-audit-backend/db"
	"genie-audit-backend/helpers"
	httpHandlers "genie-audit-backend/http"
	"genie-audit-backend/logger"
)

func main() {
	logger.Initialize()

	godotenv.Load()

	addr := os.Getenv(helpers.API_ADDR)
	if addr == "" {
		addr = ":8080"
	}

	dbURL := os.Getenv(helpers.DATABASE_URL)
	if dbURL == "" {
		logger.Log.Fatal("DATABASE_URL is required")
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Log.Infof("connecting to database")
	pool, err := db.GetPool(ctx, dbURL)
	if err != nil {
		logger.Log.Fatalf("failed to create database pool: %v", err)
	}
	defer db.Close()

	router := httpHandlers.NewRouter(pool)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Log.Errorf("graceful shutdown failed: %v", err)
		} else {
			logger.Log.Info("graceful shutdown complete")
		}
	}()

	logger.Log.Infof("API server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Log.Fatalf("server error: %v", err)
	}
	logger.Log.Info("API server stopped")
}
