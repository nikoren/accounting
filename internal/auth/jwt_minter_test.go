package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTMinter(t *testing.T) {
	// Create a new minter
	users := map[string]User{
		"admin": {Username: "admin", Password: "admin123"},
		"user":  {Username: "user", Password: "user123"},
	}
	minter, err := NewJWTMinter(users)
	require.NoError(t, err)
	require.NotNil(t, minter)

	// Create a test server
	mux := http.NewServeMux()
	minter.Mount(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Run("invalid method", func(t *testing.T) {
		// Test GET request (should fail)
		resp, err := http.Get(server.URL + "/auth/login")
		require.NoError(t, err)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		// Test invalid credentials
		req := LoginRequest{
			Username: "invalid",
			Password: "invalid",
		}
		body, err := json.Marshal(req)
		require.NoError(t, err)

		resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("valid credentials", func(t *testing.T) {
		// Test valid credentials
		req := LoginRequest{
			Username: "admin",
			Password: "admin123",
		}
		body, err := json.Marshal(req)
		require.NoError(t, err)

		resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var loginResp LoginResponse
		err = json.NewDecoder(resp.Body).Decode(&loginResp)
		require.NoError(t, err)
		assert.NotEmpty(t, loginResp.Token)

		// Verify token
		token, err := jwt.ParseString(loginResp.Token, jwt.WithVerify(false))
		require.NoError(t, err)

		// Check token claims
		issuer, ok := token.Issuer()
		require.True(t, ok)
		assert.Equal(t, "accounting-service", issuer)

		subject, ok := token.Subject()
		require.True(t, ok)
		assert.Equal(t, "admin", subject)

		exp, ok := token.Expiration()
		require.True(t, ok)
		assert.True(t, exp.After(time.Now()))
		assert.True(t, exp.Before(time.Now().Add(25*time.Hour))) // 24h + 1h buffer
	})

	t.Run("different user", func(t *testing.T) {
		// Test different user
		req := LoginRequest{
			Username: "user",
			Password: "user123",
		}
		body, err := json.Marshal(req)
		require.NoError(t, err)

		resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var loginResp LoginResponse
		err = json.NewDecoder(resp.Body).Decode(&loginResp)
		require.NoError(t, err)
		assert.NotEmpty(t, loginResp.Token)

		// Verify token
		token, err := jwt.ParseString(loginResp.Token, jwt.WithVerify(false))
		require.NoError(t, err)
		subject, ok := token.Subject()
		require.True(t, ok)
		assert.Equal(t, "user", subject)
	})
}
