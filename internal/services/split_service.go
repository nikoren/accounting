package services

import (
	"accounting/internal/domain"
	"accounting/internal/domain/ports"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Assert that *SplitService implements SplitServiceInterface interface
var _ SplitServiceInterface = (*SplitService)(nil)

// SplitService handles business logic for document splitting
type SplitService struct {
	uowFactory func() (ports.UnitOfWork, error)
	renderSvc  ports.RenderService
}

// NewSplitService creates a new SplitService
func NewSplitService(uowFactory func() (ports.UnitOfWork, error), renderSvc ports.RenderService) *SplitService {
	return &SplitService{
		uowFactory: uowFactory,
		renderSvc:  renderSvc,
	}
}

// convertPageToResponse converts a domain page to a page response
func convertPageToResponse(page *domain.Page) *PageResponse {
	return &PageResponse{
		ID:         page.ID,
		PageNumber: fmt.Sprintf("%d", page.PageNumber),
	}
}

// convertDocumentToResponse converts a domain document to a document response
func convertDocumentToResponse(doc *domain.Document) *DocumentResponse {
	pages := make([]*PageResponse, len(doc.Pages))
	for i, page := range doc.Pages {
		pages[i] = &PageResponse{
			ID:         page.ID,
			PageNumber: strconv.Itoa(page.PageNumber),
			URL:        page.URL,
		}
	}
	return &DocumentResponse{
		ID:               doc.ID,
		SplitID:          doc.SplitID,
		Name:             doc.Name,
		Classification:   doc.Classification,
		Filename:         doc.Filename,
		ShortDescription: doc.ShortDescription,
		StartPage:        doc.StartPage,
		EndPage:          doc.EndPage,
		Pages:            pages,
	}
}

// LoadSplit loads a split by ID
func (s *SplitService) LoadSplit(ctx context.Context, id string) (*LoadSplitResponse, error) {
	uow, err := s.uowFactory()
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	split, err := uow.SplitRepository().Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if split == nil {
		return nil, domain.ErrNotFound
	}

	// Convert domain documents to response documents
	documents := make([]*DocumentResponse, len(split.Documents))
	for i, doc := range split.Documents {
		documents[i] = convertDocumentToResponse(&doc)
	}

	// Convert unassigned pages to response pages
	unassignedPages := make([]*PageResponse, len(split.UnassignedPages))
	for i, page := range split.UnassignedPages {
		unassignedPages[i] = convertPageToResponse(page)
	}

	return &LoadSplitResponse{
		ID:              split.ID,
		ClientID:        split.ClientID,
		Status:          split.Status,
		Documents:       documents,
		UnassignedPages: unassignedPages,
	}, nil
}

// UpdateDocumentMetadata updates document metadata
func (s *SplitService) UpdateDocumentMetadata(ctx context.Context, id string, req UpdateDocumentMetadataRequest) (*DocumentResponse, error) {
	uow, err := s.uowFactory()
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	// Get split ID for the document
	splitID, err := uow.SplitRepository().GetSplitIDByDocumentID(ctx, id)
	if err != nil {
		return nil, err
	}
	if splitID == "" {
		return nil, domain.ErrNotFound
	}

	// Load split aggregate
	split, err := uow.SplitRepository().Get(ctx, splitID)
	if err != nil {
		return nil, err
	}
	if split == nil {
		return nil, domain.ErrNotFound
	}

	// Convert request to domain metadata
	metadata := domain.DocumentMetadata{
		Name:             req.Name,
		Classification:   req.Classification,
		ShortDescription: req.ShortDescription,
	}

	// Update document metadata using domain logic
	if err := split.UpdateDocumentMetadata(id, metadata); err != nil {
		return nil, err
	}

	// Save the aggregate
	if err := uow.SplitRepository().Save(ctx, split); err != nil {
		return nil, err
	}

	if err := uow.Commit(ctx); err != nil {
		return nil, err
	}

	// Find the updated document
	for _, doc := range split.Documents {
		if doc.ID == id {
			return convertDocumentToResponse(&doc), nil
		}
	}

	return nil, domain.ErrNotFound
}

// MovePages moves pages between documents
func (s *SplitService) MovePages(ctx context.Context, req MovePagesRequest) (*MovePagesResponse, error) {
	uow, err := s.uowFactory()
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	split, err := uow.SplitRepository().Get(ctx, req.SplitID)
	if err != nil {
		return nil, err
	}
	if split == nil {
		return nil, domain.ErrNotFound
	}

	// Use domain logic to move pages
	if err := split.MovePages(req.FromDocumentID, req.ToDocumentID, req.PageIDs); err != nil {
		return nil, err
	}

	// Save the aggregate
	if err := uow.SplitRepository().Save(ctx, split); err != nil {
		return nil, err
	}

	if err := uow.Commit(ctx); err != nil {
		return nil, err
	}

	// Find the updated documents
	var fromDoc, toDoc *domain.Document
	for _, doc := range split.Documents {
		if doc.ID == req.FromDocumentID {
			fromDoc = &doc
		}
		if doc.ID == req.ToDocumentID {
			toDoc = &doc
		}
	}

	if fromDoc == nil || toDoc == nil {
		return nil, domain.ErrNotFound
	}

	return &MovePagesResponse{
		FromDocument: convertDocumentToResponse(fromDoc),
		ToDocument:   convertDocumentToResponse(toDoc),
	}, nil
}

// CreateDocument creates a new document
func (s *SplitService) CreateDocument(ctx context.Context, req CreateDocumentRequest) (*DocumentResponse, error) {
	uow, err := s.uowFactory()
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	split, err := uow.SplitRepository().Get(ctx, req.SplitID)
	if err != nil {
		return nil, err
	}
	if split == nil {
		return nil, domain.ErrNotFound
	}

	// Generate a new UUID for the document ID
	docID := uuid.NewString()

	// Collect pages by ID from split's unassigned pages
	var pages []*domain.Page
	remainingUnassigned := make([]*domain.Page, 0, len(split.UnassignedPages))
	pageIDSet := make(map[string]struct{}, len(req.PageIDs))
	for _, pid := range req.PageIDs {
		pageIDSet[pid] = struct{}{}
	}
	for _, page := range split.UnassignedPages {
		if _, ok := pageIDSet[page.ID]; ok {
			pages = append(pages, page)
		} else {
			remainingUnassigned = append(remainingUnassigned, page)
		}
	}
	if len(pages) == 0 {
		return nil, domain.NewValidationError("no valid pages specified for new document", nil)
	}

	// Remove assigned pages from unassigned list
	split.UnassignedPages = remainingUnassigned

	// Create document using domain logic
	doc := &domain.Document{
		ID:               docID,
		SplitID:          req.SplitID,
		Name:             req.Name,
		Classification:   req.Classification,
		Filename:         req.Filename,
		ShortDescription: req.ShortDescription,
		Pages:            pages,
	}

	if err := split.AddDocument(doc); err != nil {
		return nil, err
	}

	// Save the aggregate
	if err := uow.SplitRepository().Save(ctx, split); err != nil {
		return nil, err
	}

	if err := uow.Commit(ctx); err != nil {
		return nil, err
	}

	return convertDocumentToResponse(doc), nil
}

// DeleteDocument deletes a document
func (s *SplitService) DeleteDocument(ctx context.Context, id string) error {
	uow, factErr := s.uowFactory()
	if factErr != nil {
		return factErr
	}
	defer uow.Rollback(ctx)

	// Get split ID for the document
	splitID, getByErr := uow.SplitRepository().GetSplitIDByDocumentID(ctx, id)
	if getByErr != nil {
		return getByErr
	}
	if splitID == "" {
		return domain.ErrNotFound
	}

	// Load split aggregate
	split, getErr := uow.SplitRepository().Get(ctx, splitID)
	if getErr != nil {
		return getErr
	}
	if split == nil {
		return domain.ErrNotFound
	}

	// Delete document using domain logic
	if remErr := split.RemoveDocument(id); remErr != nil {
		return remErr
	}

	// Save the aggregate
	if saveErr := uow.SplitRepository().Save(ctx, split); saveErr != nil {
		return saveErr
	}

	return uow.Commit(ctx)
}

// FinalizeSplit finalizes a split
func (s *SplitService) FinalizeSplit(ctx context.Context, id string) error {
	uow, err := s.uowFactory()
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)

	split, err := uow.SplitRepository().Get(ctx, id)
	if err != nil {
		return err
	}
	if split == nil {
		return domain.ErrNotFound
	}

	// Finalize split using domain logic
	if err := split.Finalize(time.Now()); err != nil {
		return err
	}

	// Save the aggregate
	if err := uow.SplitRepository().Save(ctx, split); err != nil {
		return err
	}

	return uow.Commit(ctx)
}

// DownloadDocument downloads a document
func (s *SplitService) DownloadDocument(ctx context.Context, id string) (*DownloadDocumentResponse, error) {
	uow, err := s.uowFactory()
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	// Get split ID for the document
	splitID, err := uow.SplitRepository().GetSplitIDByDocumentID(ctx, id)
	if err != nil {
		return nil, err
	}
	if splitID == "" {
		return nil, domain.ErrNotFound
	}

	// Load split aggregate
	split, err := uow.SplitRepository().Get(ctx, splitID)
	if err != nil {
		return nil, err
	}
	if split == nil {
		return nil, domain.ErrNotFound
	}

	// Find the document
	var doc *domain.Document
	for _, d := range split.Documents {
		if d.ID == id {
			doc = &d
			break
		}
	}

	if doc == nil {
		return nil, domain.ErrNotFound
	}

	// Download document using render service
	resp, err := s.renderSvc.RenderDocument(ctx, ports.RenderDocumentRequest{
		Document: doc,
	})
	if err != nil {
		return nil, err
	}

	return &DownloadDocumentResponse{
		Data:        resp.Data,
		Filename:    doc.Filename,
		ContentType: "application/pdf",
	}, nil
}
