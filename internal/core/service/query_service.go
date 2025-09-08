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
	traceDataPort       ports.TraceDataPort
	relationshipService ports.RelationshipService
	logger              *slog.Logger
}

// NewQueryService creates a new instance of the query service.
func NewQueryService(logDataPort ports.LogDataPort, traceDataPort ports.TraceDataPort, relationshipService ports.RelationshipService, logger *slog.Logger) ports.QueryService {
	return &queryService{
		logDataPort:         logDataPort,
		traceDataPort:       traceDataPort,
		relationshipService: relationshipService,
		logger:              logger,
	}
}

// ExecuteQuery orchestrates the query execution.
func (s *queryService) ExecuteQuery(ctx context.Context, req domain.QueryRequest) (*domain.QueryResult, error) {
	s.logger.Info("starting query execution", "query", req.Query)
	startTime := time.Now()

	// 1. Parse the request query to separate rules from the final query atom.
	s.logger.Debug("parsing query request")
	requestUnit, err := parse.Unit(strings.NewReader(req.Query))
	if err != nil {
		return nil, fmt.Errorf("failed to parse request query unit: %w", err)
	}
	if len(requestUnit.Clauses) == 0 {
		return nil, fmt.Errorf("empty query request")
	}
	lastClause := requestUnit.Clauses[len(requestUnit.Clauses)-1]
	if len(lastClause.Premises) > 0 {
		return nil, fmt.Errorf("last clause in query must be a simple atom, not a rule")
	}
	queryAtom := lastClause.Head
	requestRules := requestUnit.Clauses[:len(requestUnit.Clauses)-1]

	// 2. Fetch log facts
	s.logger.Debug("fetching log facts")
	logFacts, err := s.logDataPort.FetchLogs(make(map[string]string))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}
	s.logger.Debug("fetched log facts", "count", len(logFacts))

	// 3. Fetch relationship facts and rules
	s.logger.Debug("fetching relationship facts and rules")
	relationshipFacts, err := s.relationshipService.GetMangleFacts()
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship facts: %w", err)
	}
	relationshipRulesStr, err := s.relationshipService.GetMangleRulesAsString()
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship rules: %w", err)
	}
	s.logger.Debug("fetched relationship info", "fact_count", len(relationshipFacts), "rule_char_count", len(relationshipRulesStr))
	relationshipRulesUnit, err := parse.Unit(strings.NewReader(relationshipRulesStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse relationship rules: %w", err)
	}

	// 4. Fetch trace facts
	s.logger.Debug("fetching trace facts")
	// TODO(Jules): The serviceName is currently hardcoded. In a real-world scenario,
	// this should be made dynamic, for example, by extracting it from the
	// query itself or from a configuration file.
	traceFacts, err := s.traceDataPort.FetchTraces(ctx, "mangle-service")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch traces: %w", err)
	}
	s.logger.Debug("fetched trace facts", "count", len(traceFacts))

	// 5. Combine facts and rules
	allFacts := append(logFacts, relationshipFacts...)
	allFacts = append(allFacts, traceFacts...)
	allRules := append(relationshipRulesUnit.Clauses, requestRules...)
	s.logger.Debug("combined facts and rules", "total_facts", len(allFacts), "total_rules", len(allRules))

	// 5. Initialize Mangle engine
	s.logger.Debug("initializing mangle engine")
	store := factstore.NewSimpleInMemoryStore()
	sourceUnit := parse.SourceUnit{
		Clauses: append(allRules, domain.FactsToClauses(allFacts)...),
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

	// 6. Execute query
	s.logger.Debug("retrieving facts from store for query", "query_atom", queryAtom.String())
	var results []domain.LogEntry
	// Extract variable names from the query, ignoring wildcards.
	varNames := make(map[int]string)
	for i, arg := range queryAtom.Args {
		if v, ok := arg.(ast.Variable); ok && v.Symbol != "_" {
			varNames[i] = v.Symbol
		}
	}

	store.GetFacts(queryAtom, func(a ast.Atom) error {
		resultMap := make(domain.LogEntry)
		// This assumes that the bound atom `a` has the same structure as the query atom.
		for i, term := range a.Args {
			varName, ok := varNames[i]
			if !ok {
				// This case handles results for parts of the query that were not variables (e.g. constants)
				// or wildcards. We skip them in the output.
				continue
			}
			// The ast.Term interface has a String() method that gives us the value.
			// For ast.String, it includes quotes, which we need to remove for clean output.
			resultMap[varName] = strings.Trim(term.String(), `"`)
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
