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

// FetchLogs fetches logs from Elasticsearch and transforms them into Mangle facts.
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
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	res, err := a.client.Search(
		a.client.Search.WithContext(context.Background()),
		a.client.Search.WithIndex("logs"),
		a.client.Search.WithBody(&buf),
		a.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %w", err)
	}

	var facts []domain.Fact
	hits, ok := r["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return []domain.Fact{}, nil // No hits found
	}

	for _, hit := range hits {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		docID, ok := hitMap["_id"].(string)
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		flattenedSource := flattenSource(source, "")
		for key, value := range flattenedSource {
			fact := ast.NewAtom(
				"log.field",
				ast.String(docID),
				ast.String(key),
				ast.String(value),
			)
			facts = append(facts, fact)
		}
	}

	return facts, nil
}

// flattenSource recursively flattens a nested map into a single-level map with dot-separated keys.
func flattenSource(source map[string]interface{}, prefix string) map[string]string {
	flattened := make(map[string]string)
	for key, value := range source {
		newKey := key
		if prefix != "" {
			newKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// If the value is a map, recurse.
			for k, val := range flattenSource(v, newKey) {
				flattened[k] = val
			}
		case []interface{}:
			// If the value is a slice, marshal it to a JSON string.
			// This is a simple way to handle arrays of objects or values.
			jsonValue, err := json.Marshal(v)
			if err == nil {
				flattened[newKey] = string(jsonValue)
			}
		default:
			// For simple values, convert to string.
			flattened[newKey] = fmt.Sprintf("%v", v)
		}
	}
	return flattened
}
