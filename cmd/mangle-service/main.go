package main

import (
	"context"
	"flag"
	"log/slog"
	"mangle-service/internal/adapters/elasticsearch"
	"mangle-service/internal/adapters/file"
	httphandler "mangle-service/internal/adapters/http"
	"mangle-service/internal/adapters/mock"
	"mangle-service/internal/core/ports"
	"mangle-service/internal/core/service"
	"mangle-service/pkg/logger"
	"net/http"

	mocktrace "mangle-service/internal/adapters"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Configuration
	env := flag.String("env", "prod", "environment (dev, prod, test)")
	flag.Parse()

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

	jaegerURL := os.Getenv("JAEGER_QUERY_URL")
	if jaegerURL == "" && *env != "test" {
		log.Error("JAEGER_QUERY_URL must be set in non-test environments")
		os.Exit(1)
	}

	// 3. Adapters
	var (
		logAdapter   ports.LogDataPort
		traceAdapter ports.TraceDataPort
	)
	if *env == "test" {
		log.Info("using mock adapters")
		logAdapter = mock.NewMockLogAdapter()
		traceAdapter = mocktrace.NewMockTraceAdapter()
	} else {
		log.Info("using real adapters")
		logAdapter = elasticsearch.NewElasticsearchAdapter()
		traceAdapter = mocktrace.NewJaegerAdapter(&http.Client{Timeout: 30 * time.Second}, jaegerURL)
	}
	fileAdapter := file.NewConfigLoader()

	// 4. Core Services
	logService := service.NewLogService(logAdapter)
	relationshipService := service.NewRelationshipService(fileAdapter)
	if err := relationshipService.LoadRelationships(relationshipConfigPath); err != nil {
		log.Error("failed to load relationships", "error", err)
		os.Exit(1)
	}
	queryService := service.NewQueryService(logService, traceAdapter, relationshipService, log)

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
