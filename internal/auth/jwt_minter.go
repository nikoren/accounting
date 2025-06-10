package auth

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// User represents an authenticated user
type User struct {
	Username string
	Password string
}

// JWTMinter handles JWT token minting
type JWTMinter struct {
	// In a real implementation, this would be a database
	users map[string]User
	// Secret key for signing JWT tokens
	secretKey []byte
}

// NewJWTMinter creates a new JWT minter
func NewJWTMinter(users map[string]User) (*JWTMinter, error) {
	// Generate a random secret key
	secretKey := make([]byte, 32)
	if _, err := rand.Read(secretKey); err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}

	return &JWTMinter{
		users:     users,
		secretKey: secretKey,
	}, nil
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string `json:"token"`
}

// LoginHandler handles login requests and mints JWT tokens
func (m *JWTMinter) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate credentials
	user, exists := m.users[req.Username]
	if !exists || user.Password != req.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create token
	token, err := jwt.NewBuilder().
		Issuer("accounting-service").
		Subject(req.Username).
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(24 * time.Hour)).
		Build()
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	// Sign token
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.HS256(), m.secretKey))
	if err != nil {
		http.Error(w, "Failed to sign token", http.StatusInternalServerError)
		return
	}

	// Return token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: string(signed),
	})
}

// Mount mounts the JWT minter to the given mux
func (m *JWTMinter) Mount(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/login", m.LoginHandler)
}

// Verifier defines the interface for token verification
type Verifier interface {
	VerifyToken(token string) (jwt.Token, error)
}

// VerifyToken verifies a JWT token and returns the claims if valid
func (m *JWTMinter) VerifyToken(token string) (jwt.Token, error) {
	parsed, err := jwt.Parse([]byte(token), jwt.WithKey(jwa.HS256(), m.secretKey))
	if err != nil {
		return nil, err
	}
	// Optionally check claims (e.g., expiration)
	exp, ok := parsed.Expiration()
	if !ok || exp.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	return parsed, nil
}
