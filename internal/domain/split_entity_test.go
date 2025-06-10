package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSplit(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		check   func(t *testing.T, split *Split)
	}{
		{
			name: "valid split with one document",
			json: `{
				"split_id": "split123",
				"client_id": "client456",
				"status": "draft",
				"documents": [
					{
						"id": "doc1",
						"classification": "W-2",
						"file_name": "w2_2023.pdf",
						"name": "John W2 Form",
						"short_description": "W2 from employer",
						"page_urls": ["page_1.png", "page_2.png", "page_3.png"]
					}
				]
			}`,
			wantErr: false,
			check: func(t *testing.T, split *Split) {
				assert.Equal(t, "split123", split.ID)
				assert.Equal(t, "client456", split.ClientID)
				assert.Equal(t, SplitStatusDraft, split.Status)
				assert.Len(t, split.Documents, 1)
				assert.Len(t, split.UnassignedPages, 0)

				doc := split.Documents[0]
				assert.Equal(t, "doc1", doc.ID)
				assert.Equal(t, "W-2", doc.Classification)
				assert.Equal(t, "w2_2023.pdf", doc.Filename)
				assert.Equal(t, "John W2 Form", doc.Name)
				assert.Equal(t, "W2 from employer", doc.ShortDescription)
				assert.Len(t, doc.Pages, 3)
			},
		},
		{
			name: "valid split with multiple documents",
			json: `{
				"split_id": "split123",
				"client_id": "client456",
				"status": "draft",
				"documents": [
					{
						"id": "doc1",
						"classification": "W-2",
						"file_name": "w2_2023.pdf",
						"name": "John W2 Form",
						"short_description": "W2 from employer",
						"page_urls": ["page_1.png", "page_2.png"]
					},
					{
						"id": "doc2",
						"classification": "Invoice",
						"file_name": "invoice_2023.pdf",
						"name": "Invoice #123",
						"short_description": "Monthly invoice",
						"page_urls": ["page_3.png", "page_4.png"]
					}
				]
			}`,
			wantErr: false,
			check: func(t *testing.T, split *Split) {
				assert.Equal(t, "split123", split.ID)
				assert.Equal(t, "client456", split.ClientID)
				assert.Equal(t, SplitStatusDraft, split.Status)
				assert.Len(t, split.Documents, 2)
				assert.Len(t, split.UnassignedPages, 0)

				// Check first document
				doc1 := split.Documents[0]
				assert.Equal(t, "doc1", doc1.ID)
				assert.Equal(t, "W-2", doc1.Classification)
				assert.Len(t, doc1.Pages, 2)

				// Check second document
				doc2 := split.Documents[1]
				assert.Equal(t, "doc2", doc2.ID)
				assert.Equal(t, "Invoice", doc2.Classification)
				assert.Len(t, doc2.Pages, 2)
			},
		},
		{
			name:    "invalid JSON",
			json:    `{invalid json}`,
			wantErr: true,
		},
		{
			name: "empty documents",
			json: `{
				"split_id": "split123",
				"client_id": "client456",
				"status": "draft",
				"documents": []
			}`,
			wantErr: false,
			check: func(t *testing.T, split *Split) {
				assert.Equal(t, "split123", split.ID)
				assert.Equal(t, "client456", split.ClientID)
				assert.Equal(t, SplitStatusDraft, split.Status)
				assert.Len(t, split.Documents, 0)
				assert.Len(t, split.UnassignedPages, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split, err := NewSplit(tt.json)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, split)

			// Check timestamps
			assert.False(t, split.CreatedAt.IsZero())
			assert.False(t, split.UpdatedAt.IsZero())
			assert.Nil(t, split.FinalizedAt)

			// Run custom checks
			if tt.check != nil {
				tt.check(t, split)
			}

			// Validate the split
			assert.NoError(t, split.Valid())
		})
	}
}

