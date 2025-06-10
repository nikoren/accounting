package splits

import (
	"database/sql"
	"log"

	"accounting/internal/domain"
)

// SQLiteUnitOfWork implements domain.UnitOfWork for SQLite
type SQLiteUnitOfWork struct {
	tx *sql.Tx
}

// NewSQLiteUnitOfWork creates a new SQLiteUnitOfWork
func NewSQLiteUnitOfWork(tx *sql.Tx) *SQLiteUnitOfWork {
	return &SQLiteUnitOfWork{
		tx: tx,
	}
}

// SplitRepository returns the split repository
func (u *SQLiteUnitOfWork) SplitRepository() domain.SplitRepository {
	return NewSplitRepositorySQL(u.tx)
}

// Commit commits the transaction
func (u *SQLiteUnitOfWork) Commit() error {
	return u.tx.Commit()
}

// Rollback rolls back the transaction (no return value)
func (u *SQLiteUnitOfWork) Rollback() {
	if err := u.tx.Rollback(); err != nil {
		log.Printf("failed to rollback transaction: %v", err)
	}
}
