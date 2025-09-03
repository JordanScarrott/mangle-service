package ports

import (
	"context"
	"mangle-service/internal/core/domain"
)

// MangleRepository defines the port for interacting with the Mangle data store.
type MangleRepository interface {
	ExecuteQuery(ctx context.Context, query string) (*domain.QueryResult, error)
}
