package domain

// QueryRequest represents the incoming request for a Mangle query.
type QueryRequest struct {
	Query string `json:"query"`
}
