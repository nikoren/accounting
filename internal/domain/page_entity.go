package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// Page represents a single page in a document or split
// The actual content is stored on the filesystem, this entity maintains only metadata
type Page struct {
	ID         string  // Unique identifier for the page
	SplitID    string  // ID of the split this page belongs to
	DocumentID *string // ID of the document this page belongs to (nil if unassigned)
	PageNumber int     // Original page number from the PDF
	URL        string  // URL to the page content on the filesystem
}

func NewPage(splitID, url string) (*Page, error) {
	// extract page number from URL
	var pageNumber int
	// assuming URL is in the format "page_1.png"
	_, err := fmt.Sscanf(url, "page_%d.png", &pageNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid page URL format: %w", err)
	}
	p := &Page{
		ID:         uuid.New().String(),
		SplitID:    splitID,
		URL:        url,
		PageNumber: pageNumber,
	}
	if p.Valid() != nil {
		return nil, fmt.Errorf("invalid page: %w", p.Valid())

	}
	return p, nil
}

func (p *Page) Valid() error {
	if p.ID == "" {
		return NewValidationError("page id is required", nil)
	}
	if p.SplitID == "" {
		return NewValidationError("split id is required", nil)
	}
	if p.URL == "" {
		return NewValidationError("url is required", nil)
	}
	return nil
}

func (p *Page) AssignToDocument(docID string) error {
	if p.IsAssigned() {
		return NewConflictError("page is already assigned to a document", nil)
	}
	p.DocumentID = &docID
	return nil
}

func (p *Page) Unassign() {
	p.DocumentID = nil
}

func (p *Page) IsAssigned() bool {
	return p.DocumentID != nil
}

// UnassignFromDocument unassigns the page from its document
func (p *Page) UnassignFromDocument() error {
	p.DocumentID = nil
	return nil
}
