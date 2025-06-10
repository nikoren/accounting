package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

// Login authenticates with the API and sets the token
func (c *Client) Login(ctx context.Context, username, password string) error {
	req := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	}

	resp := struct {
		Token string `json:"token"`
	}{}

	if err := c.do(ctx, "POST", "/auth/login", req, &resp); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	c.SetToken(resp.Token)
	return nil
}

// LoadSplit retrieves a split by ID
func (c *Client) LoadSplit(ctx context.Context, splitID string) (*Split, error) {
	var split Split
	if err := c.do(ctx, "GET", fmt.Sprintf("/splits/%s", splitID), nil, &split); err != nil {
		return nil, fmt.Errorf("failed to load split: %w", err)
	}
	return &split, nil
}

// UpdateDocumentMetadata updates a document's metadata
func (c *Client) UpdateDocumentMetadata(ctx context.Context, documentID string, req UpdateDocumentMetadataRequest) (*DocumentResponse, error) {
	var resp DocumentResponse
	if err := c.do(ctx, "PATCH", fmt.Sprintf("/documents/%s", documentID), req, &resp); err != nil {
		return nil, fmt.Errorf("failed to update document metadata: %w", err)
	}
	return &resp, nil
}

// MovePages moves pages between documents
func (c *Client) MovePages(ctx context.Context, req MovePagesRequest) (*MovePagesResponse, error) {
	var resp MovePagesResponse
	if err := c.do(ctx, "POST", "/pages/move", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to move pages: %w", err)
	}
	return &resp, nil
}

// CreateDocument creates a new document
func (c *Client) CreateDocument(ctx context.Context, req CreateDocumentRequest) (*DocumentResponse, error) {
	var resp DocumentResponse
	if err := c.do(ctx, "POST", "/documents", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}
	return &resp, nil
}

// DeleteDocument deletes a document
func (c *Client) DeleteDocument(ctx context.Context, documentID string) error {
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/documents/%s", documentID), nil, nil); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// FinalizeSplit finalizes a split
func (c *Client) FinalizeSplit(ctx context.Context, splitID string) error {
	if err := c.do(ctx, "POST", fmt.Sprintf("/splits/%s/finalize", splitID), nil, nil); err != nil {
		return fmt.Errorf("failed to finalize split: %w", err)
	}
	return nil
}

// DownloadDocument downloads a document
func (c *Client) DownloadDocument(ctx context.Context, documentID string) ([]byte, error) {
	var data []byte
	if err := c.do(ctx, "GET", fmt.Sprintf("/documents/%s/download", documentID), nil, &data); err != nil {
		return nil, fmt.Errorf("failed to download document: %w", err)
	}
	return data, nil
}

// GetMetrics retrieves server metrics
func (c *Client) GetMetrics(ctx context.Context) (*MetricsResponse, error) {
	var metrics MetricsResponse
	if err := c.do(ctx, "GET", "/metrics", nil, &metrics); err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	return &metrics, nil
}

// do performs an HTTP request
func (c *Client) do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("request failed: %s", errResp.Error)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
