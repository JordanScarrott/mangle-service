package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mangle-service/internal/adapters/file"
	httphandler "mangle-service/internal/adapters/http"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/service"
	"mangle-service/pkg/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/mangle/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cascadingFailureLogAdapter is a mock implementation of LogDataPort for this specific test case.
type cascadingFailureLogAdapter struct{}

func (a *cascadingFailureLogAdapter) FetchLogs(queryCriteria map[string]string) ([]domain.Fact, error) {
	return []domain.Fact{
		// Successful transaction (noise)
		// logs("trace-abc", "api-gateway", 200, "Request processed successfully")
		ast.NewAtom(
			"logs",
			ast.String("trace-abc"),
			ast.String("api-gateway"),
			ast.Number(200),
			ast.String("Request processed successfully"),
		),
		// Failing transaction
		// logs("trace-xyz", "api-gateway", 200, "Forwarding request to order-service")
		ast.NewAtom(
			"logs",
			ast.String("trace-xyz"),
			ast.String("api-gateway"),
			ast.Number(200),
			ast.String("Forwarding request to order-service"),
		),
		// logs("trace-xyz", "order-service", 500, "Database connection failed")
		ast.NewAtom(
			"logs",
			ast.String("trace-xyz"),
			ast.String("order-service"),
			ast.Number(500),
			ast.String("Database connection failed"),
		),
		// logs("trace-xyz", "api-gateway", 500, "Internal Server Error on response")
		ast.NewAtom(
			"logs",
			ast.String("trace-xyz"),
			ast.String("api-gateway"),
			ast.Number(500),
			ast.String("Internal Server Error on response"),
		),
	}, nil
}

func TestEndToEndCascadingFailureQuery(t *testing.T) {
	// 1. Setup
	log := logger.New(slog.LevelDebug)
	logAdapter := &cascadingFailureLogAdapter{}
	fileAdapter := file.NewConfigLoader()

	// Create a temporary relationships file for the test
	// NOTE: Using the format the code actually supports, not the one from the prompt.
	relationshipContent := `
relationships:
  - service: "api-gateway"
    depends_on: ["order-service"]
`
	tmpfile, err := os.CreateTemp("", "relationships.*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.Write([]byte(relationshipContent))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	relationshipService := service.NewRelationshipService(fileAdapter)
	err = relationshipService.LoadRelationships(tmpfile.Name())
	require.NoError(t, err)

	logService := service.NewLogService(logAdapter)
	queryService := service.NewQueryService(logService, relationshipService, log)

	httpAdapter := httphandler.NewAdapter(queryService, log, "8080")

	// 2. Test Server
	server := httptest.NewServer(httpAdapter.GetRouter())
	defer server.Close()

	// 3. Mangle Query
	mangleQuery := `
		gateway_crashed(TraceID) :- logs(TraceID, "api-gateway", 500, _).
		root_cause_service(Service, TraceID) :- gateway_crashed(TraceID), calls("api-gateway", Service), logs(TraceID, Service, 500, _).
	`
	queryReq := domain.QueryRequest{Query: mangleQuery + "root_cause_service(Service, TraceID)."}
	reqBodyBytes, err := json.Marshal(queryReq)
	require.NoError(t, err)
	reqBody := bytes.NewBuffer(reqBodyBytes)

	// 4. HTTP Request
	resp, err := http.Post(server.URL+"/query", "application/json", reqBody)
	require.NoError(t, err)
	defer resp.Body.Close()

	// 5. Assertions
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result domain.QueryResult
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	// We expect one result: the order-service on trace-xyz
	expectedBindings := []domain.LogEntry{
		{"Service": "order-service", "TraceID": "trace-xyz"},
	}

	assert.Equal(t, 1, result.Count)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, expectedBindings, result.Results)
}
