package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/middleware"
)

// Simplified router for HTMX demo
func NewSimpleRouter() *gin.Engine {
	// Create router without default middleware
	r := gin.New()
	
	// Add logging and recovery middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// Add i18n middleware for language detection
	i18nMiddleware := middleware.NewI18nMiddleware()
	r.Use(i18nMiddleware.Handle())
	
	// Setup HTMX routes with dynamic template loading
	SetupHTMXRoutes(r)
	
	// Setup LDAP API routes
	SetupLDAPRoutes(r)
	
	// Setup i18n API routes
	i18nHandlers := NewI18nHandlers()
	apiV1 := r.Group("/api/v1")
	i18nHandlers.RegisterRoutes(apiV1)
	
	return r
}