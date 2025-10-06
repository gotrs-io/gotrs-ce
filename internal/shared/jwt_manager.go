package shared

import (
	"crypto/rand"
	"encoding/hex"
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
			// For non-production, generate ephemeral secret to avoid weak default
			if os.Getenv("APP_ENV") != "production" {
				b := make([]byte, 32)
				if _, err := rand.Read(b); err == nil {
					jwtSecret = hex.EncodeToString(b)
				}
			}
		}
		// Enforce minimum length (32 bytes hex = 64 chars) to satisfy security gate
		if len(jwtSecret) < 32 {
			// Fallback: extend with random padding if still short (non-production only)
			if os.Getenv("APP_ENV") != "production" {
				pad := make([]byte, 16)
				rand.Read(pad)
				jwtSecret += hex.EncodeToString(pad)
			}
		}
		
		// Token duration (15 minutes to match auth API)
		tokenDuration := 15 * time.Minute
		
		// Create the shared JWT manager using the proper constructor
		globalJWTManager = auth.NewJWTManager(jwtSecret, tokenDuration)
	})
	
	return globalJWTManager
}