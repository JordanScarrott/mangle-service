package service

import (
	"context"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"
)

type queryService struct {
	repo ports.MangleRepository
}

// NewQueryService creates a new instance of the query service.
func NewQueryService(repo ports.MangleRepository) ports.QueryService {
	return &queryService{
		repo: repo,
	}
}

// ExecuteQuery orchestrates the query execution.
func (s *queryService) ExecuteQuery(ctx context.Context, req domain.QueryRequest) (*domain.QueryResult, error) {
	// Here you could add validation or other business logic.
	return s.repo.ExecuteQuery(ctx, req.Query)
}
