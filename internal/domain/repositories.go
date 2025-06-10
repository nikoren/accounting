package domain

import "context"

// SplitRepository handles split aggregate persistence
type SplitRepository interface {
	// Get retrieves a split by ID
	Get(ctx context.Context, id string) (*Split, error)
	// Save persists a split aggregate
	Save(ctx context.Context, split *Split) error
	// Delete removes a split
	Delete(ctx context.Context, id string) error
	// ListByClientID retrieves all splits for a client
	ListByClientID(ctx context.Context, clientID string) ([]*Split, error)
	// GetSplitIDByDocumentID retrieves the split ID for a given document ID
	GetSplitIDByDocumentID(ctx context.Context, documentID string) (string, error)
}
