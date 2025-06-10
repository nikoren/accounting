package services

import (
	"accounting/internal/domain"
	"context"
)

// PageResponse represents a page in the API
type PageResponse struct {
	ID         string `json:"id"`
	PageNumber string `json:"page_number"`
	URL        string `json:"url"`
}

// DocumentResponse represents a document in the API
type DocumentResponse struct {
	ID               string          `json:"id"`
	SplitID          string          `json:"split_id"`
	Name             string          `json:"name"`
	Classification   string          `json:"classification"`
	Filename         string          `json:"filename"`
	ShortDescription string          `json:"short_description"`
	StartPage        string          `json:"start_page"`
	EndPage          string          `json:"end_page"`
	Pages            []*PageResponse `json:"pages"`
}

// LoadSplitResponse represents a split in the API
type LoadSplitResponse struct {
	ID              string              `json:"id"`
	ClientID        string              `json:"client_id"`
	Status          domain.SplitStatus  `json:"status"`
	Documents       []*DocumentResponse `json:"documents"`
	UnassignedPages []*PageResponse     `json:"unassigned_pages"`
}

// UpdateDocumentMetadataRequest represents a request to update document metadata
type UpdateDocumentMetadataRequest struct {
	Name             *string `json:"name,omitempty"`
	Classification   *string `json:"classification,omitempty"`
	ShortDescription *string `json:"short_description,omitempty"`
}

// MovePagesRequest represents a request to move pages between documents
type MovePagesRequest struct {
	SplitID        string   `json:"split_id"`
	FromDocumentID string   `json:"from_document_id"`
	ToDocumentID   string   `json:"to_document_id"`
	PageIDs        []string `json:"page_ids"`
}

// MovePagesResponse represents the response from moving pages
type MovePagesResponse struct {
	FromDocument *DocumentResponse `json:"from_document"`
	ToDocument   *DocumentResponse `json:"to_document"`
}

// CreateDocumentRequest represents a request to create a document
type CreateDocumentRequest struct {
	SplitID          string   `json:"split_id"`
	Name             string   `json:"name"`
	Classification   string   `json:"classification"`
	Filename         string   `json:"filename"`
	ShortDescription string   `json:"short_description"`
	PageIDs          []string `json:"page_ids"`
}

// DeleteDocumentRequest represents a request to delete a document
type DeleteDocumentRequest struct {
	DocumentID string
}

// DownloadDocumentResponse represents the response from downloading a document
type DownloadDocumentResponse struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}

// SplitServiceInterface defines the interface for split operations (for handler and tests)
type SplitServiceInterface interface {
	LoadSplit(ctx context.Context, id string) (*LoadSplitResponse, error)
	UpdateDocumentMetadata(ctx context.Context, documentID string, req UpdateDocumentMetadataRequest) (*DocumentResponse, error)
	MovePages(ctx context.Context, req MovePagesRequest) (*MovePagesResponse, error)
	CreateDocument(ctx context.Context, req CreateDocumentRequest) (*DocumentResponse, error)
	DeleteDocument(ctx context.Context, documentID string) error
	FinalizeSplit(ctx context.Context, splitID string) error
	DownloadDocument(ctx context.Context, documentID string) (*DownloadDocumentResponse, error)
}

// ErrNotFound is returned when a requested resource is not found
var ErrNotFound = domain.ErrNotFound
