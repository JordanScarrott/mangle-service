package service

import (
	"fmt"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
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

// GenerateFacts generates a slice of facts from the loaded service relationships.
func (s *RelationshipService) GenerateFacts() ([]ast.Clause, error) {
	if s.config == nil {
		return nil, fmt.Errorf("config is not loaded")
	}

	var facts []ast.Clause
	for _, rel := range s.config.Relationships {
		for _, dep := range rel.DependsOn {
			factStr := fmt.Sprintf(`calls("%s", "%s").`, rel.Service, dep)
			fact, err := parse.Clause(factStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse fact '%s': %w", factStr, err)
			}
			facts = append(facts, fact)
		}
	}
	return facts, nil
}
