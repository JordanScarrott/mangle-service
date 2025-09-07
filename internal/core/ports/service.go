package ports

import (
	"context"
	"mangle-service/internal/core/domain"
)

// QueryService defines the port for the core application service.
type QueryService interface {
	ExecuteQuery(ctx context.Context, req domain.QueryRequest) (*domain.QueryResult, error)
}

// RelationshipService defines the port for the relationship service.
type RelationshipService interface {
	LoadRelationships(path string) error
	GetRelationships() []domain.ServiceRelationship
	GetMangleRules() ([]domain.Fact, error)
}
