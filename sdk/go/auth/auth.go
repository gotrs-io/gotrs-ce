package auth

import (
	"fmt"
	"time"
)

// AuthMethod represents different authentication methods
type AuthMethod int

const (
	// AuthMethodAPIKey uses API key authentication
	AuthMethodAPIKey AuthMethod = iota
	// AuthMethodJWT uses JWT token authentication
	AuthMethodJWT
	// AuthMethodOAuth2 uses OAuth2 token authentication
	AuthMethodOAuth2
)

// Authenticator interface for different auth methods
type Authenticator interface {
	// GetAuthHeader returns the authorization header value
	GetAuthHeader() string
	// IsExpired checks if the authentication is expired
	IsExpired() bool
	// Refresh refreshes the authentication if possible
	Refresh() error
	// Type returns the authentication method type
	Type() AuthMethod
}

// APIKeyAuth implements API key authentication
type APIKeyAuth struct {
	APIKey string
	Header string // Default: "X-API-Key"
}

// NewAPIKeyAuth creates a new API key authenticator
func NewAPIKeyAuth(apiKey string) *APIKeyAuth {
	return &APIKeyAuth{
		APIKey: apiKey,
		Header: "X-API-Key",
	}
}

// GetAuthHeader returns the API key header
func (a *APIKeyAuth) GetAuthHeader() string {
	return a.APIKey
}

// IsExpired always returns false for API keys
func (a *APIKeyAuth) IsExpired() bool {
	return false
}

// Refresh is not applicable for API keys
func (a *APIKeyAuth) Refresh() error {
	return nil
}

// Type returns the authentication method type
func (a *APIKeyAuth) Type() AuthMethod {
	return AuthMethodAPIKey
}

// JWTAuth implements JWT token authentication
type JWTAuth struct {
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
	RefreshFunc  func(refreshToken string) (string, string, time.Time, error)
}

// NewJWTAuth creates a new JWT authenticator
func NewJWTAuth(token, refreshToken string, expiresAt time.Time) *JWTAuth {
	return &JWTAuth{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
}

// GetAuthHeader returns the JWT token in Bearer format
func (a *JWTAuth) GetAuthHeader() string {
	return fmt.Sprintf("Bearer %s", a.Token)
}

// IsExpired checks if the JWT token is expired
func (a *JWTAuth) IsExpired() bool {
	return time.Now().After(a.ExpiresAt.Add(-1 * time.Minute)) // 1 minute buffer
}

// Refresh refreshes the JWT token using the refresh token
func (a *JWTAuth) Refresh() error {
	if a.RefreshFunc == nil {
		return fmt.Errorf("no refresh function configured")
	}

	token, refreshToken, expiresAt, err := a.RefreshFunc(a.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	a.Token = token
	a.RefreshToken = refreshToken
	a.ExpiresAt = expiresAt
	return nil
}

// Type returns the authentication method type
func (a *JWTAuth) Type() AuthMethod {
	return AuthMethodJWT
}

// OAuth2Auth implements OAuth2 token authentication
type OAuth2Auth struct {
	AccessToken  string
	RefreshToken string
	TokenType    string // Usually "Bearer"
	ExpiresAt    time.Time
	RefreshFunc  func(refreshToken string) (string, string, time.Time, error)
}

// NewOAuth2Auth creates a new OAuth2 authenticator
func NewOAuth2Auth(accessToken, refreshToken, tokenType string, expiresAt time.Time) *OAuth2Auth {
	if tokenType == "" {
		tokenType = "Bearer"
	}
	return &OAuth2Auth{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresAt:    expiresAt,
	}
}

// GetAuthHeader returns the OAuth2 token
func (a *OAuth2Auth) GetAuthHeader() string {
	return fmt.Sprintf("%s %s", a.TokenType, a.AccessToken)
}

// IsExpired checks if the OAuth2 token is expired
func (a *OAuth2Auth) IsExpired() bool {
	return time.Now().After(a.ExpiresAt.Add(-1 * time.Minute)) // 1 minute buffer
}

// Refresh refreshes the OAuth2 token using the refresh token
func (a *OAuth2Auth) Refresh() error {
	if a.RefreshFunc == nil {
		return fmt.Errorf("no refresh function configured")
	}

	accessToken, refreshToken, expiresAt, err := a.RefreshFunc(a.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	a.AccessToken = accessToken
	a.RefreshToken = refreshToken
	a.ExpiresAt = expiresAt
	return nil
}

// Type returns the authentication method type
func (a *OAuth2Auth) Type() AuthMethod {
	return AuthMethodOAuth2
}

// NoAuth represents no authentication
type NoAuth struct{}

// NewNoAuth creates a new no-auth authenticator
func NewNoAuth() *NoAuth {
	return &NoAuth{}
}

// GetAuthHeader returns empty string
func (a *NoAuth) GetAuthHeader() string {
	return ""
}

// IsExpired always returns false
func (a *NoAuth) IsExpired() bool {
	return false
}

// Refresh is not applicable
func (a *NoAuth) Refresh() error {
	return nil
}

// Type returns a special value for no auth
func (a *NoAuth) Type() AuthMethod {
	return -1
}