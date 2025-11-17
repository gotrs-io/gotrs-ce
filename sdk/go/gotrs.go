// Package gotrs provides a Go SDK for the GOTRS ticketing system API.
//
// This SDK provides a complete interface to the GOTRS API, including:
//   - Ticket management (create, read, update, delete, search)
//   - User management
//   - Queue management
//   - Dashboard statistics
//   - LDAP integration
//   - Webhook management
//   - Internal notes
//   - Real-time events
//
// Basic usage:
//
//	client := gotrs.NewClientWithAPIKey("https://your-gotrs-instance.com", "your-api-key")
//	tickets, err := client.Tickets.List(ctx, nil)
//
// Authentication methods:
//
//	// API Key (recommended for server-to-server)
//	client := gotrs.NewClientWithAPIKey(baseURL, apiKey)
//
//	// JWT Token
//	client := gotrs.NewClientWithJWT(baseURL, token, refreshToken, expiresAt)
//
//	// Custom authentication
//	auth := gotrs.NewOAuth2Auth(accessToken, refreshToken, "Bearer", expiresAt)
//	client := gotrs.NewClient(&gotrs.Config{
//		BaseURL: baseURL,
//		Auth:    auth,
//	})
package gotrs

import (
	"time"

	"github.com/gotrs-io/gotrs-ce/sdk/go/auth"
	"github.com/gotrs-io/gotrs-ce/sdk/go/client"
)

// Client represents the GOTRS API client
type Client = client.Client

// Config represents client configuration
type Config = client.Config

// NewClient creates a new GOTRS API client with custom configuration
func NewClient(config *Config) *Client {
	return client.NewClient(config)
}

// NewClientWithAPIKey creates a new client with API key authentication
func NewClientWithAPIKey(baseURL, apiKey string) *Client {
	return client.NewClientWithAPIKey(baseURL, apiKey)
}

// NewClientWithJWT creates a new client with JWT authentication
func NewClientWithJWT(baseURL, token, refreshToken string, expiresAt time.Time) *Client {
	return client.NewClientWithJWT(baseURL, token, refreshToken, expiresAt)
}

// Authentication helpers
var (
	// NewAPIKeyAuth creates a new API key authenticator
	NewAPIKeyAuth = auth.NewAPIKeyAuth

	// NewJWTAuth creates a new JWT authenticator
	NewJWTAuth = auth.NewJWTAuth

	// NewOAuth2Auth creates a new OAuth2 authenticator
	NewOAuth2Auth = auth.NewOAuth2Auth

	// NewNoAuth creates a new no-auth authenticator
	NewNoAuth = auth.NewNoAuth
)

// Version information
const (
	// Version is the current SDK version
	Version = "1.0.0"

	// UserAgent is the default user agent string
	UserAgent = "gotrs-go-sdk/" + Version
)
