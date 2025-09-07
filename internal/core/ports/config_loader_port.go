package ports

import "mangle-service/internal/core/domain"

// ConfigLoaderPort is an interface for loading service relationship configurations.
type ConfigLoaderPort interface {
	Load(path string) (*domain.RelationshipConfig, error)
}
