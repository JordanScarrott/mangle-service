package domain

// ServiceRelationship defines a single service and its dependencies.
type ServiceRelationship struct {
	Service   string   `yaml:"service"`
	DependsOn []string `yaml:"depends_on"`
}

// RelationshipConfig represents the entire service relationship configuration.
type RelationshipConfig struct {
	Relationships []ServiceRelationship `yaml:"relationships"`
}
