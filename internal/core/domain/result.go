package domain

// LogEntry represents a single log entry returned from a query.
// It's a map to handle dynamic fields.
type LogEntry map[string]interface{}

// QueryResult represents the result of a Mangle query.
type QueryResult struct {
	Results []LogEntry `json:"results"`
	Count   int        `json:"count"`
}
