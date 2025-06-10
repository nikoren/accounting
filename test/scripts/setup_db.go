package main

import (
	"accounting/internal/client"
	"context"
	"database/sql"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Create split and initial document directly in the database
	dbPath := os.Getenv("APP_DB_PATH")
	if dbPath == "" {
		dbPath = "test_accounting.db"
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Insert test split
	now := time.Now()
	_, err = db.Exec(`INSERT INTO splits (id, client_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		"test-split", "test-client", "draft", now, now)
	if err != nil {
		panic(err)
	}

	// Insert initial document
	_, err = db.Exec(`INSERT INTO documents (id, split_id, name, classification, filename, short_description, start_page, end_page) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"initial-doc", "test-split", "Initial Document", "Test Class", "initial.pdf", "Initial Description", "page1", "page4")
	if err != nil {
		panic(err)
	}

	// Insert test pages (assigned to the initial document)
	pageIDs := []string{"page1", "page2", "page3", "page4"}
	for i, pid := range pageIDs {
		_, err = db.Exec(`INSERT INTO pages (id, split_id, document_id, page_number, url) VALUES (?, ?, ?, ?, ?)`,
			pid, "test-split", "initial-doc", i+1, "http://test.com/"+pid)
		if err != nil {
			panic(err)
		}
	}

	// Create client for any additional API operations
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}
	c := client.NewClient(baseURL)

	// Login
	ctx := context.Background()
	if err := c.Login(ctx, "test", "test"); err != nil {
		panic(err)
	}

	// Wait for everything to be processed
	time.Sleep(time.Second)
}
