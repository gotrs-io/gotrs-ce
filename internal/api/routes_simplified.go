package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// NewSimpleRouter creates a router with basic routes
func NewSimpleRouter() *gin.Engine {
	log.Println("🔧 Starting NewSimpleRouter initialization")

	// Create router with default middleware
	r := gin.Default()
	log.Println("✅ Gin router created")

	// Initialize pongo2 renderer for templates, but only if templates exist
	// Determine template directory with fallbacks
	templateDir := os.Getenv("TEMPLATES_DIR")
	if templateDir == "" {
		// Try local templates then web/templates
		candidates := []string{"./templates", "./web/templates"}
		for _, c := range candidates {
			if fi, err := os.Stat(c); err == nil && fi.IsDir() {
				templateDir = c
				break
			}
		}
	}
	if templateDir != "" {
		if _, err := os.Stat(templateDir); err == nil {
			// Normalize path
			abs, _ := filepath.Abs(templateDir)
			log.Printf("📂 Initializing pongo2 renderer with template dir: %s", abs)
			pongo2Renderer = NewPongo2Renderer(templateDir)
			log.Println("✅ Pongo2 template renderer initialized")
		} else {
			log.Printf("⚠️ Templates directory resolved but not accessible (%s): %v", templateDir, err)
		}
	} else {
		log.Printf("⚠️ No template directory found; renderer disabled (OK for route-only tests)")
	}

	// Static files will be served by SetupHTMXRoutes
	log.Println("📁 Static file serving will be handled by SetupHTMXRoutes")

	log.Println("🔧 About to call SetupHTMXRoutes")
	// Setup HTMX routes
	SetupHTMXRoutes(r)
	log.Println("✅ HTMX routes registered successfully")

	// Test route to verify basic routing works
	log.Println("🧪 Adding test route")
	r.GET("/test", func(c *gin.Context) {
		log.Println("🧪 Test route called")
		c.String(200, "Test route working!")
	})
	log.Println("✅ Test route added")

	// Minimal logout handlers for tests
	ensureRoute(r, http.MethodGet, "/logout", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})
	ensureRoute(r, http.MethodPost, "/logout", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	log.Println("🎉 NewSimpleRouter initialization complete")
	return r
}

func ensureRoute(r *gin.Engine, method, path string, handler gin.HandlerFunc) {
	for _, ri := range r.Routes() {
		if ri.Method == method && ri.Path == path {
			log.Printf("ℹ️ route %s %s already registered; keeping existing handler", method, path)
			return
		}
	}
	r.Handle(method, path, handler)
}

// SetupBasicRoutes adds basic routes to an existing router
func SetupBasicRoutes(r *gin.Engine) {
	log.Println("🔧 SetupBasicRoutes called - adding manual routes")
	log.Println("Basic routes disabled - using YAML routing system")

	// Add a simple manual route to test if basic routing works
	r.GET("/manual-test", func(c *gin.Context) {
		log.Println("🧪 Manual test route called")
		c.String(200, "Manual route working!")
	})
	log.Println("✅ Manual test route added")
}
