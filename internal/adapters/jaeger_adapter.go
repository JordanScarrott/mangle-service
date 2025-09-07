package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"mangle-service/internal/core/domain"
	"mangle-service/internal/core/ports"
	"net/http"
	"net/url"

	"github.com/google/mangle/ast"
)

// JaegerResponse is the top-level structure of the Jaeger API response.
type JaegerResponse struct {
	Data []JaegerTrace `json:"data"`
}

// JaegerTrace is a single trace.
type JaegerTrace struct {
	TraceID   string                `json:"traceID"`
	Spans     []JaegerSpan          `json:"spans"`
	Processes map[string]JaegerProcess `json:"processes"`
}

// JaegerSpan is a single span.
type JaegerSpan struct {
	TraceID       string            `json:"traceID"`
	SpanID        string            `json:"spanID"`
	OperationName string            `json:"operationName"`
	References    []JaegerReference `json:"references"`
	StartTime     int64             `json:"startTime"`
	Duration      int64             `json:"duration"`
	Tags          []JaegerTag       `json:"tags"`
	ProcessID     string            `json:"processID"`
}

// JaegerReference is a reference to another span.
type JaegerReference struct {
	RefType string `json:"refType"`
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

// JaegerTag is a key-value pair of metadata.
type JaegerTag struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// JaegerProcess is a process that originated a span.
type JaegerProcess struct {
	ProcessID   string      `json:"processID"`
	ServiceName string      `json:"serviceName"`
	Tags        []JaegerTag `json:"tags"`
}

// JaegerAdapter fetches trace data from a Jaeger instance.
type JaegerAdapter struct {
	client  *http.Client
	baseURL string
}

// NewJaegerAdapter creates a new instance of the Jaeger adapter.
func NewJaegerAdapter(client *http.Client, baseURL string) ports.TraceDataPort {
	return &JaegerAdapter{
		client:  client,
		baseURL: baseURL,
	}
}

// FetchTraces fetches trace data for a given service and returns it as Mangle facts.
func (a *JaegerAdapter) FetchTraces(ctx context.Context, serviceName string) ([]domain.Fact, error) {
	// Construct the URL with query parameters
	u, err := url.Parse(a.baseURL + "/api/traces")
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	q := u.Query()
	q.Set("service", serviceName)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch traces from Jaeger: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jaeger API returned non-OK status: %d", resp.StatusCode)
	}

	var response JaegerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Jaeger response: %w", err)
	}

	var facts []domain.Fact
	for _, trace := range response.Data {
		for _, span := range trace.Spans {
			process, ok := trace.Processes[span.ProcessID]
			if !ok {
				continue // Should not happen in valid Jaeger data
			}

			parentSpanID := ""
			if len(span.References) > 0 {
				parentSpanID = span.References[0].SpanID
			}

			// Create span/6 fact
			facts = append(facts, ast.NewAtom(
				"span",
				ast.String(span.TraceID),
				ast.String(span.SpanID),
				ast.String(parentSpanID),
				ast.String(process.ServiceName),
				ast.String(span.OperationName),
				ast.Number(span.Duration/1000), // Convert microseconds to milliseconds
			))

			// Create tag/4 facts
			for _, tag := range span.Tags {
				facts = append(facts, ast.NewAtom(
					"tag",
					ast.String(span.TraceID),
					ast.String(span.SpanID),
					ast.String(tag.Key),
					ast.String(fmt.Sprintf("%v", tag.Value)),
				))
			}
		}
	}

	return facts, nil
}
