package service

import (
	"context"
	"fmt"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"
	"strings"

	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/engine"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/parse"
)

type queryService struct {
	logDataPort         ports.LogDataPort
	relationshipService ports.RelationshipService
}

// NewQueryService creates a new instance of the query service.
func NewQueryService(logDataPort ports.LogDataPort, relationshipService ports.RelationshipService) ports.QueryService {
	return &queryService{
		logDataPort:         logDataPort,
		relationshipService: relationshipService,
	}
}

// ExecuteQuery orchestrates the query execution.
func (s *queryService) ExecuteQuery(ctx context.Context, req domain.QueryRequest) (*domain.QueryResult, error) {
	// 1. Fetch log facts
	logFacts, err := s.logDataPort.FetchLogs(make(map[string]string))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}

	// 2. Fetch relationship facts and rules
	relationshipFacts, err := s.relationshipService.GetMangleFacts()
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship facts: %w", err)
	}
	rulesStr, err := s.relationshipService.GetMangleRulesAsString()
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship rules: %w", err)
	}

	// 3. Combine facts
	allFacts := append(logFacts, relationshipFacts...)

	// 4. Initialize Mangle engine
	store := factstore.NewSimpleInMemoryStore()
	rulesUnit, err := parse.Unit(strings.NewReader(rulesStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse rules: %w", err)
	}

	sourceUnit := parse.SourceUnit{
		Clauses: append(rulesUnit.Clauses, domain.FactsToClauses(allFacts)...),
	}

	program, err := analysis.AnalyzeOneUnit(sourceUnit, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create program: %w", err)
	}

	if err := engine.EvalProgram(program, store); err != nil {
		return nil, fmt.Errorf("program evaluation failed: %w", err)
	}

	// 5. Execute query
	query, err := parse.Atom(req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	var results []domain.LogEntry
	store.GetFacts(query, func(a ast.Atom) error {
		resultMap := make(domain.LogEntry)
		for i, term := range a.Args {
			resultMap[fmt.Sprintf("var%d", i)] = term.String()
		}
		results = append(results, resultMap)
		return nil
	})

	return &domain.QueryResult{
		Results: results,
		Count:   len(results),
	}, nil
}
