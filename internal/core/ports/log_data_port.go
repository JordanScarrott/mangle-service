package ports

import "mangle-service/internal/core/domain"

// LogDataPort is an interface for fetching log data from a data source.
type LogDataPort interface {
	FetchLogs(queryCriteria map[string]string) ([]domain.Fact, error)
}
