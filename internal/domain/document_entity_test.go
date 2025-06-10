package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocument_UpdateMetadata(t *testing.T) {
	pages := []*Page{}
	page, err := NewPage("split1", "page_1.png")
	assert.NoError(t, err)
	pages = append(pages, page)
	doc, err := NewDocument(
		"doc1",
		"split1",
		"Original Name",
		"W-2",
		"original.pdf",
		"Original description",
		pages,
	)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		metadata    DocumentMetadata
		expectName  string
		expectClass string
		expectDesc  string
	}{
		{
			name: "update all fields",
			metadata: DocumentMetadata{
				Name:             ptrString("Updated Name"),
				Classification:   ptrString("1099"),
				ShortDescription: ptrString("Updated description"),
			},
			expectName:  "Updated Name",
			expectClass: "1099",
			expectDesc:  "Updated description",
		},
		{
			name: "update some fields",
			metadata: DocumentMetadata{
				Name:             ptrString("Partial Name"),
				Classification:   nil,
				ShortDescription: ptrString("Partial description"),
			},
			expectName:  "Partial Name",
			expectClass: doc.Classification,
			expectDesc:  "Partial description",
		},
		{
			name: "empty values (no update)",
			metadata: DocumentMetadata{
				Name:             nil,
				Classification:   nil,
				ShortDescription: nil,
			},
			expectName:  doc.Name,
			expectClass: doc.Classification,
			expectDesc:  doc.ShortDescription,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := *doc // copy original doc
			err := d.UpdateMetadata(tt.metadata)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectName, d.Name)
			assert.Equal(t, tt.expectClass, d.Classification)
			assert.Equal(t, tt.expectDesc, d.ShortDescription)
		})
	}
}

// ptrString is a helper to get a pointer to a string literal
func ptrString(s string) *string { return &s }
