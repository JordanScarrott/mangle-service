package main

import (
	"context"
	"log/slog"
	"mangle-service/internal/adapters/elasticsearch"
	"mangle-service/internal/adapters/file"
	httphandler "mangle-service/internal/adapters/http"
	"mangle-service/internal/core/service"
	"mangle-service/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	relationshipConfigPath := os.Getenv("RELATIONSHIP_CONFIG_PATH")
	if relationshipConfigPath == "" {
		relationshipConfigPath = "relationships.json"
	}

	// 2. Logger
	log := logger.New(slog.LevelDebug)

	// 3. Adapters
	elasticsearchAdapter := elasticsearch.NewElasticsearchAdapter()
	fileAdapter := file.NewConfigLoader()

	// 4. Core Services
	logService := service.NewLogService(elasticsearchAdapter)
	relationshipService := service.NewRelationshipService(fileAdapter)
	if err := relationshipService.LoadRelationships(relationshipConfigPath); err != nil {
		log.Error("failed to load relationships", "error", err)
		os.Exit(1)
	}
	queryService := service.NewQueryService(logService, relationshipService, log)

	// 5. HTTP Server
	httpAdapter := httphandler.NewAdapter(queryService, log, port)

	// 6. Start Server & Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := httpAdapter.Start(ctx); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start http server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	// Shutdown the server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpAdapter.Stop(shutdownCtx); err != nil {
		log.Error("failed to gracefully shutdown server", "error", err)
		os.Exit(1)
	}

	log.Info("server shutdown complete")
}