func TestSplit_AddDocument(t *testing.T) {
	// Helper function to create a test split
	createTestSplit := func(status SplitStatus) *Split {
		return &Split{
			ID:              "split123",
			ClientID:        "client456",
			Status:          status,
			Documents:       make([]Document, 0),
			UnassignedPages: make([]*Page, 0),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
	}

	// Helper function to create a test document
	createTestDocument := func(id string, pages []*Page) *Document {
		doc, err := NewDocument(
			id,
			"split123",
			"Test Document",
			"W-2",
			"test.pdf",
			"Test Description",
			pages,
		)
		require.NoError(t, err)
		return doc
	}

	// Helper function to create test pages
	createTestPages := func(count int) []*Page {
		pages := make([]*Page, count)
		for i := 0; i < count; i++ {
			page, err := NewPage("split123", fmt.Sprintf("page_%d.png", i+1))
			require.NoError(t, err)
			pages[i] = page
		}
		return pages
	}

	tests := []struct {
		name        string
		setup       func() (*Split, *Document)
		wantErr     bool
		errContains string
		check       func(t *testing.T, split *Split)
	}{
		{
			name: "successfully add document to draft split",
			setup: func() (*Split, *Document) {
				split := createTestSplit(SplitStatusDraft)
				pages := createTestPages(2)
				doc := createTestDocument("doc1", pages)
				return split, doc
			},
			wantErr: false,
			check: func(t *testing.T, split *Split) {
				assert.Len(t, split.Documents, 1)
				doc := split.Documents[0]
				assert.Equal(t, "doc1", doc.ID)
				assert.Equal(t, "split123", doc.SplitID)
				assert.Len(t, doc.Pages, 2)
			},
		},
		{
			name: "cannot add document to finalized split",
			setup: func() (*Split, *Document) {
				split := createTestSplit(SplitStatusFinalized)
				pages := createTestPages(2)
				doc := createTestDocument("doc1", pages)
				return split, doc
			},
			wantErr:     true,
			errContains: "cannot add document to finalized split",
		},
		{
			name: "cannot add document with missing required fields",
			setup: func() (*Split, *Document) {
				split := createTestSplit(SplitStatusDraft)
				pages := createTestPages(2)
				// Create a document with missing required fields
				doc := &Document{
					ID:      "doc1",
					SplitID: "split123",
					// Missing Name
					Classification: "W-2",
					// Missing Filename
					ShortDescription: "Test Description",
					Pages:            pages,
				}
				return split, doc
			},
			wantErr:     true,
			errContains: "invalid document",
		},
		{
			name: "cannot add document with duplicate ID",
			setup: func() (*Split, *Document) {
				split := createTestSplit(SplitStatusDraft)
				pages := createTestPages(2)
				doc := createTestDocument("doc1", pages)
				require.NoError(t, split.AddDocument(doc))
				doc2 := createTestDocument("doc1", createTestPages(2))
				return split, doc2
			},
			wantErr:     true,
			errContains: "document with ID already exists",
		},
		{
			name: "cannot add document with already assigned pages",
			setup: func() (*Split, *Document) {
				split := createTestSplit(SplitStatusDraft)
				page, err := NewPage("split1", "page_1.png")
				require.NoError(t, err)
				require.NoError(t, page.AssignToDocument("doc1"))
				doc := createTestDocument("doc1", []*Page{page})
				return split, doc
			},
			wantErr:     true,
			errContains: "cannot add document with already assigned pages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split, doc := tt.setup()
			err := split.AddDocument(doc)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, split)
			}
		})
	}
}

