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

// GetMangleRules transforms the loaded service relationships into Mangle rules.
func (s *RelationshipService) GetMangleRules() ([]domain.Fact, error) {
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

func (s *RelationshipService) createCallsFact(service, dependency string) (domain.Fact, error) {
	predicate := ast.PredicateSym{Symbol: "calls", Arity: 2}
	args := []ast.Term{
		ast.String(service),
		ast.String(dependency),
	}

	return ast.NewAtom(predicate, args...), nil
}
