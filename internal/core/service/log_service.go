package service

import (
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"
)

// LogService is the core service for handling log-related operations.
type LogService struct {
	logDataPort ports.LogDataPort
}

// NewLogService creates a new instance of the LogService.
func NewLogService(logDataPort ports.LogDataPort) *LogService {
	return &LogService{
		logDataPort: logDataPort,
	}
}

// FetchLogs fetches logs based on the provided criteria.
func (s *LogService) FetchLogs(queryCriteria map[string]string) ([]domain.Fact, error) {
	return s.logDataPort.FetchLogs(queryCriteria)
}
