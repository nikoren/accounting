package splits

import (
	"accounting/internal/domain"
	"context"
	"database/sql"
	"fmt"
)

// SplitRepositorySQL implements domain.SplitRepository using SQLite
type SplitRepositorySQL struct {
	tx *sql.Tx
}

// NewSplitRepositorySQL creates a new SQLite-based split repository
func NewSplitRepositorySQL(tx *sql.Tx) *SplitRepositorySQL {
	return &SplitRepositorySQL{tx: tx}
}

// Get retrieves a split by ID
func (r *SplitRepositorySQL) Get(ctx context.Context, id string) (*domain.Split, error) {
	// Get split
	var split domain.Split
	err := r.tx.QueryRowContext(ctx, `
		SELECT id, client_id, status, created_at, updated_at
		FROM splits
		WHERE id = ?
	`, id).Scan(&split.ID, &split.ClientID, &split.Status, &split.CreatedAt, &split.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting split: %w", err)
	}

	// Get documents
	documents, err := r.getDocuments(ctx, id)
	if err != nil {
		return nil, err
	}
	split.Documents = documents

	// Get unassigned pages
	unassignedPages, err := r.getUnassignedPages(ctx, id)
	if err != nil {
		return nil, err
	}
	split.UnassignedPages = unassignedPages

	return &split, nil
}

// Save persists a split aggregate
func (r *SplitRepositorySQL) Save(ctx context.Context, split *domain.Split) error {
	// Save split
	_, err := r.tx.ExecContext(ctx, `
		INSERT INTO splits (id, client_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			client_id = excluded.client_id,
			status = excluded.status,
			updated_at = excluded.updated_at
	`, split.ID, split.ClientID, split.Status, split.CreatedAt, split.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error saving split: %w", err)
	}

	// Delete documents not present in split.Documents
	docIDs := make(map[string]struct{}, len(split.Documents))
	for _, doc := range split.Documents {
		docIDs[doc.ID] = struct{}{}
	}
	rows, err := r.tx.QueryContext(ctx, "SELECT id FROM documents WHERE split_id = ?", split.ID)
	if err != nil {
		return fmt.Errorf("error querying documents for deletion: %w", err)
	}
	var toDeleteDocIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("error scanning document id: %w", err)
		}
		if _, ok := docIDs[id]; !ok {
			toDeleteDocIDs = append(toDeleteDocIDs, id)
		}
	}
	rows.Close()
	for _, id := range toDeleteDocIDs {
		_, err := r.tx.ExecContext(ctx, "DELETE FROM documents WHERE id = ?", id)
		if err != nil {
			return fmt.Errorf("error deleting document: %w", err)
		}
		_, err = r.tx.ExecContext(ctx, "DELETE FROM pages WHERE document_id = ?", id)
		if err != nil {
			return fmt.Errorf("error deleting pages for document: %w", err)
		}
	}

	// Delete pages not present in split.Documents or split.UnassignedPages
	pageIDs := make(map[string]struct{})
	for _, doc := range split.Documents {
		for _, page := range doc.Pages {
			pageIDs[page.ID] = struct{}{}
		}
	}
	for _, page := range split.UnassignedPages {
		pageIDs[page.ID] = struct{}{}
	}
	rows, err = r.tx.QueryContext(ctx, "SELECT id FROM pages WHERE split_id = ?", split.ID)
	if err != nil {
		return fmt.Errorf("error querying pages for deletion: %w", err)
	}
	var toDeletePageIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("error scanning page id: %w", err)
		}
		if _, ok := pageIDs[id]; !ok {
			toDeletePageIDs = append(toDeletePageIDs, id)
		}
	}
	rows.Close()
	for _, id := range toDeletePageIDs {
		_, err := r.tx.ExecContext(ctx, "DELETE FROM pages WHERE id = ?", id)
		if err != nil {
			return fmt.Errorf("error deleting page: %w", err)
		}
	}

	// Save documents
	for _, doc := range split.Documents {
		_, err = r.tx.ExecContext(ctx, `
			INSERT INTO documents (id, split_id, name, classification, filename, short_description, start_page, end_page)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				split_id = excluded.split_id,
				name = excluded.name,
				classification = excluded.classification,
				filename = excluded.filename,
				short_description = excluded.short_description,
				start_page = excluded.start_page,
				end_page = excluded.end_page
		`, doc.ID, doc.SplitID, doc.Name, doc.Classification, doc.Filename, doc.ShortDescription, doc.StartPage, doc.EndPage)
		if err != nil {
			return fmt.Errorf("error saving document: %w", err)
		}

		// Save pages
		for _, page := range doc.Pages {
			_, err = r.tx.ExecContext(ctx, `
				INSERT INTO pages (id, split_id, document_id, page_number, url)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET
					split_id = excluded.split_id,
					document_id = excluded.document_id,
					page_number = excluded.page_number,
					url = excluded.url
			`, page.ID, page.SplitID, doc.ID, page.PageNumber, page.URL)
			if err != nil {
				return fmt.Errorf("error saving page: %w", err)
			}
		}
	}

	// Save unassigned pages
	for _, page := range split.UnassignedPages {
		_, err = r.tx.ExecContext(ctx, `
			INSERT INTO pages (id, split_id, document_id, page_number, url)
			VALUES (?, ?, NULL, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				split_id = excluded.split_id,
				document_id = excluded.document_id,
				page_number = excluded.page_number,
				url = excluded.url
		`, page.ID, page.SplitID, page.PageNumber, page.URL)
		if err != nil {
			return fmt.Errorf("error saving unassigned page: %w", err)
		}
	}

	return nil
}

