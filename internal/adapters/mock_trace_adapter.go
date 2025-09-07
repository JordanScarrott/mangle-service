package adapters

import (
	"context"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"

	"github.com/google/mangle/ast"
)

// MockTraceAdapter is a mock implementation of the TraceDataPort for testing.
type MockTraceAdapter struct{}

// NewMockTraceAdapter creates a new instance of the mock trace adapter.
func NewMockTraceAdapter() ports.TraceDataPort {
	return &MockTraceAdapter{}
}

// FetchTraces returns a hardcoded set of trace data as Mangle facts.
func (a *MockTraceAdapter) FetchTraces(ctx context.Context, serviceName string) ([]domain.Fact, error) {
	facts := []domain.Fact{
		// span("t123", "spanA", "", "api-gateway", "GET /users", 150)
		ast.NewAtom(
			"span",
			ast.String("t123"),
			ast.String("spanA"),
			ast.String(""),
			ast.String("api-gateway"),
			ast.String("GET /users"),
			ast.Number(150),
		),
		// span("t123", "spanB", "spanA", "user-service", "db_query", 100)
		ast.NewAtom(
			"span",
			ast.String("t123"),
			ast.String("spanB"),
			ast.String("spanA"),
			ast.String("user-service"),
			ast.String("db_query"),
			ast.Number(100),
		),
		// tag("t123", "spanA", "http.status_code", "200")
		ast.NewAtom(
			"tag",
			ast.String("t123"),
			ast.String("spanA"),
			ast.String("http.status_code"),
			ast.String("200"),
		),
	}
	return facts, nil
}
