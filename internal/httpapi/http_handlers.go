package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"accounting/internal/domain"
	"accounting/internal/services"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// TokenVerifier is a local interface for verifying JWT tokens
type TokenVerifier interface {
	VerifyToken(token string) (any, error)
}

// SplitHandler handles HTTP requests for split operations
type SplitHandler struct {
	splitSvc      services.SplitServiceInterface
	tokenVerifier TokenVerifier
}

// NewSplitHandler creates a new SplitHandler
func NewSplitHandler(splitSvc services.SplitServiceInterface, tokenVerifier TokenVerifier) *SplitHandler {
	return &SplitHandler{
		splitSvc:      splitSvc,
		tokenVerifier: tokenVerifier,
	}
}

// Helper to extract the ID from the path (second segment)
func getIDFromPath(r *http.Request) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// Helper to write JSON error without trailing newline
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	b, _ := json.Marshal(map[string]string{"error": msg})
	w.Write(b)
}

// LoadSplitHandler handles GET requests to load a split
func (h *SplitHandler) LoadSplitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Verify the token
	_, err := h.tokenVerifier.VerifyToken(parts[1])
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	id := getIDFromPath(r)
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "split ID is required")
		return
	}

	resp, err := h.splitSvc.LoadSplit(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateDocumentMetadataHandler handles PATCH requests to update document metadata
func (h *SplitHandler) UpdateDocumentMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Verify the token
	_, err := h.tokenVerifier.VerifyToken(parts[1])
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	id := getIDFromPath(r)
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "document ID is required")
		return
	}

	var req services.UpdateDocumentMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.splitSvc.UpdateDocumentMetadata(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// MovePagesHandler handles POST requests to move pages between documents
func (h *SplitHandler) MovePagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Verify the token
	_, err := h.tokenVerifier.VerifyToken(parts[1])
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var req services.MovePagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.PageIDs) == 0 {
		writeJSONError(w, http.StatusBadRequest, "page IDs are required")
		return
	}

	resp, err := h.splitSvc.MovePages(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateDocumentHandler handles POST requests to create a new document
func (h *SplitHandler) CreateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Verify the token
	_, err := h.tokenVerifier.VerifyToken(parts[1])
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var req services.CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.PageIDs) == 0 {
		writeJSONError(w, http.StatusBadRequest, "page IDs are required")
		return
	}

	resp, err := h.splitSvc.CreateDocument(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// DeleteDocumentHandler handles DELETE requests to remove a document
func (h *SplitHandler) DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Verify the token
	_, err := h.tokenVerifier.VerifyToken(parts[1])
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	id := getIDFromPath(r)
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "document ID is required")
		return
	}

	err = h.splitSvc.DeleteDocument(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// FinalizeSplitHandler handles POST requests to finalize a split
func (h *SplitHandler) FinalizeSplitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Verify the token
	_, err := h.tokenVerifier.VerifyToken(parts[1])
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	id := getIDFromPath(r)
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "split ID is required")
		return
	}

	err = h.splitSvc.FinalizeSplit(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DownloadDocumentHandler handles GET requests to download a document
func (h *SplitHandler) DownloadDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Verify the token
	_, err := h.tokenVerifier.VerifyToken(parts[1])
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	id := getIDFromPath(r)
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "document ID is required")
		return
	}

	resp, err := h.splitSvc.DownloadDocument(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}
