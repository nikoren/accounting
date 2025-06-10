package integration

import (
	"accounting/internal/client"
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	apiClient *client.Client
	baseURL   string
)

func setupTestDB() error {
	dbPath := os.Getenv("APP_DB_PATH")
	if dbPath == "" {
		dbPath = "test_accounting.db"
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Drop existing tables to ensure a clean schema
	_, err = db.Exec(`
	DROP TABLE IF EXISTS pages;
	DROP TABLE IF EXISTS documents;
	DROP TABLE IF EXISTS splits;
	`)
	if err != nil {
		return err
	}

	// Create tables with full schema
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
	if err != nil {
		return err
	}

	// Insert test splits
	now := time.Now()
	_, err = db.Exec(`INSERT INTO splits (id, client_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		"test-split", "test-client", "draft", now, now)
	if err != nil {
		return err
	}

	// Insert test pages (unassigned)
	pageIDs := []string{"page1", "page2", "page3", "page4"}
	for i, pid := range pageIDs {
		_, err = db.Exec(`INSERT INTO pages (id, split_id, page_number, url) VALUES (?, ?, ?, ?)`,
			pid, "test-split", i+1, "http://test.com/"+pid)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestMain(m *testing.M) {
	// Get base URL from environment or use default
	baseURL = os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081" // Match the port in run_tests.sh
	}

	// Setup test DB
	if err := setupTestDB(); err != nil {
		panic(err)
	}

	// Create client
	apiClient = client.NewClient(baseURL)

	// Run tests
	os.Exit(m.Run())
}

func TestAuthentication(t *testing.T) {
	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		err := apiClient.Login(ctx, "test", "test")
		require.NoError(t, err)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		err := apiClient.Login(ctx, "invalid", "invalid")
		assert.Error(t, err)
	})
}

func TestSplitOperations(t *testing.T) {
	ctx := context.Background()

	// Login first
	err := apiClient.Login(ctx, "test", "test")
	require.NoError(t, err)

	// Create a test split first
	splitID := "test-split"
	pageIDs := []string{"page1", "page2", "page3", "page4"}

	// Delete the initial document first to free up the pages
	err = apiClient.DeleteDocument(ctx, "initial-doc")
	require.NoError(t, err)

	t.Run("create and delete document", func(t *testing.T) {
		// Create document
		doc, err := apiClient.CreateDocument(ctx, client.CreateDocumentRequest{
			SplitID:          splitID,
			Name:             "Test Document",
			Classification:   "Test Class",
			Filename:         "test.pdf",
			ShortDescription: "Test Description",
			PageIDs:          pageIDs[:2], // Use first two pages
		})
		require.NoError(t, err)
		assert.NotEmpty(t, doc.ID)

		// Delete document
		err = apiClient.DeleteDocument(ctx, doc.ID)
		require.NoError(t, err)
	})

	t.Run("move pages between documents", func(t *testing.T) {
		// Create source document
		sourceDoc, err := apiClient.CreateDocument(ctx, client.CreateDocumentRequest{
			SplitID:          splitID,
			Name:             "Source Document",
			Classification:   "Source Class",
			Filename:         "source.pdf",
			ShortDescription: "Source Description",
			PageIDs:          pageIDs[:3], // Use first three pages
		})
		require.NoError(t, err)

		// Create target document
		targetDoc, err := apiClient.CreateDocument(ctx, client.CreateDocumentRequest{
			SplitID:          splitID,
			Name:             "Target Document",
			Classification:   "Target Class",
			Filename:         "target.pdf",
			ShortDescription: "Target Description",
			PageIDs:          pageIDs[3:], // Use last page
		})
		require.NoError(t, err)

		// Move pages
		_, err = apiClient.MovePages(ctx, client.MovePagesRequest{
			SplitID:        splitID,
			FromDocumentID: sourceDoc.ID,
			ToDocumentID:   targetDoc.ID,
			PageIDs:        pageIDs[:2], // Move first two pages
		})
		require.NoError(t, err)

		// Cleanup
		err = apiClient.DeleteDocument(ctx, sourceDoc.ID)
		require.NoError(t, err)
		err = apiClient.DeleteDocument(ctx, targetDoc.ID)
		require.NoError(t, err)
	})

	t.Run("finalize split", func(t *testing.T) {
		// Create a document first
		_, err := apiClient.CreateDocument(ctx, client.CreateDocumentRequest{
			SplitID:          splitID,
			Name:             "Final Document",
			Classification:   "Final Class",
			Filename:         "final.pdf",
			ShortDescription: "Final Description",
			PageIDs:          pageIDs, // Use all pages
		})
		require.NoError(t, err)

		// Finalize split
		err = apiClient.FinalizeSplit(ctx, splitID)
		require.NoError(t, err)
	})
}

func TestMetrics(t *testing.T) {
	ctx := context.Background()

	t.Run("get metrics", func(t *testing.T) {
		metrics, err := apiClient.GetMetrics(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, metrics.RequestsTotal, int64(0))
		assert.GreaterOrEqual(t, metrics.ActiveConnections, int32(0))
	})
}

func TestErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Login first
	err := apiClient.Login(ctx, "test", "test")
	require.NoError(t, err)

	t.Run("non-existent split", func(t *testing.T) {
		_, err := apiClient.LoadSplit(ctx, "non-existent-split")
		assert.Error(t, err)
	})

	t.Run("non-existent document", func(t *testing.T) {
		err := apiClient.DeleteDocument(ctx, "non-existent-doc")
		assert.Error(t, err)
	})
}
