package services

import (
	"accounting/internal/domain/ports"
	"context"
	"fmt"
)

// RenderService handles document rendering operations
type RenderService struct{}

// NewRenderService creates a new instance of RenderService
func NewRenderService() ports.RenderService {
	return &RenderService{}
}

// RenderDocument implements the ports.RenderService interface
func (s *RenderService) RenderDocument(ctx context.Context, req ports.RenderDocumentRequest) (*ports.RenderDocumentResponse, error) {
	// Create a simple PDF with the document name
	// For now, we'll just return a placeholder PDF
	// In a real implementation, you would use a PDF generation library
	// like github.com/jung-kurt/gofpdf or github.com/unidoc/unipdf

	// Create a simple PDF with the document name as text
	pdfContent := fmt.Sprintf("%%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /Resources << /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >> /Contents 4 0 R >>\nendobj\n4 0 obj\n<< /Length 44 >>\nstream\nBT\n/F1 24 Tf\n100 700 Td\n(%s) Tj\nET\nendstream\nendobj\nxref\n0 5\n0000000000 65535 f\n0000000009 00000 n\n0000000056 00000 n\n0000000111 00000 n\n0000000256 00000 n\ntrailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n364\n%%EOF", req.Document.Filename)

	return &ports.RenderDocumentResponse{
		Filename:    req.Document.Filename,
		ContentType: "application/pdf",
		Data:        []byte(pdfContent),
	}, nil
}
