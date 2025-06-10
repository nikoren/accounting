package uow

import (
	"accounting/internal/domain"
	"accounting/internal/infrastructure/db/repositories/splits"
	"context"
	"database/sql"
)

// UnitOfWorkSQL implements domain.UnitOfWork using SQLite
type UnitOfWorkSQL struct {
	db *sql.DB
	tx *sql.Tx
}

// NewUnitOfWorkSQL creates a new SQLite-based unit of work
func NewUnitOfWorkSQL(db *sql.DB) *UnitOfWorkSQL {
	return &UnitOfWorkSQL{db: db}
}

// Begin starts a new transaction
func (u *UnitOfWorkSQL) Begin() error {
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}
	u.tx = tx
	return nil
}

// Commit commits the transaction
func (u *UnitOfWorkSQL) Commit(ctx context.Context) error {
	if u.tx == nil {
		return nil
	}
	return u.tx.Commit()
}

// Rollback rolls back the transaction
func (u *UnitOfWorkSQL) Rollback(ctx context.Context) error {
	if u.tx == nil {
		return nil
	}
	return u.tx.Rollback()
}

// SplitRepository returns a new split repository instance
func (u *UnitOfWorkSQL) SplitRepository() domain.SplitRepository {
	return splits.NewSplitRepositorySQL(u.tx)
}
