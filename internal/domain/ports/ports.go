package ports

import (
	"accounting/internal/domain"
	"context"
	"io"
)

// SplitIngestionService handles the ingestion of new splits
type SplitIngestionService interface {
	// IngestSplit processes a new split request
	IngestSplit(ctx context.Context, req IngestSplitRequest) (*IngestSplitResponse, error)
}

// IngestSplitRequest represents a request to ingest a new split
type IngestSplitRequest struct {
	ClientID string
	File     io.Reader
}

// IngestSplitResponse represents the response from ingesting a split
type IngestSplitResponse struct {
	SplitID string
}

// UnitOfWork defines the interface for managing transactions
type UnitOfWork interface {
	// SplitRepository returns the split repository
	SplitRepository() domain.SplitRepository
	// Commit commits the transaction
	Commit(ctx context.Context) error
	// Rollback rolls back the transaction
	Rollback(ctx context.Context) error
}

// RenderService handles document rendering
type RenderService interface {
	// RenderDocument renders a document to a downloadable format
	RenderDocument(ctx context.Context, req RenderDocumentRequest) (*RenderDocumentResponse, error)
}

// RenderDocumentRequest represents a request to render a document
type RenderDocumentRequest struct {
	Document *domain.Document
}

// RenderDocumentResponse represents the response from rendering a document
type RenderDocumentResponse struct {
	Filename    string
	ContentType string
	Data        []byte
}
