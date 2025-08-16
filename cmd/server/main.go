package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/api"
)

func main() {
	// Set Gin mode
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create a simple router with HTMX routes
	r := api.NewSimpleRouter()
	
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "gotrs-htmx"})
	})

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting GOTRS HTMX server on port %s\n", port)
	fmt.Println("Available routes:")
	fmt.Println("  GET  /          -> Redirect to /login")
	fmt.Println("  GET  /login     -> Login page") 
	fmt.Println("  GET  /dashboard -> Dashboard (demo)")
	fmt.Println("  GET  /tickets   -> Tickets list (demo)")
	fmt.Println("  POST /api/auth/login -> HTMX login")
	
	log.Fatal(r.Run(":" + port))
}