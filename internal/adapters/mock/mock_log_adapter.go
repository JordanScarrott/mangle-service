package mock

import (
	"mangle-service/internal/core/domain"

	"github.com/google/mangle/ast"
)

// MockLogAdapter is a mock implementation of the LogDataPort for testing.
type MockLogAdapter struct{}

// NewMockLogAdapter creates a new MockLogAdapter.
func NewMockLogAdapter() *MockLogAdapter {
	return &MockLogAdapter{}
}

// FetchLogs returns a hardcoded list of log facts for testing purposes.
func (a *MockLogAdapter) FetchLogs(queryCriteria map[string]string) ([]domain.Fact, error) {
	facts := []domain.Fact{
		// logs('A', 200, 'call to B')
		ast.NewAtom(
			"logs",
			ast.String("A"),
			ast.Number(200),
			ast.String("call to B"),
		),
		// logs('B', 500, 'database error')
		ast.NewAtom(
			"logs",
			ast.String("B"),
			ast.Number(500),
			ast.String("database error"),
		),
	}
	return facts, nil
}
