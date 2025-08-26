package shared

import (
	"os"
	"sync"
	"time"
	
	"github.com/gotrs-io/gotrs-ce/internal/auth"
)

var (
	globalJWTManager *auth.JWTManager
	jwtOnce         sync.Once
)

// GetJWTManager returns the singleton JWT manager instance
// This ensures auth service and middleware use the same JWT configuration
func GetJWTManager() *auth.JWTManager {
	jwtOnce.Do(func() {
		// Get JWT secret from environment or use default for development
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "development-secret-change-in-production"
		}
		
		// Token duration (24 hours by default)
		tokenDuration := 24 * time.Hour
		
		// Create the shared JWT manager using the proper constructor
		globalJWTManager = auth.NewJWTManager(jwtSecret, tokenDuration)
	})
	
	return globalJWTManager
}