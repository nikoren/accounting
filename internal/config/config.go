package config

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port            int    `envconfig:"PORT" default:"8080"`
	Host            string `envconfig:"HOST" default:"localhost"`
	ShutdownTimeout int    `envconfig:"SHUTDOWN_TIMEOUT" default:"10"` // in seconds

	// Database configuration
	DatabasePath string `envconfig:"DB_PATH" default:"accounting.db"`

	// Rate limiting
	RequestsPerSecond int `envconfig:"REQUESTS_PER_SECOND" default:"100"`
	BurstSize         int `envconfig:"BURST_SIZE" default:"200"`

	// Users configuration
	Users []User `envconfig:"USERS" required:"true"`
}

// User represents a user in the system
type User struct {
	Username string
	Password string
}

// Decode implements envconfig.Decoder for User
func (u *User) Decode(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid user format, expected username:password, got: %s", value)
	}
	u.Username = parts[0]
	u.Password = parts[1]
	return nil
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("APP", &cfg)
	if err != nil {
		return nil, fmt.Errorf("env config error: %w", err)
	}
	return &cfg, nil
}

// GetUsersMap converts the users slice to a map for easier lookup
func (c *Config) GetUsersMap() map[string]User {
	users := make(map[string]User)
	for _, u := range c.Users {
		users[u.Username] = u
	}
	return users
}
