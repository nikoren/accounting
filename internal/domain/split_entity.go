package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Split is the aggregate root for one AI‐generated split of a PDF bundle.
type Split struct {
	ID              string      // unique split identifier
	ClientID        string      // who owns this split
	Status          SplitStatus // draft | finalized
	Documents       []Document  // all docs in this split
	UnassignedPages []*Page     // pages not yet in any document

	CreatedAt   time.Time  // when split was created
	UpdatedAt   time.Time  // when split was last updated
	FinalizedAt *time.Time // set when Status == Finalized
}

func NewSplit(jsonRepr string) (*Split, error) {
	var splitData struct {
		ID        string      `json:"split_id"`
		ClientID  string      `json:"client_id"`
		Status    SplitStatus `json:"status"`
		Documents []struct {
			ID               string   `json:"id"`
			Classification   string   `json:"classification"`
			Filename         string   `json:"file_name"`
			Name             string   `json:"name"`
			ShortDescription string   `json:"short_description"`
			PageURLs         []string `json:"page_urls"`
		} `json:"documents"`
	}

	if err := json.Unmarshal([]byte(jsonRepr), &splitData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal split JSON: %w", err)
	}

	// Create the split
	split := &Split{
		ID:              splitData.ID,
		ClientID:        splitData.ClientID,
		Status:          splitData.Status,
		Documents:       make([]Document, 0, len(splitData.Documents)),
		UnassignedPages: make([]*Page, 0),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Process each document
	for _, docData := range splitData.Documents {
		// Create pages for this document
		pages := make([]*Page, 0, len(docData.PageURLs))
		for _, url := range docData.PageURLs {
			page, err := NewPage(splitData.ID, url)
			if err != nil {
				return nil, fmt.Errorf("failed to create page from URL %s: %w", url, err)
			}
			pages = append(pages, page)
		}

		// Create the document
		doc, err := NewDocument(
			docData.ID,
			splitData.ID,
			docData.Name,
			docData.Classification,
			docData.Filename,
			docData.ShortDescription,
			pages,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create document: %w", err)
		}

		split.Documents = append(split.Documents, *doc)
	}

	// Validate the split
	if err := split.Valid(); err != nil {
		return nil, fmt.Errorf("invalid split: %w", err)
	}

	return split, nil
}

func (s *Split) String() string {
	return fmt.Sprintf("Split %v (%s): %d documents, %d unassigned pages, status: %s",
		s.ID, s.ClientID, len(s.Documents), len(s.UnassignedPages), s.Status)
}

func (s *Split) Valid() error {
	if s.ID == "" {
		return NewValidationError("split ID is required", nil)
	}
	if s.ClientID == "" {
		return NewValidationError("client ID is required", nil)
	}
	for _, doc := range s.Documents {
		if docErr := doc.Valid(); docErr != nil {
			return fmt.Errorf("invalid document in split %v: %w", s.ID, docErr)
		}
	}
	return nil
}

// Finalize marks the split as finalized
func (s *Split) Finalize(finalizedAt time.Time) error {
	if s.Status == SplitStatusFinalized {
		return NewConflictError("split already finalized", nil)
	}

	if len(s.UnassignedPages) > 0 {
		return NewValidationError("cannot finalize split with unassigned pages", nil)
	}

	if validErr := s.Valid(); validErr != nil {
		return NewValidationError("invalid split", validErr)
	}

	s.Status = SplitStatusFinalized
	s.FinalizedAt = &finalizedAt
	return nil
}

// AddDocument adds a new document to the split
func (s *Split) AddDocument(doc *Document) error {
	if s.Status == SplitStatusFinalized {
		return NewConflictError("cannot add document to finalized split", nil)
	}
	if err := doc.Valid(); err != nil {
		return NewValidationError("invalid document", err)
	}
	for _, existingDoc := range s.Documents {
		if existingDoc.ID == doc.ID {
			return NewConflictError("document with ID already exists", nil)
		}
	}
	for _, page := range doc.Pages {
		if page.IsAssigned() {
			return NewConflictError("cannot add document with already assigned pages", nil)
		}
	}
	s.Documents = append(s.Documents, *doc)
	return nil
}

// RemoveDocument removes a document from the split
func (s *Split) RemoveDocument(docID string) error {
	if s.Status == SplitStatusFinalized {
		return NewConflictError("cannot remove document from finalized split", nil)
	}
	for i, doc := range s.Documents {
		if doc.ID == docID {
			// Remove all pages from the document and get them as unassigned
			pageIDs := make([]string, len(doc.Pages))
			for idx, p := range doc.Pages {
				pageIDs[idx] = p.ID
			}
			removedPages, err := doc.RemovePages(pageIDs)
			if err != nil {
				var domainErr *DomainError
				if errors.As(err, &domainErr) {
					return err
				}
				return NewValidationError("failed to remove pages from document", err)
			}
			// Remove the document from the list
			s.Documents = append(s.Documents[:i], s.Documents[i+1:]...)
			// Add the removed pages to unassigned pages
			s.UnassignedPages = append(s.UnassignedPages, removedPages...)
			return nil
		}
	}
	return NewNotFoundError("document not found in split", nil)
}

// MovePages moves pages between documents
func (s *Split) MovePages(fromDocID, toDocID string, pageIDs []string) error {
	if s.Status == SplitStatusFinalized {
		return NewConflictError("cannot move pages in finalized split", nil)
	}

	var fromDoc, toDoc *Document
	for i := range s.Documents {
		if s.Documents[i].ID == fromDocID {
			fromDoc = &s.Documents[i]
		}
		if s.Documents[i].ID == toDocID {
			toDoc = &s.Documents[i]
		}
	}
	if fromDoc == nil {
		return NewNotFoundError("source document not found", nil)
	}
	if toDoc == nil {
		return NewNotFoundError("target document not found", nil)
	}

	// Check if any page to be moved is already present in the target document
	for _, pid := range pageIDs {
		for _, p := range toDoc.Pages {
			if p.ID == pid {
				return NewValidationError("cannot move pages that are already assigned to target document", nil)
			}
		}
	}

	removedPages, err := fromDoc.RemovePages(pageIDs)
	if err != nil {
		return NewValidationError("failed to remove pages from source document", err)
	}
	if err := toDoc.AddPages(removedPages); err != nil {
		return NewValidationError("failed to add pages to target document", err)
	}
	return nil
}

// UpdateDocumentMetadata updates document metadata
func (s *Split) UpdateDocumentMetadata(docID string, meta DocumentMetadata) error {
	if s.Status == SplitStatusFinalized {
		return fmt.Errorf("cannot update document in finalized split %v", s.ID)
	}

	// Find document
	for i := range s.Documents {
		if s.Documents[i].ID == docID {
			return s.Documents[i].UpdateMetadata(meta)
		}
	}

	return fmt.Errorf("document %v not found in split %v", docID, s.ID)
}

func (s *Split) findDoc(fromDocID string) (*Document, error) {
	// Find source document
	var fromDoc *Document
	for i := range s.Documents {
		if s.Documents[i].ID == fromDocID {
			fromDoc = &s.Documents[i]
			break
		}
	}
	if fromDoc == nil {
		return nil, fmt.Errorf("document %v not found in split %v", fromDocID, s.ID)
	}
	return fromDoc, nil
}

func Map[T any, R any](slice []T, f func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = f(v)
	}
	return result
}
