package service

import (
	"fmt"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"

	"github.com/google/mangle/ast"
)

var _ ports.RelationshipService = (*RelationshipService)(nil)

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

// GetMangleFacts transforms the loaded service relationships into Mangle facts.
func (s *RelationshipService) GetMangleFacts() ([]domain.Fact, error) {
	if s.config == nil {
		return nil, fmt.Errorf("relationships not loaded")
	}

	var facts []domain.Fact
	for _, rel := range s.config.Relationships {
		for _, dep := range rel.DependsOn {
			fact, err := s.createCallsFact(rel.Service, dep)
			if err != nil {
				return nil, err
			}
			facts = append(facts, fact)
		}
	}
	return facts, nil
}

// GetMangleRulesAsString returns the Mangle rules for service dependencies.
func (s *RelationshipService) GetMangleRulesAsString() (string, error) {
	return `
		depends_on(X, Y) :- calls(X, Y).
		depends_on(X, Z) :- calls(X, Y), depends_on(Y, Z).
	`, nil
}

func (s *RelationshipService) createCallsFact(service, dependency string) (domain.Fact, error) {
	return ast.NewAtom(
		"calls",
		ast.String(service),
		ast.String(dependency),
	), nil
}
