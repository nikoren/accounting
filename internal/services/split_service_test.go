package services

import (
	"accounting/internal/domain"
	"accounting/internal/domain/ports"
	"accounting/internal/infrastructure/db/uow"
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRenderService implements ports.RenderService for testing
type mockRenderService struct{}

func (m *mockRenderService) RenderDocument(ctx context.Context, req ports.RenderDocumentRequest) (*ports.RenderDocumentResponse, error) {
	return &ports.RenderDocumentResponse{
		Filename:    req.Document.Filename,
		ContentType: "application/pdf",
		Data:        []byte("test data"),
	}, nil
}

func setupTestDB(t *testing.T) (*sql.DB, func() (ports.UnitOfWork, error)) {
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

	uowFactory := func() (ports.UnitOfWork, error) {
		uow := uow.NewUnitOfWorkSQL(db)
		if err := uow.Begin(); err != nil {
			return nil, err
		}
		return uow, nil
	}

	return db, uowFactory
}

func TestSplitService_LoadSplit(t *testing.T) {
	db, uowFactory := setupTestDB(t)
	defer db.Close()

	service := NewSplitService(uowFactory, &mockRenderService{})
	ctx := context.Background()

	// Test loading non-existent split
	_, err := service.LoadSplit(ctx, "non-existent")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotFound, err)

	// Create test split
	uow, err := uowFactory()
	require.NoError(t, err)
	defer uow.Rollback(ctx)

	now := time.Now()
	split := &domain.Split{
		ID:        "test-split",
		ClientID:  "test-client",
		Status:    domain.SplitStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = uow.SplitRepository().Save(ctx, split)
	require.NoError(t, err)
	err = uow.Commit(ctx)
	require.NoError(t, err)

	// Test loading existing split
	response, err := service.LoadSplit(ctx, "test-split")
	require.NoError(t, err)
	assert.Equal(t, "test-split", response.ID)
	assert.Equal(t, "test-client", response.ClientID)
	assert.Equal(t, domain.SplitStatusDraft, response.Status)
}

func TestSplitService_UpdateDocumentMetadata(t *testing.T) {
	db, uowFactory := setupTestDB(t)
	defer db.Close()

	service := NewSplitService(uowFactory, &mockRenderService{})
	ctx := context.Background()

	// Create test split with document
	uow, err := uowFactory()
	require.NoError(t, err)
	defer uow.Rollback(ctx)

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
				Name:             "Original Name",
				Classification:   "Original Class",
				Filename:         "test.pdf",
				ShortDescription: "Original Description",
				StartPage:        "1",
				EndPage:          "2",
			},
		},
	}
	err = uow.SplitRepository().Save(ctx, split)
	require.NoError(t, err)
	err = uow.Commit(ctx)
	require.NoError(t, err)

	// Test updating document metadata
	newName := "Updated Name"
	newClass := "Updated Class"
	newDesc := "Updated Description"
	req := UpdateDocumentMetadataRequest{
		Name:             &newName,
		Classification:   &newClass,
		ShortDescription: &newDesc,
	}

	response, err := service.UpdateDocumentMetadata(ctx, "doc1", req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", response.Name)
	assert.Equal(t, "Updated Class", response.Classification)
	assert.Equal(t, "Updated Description", response.ShortDescription)
}

func TestSplitService_MovePages(t *testing.T) {
	db, uowFactory := setupTestDB(t)
	defer db.Close()

	service := NewSplitService(uowFactory, &mockRenderService{})
	ctx := context.Background()

	// Create test split with two documents and pages
	uow, err := uowFactory()
	require.NoError(t, err)
	defer uow.Rollback(ctx)

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
				Name:             "Document 1",
				Classification:   "Class 1",
				Filename:         "doc1.pdf",
				ShortDescription: "Description 1",
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
			{
				ID:               "doc2",
				SplitID:          "test-split",
				Name:             "Document 2",
				Classification:   "Class 2",
				Filename:         "doc2.pdf",
				ShortDescription: "Description 2",
				StartPage:        "3",
				EndPage:          "4",
				Pages: []*domain.Page{
					{
						ID:         "page2",
						SplitID:    "test-split",
						DocumentID: stringPtr("doc2"),
						PageNumber: 2,
						URL:        "http://test.com/2",
					},
				},
			},
		},
	}
	err = uow.SplitRepository().Save(ctx, split)
	require.NoError(t, err)
	err = uow.Commit(ctx)
	require.NoError(t, err)

	// Test moving pages
	req := MovePagesRequest{
		SplitID:        "test-split",
		FromDocumentID: "doc1",
		ToDocumentID:   "doc2",
		PageIDs:        []string{"page1"},
	}

	response, err := service.MovePages(ctx, req)
	require.NoError(t, err)
	assert.Len(t, response.FromDocument.Pages, 0)
	assert.Len(t, response.ToDocument.Pages, 2)
}

