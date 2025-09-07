package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
func (a *ElasticsearchAdapter) FetchLogs(queryCriteria map[string]string) ([]domain.Fact, error) {
	var mustClauses []interface{}
	for key, value := range queryCriteria {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				key: value,
			},
		})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	res, err := a.client.Search(
		a.client.Search.WithContext(context.Background()),
		a.client.Search.WithIndex("logs"), // Assuming logs are in an index named "logs"
		a.client.Search.WithBody(&buf),
		a.client.Search.WithTrackTotalHits(true),
		a.client.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch error: %s", res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("Error parsing the response body: %s", err)
	}

	var facts []domain.Fact
	hits, ok := r["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return []domain.Fact{}, nil // No hits found
	}

	for _, hit := range hits {
		source, ok := hit.(map[string]interface{})["_source"]
		if !ok {
			continue
		}
		jsonSource, err := json.Marshal(source)
		if err != nil {
			continue
		}
		fact := ast.NewAtom("log", ast.String(string(jsonSource)))
		facts = append(facts, fact)
	}

	return facts, nil
}
