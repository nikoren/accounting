package domain

import (
	"fmt"
	"slices"
)

// Document represents one contiguous chunk of pages within a Split.
type Document struct {
	ID               string  // unique document identifier
	SplitID          string  // back‐reference to parent Split
	Name             string  // e.g. "John's W-2"
	Classification   string  // e.g. "W-2", "Invoice", etc.
	Filename         string  // original file name, e.g. "w2_2023.pdf"
	ShortDescription string  // human‐friendly summary
	Pages            []*Page // the actual page entities
	StartPage        string
	EndPage          string // lowest and highest page numbers in Pages
}

func NewDocument(
	id string,
	splitID,
	name,
	classification,
	filename,
	shortDescription string,
	pages []*Page,
) (*Document, error) {
	d := &Document{
		ID:               id,
		SplitID:          splitID,
		Name:             name,
		Classification:   classification,
		Filename:         filename,
		ShortDescription: shortDescription,
		Pages:            pages,
	}
	d.updatePageNumbers()

	if err := d.Valid(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	return d, nil
}

// DocumentMetadata represents optional fields that can be updated on a document
type DocumentMetadata struct {
	Name             *string // optional new name for the document
	Classification   *string // optional new classification
	ShortDescription *string // optional new description
}

// AddPages adds pages to the document
func (d *Document) AddPages(pages []*Page) error {
	for _, newPage := range pages {
		if assignErr := newPage.AssignToDocument(d.ID); assignErr != nil {
			return NewValidationError("failed to assign page to document", assignErr)
		}
		d.Pages = append(d.Pages, newPage)
	}
	d.updatePageNumbers()
	return nil
}

// RemovePages removes pages from the document and returns the removed pages
func (d *Document) RemovePages(pageIDs []string) ([]*Page, error) {
	if len(pageIDs) == 0 {
		return nil, nil
	}
	idsToRemove := make(map[string]struct{}, len(pageIDs))
	for _, id := range pageIDs {
		idsToRemove[id] = struct{}{}
	}
	removed := make([]*Page, 0)
	remaining := make([]*Page, 0, len(d.Pages))
	for _, page := range d.Pages {
		if _, shouldRemove := idsToRemove[page.ID]; shouldRemove {
			removed = append(removed, page)
		} else {
			remaining = append(remaining, page)
		}
	}
	if len(removed) == 0 {
		return nil, NewNotFoundError("none of the specified pages found in document", nil)
	}
	d.Pages = remaining
	for _, page := range removed {
		page.Unassign()
	}

	d.updatePageNumbers()
	return removed, nil
}

// UpdateMetadata updates document metadata
func (d *Document) UpdateMetadata(metadata DocumentMetadata) error {
	if metadata.Name != nil {
		d.Name = *metadata.Name
	}
	if metadata.Classification != nil {
		d.Classification = *metadata.Classification
	}
	if metadata.ShortDescription != nil {
		d.ShortDescription = *metadata.ShortDescription
	}
	return nil
}

func (d *Document) Valid() error {
	if d.ID == "" {
		return NewValidationError("document ID is required", nil)
	}
	if d.SplitID == "" {
		return NewValidationError("split ID is required", nil)
	}
	if d.Name == "" {
		return NewValidationError("document name is required", nil)
	}
	if d.Classification == "" {
		return NewValidationError("document classification is required", nil)
	}
	if d.Filename == "" {
		return NewValidationError("document filename is required", nil)
	}
	if len(d.Pages) == 0 {
		return NewValidationError("document must have at least one page", nil)
	}
	for _, page := range d.Pages {
		if err := page.Valid(); err != nil {
			return NewValidationError("invalid page in document", err)
		}
	}
	return nil
}

func (d *Document) updatePageNumbers() {

	// sort pages by PageNumber
	slices.SortFunc(d.Pages, func(a, b *Page) int {
		if a.PageNumber < b.PageNumber {
			return -1
		}
		if a.PageNumber > b.PageNumber {
			return 1
		}
		return 0
	})

	// Update StartPage and EndPage based on current pages
	if len(d.Pages) > 0 {
		d.StartPage = d.Pages[0].URL
		d.EndPage = d.Pages[len(d.Pages)-1].URL
	} else {
		d.StartPage = ""
		d.EndPage = ""
	}
}

// AssignToSplit assigns the document to a split
func (d *Document) AssignToSplit(splitID string) error {
	if d.SplitID != "" && d.SplitID != splitID {
		return NewConflictError("document is already assigned to a different split", nil)
	}
	d.SplitID = splitID
	for _, page := range d.Pages {
		page.SplitID = splitID
	}
	return nil
}
