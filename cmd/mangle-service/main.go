package main

import (
	"context"
	"fmt"
	"log/slog"
	httphandler "mangle-service/internal/adapters/http"
	"mangle-service/internal/adapters/mangle"
	"mangle-service/internal/core/service"
	"mangle-service/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/engine"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/parse"
)

func main() {
	// 1. Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. Logger
	log := logger.New(slog.LevelDebug)

	// Mangle POC
	runManglePOC(log)

	// 3. Adapters
	mangleAdapter, err := mangle.NewGoogleMangleAdapter()
	if err != nil {
		log.Error("failed to create mangle adapter", "error", err)
		os.Exit(1)
	}

	// 4. Core Service
	queryService := service.NewQueryService(mangleAdapter)

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

func runManglePOC(log *slog.Logger) {
	log.Info("--- Running Mangle Proof-of-Concept ---")

	// Hardcoded facts for demonstration purposes.
	fact1, err := parse.Clause(`calls("service-a", "service-b").`)
	if err != nil {
		log.Error("failed to parse fact1", "error", err)
		return
	}
	fact2, err := parse.Clause(`calls("service-b", "service-c").`)
	if err != nil {
		log.Error("failed to parse fact2", "error", err)
		return
	}

	const dependsOnRules = `
		depends_on(X, Y) :- calls(X, Y).
		depends_on(X, Z) :- calls(X, Y), depends_on(Y, Z).
	`
	rulesUnit, err := parse.Unit(strings.NewReader(dependsOnRules))
	if err != nil {
		log.Error("failed to parse rules", "error", err)
		return
	}

	sourceUnit := parse.SourceUnit{
		Clauses: append([]ast.Clause{fact1, fact2}, rulesUnit.Clauses...),
	}

	program, err := analysis.AnalyzeOneUnit(sourceUnit, nil)
	if err != nil {
		log.Error("failed to create program", "error", err)
		return
	}

	store := factstore.NewSimpleInMemoryStore()
	if err := engine.EvalProgram(program, store); err != nil {
		log.Error("program evaluation failed", "error", err)
		return
	}

	query, err := parse.Atom(`depends_on("service-a", X)`)
	if err != nil {
		log.Error("failed to parse query", "error", err)
		return
	}

	log.Info("--- Querying for: depends_on(\"service-a\", X) ---")
	var results []ast.Atom
	store.GetFacts(query, func(a ast.Atom) error {
		results = append(results, a)
		return nil
	})

	log.Info(fmt.Sprintf("Found %d results:", len(results)))
	for _, fact := range results {
		log.Info(fact.String())
	}
	log.Info("--- Mangle Proof-of-Concept Finished ---")
}
