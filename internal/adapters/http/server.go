package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"
	"net/http"
	"time"
)

type Adapter struct {
	service ports.QueryService
	logger  *slog.Logger
	server  *http.Server
}

func NewAdapter(service ports.QueryService, logger *slog.Logger, port string) *Adapter {
	return &Adapter{
		service: service,
		logger:  logger,
		server: &http.Server{
			Addr: ":" + port,
		},
	}
}

func (a *Adapter) Start(ctx context.Context) error {
	a.logger.Info("starting server", "addr", a.server.Addr)
	mux := http.NewServeMux()
	mux.HandleFunc("/query", a.handleQuery)
	mux.HandleFunc("/healthz", a.handleHealthCheck)

	a.server.Handler = mux
	return a.server.ListenAndServe()
}

func (a *Adapter) Stop(ctx context.Context) error {
	a.logger.Info("stopping server")
	return a.server.Shutdown(ctx)
}

func (a *Adapter) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	var req domain.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := a.service.ExecuteQuery(r.Context(), req)
	if err != nil {
		a.logger.Error("error executing query", "error", err)
		a.writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, result, http.StatusOK)
	a.logger.Info("processed query", "duration", time.Since(start), "query", req.Query, "results", result.Count)
}

func (a *Adapter) writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		a.logger.Error("failed to write json response", "error", err)
	}
}

func (a *Adapter) writeError(w http.ResponseWriter, message string, status int) {
	a.writeJSON(w, map[string]string{"error": message}, status)
}