func TestSplitService_CreateDocument(t *testing.T) {
	db, uowFactory := setupTestDB(t)
	defer db.Close()

	service := NewSplitService(uowFactory, &mockRenderService{})
	ctx := context.Background()

	// Create test split
	uow, err := uowFactory()
	require.NoError(t, err)
	defer uow.Rollback(ctx)

	now := time.Now()
	split := &domain.Split{
		ID:        "test-split",
		ClientID:  "test-client",
		Status:    domain.SplitStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
		UnassignedPages: []*domain.Page{
			{
				ID:         "page1",
				SplitID:    "test-split",
				PageNumber: 1,
				URL:        "http://test.com/1",
			},
		},
	}
	err = uow.SplitRepository().Save(ctx, split)
	require.NoError(t, err)
	err = uow.Commit(ctx)
	require.NoError(t, err)

	// Test creating document
	req := CreateDocumentRequest{
		SplitID:          "test-split",
		Name:             "New Document",
		Classification:   "New Class",
		Filename:         "new.pdf",
		ShortDescription: "New Description",
		PageIDs:          []string{"page1"},
	}

	response, err := service.CreateDocument(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "New Document", response.Name)
	assert.Equal(t, "New Class", response.Classification)
	assert.Equal(t, "New Description", response.ShortDescription)
	assert.Len(t, response.Pages, 1)
}

func TestSplitService_DeleteDocument(t *testing.T) {
	db, uowFactory := setupTestDB(t)
	defer db.Close()

	service := NewSplitService(uowFactory, &mockRenderService{})
	ctx := context.Background()

	// Create test split with document
	uow, err := uowFactory()
	require.NoError(t, err)
	defer uow.Rollback(ctx)

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
				Name:             "Test Document",
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
	err = uow.SplitRepository().Save(ctx, split)
	require.NoError(t, err)
	err = uow.Commit(ctx)
	require.NoError(t, err)

	// Test deleting document
	err = service.DeleteDocument(ctx, "doc1")
	require.NoError(t, err)

	// Verify document is deleted
	loadedSplit, err := service.LoadSplit(ctx, "test-split")
	require.NoError(t, err)
	assert.Len(t, loadedSplit.Documents, 0)
}

func TestSplitService_FinalizeSplit(t *testing.T) {
	db, uowFactory := setupTestDB(t)
	defer db.Close()

	service := NewSplitService(uowFactory, &mockRenderService{})
	ctx := context.Background()

	// Create test split
	uow, err := uowFactory()
	require.NoError(t, err)
	defer uow.Rollback(ctx)

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
				Name:             "Test Document",
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
	err = uow.SplitRepository().Save(ctx, split)
	require.NoError(t, err)
	err = uow.Commit(ctx)
	require.NoError(t, err)

	// Test finalizing split
	err = service.FinalizeSplit(ctx, "test-split")
	require.NoError(t, err)

	// Verify split is finalized
	loadedSplit, err := service.LoadSplit(ctx, "test-split")
	require.NoError(t, err)
	assert.Equal(t, domain.SplitStatusFinalized, loadedSplit.Status)
}

func TestSplitService_DownloadDocument(t *testing.T) {
	db, uowFactory := setupTestDB(t)
	defer db.Close()

	service := NewSplitService(uowFactory, &mockRenderService{})
	ctx := context.Background()

	// Create test split with document
	uow, err := uowFactory()
	require.NoError(t, err)
	defer uow.Rollback(ctx)

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
				Name:             "Test Document",
				Classification:   "Test Class",
				Filename:         "test.pdf",
				ShortDescription: "Test Description",
				StartPage:        "1",
				EndPage:          "2",
			},
		},
	}
	err = uow.SplitRepository().Save(ctx, split)
	require.NoError(t, err)
	err = uow.Commit(ctx)
	require.NoError(t, err)

	// Test downloading document
	response, err := service.DownloadDocument(ctx, "doc1")
	require.NoError(t, err)
	assert.Equal(t, "test.pdf", response.Filename)
	assert.Equal(t, "application/pdf", response.ContentType)
	assert.Equal(t, []byte("test data"), response.Data)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
