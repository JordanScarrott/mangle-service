package ports

import (
	"context"
	"mangle-service/internal/core/domain"
)

// TraceDataPort is the port for fetching trace data.
type TraceDataPort interface {
	FetchTraces(ctx context.Context, serviceName string) ([]domain.Fact, error)
}
