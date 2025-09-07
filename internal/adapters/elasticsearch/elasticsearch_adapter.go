package elasticsearch

import (
	"log"
	"os"

	"mangle-service/internal/core/domain"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/mangle/ast"
)

// ElasticsearchAdapter implements the LogDataPort interface.
type ElasticsearchAdapter struct {
	client *elasticsearch.Client
}

// NewElasticsearchAdapter creates a new ElasticsearchAdapter.
func NewElasticsearchAdapter() *ElasticsearchAdapter {
	cfg := elasticsearch.Config{
		Addresses: []string{
			os.Getenv("ELASTICSEARCH_ADDRESS"),
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the Elasticsearch client: %s", err)
	}
	return &ElasticsearchAdapter{client: es}
}

// FetchLogs fetches logs from Elasticsearch.
// This is a stub implementation that returns dummy data.
func (a *ElasticsearchAdapter) FetchLogs(queryCriteria map[string]string) ([]domain.Fact, error) {
	// Create a dummy fact for demonstration purposes.
	// A real implementation would query Elasticsearch and transform
	// the results into facts.
	dummyFact := ast.NewAtom("log", ast.String("dummy log entry"))

	return []domain.Fact{
		dummyFact,
	}, nil
}