func TestSplit_RemoveDocument(t *testing.T) {
	// Helper function to create a test split
	createTestSplit := func(status SplitStatus) *Split {
		return &Split{
			ID:              "split123",
			ClientID:        "client456",
			Status:          status,
			Documents:       make([]Document, 0),
			UnassignedPages: make([]*Page, 0),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
	}

	// Helper function to create a test document
	createTestDocument := func(id string, pages []*Page) *Document {
		doc, err := NewDocument(
			id,
			"split123",
			"Test Document",
			"W-2",
			"test.pdf",
			"Test Description",
			pages,
		)
		require.NoError(t, err)
		return doc
	}

	// Helper function to create test pages
	createTestPages := func(count int) []*Page {
		pages := make([]*Page, count)
		for i := 0; i < count; i++ {
			page, err := NewPage("split123", fmt.Sprintf("page_%d.png", i+1))
			require.NoError(t, err)
			pages[i] = page
		}
		return pages
	}

	tests := []struct {
		name        string
		setup       func() *Split
		docID       string
		wantErr     bool
		errContains string
		check       func(t *testing.T, split *Split)
	}{
		{
			name: "successfully remove document from draft split",
			setup: func() *Split {
				split := createTestSplit(SplitStatusDraft)
				pages := createTestPages(2)
				doc := createTestDocument("doc1", pages)
				require.NoError(t, split.AddDocument(doc))
				return split
			},
			docID:   "doc1",
			wantErr: false,
			check: func(t *testing.T, split *Split) {
				assert.Len(t, split.Documents, 0)
				for _, page := range split.UnassignedPages {
					assert.Nil(t, page.DocumentID)
				}
			},
		},
		{
			name: "cannot remove document from finalized split",
			setup: func() *Split {
				split := createTestSplit(SplitStatusFinalized)
				// Do not add any document to the finalized split
				return split
			},
			docID:       "doc1",
			wantErr:     true,
			errContains: "cannot remove document from finalized split",
		},
		{
			name: "cannot remove non-existent document",
			setup: func() *Split {
				split := createTestSplit(SplitStatusDraft)
				return split
			},
			docID:       "nonexistent",
			wantErr:     true,
			errContains: "not found in split",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := tt.setup()
			err := split.RemoveDocument(tt.docID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, split)
			}
		})
	}
}

