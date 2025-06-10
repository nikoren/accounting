package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Set up test environment variables
	os.Setenv("APP_USERS", "admin:admin123,user:user123")
	defer os.Unsetenv("APP_USERS")

	// Load configuration
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify default values
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 10, cfg.ShutdownTimeout)
	assert.Equal(t, "accounting.db", cfg.DatabasePath)
	assert.Equal(t, 100, cfg.RequestsPerSecond)
	assert.Equal(t, 200, cfg.BurstSize)

	// Verify users
	require.Len(t, cfg.Users, 2)
	assert.Equal(t, "admin", cfg.Users[0].Username)
	assert.Equal(t, "admin123", cfg.Users[0].Password)
	assert.Equal(t, "user", cfg.Users[1].Username)
	assert.Equal(t, "user123", cfg.Users[1].Password)

	// Test GetUsersMap
	usersMap := cfg.GetUsersMap()
	assert.Len(t, usersMap, 2)
	assert.Equal(t, "admin123", usersMap["admin"].Password)
	assert.Equal(t, "user123", usersMap["user"].Password)
}

func TestLoadConfigWithCustomValues(t *testing.T) {
	// Set up test environment variables
	os.Setenv("APP_PORT", "9090")
	os.Setenv("APP_HOST", "0.0.0.0")
	os.Setenv("APP_DB_PATH", "test.db")
	os.Setenv("APP_USERS", "test:test123")
	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_HOST")
		os.Unsetenv("APP_DB_PATH")
		os.Unsetenv("APP_USERS")
	}()

	// Load configuration
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify custom values
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, "test.db", cfg.DatabasePath)

	// Verify users
	require.Len(t, cfg.Users, 1)
	assert.Equal(t, "test", cfg.Users[0].Username)
	assert.Equal(t, "test123", cfg.Users[0].Password)
}

func TestLoadConfigWithInvalidUserFormat(t *testing.T) {
	// Set up test environment variables with invalid user format
	os.Setenv("APP_USERS", "invalidformat")
	defer os.Unsetenv("APP_USERS")

	// Load configuration
	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user format")
}

func TestLoadConfigWithoutRequiredUsers(t *testing.T) {
	// Ensure APP_USERS is not set
	os.Unsetenv("APP_USERS")

	// Load configuration
	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required key USERS missing value")
}
