package domain

import "context"

// UnitOfWork represents a unit of work pattern for managing transactions
type UnitOfWork interface {
	// SplitRepository returns the split repository
	SplitRepository() SplitRepository
	// Commit commits the transaction
	Commit(ctx context.Context) error
	// Rollback rolls back the transaction
	Rollback(ctx context.Context) error
}
