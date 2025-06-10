package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/login":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
			if req.Username != "test" || req.Password != "test" {
				http.Error(w, "invalid credentials", http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})

		case "/splits/test":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if r.Header.Get("Authorization") != "Bearer test-token" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(Split{
				SplitID:  "test",
				ClientID: "client1",
				Documents: []Document{
					{
						ID:       "doc1",
						Name:     "Test Document",
						PageURLs: []string{"page1.png"},
					},
				},
			})

		case "/metrics":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			json.NewEncoder(w).Encode(MetricsResponse{
				RequestsTotal:     100,
				ActiveConnections: 5,
			})

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL)

	t.Run("login", func(t *testing.T) {
		err := client.Login(context.Background(), "test", "test")
		require.NoError(t, err)
		assert.Equal(t, "test-token", client.token)
	})

	t.Run("login invalid credentials", func(t *testing.T) {
		err := client.Login(context.Background(), "invalid", "invalid")
		assert.Error(t, err)
	})

	t.Run("load split", func(t *testing.T) {
		// Login first
		err := client.Login(context.Background(), "test", "test")
		require.NoError(t, err)

		// Load split
		split, err := client.LoadSplit(context.Background(), "test")
		require.NoError(t, err)
		assert.Equal(t, "test", split.SplitID)
		assert.Equal(t, "client1", split.ClientID)
		assert.Len(t, split.Documents, 1)
		assert.Equal(t, "doc1", split.Documents[0].ID)
	})

	t.Run("load split unauthorized", func(t *testing.T) {
		client.token = "" // Clear token
		_, err := client.LoadSplit(context.Background(), "test")
		assert.Error(t, err)
	})

	t.Run("get metrics", func(t *testing.T) {
		metrics, err := client.GetMetrics(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(100), metrics.RequestsTotal)
		assert.Equal(t, int32(5), metrics.ActiveConnections)
	})
}
