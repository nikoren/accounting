package client

// Split represents a document split
type Split struct {
	SplitID   string     `json:"split_id"`
	ClientID  string     `json:"client_id"`
	Status    string     `json:"status"`
	Documents []Document `json:"documents"`
}

// Document represents a document in a split
type Document struct {
	ID               string         `json:"id"`
	Classification   string         `json:"classification"`
	FileName         string         `json:"file_name"`
	Name             string         `json:"name"`
	StartPage        string         `json:"start_page"`
	EndPage          string         `json:"end_page"`
	ShortDescription string         `json:"short_description"`
	Pages            []PageResponse `json:"pages"`
}

// UpdateDocumentMetadataRequest represents a request to update document metadata
type UpdateDocumentMetadataRequest struct {
	Name           string `json:"name,omitempty"`
	Classification string `json:"classification,omitempty"`
	FileName       string `json:"filename,omitempty"`
}

// MovePagesRequest represents a request to move pages between documents
type MovePagesRequest struct {
	SplitID        string   `json:"split_id"`
	FromDocumentID string   `json:"from_document_id"`
	ToDocumentID   string   `json:"to_document_id"`
	PageIDs        []string `json:"page_ids"`
}

// CreateDocumentRequest represents a request to create a new document
type CreateDocumentRequest struct {
	SplitID          string   `json:"split_id"`
	Name             string   `json:"name"`
	Classification   string   `json:"classification"`
	Filename         string   `json:"filename"`
	ShortDescription string   `json:"short_description"`
	PageIDs          []string `json:"page_ids"`
}

// DocumentResponse represents a document response
type DocumentResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MovePagesResponse represents a response to a move pages request
type MovePagesResponse struct {
	FromDocument *DocumentResponse `json:"fromDocument"`
	ToDocument   *DocumentResponse `json:"toDocument"`
}

// MetricsResponse represents server metrics
type MetricsResponse struct {
	UptimeSeconds     float64     `json:"uptime_seconds"`
	RequestsTotal     int64       `json:"requests_total"`
	ErrorsTotal       int64       `json:"errors_total"`
	LastError         interface{} `json:"last_error"`
	AvgDurationMs     float64     `json:"avg_duration_ms"`
	TotalResponseMB   float64     `json:"total_response_mb"`
	ActiveConnections int32       `json:"active_connections"`
	RateLimitHits     int64       `json:"rate_limit_hits"`
}

// PageResponse represents a page response
type PageResponse struct {
	ID         string `json:"id"`
	PageNumber string `json:"page_number"`
	URL        string `json:"url"`
}
