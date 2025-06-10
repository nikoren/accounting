package splits

import (
	"accounting/internal/domain"
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, *sql.Tx) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE splits (
			id TEXT PRIMARY KEY,
			client_id TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
		CREATE TABLE documents (
			id TEXT PRIMARY KEY,
			split_id TEXT NOT NULL,
			name TEXT NOT NULL,
			classification TEXT,
			filename TEXT,
			short_description TEXT,
			start_page TEXT,
			end_page TEXT,
			FOREIGN KEY (split_id) REFERENCES splits(id)
		);
		CREATE TABLE pages (
			id TEXT PRIMARY KEY,
			split_id TEXT NOT NULL,
			document_id TEXT,
			page_number TEXT NOT NULL,
			url TEXT NOT NULL,
			FOREIGN KEY (split_id) REFERENCES splits(id),
			FOREIGN KEY (document_id) REFERENCES documents(id)
		);
	`)
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)

	return db, tx
}

func TestSplitRepositorySQL_Get(t *testing.T) {
	db, tx := setupTestDB(t)
	defer db.Close()
	defer tx.Rollback()

	repo := NewSplitRepositorySQL(tx)
	ctx := context.Background()

	// Test getting non-existent split
	split, err := repo.Get(ctx, "non-existent")
	require.NoError(t, err)
	assert.Nil(t, split)

	// Insert test data
	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO splits (id, client_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "test-split", "test-client", domain.SplitStatusDraft, now, now)
	require.NoError(t, err)

	// Test getting existing split
	split, err = repo.Get(ctx, "test-split")
	require.NoError(t, err)
	assert.NotNil(t, split)
	assert.Equal(t, "test-split", split.ID)
	assert.Equal(t, "test-client", split.ClientID)
	assert.Equal(t, domain.SplitStatusDraft, split.Status)
}

func TestSplitRepositorySQL_Save(t *testing.T) {
	db, tx := setupTestDB(t)
	defer db.Close()
	defer tx.Rollback()

	repo := NewSplitRepositorySQL(tx)
	ctx := context.Background()

	// Create test split
	now := time.Now()
	split := &domain.Split{
		ID:        "test-split",
		ClientID:  "test-client",
		Status:    domain.SplitStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
		Documents: []domain.Document{
			{
				ID:               "doc1",
				SplitID:          "test-split",
				Name:             "Test Doc",
				Classification:   "Test Class",
				Filename:         "test.pdf",
				ShortDescription: "Test Description",
				StartPage:        "1",
				EndPage:          "2",
				Pages: []*domain.Page{
					{
						ID:         "page1",
						SplitID:    "test-split",
						DocumentID: stringPtr("doc1"),
						PageNumber: 1,
						URL:        "http://test.com/1",
					},
				},
			},
		},
	}

	// Save split
	err := repo.Save(ctx, split)
	require.NoError(t, err)

	// Verify saved data
	savedSplit, err := repo.Get(ctx, "test-split")
	require.NoError(t, err)
	assert.Equal(t, split.ID, savedSplit.ID)
	assert.Equal(t, split.ClientID, savedSplit.ClientID)
	assert.Equal(t, split.Status, savedSplit.Status)
	assert.Len(t, savedSplit.Documents, 1)
	assert.Equal(t, split.Documents[0].ID, savedSplit.Documents[0].ID)
	assert.Len(t, savedSplit.Documents[0].Pages, 1)
}

func TestSplitRepositorySQL_Delete(t *testing.T) {
	db, tx := setupTestDB(t)
	defer db.Close()
	defer tx.Rollback()

	repo := NewSplitRepositorySQL(tx)
	ctx := context.Background()

	// Insert test data
	now := time.Now()
	_, err := tx.Exec(`
		INSERT INTO splits (id, client_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "test-split", "test-client", domain.SplitStatusDraft, now, now)
	require.NoError(t, err)

	// Delete split
	err = repo.Delete(ctx, "test-split")
	require.NoError(t, err)

	// Verify deletion
	split, err := repo.Get(ctx, "test-split")
	require.NoError(t, err)
	assert.Nil(t, split)
}

func TestSplitRepositorySQL_ListByClientID(t *testing.T) {
	db, tx := setupTestDB(t)
	defer db.Close()
	defer tx.Rollback()

	repo := NewSplitRepositorySQL(tx)
	ctx := context.Background()

	// Insert test data
	now := time.Now()
	_, err := tx.Exec(`
		INSERT INTO splits (id, client_id, status, created_at, updated_at)
		VALUES 
			(?, ?, ?, ?, ?),
			(?, ?, ?, ?, ?)
	`, "split1", "client1", domain.SplitStatusDraft, now, now,
		"split2", "client1", domain.SplitStatusDraft, now, now)
	require.NoError(t, err)

	// List splits
	splits, err := repo.ListByClientID(ctx, "client1")
	require.NoError(t, err)
	assert.Len(t, splits, 2)
	assert.Equal(t, "split1", splits[0].ID)
	assert.Equal(t, "split2", splits[1].ID)
}

func TestSplitRepositorySQL_GetSplitIDByDocumentID(t *testing.T) {
	db, tx := setupTestDB(t)
	defer db.Close()
	defer tx.Rollback()

	repo := NewSplitRepositorySQL(tx)
	ctx := context.Background()

	// Insert test data
	now := time.Now()
	_, err := tx.Exec(`
		INSERT INTO splits (id, client_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "test-split", "test-client", domain.SplitStatusDraft, now, now)
	require.NoError(t, err)

	_, err = tx.Exec(`
		INSERT INTO documents (id, split_id, name, classification, filename, short_description, start_page, end_page)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "test-doc", "test-split", "Test Doc", "Test Class", "test.pdf", "Test Description", "1", "2")
	require.NoError(t, err)

	// Get split ID
	splitID, err := repo.GetSplitIDByDocumentID(ctx, "test-doc")
	require.NoError(t, err)
	assert.Equal(t, "test-split", splitID)

	// Test non-existent document
	_, err = repo.GetSplitIDByDocumentID(ctx, "non-existent")
	assert.Error(t, err)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
