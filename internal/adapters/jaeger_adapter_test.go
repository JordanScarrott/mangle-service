package adapters

import (
	"context"
	"io"
	"mangle-service/internal/core/domain"
	"net/http"
	"strings"
	"testing"

	"github.com/google/mangle/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRoundTripper is a mock implementation of http.RoundTripper for testing.
type mockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestJaegerAdapter_FetchTraces(t *testing.T) {
	// 1. Setup
	mockResponse := `{
		"data": [
			{
				"traceID": "t1",
				"spans": [
					{
						"traceID": "t1",
						"spanID": "s1",
						"operationName": "op1",
						"references": [],
						"startTime": 1672531200000000,
						"duration": 100000,
						"tags": [
							{"key": "k1", "type": "string", "value": "v1"}
						],
						"processID": "p1"
					}
				],
				"processes": {
					"p1": {
						"processID": "p1",
						"serviceName": "service1"
					}
				}
			}
		]
	}`

	client := &http.Client{
		Transport: &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(mockResponse)),
			},
		},
	}
	adapter := NewJaegerAdapter(client, "http://localhost:16686")

	// 2. Execute
	facts, err := adapter.FetchTraces(context.Background(), "service1")

	// 3. Assert
	require.NoError(t, err)
	require.Len(t, facts, 2)

	expectedFacts := []domain.Fact{
		ast.NewAtom(
			"span",
			ast.String("t1"),
			ast.String("s1"),
			ast.String(""),
			ast.String("service1"),
			ast.String("op1"),
			ast.Number(100), // 100000 microseconds -> 100 milliseconds
		),
		ast.NewAtom(
			"tag",
			ast.String("t1"),
			ast.String("s1"),
			ast.String("k1"),
			ast.String("v1"),
		),
	}

	assert.ElementsMatch(t, expectedFacts, facts)
}
