package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"accounting/internal/domain"
	"accounting/internal/services"

	"github.com/stretchr/testify/assert"
)

// MockSplitService is a mock implementation of SplitServiceInterface
type MockSplitService struct {
	loadSplitFunc              func(ctx context.Context, id string) (*services.LoadSplitResponse, error)
	updateDocumentMetadataFunc func(ctx context.Context, documentID string, req services.UpdateDocumentMetadataRequest) (*services.DocumentResponse, error)
	movePagesFunc              func(ctx context.Context, req services.MovePagesRequest) (*services.MovePagesResponse, error)
	createDocumentFunc         func(ctx context.Context, req services.CreateDocumentRequest) (*services.DocumentResponse, error)
	deleteDocumentFunc         func(ctx context.Context, documentID string) error
	finalizeSplitFunc          func(ctx context.Context, splitID string) error
	downloadDocumentFunc       func(ctx context.Context, documentID string) (*services.DownloadDocumentResponse, error)
}

func (m *MockSplitService) LoadSplit(ctx context.Context, id string) (*services.LoadSplitResponse, error) {
	return m.loadSplitFunc(ctx, id)
}

func (m *MockSplitService) UpdateDocumentMetadata(ctx context.Context, documentID string, req services.UpdateDocumentMetadataRequest) (*services.DocumentResponse, error) {
	return m.updateDocumentMetadataFunc(ctx, documentID, req)
}

func (m *MockSplitService) MovePages(ctx context.Context, req services.MovePagesRequest) (*services.MovePagesResponse, error) {
	return m.movePagesFunc(ctx, req)
}

func (m *MockSplitService) CreateDocument(ctx context.Context, req services.CreateDocumentRequest) (*services.DocumentResponse, error) {
	return m.createDocumentFunc(ctx, req)
}

func (m *MockSplitService) DeleteDocument(ctx context.Context, documentID string) error {
	return m.deleteDocumentFunc(ctx, documentID)
}

func (m *MockSplitService) FinalizeSplit(ctx context.Context, splitID string) error {
	return m.finalizeSplitFunc(ctx, splitID)
}

func (m *MockSplitService) DownloadDocument(ctx context.Context, documentID string) (*services.DownloadDocumentResponse, error) {
	return m.downloadDocumentFunc(ctx, documentID)
}

// mockVerifier is a mock implementation of TokenVerifier
type mockVerifier struct{}

func (m *mockVerifier) VerifyToken(token string) (any, error) {
	// Mock implementation for testing
	return nil, nil
}

func TestLoadSplitHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		mockResponse   *services.LoadSplitResponse
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "success",
			method:         http.MethodGet,
			path:           "/splits/123/load",
			mockResponse:   &services.LoadSplitResponse{ID: "123"},
			expectedStatus: http.StatusOK,
			expectedBody:   &services.LoadSplitResponse{ID: "123"},
		},
		{
			name:           "not found",
			method:         http.MethodGet,
			path:           "/splits/nonexistent/load",
			mockError:      domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "not found"},
		},
		{
			name:           "empty id",
			method:         http.MethodGet,
			path:           "/splits//load",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "split ID is required"},
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			path:           "/splits/123/load",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   map[string]interface{}{"error": "method not allowed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSplitService{
				loadSplitFunc: func(ctx context.Context, id string) (*services.LoadSplitResponse, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}
			handler := NewSplitHandler(mockService, &mockVerifier{})
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			handler.LoadSplitHandler(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response services.LoadSplitResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, &response)
			} else {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestUpdateDocumentMetadataHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		mockResponse   *services.DocumentResponse
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "success",
			method:         http.MethodPatch,
			path:           "/documents/123/metadata",
			body:           map[string]interface{}{"name": "Updated Document"},
			mockResponse:   &services.DocumentResponse{ID: "123", Name: "Updated Document"},
			expectedStatus: http.StatusOK,
			expectedBody:   &services.DocumentResponse{ID: "123", Name: "Updated Document"},
		},
		{
			name:           "not found",
			method:         http.MethodPatch,
			path:           "/documents/non-existent/metadata",
			body:           map[string]interface{}{"name": "Updated Document"},
			mockError:      domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "not found"},
		},
		{
			name:           "empty id",
			method:         http.MethodPatch,
			path:           "/documents//metadata",
			body:           map[string]interface{}{"name": "Updated Document"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "document ID is required"},
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			path:           "/documents/123/metadata",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   map[string]interface{}{"error": "method not allowed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSplitService{
				updateDocumentMetadataFunc: func(ctx context.Context, documentID string, req services.UpdateDocumentMetadataRequest) (*services.DocumentResponse, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}
			handler := NewSplitHandler(mockService, &mockVerifier{})
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			handler.UpdateDocumentMetadataHandler(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response services.DocumentResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, &response)
			} else {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestMovePagesHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		mockResponse   *services.MovePagesResponse
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "success",
			method: http.MethodPost,
			path:   "/documents/123/pages/move",
			body: services.MovePagesRequest{
				PageIDs:        []string{"1", "2"},
				FromDocumentID: "123",
				ToDocumentID:   "456",
			},
			mockResponse: &services.MovePagesResponse{
				FromDocument: &services.DocumentResponse{ID: "123"},
				ToDocument:   &services.DocumentResponse{ID: "456"},
			},
			expectedStatus: http.StatusOK,
			expectedBody: &services.MovePagesResponse{
				FromDocument: &services.DocumentResponse{ID: "123"},
				ToDocument:   &services.DocumentResponse{ID: "456"},
			},
		},
		{
			name:   "not found",
			method: http.MethodPost,
			path:   "/documents/123/pages/move",
			body: services.MovePagesRequest{
				PageIDs:        []string{"1", "2"},
				FromDocumentID: "123",
				ToDocumentID:   "456",
			},
			mockError:      domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "not found"},
		},
		{
			name:   "empty page ids",
			method: http.MethodPost,
			path:   "/documents/123/pages/move",
			body: services.MovePagesRequest{
				PageIDs:        []string{},
				FromDocumentID: "123",
				ToDocumentID:   "456",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "page IDs are required"},
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			path:           "/documents/123/pages/move",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   map[string]interface{}{"error": "method not allowed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSplitService{
				movePagesFunc: func(ctx context.Context, req services.MovePagesRequest) (*services.MovePagesResponse, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}
			handler := NewSplitHandler(mockService, &mockVerifier{})
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			handler.MovePagesHandler(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response services.MovePagesResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, &response)
			} else {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestCreateDocumentHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		mockResponse   *services.DocumentResponse
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "success",
			method: http.MethodPost,
			path:   "/splits/123/documents",
			body: services.CreateDocumentRequest{
				Name:    "New Document",
				PageIDs: []string{"1", "2"},
			},
			mockResponse: &services.DocumentResponse{
				ID:   "123",
				Name: "New Document",
			},
			expectedStatus: http.StatusCreated,
			expectedBody: &services.DocumentResponse{
				ID:   "123",
				Name: "New Document",
			},
		},
		{
			name:   "empty page ids",
			method: http.MethodPost,
			path:   "/splits/123/documents",
			body: services.CreateDocumentRequest{
				Name:    "New Document",
				PageIDs: []string{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "page IDs are required"},
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			path:           "/splits/123/documents",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   map[string]interface{}{"error": "method not allowed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSplitService{
				createDocumentFunc: func(ctx context.Context, req services.CreateDocumentRequest) (*services.DocumentResponse, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}
			handler := NewSplitHandler(mockService, &mockVerifier{})
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			handler.CreateDocumentHandler(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				var response services.DocumentResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, &response)
			} else {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestDeleteDocumentHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "success",
			method:         http.MethodDelete,
			path:           "/documents/123",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "not found",
			method:         http.MethodDelete,
			path:           "/documents/non-existent",
			mockError:      domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "not found"},
		},
		{
			name:           "empty id",
			method:         http.MethodDelete,
			path:           "/documents/",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "document ID is required"},
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			path:           "/documents/123",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   map[string]interface{}{"error": "method not allowed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSplitService{
				deleteDocumentFunc: func(ctx context.Context, documentID string) error {
					return tt.mockError
				},
			}
			handler := NewSplitHandler(mockService, &mockVerifier{})
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			handler.DeleteDocumentHandler(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus != http.StatusNoContent {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestFinalizeSplitHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			path:           "/splits/123/finalize",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "not found",
			method:         http.MethodPost,
			path:           "/splits/non-existent/finalize",
			mockError:      domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "not found"},
		},
		{
			name:           "empty id",
			method:         http.MethodPost,
			path:           "/splits//finalize",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "split ID is required"},
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			path:           "/splits/123/finalize",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   map[string]interface{}{"error": "method not allowed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSplitService{
				finalizeSplitFunc: func(ctx context.Context, splitID string) error {
					return tt.mockError
				},
			}
			handler := NewSplitHandler(mockService, &mockVerifier{})
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			handler.FinalizeSplitHandler(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus != http.StatusNoContent {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestDownloadDocumentHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		mockResponse   *services.DownloadDocumentResponse
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "success",
			method: http.MethodGet,
			path:   "/documents/123",
			mockResponse: &services.DownloadDocumentResponse{
				Data: []byte("PDF content"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []byte("PDF content"),
		},
		{
			name:           "not found",
			method:         http.MethodGet,
			path:           "/documents/non-existent",
			mockError:      domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "not found"},
		},
		{
			name:           "empty id",
			method:         http.MethodGet,
			path:           "/documents/",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "document ID is required"},
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			path:           "/documents/123",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   map[string]interface{}{"error": "method not allowed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockSplitService{
				downloadDocumentFunc: func(ctx context.Context, documentID string) (*services.DownloadDocumentResponse, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}
			handler := NewSplitHandler(mockService, &mockVerifier{})
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			handler.DownloadDocumentHandler(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody, w.Body.Bytes())
			} else {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

// compareMaps compares two maps recursively
func compareMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, exists := b[k]
		if !exists {
			return false
		}
		switch av := v.(type) {
		case map[string]interface{}:
			bvMap, ok := bv.(map[string]interface{})
			if !ok {
				return false
			}
			if !compareMaps(av, bvMap) {
				return false
			}
		case []interface{}:
			bvSlice, ok := bv.([]interface{})
			if !ok {
				return false
			}
			if !compareSlices(av, bvSlice) {
				return false
			}
		default:
			if v != bv {
				return false
			}
		}
	}
	return true
}

// compareSlices compares two slices recursively
func compareSlices(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		switch av := a[i].(type) {
		case map[string]interface{}:
			bv, ok := b[i].(map[string]interface{})
			if !ok {
				return false
			}
			if !compareMaps(av, bv) {
				return false
			}
		case []interface{}:
			bv, ok := b[i].([]interface{})
			if !ok {
				return false
			}
			if !compareSlices(av, bv) {
				return false
			}
		default:
			if a[i] != b[i] {
				return false
			}
		}
	}
	return true
}

