package service

import (
	"context"
	"fmt"
	"log/slog"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"
	"strings"
	"time"

	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/engine"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/parse"
)

type queryService struct {
	logDataPort         ports.LogDataPort
	relationshipService ports.RelationshipService
	logger              *slog.Logger
}

// NewQueryService creates a new instance of the query service.
func NewQueryService(logDataPort ports.LogDataPort, relationshipService ports.RelationshipService, logger *slog.Logger) ports.QueryService {
	return &queryService{
		logDataPort:         logDataPort,
		relationshipService: relationshipService,
		logger:              logger,
	}
}

// ExecuteQuery orchestrates the query execution.
func (s *queryService) ExecuteQuery(ctx context.Context, req domain.QueryRequest) (*domain.QueryResult, error) {
	s.logger.Info("starting query execution", "query", req.Query)
	startTime := time.Now()

	// 1. Fetch log facts
	s.logger.Debug("fetching log facts")
	logFacts, err := s.logDataPort.FetchLogs(make(map[string]string))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}
	s.logger.Debug("fetched log facts", "count", len(logFacts))

	// 2. Fetch relationship facts and rules
	s.logger.Debug("fetching relationship facts and rules")
	relationshipFacts, err := s.relationshipService.GetMangleFacts()
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship facts: %w", err)
	}
	rulesStr, err := s.relationshipService.GetMangleRulesAsString()
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship rules: %w", err)
	}
	s.logger.Debug("fetched relationship info", "fact_count", len(relationshipFacts), "rule_char_count", len(rulesStr))

	// 3. Combine facts
	allFacts := append(logFacts, relationshipFacts...)
	s.logger.Debug("combined facts", "total_count", len(allFacts))

	// 4. Initialize Mangle engine
	s.logger.Debug("initializing mangle engine")
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

	s.logger.Debug("evaluating program")
	if err := engine.EvalProgram(program, store); err != nil {
		return nil, fmt.Errorf("program evaluation failed: %w", err)
	}
	s.logger.Debug("program evaluation complete")

	// 5. Execute query
	s.logger.Debug("parsing query atom")
	query, err := parse.Atom(req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	s.logger.Debug("retrieving facts from store")
	var results []domain.LogEntry
	store.GetFacts(query, func(a ast.Atom) error {
		resultMap := make(domain.LogEntry)
		for i, term := range a.Args {
			resultMap[fmt.Sprintf("var%d", i)] = term.String()
		}
		results = append(results, resultMap)
		return nil
	})
	s.logger.Debug("retrieved facts", "count", len(results))

	duration := time.Since(startTime)
	s.logger.Info("query execution complete", "duration", duration, "results", len(results))

	return &domain.QueryResult{
		Results: results,
		Count:   len(results),
	}, nil
}
