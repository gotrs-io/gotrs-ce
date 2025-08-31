package api

import (
	"log"

	"github.com/gin-gonic/gin"
)

// NewSimpleRouter creates a router with basic routes
func NewSimpleRouter() *gin.Engine {
	log.Println("🔧 Starting NewSimpleRouter initialization")

	// Create router with default middleware
	r := gin.Default()
	log.Println("✅ Gin router created")

	// Initialize pongo2 renderer for templates
	templateDir := "./templates"
	log.Printf("📂 Attempting to initialize pongo2 renderer with template dir: %s", templateDir)
	pongo2Renderer = NewPongo2Renderer(templateDir)
	log.Println("✅ Pongo2 template renderer initialized")

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

	log.Println("🎉 NewSimpleRouter initialization complete")
	return r
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
