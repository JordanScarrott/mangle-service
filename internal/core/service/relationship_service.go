package service

import (
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"
)

// RelationshipService is a service for managing service relationships.
type RelationshipService struct {
	configLoader ports.ConfigLoaderPort
	config       *domain.RelationshipConfig
}

// NewRelationshipService creates a new RelationshipService.
func NewRelationshipService(configLoader ports.ConfigLoaderPort) *RelationshipService {
	return &RelationshipService{
		configLoader: configLoader,
	}
}

// LoadRelationships loads the service relationships from the given path.
func (s *RelationshipService) LoadRelationships(path string) error {
	config, err := s.configLoader.Load(path)
	if err != nil {
		return err
	}
	s.config = config
	return nil
}

// GetRelationships returns the loaded service relationships.
func (s *RelationshipService) GetRelationships() []domain.ServiceRelationship {
	if s.config == nil {
		return nil
	}
	return s.config.Relationships
}