func TestSplit_MovePages(t *testing.T) {
	// Helper function to create a test split
	createTestSplit := func(status SplitStatus) *Split {
		return &Split{
			ID:              "split123",
			ClientID:        "client456",
			Status:          status,
			Documents:       make([]Document, 0),
			UnassignedPages: make([]*Page, 0),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
	}

	// Helper function to create a test document
	createTestDocument := func(id string, pages []*Page) *Document {
		doc, err := NewDocument(
			id,
			"split123",
			"Test Document",
			"W-2",
			"test.pdf",
			"Test Description",
			pages,
		)
		require.NoError(t, err)
		return doc
	}

	// Helper function to create test pages
	createTestPages := func(count int) []*Page {
		pages := make([]*Page, count)
		for i := 0; i < count; i++ {
			page, err := NewPage("split123", fmt.Sprintf("page_%d.png", i+1))
			require.NoError(t, err)
			pages[i] = page
		}
		return pages
	}

	tests := []struct {
		name        string
		setup       func() *Split
		fromDocID   string
		toDocID     string
		pageIDs     []string
		wantErr     bool
		errContains string
		check       func(t *testing.T, split *Split)
	}{
		{
			name: "successfully move pages between documents",
			setup: func() *Split {
				split := createTestSplit(SplitStatusDraft)
				// Create source document with 3 pages
				sourcePages := createTestPages(3)
				sourceDoc := createTestDocument("doc1", sourcePages)
				require.NoError(t, split.AddDocument(sourceDoc))

				// Create target document with 2 pages
				targetPages := createTestPages(2)
				targetDoc := createTestDocument("doc2", targetPages)
				require.NoError(t, split.AddDocument(targetDoc))

				return split
			},
			fromDocID: "doc1",
			toDocID:   "doc2",
			pageIDs:   nil, // will be set in test body
			wantErr:   false,
			check: func(t *testing.T, split *Split) {
				// Find source and target documents
				var sourceDoc, targetDoc *Document
				for i := range split.Documents {
					if split.Documents[i].ID == "doc1" {
						sourceDoc = &split.Documents[i]
					}
					if split.Documents[i].ID == "doc2" {
						targetDoc = &split.Documents[i]
					}
				}
				require.NotNil(t, sourceDoc)
				require.NotNil(t, targetDoc)

				// Source document should have 1 page left
				assert.Len(t, sourceDoc.Pages, 1)
				// Target document should have 4 pages
				assert.Len(t, targetDoc.Pages, 4)
			},
		},
		{
			name: "cannot move pages in finalized split",
			setup: func() *Split {
				split := createTestSplit(SplitStatusDraft)
				// Create source document with 3 pages
				sourcePages := createTestPages(3)
				sourceDoc := createTestDocument("doc1", sourcePages)
				require.NoError(t, split.AddDocument(sourceDoc))

				// Create target document with 2 pages
				targetPages := createTestPages(2)
				targetDoc := createTestDocument("doc2", targetPages)
				require.NoError(t, split.AddDocument(targetDoc))

				// Set split to finalized after adding documents
				split.Status = SplitStatusFinalized
				return split
			},
			fromDocID:   "doc1",
			toDocID:     "doc2",
			pageIDs:     nil, // will be set in test body
			wantErr:     true,
			errContains: "cannot move pages in finalized split",
		},
		{
			name: "cannot move pages from non-existent source document",
			setup: func() *Split {
				split := createTestSplit(SplitStatusDraft)
				doc := createTestDocument("doc1", createTestPages(2))
				require.NoError(t, split.AddDocument(doc))
				return split
			},
			fromDocID:   "non_existent",
			toDocID:     "doc1",
			pageIDs:     []string{"page_1.png"},
			wantErr:     true,
			errContains: "source document not found",
		},
		{
			name: "cannot move pages to non-existent target document",
			setup: func() *Split {
				split := createTestSplit(SplitStatusDraft)
				doc := createTestDocument("doc1", createTestPages(2))
				require.NoError(t, split.AddDocument(doc))
				return split
			},
			fromDocID:   "doc1",
			toDocID:     "non_existent",
			pageIDs:     []string{"page_1.png"},
			wantErr:     true,
			errContains: "target document not found",
		},
		{
			name: "cannot move non-existent pages",
			setup: func() *Split {
				split := createTestSplit(SplitStatusDraft)
				// Create source document with 3 pages
				sourcePages := createTestPages(3)
				sourceDoc := createTestDocument("doc1", sourcePages)
				require.NoError(t, split.AddDocument(sourceDoc))

				// Create target document with 2 pages
				targetPages := createTestPages(2)
				targetDoc := createTestDocument("doc2", targetPages)
				require.NoError(t, split.AddDocument(targetDoc))

				return split
			},
			fromDocID:   "doc1",
			toDocID:     "doc2",
			pageIDs:     []string{"nonexistent.png"},
			wantErr:     true,
			errContains: "none of the specified pages found in document",
		},
		{
			name: "cannot move pages that are already in target document",
			setup: func() *Split {
				split := createTestSplit(SplitStatusDraft)
				// Create a single page instance
				page, err := NewPage("split123", "page_1.png")
				require.NoError(t, err)
				doc1 := createTestDocument("doc1", []*Page{page})
				doc2 := createTestDocument("doc2", []*Page{page})
				require.NoError(t, split.AddDocument(doc1))
				require.NoError(t, split.AddDocument(doc2))
				return split
			},
			fromDocID:   "doc1",
			toDocID:     "doc2",
			pageIDs:     []string{"page_1.png"},
			wantErr:     true,
			errContains: "already assigned to target document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := tt.setup()

			// For the "successfully move pages between documents" test, set pageIDs to the actual IDs
			if tt.name == "successfully move pages between documents" {
				var sourceDoc *Document
				for i := range split.Documents {
					if split.Documents[i].ID == tt.fromDocID {
						sourceDoc = &split.Documents[i]
						break
					}
				}
				require.NotNil(t, sourceDoc)
				if len(sourceDoc.Pages) >= 2 {
					// Move the first two pages
					tt.pageIDs = []string{sourceDoc.Pages[0].ID, sourceDoc.Pages[1].ID}
				} else {
					t.Fatal("not enough pages in source document for test")
				}
			}

			// For the "cannot move pages in finalized split" test, set pageIDs to the actual IDs
			if tt.name == "cannot move pages in finalized split" {
				var sourceDoc *Document
				for i := range split.Documents {
					if split.Documents[i].ID == tt.fromDocID {
						sourceDoc = &split.Documents[i]
						break
					}
				}
				require.NotNil(t, sourceDoc)
				if len(sourceDoc.Pages) >= 2 {
					tt.pageIDs = []string{sourceDoc.Pages[0].ID, sourceDoc.Pages[1].ID}
				} else {
					t.Fatal("not enough pages in source document for test")
				}
			}

			// For the "cannot move pages that are already in target document" test, set pageIDs to the actual IDs
			if tt.name == "cannot move pages that are already in target document" {
				var sourceDoc *Document
				for i := range split.Documents {
					if split.Documents[i].ID == tt.fromDocID {
						sourceDoc = &split.Documents[i]
						break
					}
				}
				require.NotNil(t, sourceDoc)
				if len(split.Documents) > 1 {
					var pageID string
					if len(split.Documents[1].Pages) > 0 {
						pageID = split.Documents[1].Pages[0].ID
						tt.pageIDs = []string{pageID}
					} else {
						t.Fatal("not enough pages in target document for test")
					}
				} else {
					t.Fatal("not enough documents in split for test")
				}
			}

			err := split.MovePages(tt.fromDocID, tt.toDocID, tt.pageIDs)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					if tt.name == "cannot move pages that are already in target document" {
						assert.Contains(t, err.Error(), "already assigned to target document")
					} else {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				}
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, split)
			}
		})
	}
}