// Delete removes a split
func (r *SplitRepositorySQL) Delete(ctx context.Context, id string) error {
	// Delete pages first (due to foreign key constraints)
	_, err := r.tx.ExecContext(ctx, "DELETE FROM pages WHERE split_id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting pages: %w", err)
	}

	// Delete documents
	_, err = r.tx.ExecContext(ctx, "DELETE FROM documents WHERE split_id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting documents: %w", err)
	}

	// Delete split
	_, err = r.tx.ExecContext(ctx, "DELETE FROM splits WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting split: %w", err)
	}

	return nil
}

// ListByClientID retrieves all splits for a client
func (r *SplitRepositorySQL) ListByClientID(ctx context.Context, clientID string) ([]*domain.Split, error) {
	rows, err := r.tx.QueryContext(ctx, `
		SELECT id, client_id, status, created_at, updated_at
		FROM splits
		WHERE client_id = ?
		ORDER BY created_at DESC
	`, clientID)
	if err != nil {
		return nil, fmt.Errorf("error listing splits: %w", err)
	}
	defer rows.Close()

	var splits []*domain.Split
	for rows.Next() {
		var split domain.Split
		err := rows.Scan(&split.ID, &split.ClientID, &split.Status, &split.CreatedAt, &split.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning split: %w", err)
		}

		// Get documents
		documents, err := r.getDocuments(ctx, split.ID)
		if err != nil {
			return nil, err
		}
		split.Documents = documents

		// Get unassigned pages
		unassignedPages, err := r.getUnassignedPages(ctx, split.ID)
		if err != nil {
			return nil, err
		}
		split.UnassignedPages = unassignedPages

		splits = append(splits, &split)
	}

	return splits, nil
}

// GetSplitIDByDocumentID retrieves the split ID for a given document ID
func (r *SplitRepositorySQL) GetSplitIDByDocumentID(ctx context.Context, documentID string) (string, error) {
	var splitID string
	err := r.tx.QueryRowContext(ctx, "SELECT split_id FROM documents WHERE id = ?", documentID).Scan(&splitID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("document %v not found", documentID)
	}
	if err != nil {
		return "", fmt.Errorf("error getting split ID: %w", err)
	}
	return splitID, nil
}

// getDocuments retrieves all documents for a split
func (r *SplitRepositorySQL) getDocuments(ctx context.Context, splitID string) ([]domain.Document, error) {
	rows, err := r.tx.QueryContext(ctx, `
		SELECT id, split_id, name, classification, filename, short_description, start_page, end_page
		FROM documents
		WHERE split_id = ?
		ORDER BY start_page
	`, splitID)
	if err != nil {
		return nil, fmt.Errorf("error getting documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.Document
	for rows.Next() {
		var doc domain.Document
		err := rows.Scan(&doc.ID, &doc.SplitID, &doc.Name, &doc.Classification, &doc.Filename, &doc.ShortDescription, &doc.StartPage, &doc.EndPage)
		if err != nil {
			return nil, fmt.Errorf("error scanning document: %w", err)
		}

		// Get pages
		pages, err := r.getPages(ctx, doc.ID)
		if err != nil {
			return nil, err
		}
		doc.Pages = pages

		documents = append(documents, doc)
	}

	return documents, nil
}

// getUnassignedPages retrieves all unassigned pages for a split
func (r *SplitRepositorySQL) getUnassignedPages(ctx context.Context, splitID string) ([]*domain.Page, error) {
	rows, err := r.tx.QueryContext(ctx, `
		SELECT id, split_id, page_number, url
		FROM pages
		WHERE split_id = ? AND document_id IS NULL
		ORDER BY page_number
	`, splitID)
	if err != nil {
		return nil, fmt.Errorf("error getting unassigned pages: %w", err)
	}
	defer rows.Close()

	var pages []*domain.Page
	for rows.Next() {
		var page domain.Page
		err := rows.Scan(&page.ID, &page.SplitID, &page.PageNumber, &page.URL)
		if err != nil {
			return nil, fmt.Errorf("error scanning page: %w", err)
		}
		pages = append(pages, &page)
	}

	return pages, nil
}

// getPages retrieves all pages for a document
func (r *SplitRepositorySQL) getPages(ctx context.Context, documentID string) ([]*domain.Page, error) {
	rows, err := r.tx.QueryContext(ctx, `
		SELECT id, split_id, page_number, url
		FROM pages
		WHERE document_id = ?
		ORDER BY page_number
	`, documentID)
	if err != nil {
		return nil, fmt.Errorf("error getting pages: %w", err)
	}
	defer rows.Close()

	var pages []*domain.Page
	for rows.Next() {
		var page domain.Page
		err := rows.Scan(&page.ID, &page.SplitID, &page.PageNumber, &page.URL)
		if err != nil {
			return nil, fmt.Errorf("error scanning page: %w", err)
		}
		pages = append(pages, &page)
	}

	return pages, nil
}
