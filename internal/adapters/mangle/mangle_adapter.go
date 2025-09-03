package mangle

import (
	"context"
	"fmt"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"
	"sort"
	"strings"

	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/engine"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/parse"
)

type googleMangleAdapter struct {
	// In a real application, you might have a connection pool or other dependencies here.
}

const dependsOnRules = `
  depends_on(X, Y) :- calls(X, Y).
  depends_on(X, Z) :- calls(X, Y), depends_on(Y, Z).
`

// NewGoogleMangleAdapter creates a new instance of the Google Mangle adapter.
func NewGoogleMangleAdapter() (ports.MangleRepository, error) {
	return &googleMangleAdapter{}, nil
}

// ExecuteQuery executes a query against the Mangle engine.
func (a *googleMangleAdapter) ExecuteQuery(ctx context.Context, query string) (*domain.QueryResult, error) {
	// This is a simplified implementation that re-initializes everything on each query.
	// A more optimized version would initialize the engine once.

	// Hardcoded facts for demonstration purposes.
	fact1, err := parse.Clause(`calls("service-a", "service-b").`)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fact1: %w", err)
	}
	fact2, err := parse.Clause(`calls("service-b", "service-c").`)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fact2: %w", err)
	}

	rulesUnit, err := parse.Unit(strings.NewReader(dependsOnRules))
	if err != nil {
		return nil, fmt.Errorf("failed to parse rules: %w", err)
	}

	sourceUnit := parse.SourceUnit{
		Clauses: append([]ast.Clause{fact1, fact2}, rulesUnit.Clauses...),
	}

	program, err := analysis.AnalyzeOneUnit(sourceUnit, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create program: %w", err)
	}

	store := factstore.NewSimpleInMemoryStore()
	if err := engine.EvalProgram(program, store); err != nil {
		return nil, fmt.Errorf("program evaluation failed: %w", err)
	}

	parsedQuery, err := parse.Atom(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	var results []ast.Atom
	store.GetFacts(parsedQuery, func(a ast.Atom) error {
		results = append(results, a)
		return nil
	})

	// Transform results into the domain model.
	// This is a naive transformation and might need to be more sophisticated
	// depending on the query structure.
	var logEntries []domain.LogEntry
	var resultVars []string
	for _, fact := range results {
		if len(fact.Args) > 0 {
			resultVars = append(resultVars, fact.Args[0].String())
		}
	}
	sort.Strings(resultVars)

	for _, val := range resultVars {
		logEntries = append(logEntries, domain.LogEntry{"X": val})
	}

	return &domain.QueryResult{
		Results: logEntries,
		Count:   len(logEntries),
	}, nil
}