func TestSplit_Finalize(t *testing.T) {
	// Helper function to create a test split
	createTestSplit := func(status SplitStatus) *Split {
		return &Split{
			ID:              "split123",
			ClientID:        "client456",
			Status:          status,
			Documents:       make([]Document, 0),
			UnassignedPages: make([]*Page, 0),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
	}

	// Helper function to create a test document
	createTestDocument := func(id string, pages []*Page) *Document {
		doc, err := NewDocument(
			id,
			"split123",
			"Test Document",
			"W-2",
			"test.pdf",
			"Test Description",
			pages,
		)
		require.NoError(t, err)
		return doc
	}

	// Helper function to create test pages
	createTestPages := func(count int) []*Page {
		pages := make([]*Page, count)
		for i := 0; i < count; i++ {
			page, err := NewPage("split123", fmt.Sprintf("page_%d.png", i+1))
			require.NoError(t, err)
			pages[i] = page
		}
		return pages
	}

	tests := []struct {
		name        string
		setup       func() *Split
		wantErr     bool
		errContains string
		check       func(t *testing.T, split *Split)
	}{
		{
			name: "successfully finalize valid split",
			setup: func() *Split {
				split := createTestSplit(SplitStatusDraft)
				// Add a document with pages
				pages := createTestPages(2)
				doc := createTestDocument("doc1", pages)
				require.NoError(t, split.AddDocument(doc))
				return split
			},
			wantErr: false,
			check: func(t *testing.T, split *Split) {
				assert.Equal(t, SplitStatusFinalized, split.Status)
				assert.NotNil(t, split.FinalizedAt)
				assert.True(t, split.FinalizedAt.After(split.CreatedAt))
			},
		},
		{
			name: "cannot finalize already finalized split",
			setup: func() *Split {
				split := createTestSplit(SplitStatusFinalized)
				split.FinalizedAt = &time.Time{}
				return split
			},
			wantErr:     true,
			errContains: "split already finalized",
		},
		{
			name: "cannot finalize split with unassigned pages",
			setup: func() *Split {
				split := createTestSplit(SplitStatusDraft)
				// Add unassigned pages
				pages := createTestPages(2)
				split.UnassignedPages = pages
				return split
			},
			wantErr:     true,
			errContains: "cannot finalize split with unassigned pages",
		},
		{
			name: "cannot finalize invalid split",
			setup: func() *Split {
				// Create an invalid split (missing required fields)
				split := &Split{
					ID:              "split123",
					Status:          SplitStatusDraft,
					Documents:       make([]Document, 0),
					UnassignedPages: make([]*Page, 0),
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}
				return split
			},
			wantErr:     true,
			errContains: "client ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := tt.setup()
			err := split.Finalize(time.Now())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, split)
			}
		})
	}
}
