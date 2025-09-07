package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
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

func TestEndToEndQuery(t *testing.T) {
	// 1. Setup - Mirroring the main.go setup but with the mock adapter
	log := logger.New(slog.LevelDebug)
	logAdapter := mock.NewMockLogAdapter()
	fileAdapter := file.NewConfigLoader()

	// Create a temporary relationships file for the test
	relationshipContent := `
relationships:
  - service: "service-a"
    depends_on: ["service-b"]
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

	httpAdapter := httphandler.NewAdapter(queryService, log, "8080") // Port is not used by httptest

	// 2. Test Server
	server := httptest.NewServer(httpAdapter.GetRouter())
	defer server.Close()

	// 3. Mangle Query
	mangleQuery := `logs(S, 500, _).`
	queryReq := domain.QueryRequest{Query: mangleQuery}
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

	expectedBindings := []domain.LogEntry{
		{"S": "B"},
	}

	assert.Equal(t, 1, result.Count)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, expectedBindings, result.Results)
}
