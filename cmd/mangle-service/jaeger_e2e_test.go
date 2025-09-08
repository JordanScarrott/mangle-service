package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mangle-service/internal/adapters"
	"mangle-service/internal/adapters/file"
	httphandler "mangle-service/internal/adapters/http"
	"mangle-service/internal/adapters/mock"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/service"
	"mangle-service/pkg/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEndJaegerQuery(t *testing.T) {
	// 1. Setup
	log := logger.New(slog.LevelDebug)
	traceAdapter := adapters.NewMockTraceAdapter()
	fileAdapter := file.NewConfigLoader()

	// Create a temporary relationships file for the test
	relationshipContent := `
relationships:
  - service: "api-gateway"
    depends_on: ["user-service"]
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

	logAdapter := mock.NewMockLogAdapter()
	queryService := service.NewQueryService(logAdapter, traceAdapter, relationshipService, log)

	httpAdapter := httphandler.NewAdapter(queryService, log, "8080")

	// 2. Test Server
	server := httptest.NewServer(httpAdapter.GetRouter())
	defer server.Close()

	// 3. Mangle Query
	mangleQuery := `
		slow_db_query(TraceID, Service) :-
			span(TraceID, SpanA, _, "api-gateway", _, _),
			span(TraceID, _, SpanA, Service, "db_query", Duration),
			tag(TraceID, SpanA, "http.status_code", "200"),
			Duration > 50.
	`
	queryReq := domain.QueryRequest{Query: mangleQuery + "slow_db_query(TraceID, Service)."}
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

	// We expect one result: the user-service on trace-t123
	expectedBindings := []domain.LogEntry{
		{"TraceID": "t123", "Service": "user-service"},
	}

	assert.Equal(t, 1, result.Count)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, expectedBindings, result.Results)
}
