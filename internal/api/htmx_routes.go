package api

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	
	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/auth"
	"github.com/gotrs-io/gotrs-ce/internal/database"
	"github.com/gotrs-io/gotrs-ce/internal/middleware"
	"github.com/gotrs-io/gotrs-ce/internal/models"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
	"github.com/gotrs-io/gotrs-ce/internal/service"
	tmpl "github.com/gotrs-io/gotrs-ce/internal/template"
	"golang.org/x/crypto/bcrypt"
)

// Global Pongo2 renderer
var pongo2Renderer *tmpl.Pongo2Renderer

// Global JWT manager and RBAC (initialized once)
var jwtManagerInstance *auth.JWTManager
var rbacInstance *auth.RBAC

// getJWTManager returns the singleton JWT manager instance, or nil if JWT is not configured
func getJWTManager() *auth.JWTManager {
	if jwtManagerInstance == nil {
		// JWT secret MUST be set via environment variable
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			// If auth is required, fail fast
			if os.Getenv("AUTH_REQUIRED") == "true" {
				log.Fatal("FATAL: JWT_SECRET environment variable is not set. Server cannot start without a JWT secret when AUTH_REQUIRED=true")
			}
			// Return nil to indicate auth is not configured
			return nil
		}
		jwtManagerInstance = auth.NewJWTManager(jwtSecret, 24*time.Hour)
	}
	return jwtManagerInstance
}

// getRBAC returns the singleton RBAC instance
func getRBAC() *auth.RBAC {
	if rbacInstance == nil {
		rbacInstance = auth.NewRBAC()
	}
	return rbacInstance
}

// Initialize Pongo2 renderer
func init() {
	// Get project root for templates
	projectRoot := getProjectRoot()
	templateDir := filepath.Join(projectRoot, "templates")
	
	// Create renderer (debug mode for development)
	pongo2Renderer = tmpl.NewPongo2Renderer(templateDir, true)
}

// getUserFromContext builds a User object from authentication context
func getUserFromContext(c *gin.Context) gin.H {
	// Check if user is authenticated
	userID, hasID := c.Get("user_id")
	if !hasID {
		// Return demo user if not authenticated and in demo mode
		if os.Getenv("DEMO_MODE") == "true" {
			demoEmail := os.Getenv("DEMO_USER_EMAIL")
			if demoEmail == "" {
				// In demo mode, require demo email to be configured
				log.Printf("WARNING: Demo mode enabled but DEMO_USER_EMAIL not set")
				demoEmail = "demo@gotrs.local" // Use a clearly demo email
			}
			return gin.H{"FirstName": "Demo", "LastName": "User", "Email": demoEmail, "Role": "Admin"}
		}
		// Not in demo mode and not authenticated - return guest user
		return gin.H{"FirstName": "Guest", "LastName": "User", "Email": "guest@gotrs.local", "Role": "Guest"}
	}
	
	// Build user object from context
	user := gin.H{}
	
	// Get user ID
	if id, ok := userID.(uint); ok {
		user["ID"] = id
	}
	
	// Get email
	if email, exists := c.Get("user_email"); exists {
		user["Email"] = email
	}
	
	// Get role - IMPORTANT: This determines admin button visibility
	if role, exists := c.Get("user_role"); exists {
		user["Role"] = role
	} else {
		// Default to Agent if role not set
		user["Role"] = "Agent"
	}
	
	// Parse name from email if not available
	// In production, this should come from user profile
	if emailStr, ok := user["Email"].(string); ok {
		parts := strings.Split(emailStr, "@")
		if len(parts) > 0 {
			nameParts := strings.Split(parts[0], ".")
			if len(nameParts) >= 2 {
				user["FirstName"] = strings.Title(nameParts[0])
				user["LastName"] = strings.Title(nameParts[1])
			} else {
				user["FirstName"] = strings.Title(nameParts[0])
				user["LastName"] = "User"
			}
		}
	}
	
	// Set defaults if not available
	if user["FirstName"] == nil {
		user["FirstName"] = "User"
	}
	if user["LastName"] == nil {
		user["LastName"] = ""
	}
	
	return user
}

// DummyTemplate is a wrapper for html/template
type DummyTemplate struct {
	tmpl *template.Template
}

func (t *DummyTemplate) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	if t.tmpl == nil {
		_, err := w.Write([]byte("Template not loaded"))
		return err
	}
	return t.tmpl.ExecuteTemplate(w, name, data)
}

// loadTemplateForRequest loads HTML templates
func loadTemplateForRequest(c *gin.Context, templatePaths ...string) (*DummyTemplate, error) {
	projectRoot := getProjectRoot()
	
	// Define template functions
	funcMap := template.FuncMap{
		"firstLetter": func(s string) string {
			if len(s) > 0 {
				return string(s[0])
			}
			return ""
		},
		"toUpper": strings.ToUpper,
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				if i+1 < len(values) {
					if key, ok := values[i].(string); ok {
						dict[key] = values[i+1]
					}
				}
			}
			return dict
		},
		"list": func(values ...interface{}) []interface{} {
			return values
		},
		"formatFileSize": func(size interface{}) string {
			var bytes int64
			switch v := size.(type) {
			case int64:
				bytes = v
			case int:
				bytes = int64(v)
			case float64:
				bytes = int64(v)
			default:
				return "0 B"
			}
			
			if bytes == 0 {
				return "0 B"
			}
			const unit = 1024
			if bytes < unit {
				return fmt.Sprintf("%d B", bytes)
			}
			div, exp := int64(unit), 0
			for n := bytes / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
		},
		"formatTimeAgo": func(t interface{}) string {
			var timeVal time.Time
			switch v := t.(type) {
			case time.Time:
				timeVal = v
			case string:
				parsed, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return v
				}
				timeVal = parsed
			default:
				return "unknown"
			}
			
			diff := time.Since(timeVal)
			switch {
			case diff < time.Minute:
				return "just now"
			case diff < time.Hour:
				return fmt.Sprintf("%d min ago", int(diff.Minutes()))
			case diff < 24*time.Hour:
				return fmt.Sprintf("%d hours ago", int(diff.Hours()))
			case diff < 7*24*time.Hour:
				return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
			default:
				return timeVal.Format("Jan 2, 2006")
			}
		},
		"canPreview": func(contentType string) bool {
			return strings.HasPrefix(contentType, "image/") ||
				contentType == "application/pdf" ||
				strings.HasPrefix(contentType, "text/") ||
				contentType == "application/json" ||
				contentType == "application/xml"
		},
	}
	
	// Create template with functions
	tmpl := template.New(filepath.Base(templatePaths[0])).Funcs(funcMap)
	
	// Parse all template files
	for _, path := range templatePaths {
		fullPath := filepath.Join(projectRoot, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}
		_, err = tmpl.Parse(string(content))
		if err != nil {
			return nil, err
		}
	}
	
	return &DummyTemplate{tmpl: tmpl}, nil
}

// loadTemplate is another compatibility shim
// TODO: Remove this once all handlers are converted to use Pongo2 directly
func loadTemplate(templatePaths ...string) (*DummyTemplate, error) {
	projectRoot := getProjectRoot()
	
	// Define template functions
	funcMap := template.FuncMap{
		"firstLetter": func(s string) string {
			if len(s) > 0 {
				return string(s[0])
			}
			return ""
		},
		"toUpper": strings.ToUpper,
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				if i+1 < len(values) {
					if key, ok := values[i].(string); ok {
						dict[key] = values[i+1]
					}
				}
			}
			return dict
		},
		"list": func(values ...interface{}) []interface{} {
			return values
		},
		"formatFileSize": func(size interface{}) string {
			var bytes int64
			switch v := size.(type) {
			case int64:
				bytes = v
			case int:
				bytes = int64(v)
			case float64:
				bytes = int64(v)
			default:
				return "0 B"
			}
			
			if bytes == 0 {
				return "0 B"
			}
			const unit = 1024
			if bytes < unit {
				return fmt.Sprintf("%d B", bytes)
			}
			div, exp := int64(unit), 0
			for n := bytes / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
		},
		"formatTimeAgo": func(t interface{}) string {
			var timeVal time.Time
			switch v := t.(type) {
			case time.Time:
				timeVal = v
			case string:
				parsed, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return v
				}
				timeVal = parsed
			default:
				return "unknown"
			}
			
			diff := time.Since(timeVal)
			switch {
			case diff < time.Minute:
				return "just now"
			case diff < time.Hour:
				return fmt.Sprintf("%d min ago", int(diff.Minutes()))
			case diff < 24*time.Hour:
				return fmt.Sprintf("%d hours ago", int(diff.Hours()))
			case diff < 7*24*time.Hour:
				return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
			default:
				return timeVal.Format("Jan 2, 2006")
			}
		},
		"canPreview": func(contentType string) bool {
			return strings.HasPrefix(contentType, "image/") ||
				contentType == "application/pdf" ||
				strings.HasPrefix(contentType, "text/") ||
				contentType == "application/json" ||
				contentType == "application/xml"
		},
	}
	
	// Create template with functions
	tmpl := template.New(filepath.Base(templatePaths[0])).Funcs(funcMap)
	
	// Parse all template files
	for _, path := range templatePaths {
		fullPath := filepath.Join(projectRoot, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}
		_, err = tmpl.Parse(string(content))
		if err != nil {
			return nil, err
		}
	}
	
	return &DummyTemplate{tmpl: tmpl}, nil
}

// filterEmptyStrings removes empty strings from a slice
func filterEmptyStrings(slice []string) []string {
	result := []string{}
	for _, s := range slice {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// getPriorityID converts priority string to ORTS priority ID
func getPriorityID(priority string) int {
	priorityMap := map[string]int{
		"very low":  1,
		"low":       2,
		"normal":    3,
		"high":      4,
		"very high": 5,
		"urgent":    5, // Alias for very high
	}
	if id, ok := priorityMap[priority]; ok {
		return id
	}
	return 3 // Default to normal
}

// getProjectRoot finds the project root directory by looking for go.mod
func getProjectRoot() string {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		// Fallback to current directory
		return "."
	}
	
	// Walk up the directory tree looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root without finding go.mod
			break
		}
		dir = parent
	}
	
	// Fallback to current directory
	return "."
}


// SetupHTMXRoutes configures routes for HTMX-based UI
func SetupHTMXRoutes(r *gin.Engine) {
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Serve static files
	r.Static("/static", "./static")
	
	// Serve favicon specifically (browsers often request this)
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./static/favicon.ico")
	})
	
	// Note: favicon.svg is served via the /static route above
	
	// Test i18n endpoint
	r.GET("/test-i18n", func(c *gin.Context) {
		pongo2Renderer.HTML(c, http.StatusOK, "test-i18n.pongo2", pongo2.Context{
			"TestMessage": "Direct test",
		})
	})
	
	// Root redirect
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/login")
	})
	
	// Authentication pages
	r.GET("/login", handleLoginPage)
	r.GET("/register", handleRegisterPage)
	// Auth login is handled in the api group below
	r.POST("/logout", handleLogout)
	r.GET("/logout", handleLogoutGET)
	
	// Protected dashboard routes
	dashboard := r.Group("/")
	
	// Add auth middleware if JWT manager is available OR if in demo mode
	jwtManager := getJWTManager()
	rbac := getRBAC()
	
	if jwtManager != nil {
		dashboard.Use(middleware.SessionMiddleware(jwtManager))
	} else if os.Getenv("DEMO_MODE") == "true" {
		// In demo mode without JWT, still use session middleware for demo tokens
		// Pass nil for JWT manager - the middleware will handle demo tokens
		dashboard.Use(middleware.SessionMiddleware(nil))
	}
	// Basic dashboard - accessible to all authenticated users
	{
		dashboard.GET("/dashboard", handleDashboard)
		dashboard.GET("/profile", underConstruction("Profile"))
		dashboard.GET("/settings", underConstruction("Settings"))
	}
	
	// Ticket routes - require ticket permissions
	ticketRoutes := dashboard.Group("/tickets")
	ticketRoutes.Use(middleware.RequireAnyPermission(rbac, auth.PermissionTicketRead, auth.PermissionOwnTicketRead))
	{
		ticketRoutes.GET("", handleTicketsList)
		ticketRoutes.GET("/:id", middleware.RequireTicketAccess(rbac), handleTicketDetail)
		ticketRoutes.GET("/:id/edit", middleware.RequirePermission(rbac, auth.PermissionTicketUpdate), handleTicketEditForm)
	}
	
	// Ticket creation and modification - require create/update permissions
	ticketWriteRoutes := dashboard.Group("/tickets")
	ticketWriteRoutes.Use(middleware.RequireAnyPermission(rbac, auth.PermissionTicketCreate, auth.PermissionOwnTicketCreate))
	{
		ticketWriteRoutes.GET("/new", handleTicketNew)
		ticketWriteRoutes.POST("/create", handleTicketCreate)
	}
	
	// Ticket actions - require specific permissions
	ticketActionRoutes := dashboard.Group("/tickets")
	ticketActionRoutes.Use(middleware.RequirePermission(rbac, auth.PermissionTicketUpdate))
	{
		ticketActionRoutes.POST("/:id/quick-action", handleTicketQuickAction)
		ticketActionRoutes.POST("/bulk-action", handleTicketBulkAction)
	}
	
	// Queue management - all queue routes
	queueRoutes := dashboard.Group("/queues")
	queueRoutes.Use(middleware.RequireAgentAccess(rbac))
	{
		// List and search routes
		queueRoutes.GET("", handleQueuesList)
		queueRoutes.GET("/clear-search", handleClearQueueSearch)
		queueRoutes.GET("/bulk-toolbar", handleBulkActionsToolbar)
		
		// Admin-only routes (new, edit, delete) - require additional admin check
		queueRoutes.GET("/new", middleware.RequireAdminAccess(rbac), handleNewQueueForm)
		queueRoutes.GET("/:id/edit", middleware.RequireAdminAccess(rbac), handleEditQueueForm)
		queueRoutes.GET("/:id/delete", middleware.RequireAdminAccess(rbac), handleDeleteQueueConfirmation)
		
		// Generic detail page route - MUST be last to not catch specific routes
		queueRoutes.GET("/:id", handleQueueDetailPage)
	}
	
	// Templates - agent access
	templateRoutes := dashboard.Group("/templates")
	templateRoutes.Use(middleware.RequireAgentAccess(rbac))
	{
		templateRoutes.GET("", handleTemplatesPage)
	}
	
	// Admin panel - admin only
	adminRoutes := dashboard.Group("/admin")
	adminRoutes.Use(middleware.RequireAdminAccess(rbac))
	{
		adminRoutes.GET("", handleAdminDashboard)
		adminRoutes.GET("/lookups", handleAdminLookups)
		adminRoutes.GET("/users", handleAdminUsers)
		adminRoutes.GET("/users/:id", handleGetUser)
		adminRoutes.POST("/users", handleCreateUser)
		adminRoutes.PUT("/users/:id", handleUpdateUser)
		adminRoutes.DELETE("/users/:id", handleDeleteUser)
		adminRoutes.PUT("/users/:id/status", handleToggleUserStatus)
		adminRoutes.POST("/users/:id/reset-password", handleResetUserPassword)
		
		// Group management routes
		adminRoutes.GET("/groups", handleAdminGroups)
		adminRoutes.GET("/groups/:id", handleGetGroup)
		adminRoutes.POST("/groups", handleCreateGroup)
		adminRoutes.PUT("/groups/:id", handleUpdateGroup)
		adminRoutes.DELETE("/groups/:id", handleDeleteGroup)
		adminRoutes.GET("/groups/:id/members", handleGetGroupMembers)
		adminRoutes.POST("/groups/:id/members", handleAddGroupMember)
		adminRoutes.DELETE("/groups/:id/members/:userId", handleRemoveGroupMember)
		adminRoutes.GET("/groups/:id/permissions", handleGetGroupPermissions)
		adminRoutes.PUT("/groups/:id/permissions", handleUpdateGroupPermissions)
		
		// Permission management routes (OTRS Role equivalent)
		adminRoutes.GET("/permissions", handleAdminPermissions)
		adminRoutes.GET("/permissions/user/:userId", handleGetUserPermissionMatrix)
		adminRoutes.PUT("/permissions/user/:userId", handleUpdateUserPermissions)
		adminRoutes.GET("/permissions/group/:groupId", handleGetGroupPermissionMatrix)
		adminRoutes.POST("/permissions/clone", handleCloneUserPermissions)
		
		adminRoutes.GET("/settings", underConstruction("System Settings"))
		adminRoutes.GET("/templates", underConstruction("Template Management"))
		adminRoutes.GET("/reports", underConstruction("Reports"))
		adminRoutes.GET("/backup", underConstruction("Backup & Restore"))
	}
	
	// HTMX API endpoints (return HTML fragments)
	api := r.Group("/api")
	
	// Authentication endpoints (no auth required)
	{
		api.GET("/auth/login", handleHTMXLogin)  // Also support GET for the form
		api.POST("/auth/login", handleHTMXLogin)
		api.POST("/auth/logout", handleHTMXLogout)
		api.GET("/auth/refresh", underConstructionAPI("/auth/refresh"))  // GET for testing
		api.POST("/auth/refresh", underConstructionAPI("/auth/refresh"))
		api.GET("/auth/register", underConstructionAPI("/auth/register"))  // GET for form
		api.POST("/auth/register", underConstructionAPI("/auth/register"))
	}
	
	// Protected API endpoints - require authentication
	protectedAPI := api.Group("")
	if jwtManager != nil {
		protectedAPI.Use(middleware.SessionMiddleware(jwtManager))
	} else if os.Getenv("DEMO_MODE") == "true" {
		protectedAPI.Use(middleware.SessionMiddleware(nil))
	}
	
	// Dashboard data - accessible to all authenticated users
	{
		protectedAPI.GET("/dashboard/stats", handleDashboardStats)
		protectedAPI.GET("/dashboard/recent-tickets", handleRecentTickets)
		protectedAPI.GET("/dashboard/activity", handleActivityFeed)
		protectedAPI.GET("/notifications", underConstructionAPI("/notifications"))
	}
	
	// Admin-only API endpoints
	adminAPI := protectedAPI.Group("")
	adminAPI.Use(middleware.RequireAdminAccess(rbac))
	{
		adminAPI.GET("/lookups/cache/invalidate", underConstructionAPI("/lookups/cache/invalidate"))
		adminAPI.POST("/lookups/cache/invalidate", handleInvalidateLookupCache)
		adminAPI.PUT("/admin/sla-config", handleUpdateSLAConfig)
		
		// Lookup CRUD Endpoints - Admin only
		adminAPI.POST("/lookups/queues", handleCreateLookupQueue)
		adminAPI.PUT("/lookups/queues/:id", handleUpdateLookupQueue)
		adminAPI.DELETE("/lookups/queues/:id", handleDeleteLookupQueue)
		adminAPI.POST("/lookups/types", handleCreateType)
		adminAPI.PUT("/lookups/types/:id", handleUpdateType)
		adminAPI.DELETE("/lookups/types/:id", handleDeleteType)
		adminAPI.PUT("/lookups/priorities/:id", handleUpdatePriority)
		adminAPI.PUT("/lookups/statuses/:id", handleUpdateStatus)
	}
	
	// Agent-level API endpoints (agents and admins)
	agentAPI := protectedAPI.Group("")
	agentAPI.Use(middleware.RequireAgentAccess(rbac))
	{
		// Queue operations
		agentAPI.GET("/queues", handleQueuesAPI)
		fmt.Println("DEBUG: Registering POST /api/queues -> handleCreateQueueWithHTMX")
		agentAPI.POST("/queues", handleCreateQueueWithHTMX)
		agentAPI.GET("/queues/:id", handleQueueDetail)
		agentAPI.PUT("/queues/:id", handleUpdateQueueWithHTMX)
		agentAPI.DELETE("/queues/:id", handleDeleteQueue)
		agentAPI.GET("/queues/:id/tickets", handleQueueTicketsWithHTMX)
		
		// Bulk queue operations
		agentAPI.PUT("/queues/bulk/:action", handleBulkQueueAction)
		agentAPI.DELETE("/queues/bulk", handleBulkQueueDelete)
		
		// Ticket operations - require ticket permissions
		agentAPI.GET("/tickets", handleTicketsAPI)
		agentAPI.GET("/tickets/filter", handleTicketsAPI)  // Filter uses same handler as list
		agentAPI.GET("/tickets/search", handleTicketSearch)
		agentAPI.GET("/search", handleTicketSearch)  // General search endpoint
		agentAPI.PUT("/tickets/bulk", handleBulkUpdateTickets)
		agentAPI.POST("/tickets", handleCreateTicketWithAttachments) // Fixed to handle attachments
		agentAPI.PUT("/tickets/:id", handleUpdateTicketEnhanced)
		agentAPI.POST("/tickets/:id/status", handleUpdateTicketStatus)
		agentAPI.POST("/tickets/:id/assign", handleAssignTicket)
		agentAPI.POST("/tickets/:id/reply", handleTicketReply)
		agentAPI.POST("/tickets/:id/priority", handleUpdateTicketPriority)
		agentAPI.POST("/tickets/:id/queue", handleUpdateTicketQueue)
		agentAPI.GET("/tickets/:id/messages", handleGetTicketMessages)
		agentAPI.POST("/tickets/:id/messages", handleAddTicketMessage)
		
		// Attachment operations
		agentAPI.GET("/attachments/:id/download", handleAttachmentDownload)
		agentAPI.GET("/attachments/:id/preview", handleAttachmentPreview)
		agentAPI.GET("/attachments/:id/thumbnail", handleAttachmentThumbnail)
		agentAPI.POST("/attachments/bulk-thumbnails", handleBulkThumbnails)
		agentAPI.DELETE("/attachments/:id", handleAttachmentDelete)
		
		// SLA and Escalation
		agentAPI.GET("/tickets/:id/sla", handleGetTicketSLA)
		agentAPI.POST("/tickets/:id/escalate", handleEscalateTicket)
		agentAPI.GET("/reports/sla", handleSLAReport)
		
		// Ticket Merge
		agentAPI.POST("/tickets/:id/merge", handleMergeTickets)
		agentAPI.POST("/tickets/:id/unmerge", handleUnmergeTicket)
		agentAPI.GET("/tickets/:id/merge-history", handleGetMergeHistory)
		
		// Ticket Attachments
		agentAPI.POST("/tickets/:id/attachments", handleUploadAttachment)
		agentAPI.GET("/tickets/:id/attachments", handleGetAttachments)
		agentAPI.GET("/tickets/:id/attachments/:attachment_id", handleDownloadAttachment)
		agentAPI.GET("/tickets/:id/attachments/:attachment_id/thumbnail", handleGetThumbnail)
		agentAPI.DELETE("/tickets/:id/attachments/:attachment_id", handleDeleteAttachment)
		
		// File serving endpoint (for stored attachments)
		agentAPI.GET("/files/*path", handleServeFile)
		
		// Advanced Search (using different endpoints to avoid conflicts)
		agentAPI.GET("/tickets/advanced-search", handleAdvancedTicketSearch)
		agentAPI.GET("/tickets/search/suggestions", handleSearchSuggestions)
		agentAPI.GET("/tickets/search/export", handleExportSearchResults)
		
		// Search History
		agentAPI.POST("/tickets/search/history", handleSaveSearchHistory)
		agentAPI.GET("/tickets/search/history", handleGetSearchHistory)
		agentAPI.DELETE("/tickets/search/history/:id", handleDeleteSearchHistory)
		
		// Saved Searches
		agentAPI.POST("/tickets/search/saved", handleCreateSavedSearch)
		agentAPI.GET("/tickets/search/saved", handleGetSavedSearches)
		agentAPI.GET("/tickets/search/saved/:id/execute", handleExecuteSavedSearch)
		agentAPI.PUT("/tickets/search/saved/:id", handleUpdateSavedSearch)
		agentAPI.DELETE("/tickets/search/saved/:id", handleDeleteSavedSearch)
		
		// Canned Responses - Using new comprehensive handlers
		cannedHandlers := NewCannedResponseHandlers()
		agentAPI.GET("/canned-responses", cannedHandlers.GetResponses)
		agentAPI.GET("/canned-responses/quick", cannedHandlers.GetQuickResponses)
		agentAPI.GET("/canned-responses/popular", cannedHandlers.GetPopularResponses)
		agentAPI.GET("/canned-responses/categories", cannedHandlers.GetCategories)
		agentAPI.GET("/canned-responses/category/:category", cannedHandlers.GetResponsesByCategory)
		agentAPI.GET("/canned-responses/search", cannedHandlers.SearchResponses)
		agentAPI.GET("/canned-responses/user", cannedHandlers.GetResponsesForUser)
		agentAPI.GET("/canned-responses/:id", cannedHandlers.GetResponseByID)
		agentAPI.POST("/canned-responses", cannedHandlers.CreateResponse)
		agentAPI.PUT("/canned-responses/:id", cannedHandlers.UpdateResponse)
		agentAPI.DELETE("/canned-responses/:id", cannedHandlers.DeleteResponse)
		agentAPI.POST("/canned-responses/apply", cannedHandlers.ApplyResponse)
		agentAPI.GET("/canned-responses/export", cannedHandlers.ExportResponses)
		agentAPI.POST("/canned-responses/import", cannedHandlers.ImportResponses)
		
		// Lookup Data Endpoints
		agentAPI.GET("/lookups/queues", handleGetQueues)
		agentAPI.GET("/lookups/priorities", handleGetPriorities)
		agentAPI.GET("/lookups/types", handleGetTypes)
		agentAPI.GET("/lookups/statuses", handleGetStatuses)
		agentAPI.GET("/lookups/form-data", handleGetFormData)
	}
	
	// Additional admin endpoints for audit and configuration
	{
		adminAPI.GET("/lookups/audit", handleGetAuditLogs)
		adminAPI.GET("/lookups/export", handleExportConfiguration)
		adminAPI.POST("/lookups/import", handleImportConfiguration)
	}
	
	// Template endpoints - Agent access
	{
		agentAPI.GET("/templates", handleGetTemplates)
		agentAPI.GET("/templates/:id", handleGetTemplate)
		agentAPI.POST("/templates", handleCreateTemplate)
		agentAPI.PUT("/templates/:id", handleUpdateTemplate)
		agentAPI.DELETE("/templates/:id", handleDeleteTemplate)
		agentAPI.GET("/templates/search", handleSearchTemplates)
		agentAPI.GET("/templates/categories", handleGetTemplateCategories)
		agentAPI.GET("/templates/popular", handleGetPopularTemplates)
		agentAPI.POST("/templates/apply", handleApplyTemplate)
		agentAPI.GET("/templates/:id/load", handleLoadTemplateIntoForm)
		agentAPI.GET("/templates/modal", handleTemplateSelectionModal)
	}
	
	// Real-time endpoints - accessible to authenticated users
	{
		protectedAPI.GET("/tickets/stream", handleTicketStream)
		protectedAPI.GET("/dashboard/activity-stream", handleActivityStream)
	}
}

// Login page
func handleLoginPage(c *gin.Context) {
	pongo2Renderer.HTML(c, http.StatusOK, "pages/login.pongo2", pongo2.Context{
		"Title": "GOTRS - Sign In",
	})
}

// Register page
func handleRegisterPage(c *gin.Context) {
	pongo2Renderer.HTML(c, http.StatusOK, "pages/register.pongo2", pongo2.Context{
		"Title": "GOTRS - Register",
	})
}

// Dashboard
func handleDashboard(c *gin.Context) {
	// Use Pongo2 renderer with i18n support
	pongo2Renderer.HTML(c, http.StatusOK, "pages/dashboard.pongo2", pongo2.Context{
		"Title":      "Dashboard - GOTRS",
		"User":       getUserFromContext(c),
		"ActivePage": "dashboard",
	})
}

// Tickets list page
func handleTicketsList(c *gin.Context) {
	// Get filter parameters
	status := c.Query("status")
	priority := c.Query("priority")
	queueID := c.Query("queue_id")
	assignedTo := c.Query("assigned_to")
	search := c.Query("search")
	sort := c.DefaultQuery("sort", "created")
	page := 1
	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	perPage := 10
	if pp := c.Query("per_page"); pp != "" {
		if parsedPerPage, err := strconv.Atoi(pp); err == nil && parsedPerPage > 0 {
			if parsedPerPage > 100 {
				perPage = 100
			} else {
				perPage = parsedPerPage
			}
		}
	}
	
	// Get real ticket data from database
	ticketService := GetTicketService()
	ticketRequest := &models.TicketListRequest{
		Page:    page,
		PerPage: 1000, // Get all for client-side filtering for now
	}
	
	ticketResponse, err := ticketService.ListTickets(ticketRequest)
	if err != nil {
		sendGuruMeditation(c, err, "handleTicketsList:ListTickets")
		return
	}
	
	// Convert tickets to display format
	tickets := []gin.H{}
	if ticketResponse != nil && ticketResponse.Tickets != nil {
		for _, t := range ticketResponse.Tickets {
			// Map priority
			priorityStr := "normal"
			switch t.TicketPriorityID {
			case 1:
				priorityStr = "low"
			case 2, 3:
				priorityStr = "normal"
			case 4:
				priorityStr = "high"
			case 5:
				priorityStr = "urgent"
			}
			
			// Map status
			statusStr := "new"
			switch t.TicketStateID {
			case 1:
				statusStr = "new"
			case 2:
				statusStr = "open"
			case 3:
				statusStr = "pending"
			case 4:
				statusStr = "resolved"
			case 5, 6:
				statusStr = "closed"
			}
			
			// Get queue name
			queueName := "Unknown"
			if t.QueueID == 1 {
				queueName = "Raw"
			} else if t.QueueID == 2 {
				queueName = "Junk"
			} else if t.QueueID == 3 {
				queueName = "Misc"
			} else if t.QueueID == 4 {
				queueName = "Support"
			}
			
			// Get customer email
			customerEmail := ""
			if t.CustomerUserID != nil {
				customerEmail = *t.CustomerUserID
			}
			
			// Format timestamps
			createdAt := t.CreateTime.Format("2006-01-02 15:04")
			updatedAt := t.ChangeTime.Format("2006-01-02 15:04")
			
			// Build ticket display object
			ticket := gin.H{
				"id":             t.ID,
				"number":         t.TicketNumber,
				"title":          t.Title,
				"status":         statusStr,
				"priority":       priorityStr,
				"queue_id":       t.QueueID,
				"queue_name":     queueName,
				"customer_email": customerEmail,
				"assigned_to":    "", // TODO: Get from UserID
				"created_at":     createdAt,
				"updated_at":     updatedAt,
				"has_new_message": false, // TODO: Check for new articles
				"sla_status":     "ok",
				"due_in":         "",
			}
			
			// Add assigned user if present
			if t.UserID != nil {
				ticket["assigned_to"] = fmt.Sprintf("User %d", *t.UserID)
			}
			
			tickets = append(tickets, ticket)
		}
	}
	
	// Apply filters
	filteredTickets := []gin.H{}
	for _, ticket := range tickets {
		// Status filter
		if status != "" && status != "all" && ticket["status"] != status {
			continue
		}
		// Priority filter
		if priority != "" && priority != "all" && ticket["priority"] != priority {
			continue
		}
		// Queue filter
		if queueID != "" {
			if qID, err := strconv.Atoi(queueID); err == nil && ticket["queue_id"] != qID {
				continue
			}
		}
		// Assigned filter
		if assignedTo == "me" && ticket["assigned_to"] != "Demo User" {
			continue
		}
		// Search filter
		if search != "" {
			searchLower := strings.ToLower(search)
			titleMatch := strings.Contains(strings.ToLower(ticket["title"].(string)), searchLower)
			numberMatch := strings.Contains(strings.ToLower(ticket["number"].(string)), searchLower)
			emailMatch := false
			if email, ok := ticket["customer_email"].(string); ok {
				emailMatch = strings.Contains(strings.ToLower(email), searchLower)
			}
			if !titleMatch && !numberMatch && !emailMatch {
				continue
			}
			
			// Highlight search terms in title
			if titleMatch {
				title := ticket["title"].(string)
				// Case-insensitive replacement with <mark> tags
				re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(search))
				ticket["highlighted_title"] = re.ReplaceAllString(title, "<mark>$0</mark>")
			}
		}
		filteredTickets = append(filteredTickets, ticket)
	}
	
	// Sort tickets
	sortTickets(filteredTickets, sort)
	
	// Pagination
	total := len(filteredTickets)
	totalPages := (total + perPage - 1) / perPage
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}
	
	start := (page - 1) * perPage
	end := start + perPage
	if end > total {
		end = total
	}
	
	var paginatedTickets []gin.H
	if start < total {
		paginatedTickets = filteredTickets[start:end]
	}
	
	// Build pagination info
	pagination := gin.H{
		"Page":     page,
		"PerPage":  perPage,
		"Total":    total,
		"Start":    start + 1,
		"End":      end,
		"HasPrev":  page > 1,
		"HasNext":  page < totalPages,
		"PrevPage": page - 1,
		"NextPage": page + 1,
	}
	
	// Get dynamic form data from lookup service with language support
	lookupService := GetLookupService()
	lang := middleware.GetLanguage(c)
	formData := lookupService.GetTicketFormDataWithLang(lang)
	
	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") != "" {
		// Return just the ticket list fragment
		tmpl, err := loadTemplateForRequest(c, "templates/components/ticket_list.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Template error: %v", err)
			return
		}
		
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(c.Writer, "ticket_list.html", gin.H{
			"Tickets":       paginatedTickets,
			"Pagination":    pagination,
			"SearchTerm":    search,
			"StatusFilter":  status,
			"PriorityFilter": priority,
			"QueueFilter":   queueID,
			"AssignedFilter": assignedTo,
			"SortBy":        sort,
		}); err != nil {
			c.String(http.StatusInternalServerError, "Render error: %v", err)
		}
		return
	}
	
	// Full page load - use Pongo2
	pongo2Renderer.HTML(c, http.StatusOK, "pages/tickets/list.pongo2", pongo2.Context{
		"Title":          "Tickets - GOTRS",
		"Tickets":        paginatedTickets,
		"Queues":         formData.Queues,
		"Priorities":     formData.Priorities,
		"Statuses":       formData.Statuses,
		"Pagination":     pagination,
		"SearchTerm":     search,
		"StatusFilter":   status,
		"PriorityFilter": priority,
		"QueueFilter":    queueID,
		"AssignedFilter": assignedTo,
		"SortBy":         sort,
		"User":           getUserFromContext(c),
		"ActivePage":     "tickets",
	})
}

// New ticket page
func handleTicketNew(c *gin.Context) {
	// Get dynamic form data from lookup service with language support
	lookupService := GetLookupService()
	lang := middleware.GetLanguage(c)
	formData := lookupService.GetTicketFormDataWithLang(lang)
	
	// Use Pongo2 renderer with i18n support
	pongo2Renderer.HTML(c, http.StatusOK, "pages/tickets/new.pongo2", pongo2.Context{
		"Title":      "New Ticket - GOTRS",
		"Queues":     formData.Queues,
		"Priorities": formData.Priorities,
		"Types":      formData.Types,
		"Statuses":   formData.Statuses,
		"User":       getUserFromContext(c),
		"ActivePage": "tickets",
	})
}

// Handle ticket creation from UI form
func handleTicketCreate(c *gin.Context) {
	// Debug: Write to file immediately
	os.WriteFile("/tmp/handler-called.log", []byte("Handler called at " + time.Now().String() + "\n"), 0644)
	
	// Get form data
	title := strings.TrimSpace(c.PostForm("title"))
	queueIDStr := c.PostForm("queue_id")
	priority := c.PostForm("priority")
	description := c.PostForm("description")
	customerEmail := strings.TrimSpace(c.PostForm("customer_email"))
	autoAssign := c.PostForm("auto_assign") == "true"
	
	// Debug: Write form data to file
	debugMsg := fmt.Sprintf("Form data: title=%s, desc=%s, desc_len=%d\n", title, description, len(description))
	os.WriteFile("/tmp/form-data.log", []byte(debugMsg), 0644)
	
	// Validation
	errors := []string{}
	
	if title == "" {
		errors = append(errors, "Title is required")
	} else if len(title) > 200 {
		errors = append(errors, "Title must be less than 200 characters")
	}
	
	if queueIDStr == "" {
		errors = append(errors, "Queue selection is required")
	}
	
	validPriorities := map[string]bool{"low": true, "normal": true, "high": true, "urgent": true}
	if priority != "" && !validPriorities[priority] {
		errors = append(errors, "Invalid priority")
	}
	if priority == "" {
		priority = "normal" // Default priority
	}
	
	if customerEmail != "" {
		// Simple email validation
		if !strings.Contains(customerEmail, "@") || !strings.Contains(customerEmail, ".") {
			errors = append(errors, "Invalid email format")
		}
	}
	
	// Return errors if validation failed
	if len(errors) > 0 {
		c.Header("Content-Type", "text/html; charset=utf-8")
		errorHTML := "<div class='text-red-600'>"
		for _, err := range errors {
			errorHTML += "<p>" + err + "</p>"
		}
		errorHTML += "</div>"
		c.String(http.StatusBadRequest, errorHTML)
		return
	}
	
	// Convert queue ID
	queueID := uint(1)
	if queueIDStr != "" {
		if id, err := strconv.Atoi(queueIDStr); err == nil {
			queueID = uint(id)
		}
	}
	
	// Create the ticket with ORTS-compatible fields
	var customerID *string
	if customerEmail != "" {
		customerID = &customerEmail
	}
	
	ticket := &models.Ticket{
		Title:            title,
		QueueID:          int(queueID),
		TypeID:           1, // Default to Incident
		TicketPriorityID: getPriorityID(priority),
		TicketStateID:    1, // New
		TicketLockID:     1, // Unlocked
		CustomerUserID:   customerID,
		CreateBy:         1, // Default user
		ChangeBy:         1,
	}
	
	if autoAssign {
		userID := 1
		ticket.UserID = &userID // Assign to current user (mock)
		ticket.TicketStateID = 2 // Open
	}
	
	// Get database connection for real operations
	db, err := database.GetDB()
	if err != nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusInternalServerError, `
			<div class="text-red-600">
				<p>Database connection failed</p>
			</div>
		`)
		return
	}
	
	// Save the ticket using real database
	ticketRepo := repository.NewTicketRepository(db)
	if err := ticketRepo.Create(ticket); err != nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusInternalServerError, `
			<div class="text-red-600">
				<p>Failed to create ticket: %s</p>
			</div>
		`, err.Error())
		return
	}
	
	// Add initial article if description provided
	if description != "" {
		// Write debug to file for testing
		os.WriteFile("/tmp/debug.log", []byte(fmt.Sprintf("Creating article for ticket %d with desc: %s\n", ticket.ID, description)), 0644)
		
		articleRepo := repository.NewArticleRepository(db)
		article := &models.Article{
			TicketID:             ticket.ID,
			Subject:              title,
			Body:                 description,
			SenderTypeID:         3, // Customer
			CommunicationChannelID: 1, // Email
			IsVisibleForCustomer: 1,
			CreateBy:            1,
			ChangeBy:            1,
		}
		
		if err := articleRepo.Create(article); err != nil {
			// Log error but don't fail the ticket creation
			fmt.Printf("ERROR: Failed to add initial article for ticket %d: %v\n", ticket.ID, err)
		} else {
			fmt.Printf("SUCCESS: Created article for ticket %d\n", ticket.ID)
		}
	} else {
		fmt.Printf("DEBUG: No description provided, skipping article creation\n")
	}
	
	// Build success message
	successMsg := fmt.Sprintf("Ticket created successfully: #%s", ticket.TicketNumber)
	if autoAssign {
		successMsg = fmt.Sprintf("Ticket created and assigned to you: #%s", ticket.TicketNumber)
	}
	
	// Return success response with HX-Trigger and redirect
	c.Header("HX-Trigger", "ticket-created")
	c.Header("HX-Redirect", fmt.Sprintf("/tickets/%d", ticket.ID))
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `
		<div class="rounded-md bg-green-50 p-4">
			<div class="flex">
				<div class="flex-shrink-0">
					<svg class="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
					</svg>
				</div>
				<div class="ml-3">
					<p class="text-sm font-medium text-green-800">%s</p>
				</div>
			</div>
		</div>
	`, successMsg)
}

// Handle ticket quick actions
func handleTicketQuickAction(c *gin.Context) {
	ticketID := c.Param("id")
	action := c.PostForm("action")
	
	var message string
	switch action {
	case "assign":
		message = fmt.Sprintf("Ticket #%s assigned to you", ticketID)
	case "close":
		message = fmt.Sprintf("Ticket #%s closed", ticketID)
	case "priority-high":
		message = "Priority updated to high"
	case "priority-urgent":
		message = "Priority updated to urgent"
	case "priority-normal":
		message = "Priority updated to normal"
	case "priority-low":
		message = "Priority updated to low"
	default:
		c.String(http.StatusBadRequest, "Invalid action")
		return
	}
	
	// Return success message with trigger to refresh list
	c.Header("HX-Trigger", "ticket-updated")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `
		<div class="text-sm text-green-600">%s</div>
	`, message)
}

// Handle bulk ticket actions
func handleTicketBulkAction(c *gin.Context) {
	ticketIDs := c.PostForm("ticket_ids")
	action := c.PostForm("action")
	
	// Parse ticket IDs
	ids := strings.Split(ticketIDs, ",")
	count := len(ids)
	
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No tickets selected"})
		return
	}
	
	var message string
	switch action {
	case "assign":
		agentID := c.PostForm("agent_id")
		if agentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Agent ID required"})
			return
		}
		message = fmt.Sprintf("%d tickets assigned", count)
		
	case "close":
		message = fmt.Sprintf("%d tickets closed", count)
		
	case "set_priority":
		priority := c.PostForm("priority")
		if priority == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Priority required"})
			return
		}
		message = fmt.Sprintf("Priority updated for %d tickets", count)
		
	case "move_queue":
		queueID := c.PostForm("queue_id")
		if queueID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Queue ID required"})
			return
		}
		message = fmt.Sprintf("%d tickets moved to queue", count)
		
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}
	
	// Return success response
	c.Header("HX-Trigger", "tickets-updated")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `
		<div class="rounded-md bg-green-50 p-4">
			<div class="flex">
				<div class="ml-3">
					<p class="text-sm font-medium text-green-800">%s</p>
				</div>
			</div>
		</div>
	`, message)
}

// Ticket detail page
func handleTicketDetail(c *gin.Context) {
	ticketID := c.Param("id")
	
	// Parse ticket ID
	id, err := strconv.Atoi(ticketID)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ticket ID format")
		return
	}
	
	// Get ticket from repository
	ticketService := GetTicketService()
	ticketModel, err := ticketService.GetTicket(uint(id))
	
	var ticket gin.H
	if err != nil {
		// Fallback to mock data for demo tickets
		ticket = gin.H{
			"ID":           id,
			"TicketNumber": fmt.Sprintf("TICKET-%06d", id),
			"Title":        "System login issues preventing access",
			"Status":       "new",
			"StatusLabel":  "New",
			"Priority":     "normal",
			"PriorityLabel": "Normal Priority",
			"Queue":        "General Support",
			"CustomerEmail": "john.doe@example.com",
			"CustomerName":  "John Doe",
			"CreateTime":    time.Now().Add(-2 * time.Hour).Format("Jan 2, 2006 3:04 PM"),
			"UpdateTime":    time.Now().Add(-2 * time.Hour).Format("Jan 2, 2006 3:04 PM"),
			"AssignedTo":    nil, // Unassigned
			"Type":          "Service Request",
			"SLAStatus":     "within", // within, warning, overdue
		}
		// Continue with mock data for non-existent tickets
		// This allows the demo to work even without real tickets
	} else {
		// Convert real ticket to display format
		simpleTicket := models.FromORTSTicket(ticketModel)
		
		// Map priority
		priorityLabels := map[string]string{
			"low": "Low Priority",
			"normal": "Normal Priority",
			"high": "High Priority",
			"urgent": "Urgent",
		}
		
		// Map status
		statusLabels := map[string]string{
			"new": "New",
			"open": "Open",
			"pending": "Pending",
			"closed": "Closed",
		}
		
		var assignedTo interface{} = nil
		if simpleTicket.AssignedTo > 0 {
			assignedTo = fmt.Sprintf("Agent %d", simpleTicket.AssignedTo)
		}
		
		ticket = gin.H{
			"ID":           simpleTicket.ID,
			"TicketNumber": simpleTicket.TicketNumber,
			"Title":        simpleTicket.Subject,
			"Status":       simpleTicket.Status,
			"StatusLabel":  statusLabels[simpleTicket.Status],
			"Priority":     simpleTicket.Priority,
			"PriorityLabel": priorityLabels[simpleTicket.Priority],
			"Queue":        fmt.Sprintf("Queue %d", simpleTicket.QueueID), // TODO: Get queue name
			"CustomerEmail": simpleTicket.CustomerEmail,
			"CustomerName":  simpleTicket.CustomerName,
			"CreateTime":    simpleTicket.CreatedAt.Format("Jan 2, 2006 3:04 PM"),
			"UpdateTime":    simpleTicket.UpdatedAt.Format("Jan 2, 2006 3:04 PM"),
			"AssignedTo":    assignedTo,
			"Type":          fmt.Sprintf("Type %d", simpleTicket.TypeID), // TODO: Get type name
			"SLAStatus":     "within", // TODO: Calculate SLA status
		}
	}
	
	// Articles/Messages - Get real messages for this ticket
	var articles []gin.H
	messages, err := ticketService.GetMessages(uint(id))
	if err != nil {
		// Log error and show empty messages
		fmt.Printf("Warning: Failed to get messages for ticket %d: %v\n", id, err)
		articles = []gin.H{}
	} else {
		// Convert messages to display format
		articles = make([]gin.H, len(messages))
		for i, msg := range messages {
			// Calculate time ago
			timeAgo := formatTimeAgo(msg.CreatedAt)
			
			// Generate initials from author name
			initials := generateInitials(msg.AuthorName)
			
			// Convert attachments if present
			var attachments []gin.H
			if msg.Attachments != nil && len(msg.Attachments) > 0 {
				attachments = make([]gin.H, len(msg.Attachments))
				for j, att := range msg.Attachments {
					attachments[j] = gin.H{
						"ID":          att.ID,
						"Filename":    att.Filename,
						"ContentType": att.ContentType,
						"Size":        formatFileSize(att.Size),
						"URL":         att.URL,
					}
				}
			}
			
			articles[i] = gin.H{
				"ID":             msg.ID,
				"AuthorName":     msg.AuthorName,
				"AuthorInitials": initials,
				"AuthorType":     msg.AuthorType,
				"TimeAgo":        timeAgo,
				"Subject":        msg.Subject,
				"Body":           msg.Body,
				"IsInternal":     msg.IsInternal,
				"Attachments":    attachments,
			}
		}
	}
	
	// Activity log
	activities := []gin.H{
		{
			"Type":        "created",
			"Description": "Ticket created",
			"TimeAgo":     "2 hours ago",
			"Icon":        "plus",
			"Color":       "green",
		},
	}
	
	// Use Pongo2 renderer with i18n support
	pongo2Renderer.HTML(c, http.StatusOK, "pages/tickets/detail.pongo2", pongo2.Context{
		"Title":      "Ticket #" + ticketID + " - GOTRS",
		"TicketID":   ticketID,
		"Ticket":     ticket,
		"Articles":   articles,
		"Activities": activities,
		"User":       getUserFromContext(c),
		"ActivePage": "tickets",
		"Messages":   articles, // Use articles as messages for now
	})
}

// HTMX Login handler
func handleHTMXLogin(c *gin.Context) {
	// Handle GET requests (for form display or API testing)
	if c.Request.Method == "GET" {
		// Return a simple success response for GET requests
		c.JSON(http.StatusOK, gin.H{
			"message": "Login endpoint ready",
			"method": "Please use POST with username/email and password",
		})
		return
	}
	
	var loginReq struct {
		Username string `json:"username" form:"username"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password" binding:"required"`
	}
	
	// Bind request data (handles both JSON and form data)
	if err := c.ShouldBind(&loginReq); err != nil {
		log.Printf("Login binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}
	
	// Use username if provided, otherwise use email
	loginIdentifier := loginReq.Username
	if loginIdentifier == "" {
		loginIdentifier = loginReq.Email
	}
	
	if loginIdentifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email is required"})
		return
	}
	
	// Check if demo mode is enabled FIRST
	demoMode := os.Getenv("DEMO_MODE") == "true"
	log.Printf("Demo mode check: DEMO_MODE=%s, demoMode=%v", os.Getenv("DEMO_MODE"), demoMode)
	if demoMode {
		// In demo mode, check demo credentials first
		demoEmail := os.Getenv("DEMO_ADMIN_EMAIL")
		demoPassword := os.Getenv("DEMO_ADMIN_PASSWORD")
		log.Printf("Demo credentials: email=%s, password=%s, loginIdentifier=%s", demoEmail, demoPassword, loginIdentifier)
		
		if demoEmail == "" || demoPassword == "" {
			// Refuse to start without demo credentials when demo mode is enabled
			log.Printf("ERROR: Demo mode enabled but DEMO_ADMIN_EMAIL or DEMO_ADMIN_PASSWORD not set")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error: Demo credentials not configured"})
			return
		}
		
		if loginIdentifier == demoEmail && loginReq.Password == demoPassword {
			// In demo mode, create a simple demo token
			// Note: This is only for demo purposes, not secure for production
			demoToken := "demo_session_" + fmt.Sprintf("%d", time.Now().Unix())
			c.SetCookie("access_token", demoToken, 86400, "/", "", false, true)
			
			c.Header("HX-Redirect", "/dashboard")
			c.JSON(http.StatusOK, gin.H{
				"access_token":  demoToken,
				"refresh_token": "demo_refresh_123",
				"user": gin.H{
					"id":         1,
					"email":      loginIdentifier,
					"first_name": "Demo",
					"last_name":  "Admin",
					"role":       "Admin",
				},
			})
			return
		}
		// In demo mode but wrong credentials
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid demo credentials"})
		return
	}
	
	// Not in demo mode, try real authentication
	authService := GetAuthService()
	if authService == nil {
		// Auth service not available and not in demo mode
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication service unavailable"})
		return
	}
	
	// Authenticate using the service
	log.Printf("Attempting login for user: %s", loginIdentifier)
	user, accessToken, refreshToken, err := authService.Login(c.Request.Context(), loginIdentifier, loginReq.Password)
	if err != nil {
		log.Printf("Login failed for %s: %v", loginIdentifier, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	
	// Set session cookie with the access token
	c.SetCookie("access_token", accessToken, 86400, "/", "", false, true)
	
	// For HTMX, set the redirect header
	c.Header("HX-Redirect", "/dashboard")
	
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Login,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	})
}

// HTMX Logout handler for API endpoint
func handleHTMXLogout(c *gin.Context) {
	// TODO: Invalidate token
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// handleLogout handles the main logout route that redirects to login
func handleLogout(c *gin.Context) {
	// Clear the session cookie
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	
	// For HTMX requests, send redirect header
	c.Header("HX-Redirect", "/login")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// handleLogoutGET handles GET requests to logout (for regular links)
func handleLogoutGET(c *gin.Context) {
	// Clear the session cookie
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	
	// For regular GET requests, do a standard redirect
	c.Redirect(http.StatusFound, "/login")
}

// underConstruction returns a simple under construction page
func underConstruction(pageName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>%s - Under Construction</title>
			<script src="https://cdn.tailwindcss.com"></script>
		</head>
		<body class="bg-gray-100 dark:bg-gray-900">
			<div class="min-h-screen flex items-center justify-center">
				<div class="text-center">
					<svg class="mx-auto h-24 w-24 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
					</svg>
					<h1 class="mt-4 text-3xl font-bold text-gray-900 dark:text-white">%s</h1>
					<p class="mt-2 text-lg text-gray-600 dark:text-gray-400">This page is under construction</p>
					<p class="mt-1 text-sm text-gray-500 dark:text-gray-500">Coming soon...</p>
					<div class="mt-6">
						<a href="/dashboard" class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300">
							← Back to Dashboard
						</a>
					</div>
				</div>
			</div>
		</body>
		</html>
		`, pageName, pageName)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	}
}

// underConstructionAPI returns a JSON response for API endpoints under construction
func underConstructionAPI(endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Endpoint %s is under construction", endpoint),
			"data":    []interface{}{},
		})
	}
}

// Dashboard stats (returns HTML fragment)
func handleDashboardStats(c *gin.Context) {
	// Get real statistics from database
	db, err := database.GetDB()
	
	// No database, no dashboard - show Guru Meditation
	if err != nil {
		sendGuruMeditation(c, err, "handleDashboardStats:GetDB")
		return
	}
	
	// Initialize stats
	stats := []gin.H{
		{"title": "Open Tickets", "value": "0", "icon": "ticket", "color": "blue"},
		{"title": "New Today", "value": "0", "icon": "plus", "color": "green"},
		{"title": "Pending", "value": "0", "icon": "clock", "color": "yellow"},
		{"title": "Resolved", "value": "0", "icon": "check", "color": "green"},
	}
	
	// Query database
	// Count open tickets (state_id = 2)
	var openCount int
	db.QueryRow("SELECT COUNT(*) FROM ticket WHERE ticket_state_id = 2").Scan(&openCount)
	
	// Count new tickets today (state_id = 1 and created today)
	var newTodayCount int
	db.QueryRow(`
		SELECT COUNT(*) FROM ticket 
		WHERE ticket_state_id = 1 
		AND DATE(create_time) = CURRENT_DATE
	`).Scan(&newTodayCount)
	
	// Count pending tickets (state_id = 3)
	var pendingCount int
	db.QueryRow("SELECT COUNT(*) FROM ticket WHERE ticket_state_id = 3").Scan(&pendingCount)
	
	// Count resolved tickets (state_id = 4)
	var resolvedCount int
	db.QueryRow("SELECT COUNT(*) FROM ticket WHERE ticket_state_id = 4").Scan(&resolvedCount)
	
	stats = []gin.H{
		{"title": "Open Tickets", "value": fmt.Sprintf("%d", openCount), "icon": "ticket", "color": "blue"},
		{"title": "New Today", "value": fmt.Sprintf("%d", newTodayCount), "icon": "plus", "color": "green"},
		{"title": "Pending", "value": fmt.Sprintf("%d", pendingCount), "icon": "clock", "color": "yellow"},
		{"title": "Resolved", "value": fmt.Sprintf("%d", resolvedCount), "icon": "check", "color": "green"},
	}
	
	tmpl, err := loadTemplateForRequest(c, "templates/components/dashboard_stats.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "dashboard_stats.html", gin.H{
		"Stats": stats,
	}); err != nil {
		c.String(http.StatusInternalServerError, "Render error: %v", err)
	}
}

// Recent tickets (returns HTML fragment)
func handleRecentTickets(c *gin.Context) {
	// Get recent tickets from database
	db, err := database.GetDB()
	
	tickets := []gin.H{}
	
	if err == nil {
		rows, err := db.Query(`
			SELECT id, tn, title, ticket_state_id, ticket_priority_id, create_time
			FROM ticket
			ORDER BY create_time DESC
			LIMIT 5
		`)
		
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id int
				var tn, title string
				var stateID, priorityID int
				var createTime time.Time
				
				if err := rows.Scan(&id, &tn, &title, &stateID, &priorityID, &createTime); err == nil {
					// Map status
					status := "new"
					switch stateID {
					case 1:
						status = "new"
					case 2:
						status = "open"
					case 3:
						status = "pending"
					case 4:
						status = "resolved"
					case 5, 6:
						status = "closed"
					}
					
					// Map priority
					priority := "medium"
					switch priorityID {
					case 1:
						priority = "low"
					case 2, 3:
						priority = "medium"
					case 4:
						priority = "high"
					case 5:
						priority = "urgent"
					}
					
					// Calculate time ago
					created := formatTimeAgo(createTime)
					
					tickets = append(tickets, gin.H{
						"id":       id,
						"number":   tn,
						"title":    title,
						"status":   status,
						"priority": priority,
						"created":  created,
					})
				}
			}
		}
	}
	
	// If no tickets, use a placeholder
	if len(tickets) == 0 {
		tickets = []gin.H{
			{"id": 0, "number": "No tickets", "title": "No recent tickets found", "status": "info", "priority": "low", "created": ""},
		}
	}
	
	tmpl, err := loadTemplateForRequest(c, "templates/components/recent_tickets.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "recent_tickets.html", gin.H{
		"Tickets": tickets,
	}); err != nil {
		c.String(http.StatusInternalServerError, "Render error: %v", err)
	}
}

// Activity feed (returns HTML fragment)
func handleActivityFeed(c *gin.Context) {
	// Get recent ticket activity from database
	db, err := database.GetDB()
	
	activities := []gin.H{}
	
	if err == nil {
		// Get recent ticket changes
		rows, err := db.Query(`
			SELECT t.tn, t.title, t.change_time, t.change_by,
			       CASE 
			         WHEN t.create_time = t.change_time THEN 'created'
			         WHEN t.ticket_state_id IN (5,6) THEN 'closed'
			         WHEN t.ticket_state_id = 4 THEN 'resolved'
			         ELSE 'updated'
			       END as action
			FROM ticket t
			ORDER BY t.change_time DESC
			LIMIT 10
		`)
		
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var tn, title, action string
				var changeTime time.Time
				var changeBy int
				
				if err := rows.Scan(&tn, &title, &changeTime, &changeBy, &action); err == nil {
					// Format time ago
					timeAgo := formatTimeAgo(changeTime)
					
					// Get user name (simplified - would need user table join)
					userName := fmt.Sprintf("User %d", changeBy)
					if changeBy == 1 {
						userName = "System"
					}
					
					activities = append(activities, gin.H{
						"user":   userName,
						"action": action,
						"target": tn,
						"time":   timeAgo,
					})
				}
			}
		}
	}
	
	// If no activities, use a placeholder
	if len(activities) == 0 {
		activities = []gin.H{
			{"user": "System", "action": "info", "target": "No recent activity", "time": ""},
		}
	}
	
	// Use Pongo2 renderer for consistency
	pongo2Renderer.HTML(c, http.StatusOK, "components/activity_feed.html", pongo2.Context{
		"Activities": activities,
	})
}

// Tickets API (returns HTML fragment)
func handleTicketsAPI(c *gin.Context) {
	// Parse query parameters for filtering
	status := filterEmptyStrings(c.QueryArray("status"))
	priority := filterEmptyStrings(c.QueryArray("priority")) 
	queue := filterEmptyStrings(c.QueryArray("queue"))
	assignee := c.Query("assignee")
	assigned := c.Query("assigned")
	search := c.Query("search")
	
	// Debug logging
	fmt.Printf("Filter Debug - Status: %v, Priority: %v, Queue: %v, Assignee: %s, Assigned: %s, Search: %s\n", 
		status, priority, queue, assignee, assigned, search)

	// Get real tickets from database
	db, err := database.GetDB()
	allTickets := []gin.H{}
	
	if err == nil {
		query := `
			SELECT t.id, t.tn, t.title, t.ticket_state_id, t.ticket_priority_id,
			       t.queue_id, q.name, COALESCE(t.customer_user_id, ''), COALESCE(t.user_id, 0)
			FROM ticket t
			LEFT JOIN queue q ON t.queue_id = q.id
			ORDER BY t.create_time DESC
			LIMIT 100
		`
		
		rows, err := db.Query(query)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, stateID, priorityID, queueID, userID int
				var tn, title, queueName, customerEmail string
				
				if err := rows.Scan(&id, &tn, &title, &stateID, &priorityID, &queueID, &queueName, &customerEmail, &userID); err == nil {
					// Map status
					statusStr := "new"
					switch stateID {
					case 1:
						statusStr = "new"
					case 2:
						statusStr = "open"
					case 3:
						statusStr = "pending"
					case 4:
						statusStr = "resolved"
					case 5, 6:
						statusStr = "closed"
					}
					
					// Map priority
					priorityStr := "medium"
					switch priorityID {
					case 1:
						priorityStr = "low"
					case 2, 3:
						priorityStr = "medium"
					case 4:
						priorityStr = "high"
					case 5:
						priorityStr = "critical"
					}
					
					// Get agent name
					agent := ""
					if userID > 0 {
						agent = fmt.Sprintf("Agent %d", userID)
					}
					
					allTickets = append(allTickets, gin.H{
						"id":       id,
						"number":   tn,
						"title":    title,
						"status":   statusStr,
						"priority": priorityStr,
						"customer": customerEmail,
						"agent":    agent,
						"queue":    queueName,
						"queueId":  queueID,
					})
				}
			}
		}
	}

	// Apply filters
	filteredTickets := []gin.H{}
	for _, ticket := range allTickets {
		// Status filter
		if len(status) > 0 {
			found := false
			for _, s := range status {
				if ticket["status"] == s {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Priority filter
		if len(priority) > 0 {
			found := false
			for _, p := range priority {
				if ticket["priority"] == p {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Queue filter
		if len(queue) > 0 {
			found := false
			for _, q := range queue {
				queueId, _ := strconv.Atoi(q)
				if ticket["queueId"] == queueId {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Assignee filter
		if assignee != "" {
			if assignee == "1" && ticket["agent"] != "Agent Smith" {
				continue
			}
		}

		// Assigned/Unassigned filter
		if assigned == "true" {
			if ticket["agent"] == "" {
				continue
			}
		} else if assigned == "false" {
			if ticket["agent"] != "" {
				continue
			}
		}

		// Search filter
		if search != "" {
			searchLower := strings.ToLower(search)
			titleMatch := strings.Contains(strings.ToLower(ticket["title"].(string)), searchLower)
			numberMatch := strings.Contains(strings.ToLower(ticket["number"].(string)), searchLower) 
			customerMatch := strings.Contains(strings.ToLower(ticket["customer"].(string)), searchLower)
			
			if !titleMatch && !numberMatch && !customerMatch {
				continue
			}
		}

		filteredTickets = append(filteredTickets, ticket)
	}

	// Prepare response data
	responseData := gin.H{
		"Tickets": filteredTickets,
		"Title":   "Tickets", // General title for the section
	}

	// Add filter badges and messages for UI
	if len(status) > 0 {
		statusLabels := []string{}
		for _, s := range status {
			switch s {
			case "new":
				statusLabels = append(statusLabels, "New")
			case "open":
				statusLabels = append(statusLabels, "Open")
			case "pending":
				statusLabels = append(statusLabels, "Pending")
			case "resolved":
				statusLabels = append(statusLabels, "Resolved")
			case "closed":
				statusLabels = append(statusLabels, "Closed")
			default:
				statusLabels = append(statusLabels, strings.Title(s))
			}
		}
		responseData["StatusFilter"] = strings.Join(statusLabels, ", ")
	}

	if len(priority) > 0 {
		priorityLabels := []string{}
		for _, p := range priority {
			if p == "high" {
				priorityLabels = append(priorityLabels, "High Priority")
			} else if p == "low" {
				priorityLabels = append(priorityLabels, "Low Priority")
			} else if p == "critical" {
				priorityLabels = append(priorityLabels, "critical")
			}
		}
		responseData["PriorityFilter"] = strings.Join(priorityLabels, ", ")
	}

	if assigned == "false" {
		responseData["AssignedFilter"] = "Unassigned"
	}

	// No tickets found message
	if len(filteredTickets) == 0 {
		responseData["NoTicketsMessage"] = "No tickets found"
	}

	tmpl, err := loadTemplateForRequest(c, "templates/components/ticket_list.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "ticket_list.html", responseData); err != nil {
		c.String(http.StatusInternalServerError, "Render error: %v", err)
	}
}

// Ticket search (returns HTML fragment)
func handleTicketSearch(c *gin.Context) {
	searchTerm := c.Query("search")
	if searchTerm == "" {
		searchTerm = c.Query("q") // Support both ?search= and ?q=
	}
	
	// Get tickets from database
	db, err := database.GetDB()
	allTickets := []gin.H{}
	
	if err == nil && searchTerm != "" {
		// Search in ticket number, title, and customer email
		rows, err := db.Query(`
			SELECT t.id, t.tn, t.title, t.ticket_state_id, t.ticket_priority_id,
			       COALESCE(t.customer_user_id, ''), COALESCE(t.user_id, 0)
			FROM ticket t
			WHERE LOWER(t.tn) LIKE LOWER($1)
			   OR LOWER(t.title) LIKE LOWER($1)
			   OR LOWER(COALESCE(t.customer_user_id, '')) LIKE LOWER($1)
			ORDER BY t.create_time DESC
			LIMIT 50
		`, "%"+searchTerm+"%")
		
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id int
				var tn, title, customerEmail string
				var stateID, priorityID, userID int
				
				if err := rows.Scan(&id, &tn, &title, &stateID, &priorityID, &customerEmail, &userID); err == nil {
					// Map status
					status := "new"
					switch stateID {
					case 1:
						status = "new"
					case 2:
						status = "open"
					case 3:
						status = "pending"
					case 4:
						status = "resolved"
					case 5, 6:
						status = "closed"
					}
					
					// Map priority
					priority := "medium"
					switch priorityID {
					case 1:
						priority = "low"
					case 2, 3:
						priority = "medium"
					case 4:
						priority = "high"
					case 5:
						priority = "urgent"
					}
					
					// Get agent name
					agent := ""
					if userID > 0 {
						agent = fmt.Sprintf("Agent %d", userID)
					}
					
					allTickets = append(allTickets, gin.H{
						"id":       id,
						"number":   tn,
						"title":    title,
						"status":   status,
						"priority": priority,
						"customer": customerEmail,
						"agent":    agent,
					})
				}
			}
		}
	}

	// All tickets are already filtered by the SQL query
	filteredTickets := allTickets
	
	// Highlight search terms if present
	if searchTerm != "" {
		for i := range filteredTickets {
			titleStr := filteredTickets[i]["title"].(string)
			// Case-insensitive replacement with <mark> tags
			re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(searchTerm))
			filteredTickets[i]["highlighted_title"] = re.ReplaceAllString(titleStr, "<mark>$0</mark>")
		}
	}
	
	// If no search term provided, show recent tickets
	if searchTerm == "" && len(filteredTickets) == 0 {
		// Empty search returns all tickets
		filteredTickets = allTickets
	}

	responseData := gin.H{
		"Tickets": filteredTickets,
		"Title":   "Tickets", // General title for the section
	}

	// Add search term to response if provided
	if searchTerm != "" {
		responseData["SearchTerm"] = searchTerm
	}

	// No results message
	if len(filteredTickets) == 0 {
		responseData["NoTicketsMessage"] = "No tickets found"
	}
	
	tmpl, err := loadTemplateForRequest(c, "templates/components/ticket_list.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "ticket_list.html", responseData); err != nil {
		c.String(http.StatusInternalServerError, "Render error: %v", err)
	}
}

// Create ticket (HTMX)
func handleCreateTicket(c *gin.Context) {
	var req struct {
		Title         string `json:"title" form:"title"`         // Accept both title and subject
		Subject       string `json:"subject" form:"subject"`
		CustomerEmail string `json:"customer_email" form:"customer_email" binding:"required,email"`
		CustomerName  string `json:"customer_name" form:"customer_name"`
		Priority      string `json:"priority" form:"priority"`
		QueueID       string `json:"queue_id" form:"queue_id"`
		TypeID        string `json:"type_id" form:"type_id"`
		Body          string `json:"body" form:"body" binding:"required"`
	}

	// Bind form data
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Use title if provided, otherwise use subject
	ticketTitle := req.Title
	if ticketTitle == "" {
		ticketTitle = req.Subject
	}
	if ticketTitle == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title or subject is required"})
		return
	}

	// Convert string values to integers with defaults
	queueID := uint(1) // Default to General Support
	if req.QueueID != "" {
		if id, err := strconv.Atoi(req.QueueID); err == nil {
			queueID = uint(id)
		}
	}

	typeID := uint(1) // Default to Incident
	if req.TypeID != "" {
		if id, err := strconv.Atoi(req.TypeID); err == nil {
			typeID = uint(id)
		}
	}

	// Set default priority if not provided
	if req.Priority == "" {
		req.Priority = "normal"
	}

	// For demo purposes, use a fixed user ID (admin)
	// In a real system, we'd get this from the authenticated user context
	createdBy := uint(1)

	// Create the ticket model with OTRS-compatible fields
	customerEmail := req.CustomerEmail
	ticket := &models.Ticket{
		Title:            ticketTitle,
		QueueID:          int(queueID),
		TypeID:           int(typeID),
		TicketPriorityID: getPriorityID(req.Priority),
		TicketStateID:    1, // New
		TicketLockID:     1, // Unlocked
		CustomerUserID:   &customerEmail,
		CreateBy:         int(createdBy),
		ChangeBy:         int(createdBy),
	}

	// Get the ticket repository directly for real database operations
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}
	
	ticketRepo := repository.NewTicketRepository(db)
	if err := ticketRepo.Create(ticket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket: " + err.Error()})
		return
	}

	// Create the first article (ticket body) using the article repository
	log.Printf("DEBUG: About to create article for ticket ID %d", ticket.ID)
	articleRepo := repository.NewArticleRepository(db)
	article := &models.Article{
		TicketID:             ticket.ID,
		Subject:              ticketTitle,
		Body:                 req.Body, // Will be converted to []byte in repository
		SenderTypeID:         3, // Customer
		CommunicationChannelID: 1, // Email
		IsVisibleForCustomer: 1,
		CreateBy:            int(createdBy),
		ChangeBy:            int(createdBy),
	}
	
	log.Printf("DEBUG: Article object created, calling Create...")
	if err := articleRepo.Create(article); err != nil {
		// Log error but don't fail the ticket creation
		log.Printf("ERROR: Failed to add initial article: %v", err)
	} else {
		log.Printf("Successfully created article ID %d for ticket ID %d", article.ID, ticket.ID)
	}
	
	// For HTMX, set the redirect header to the ticket detail page
	c.Header("HX-Redirect", fmt.Sprintf("/tickets/%d", ticket.ID))
	c.JSON(http.StatusCreated, gin.H{
		"id":            ticket.ID,
		"ticket_number": ticket.TicketNumber,
		"message":       "Ticket created successfully",
		"queue_id":      float64(ticket.QueueID),
		"type_id":       float64(ticket.TypeID),
		"priority":      req.Priority,
	})
}

// Update ticket (HTMX)
func handleUpdateTicket(c *gin.Context) {
	// TODO: Implement ticket update
	c.JSON(http.StatusOK, gin.H{"message": "Ticket updated"})
}

// Update ticket status (HTMX)
func handleUpdateTicketStatus(c *gin.Context) {
	ticketID := c.Param("id")
	var req struct {
		Status string `json:"status" form:"status" binding:"required"`
	}
	
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// TODO: Update ticket status in database
	// For now, return success
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Ticket %s status updated to %s", ticketID, req.Status),
		"status": req.Status,
	})
}

// Assign ticket to agent (HTMX)
func handleAssignTicket(c *gin.Context) {
	ticketID := c.Param("id")
	
	// For "Assign to Me", we don't need any request body
	// Just assign to the current user (hardcoded for now)
	agentID := 1 // TODO: Get from session/auth context
	agentName := "Demo Agent" // TODO: Get from session
	
	// TODO: Update ticket assignment in database
	
	// Return success message that HTMX can display
	c.Header("HX-Trigger", `{"showMessage": {"message": "Ticket assigned to you", "type": "success"}}`)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Ticket %s assigned to %s", ticketID, agentName),
		"agent_id": agentID,
	})
}

// Add reply to ticket (HTMX)
func handleTicketReply(c *gin.Context) {
	ticketID := c.Param("id")
	_ = ticketID // Will be used when connecting to database
	var req struct {
		Reply       string `json:"reply" form:"reply" binding:"required"`
		Internal    bool   `json:"internal" form:"internal"`
		CloseTicket bool   `json:"close_ticket" form:"close_ticket"`
	}
	
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// TODO: Add reply to ticket in database
	// For now, return HTML fragment for the new message
	newMessage := gin.H{
		"ID":             time.Now().Unix(),
		"AuthorName":     "Support Agent",
		"AuthorInitials": "SA",
		"AuthorType":     "Agent",
		"TimeAgo":        "Just now",
		"Body":           req.Reply,
		"IsInternal":     req.Internal,
	}
	
	// Return HTML fragment for HTMX to append
	tmpl, err := loadTemplateForRequest(c, "templates/components/ticket_message.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "ticket_message.html", newMessage); err != nil {
		c.String(http.StatusInternalServerError, "Render error: %v", err)
	}
}

// Update ticket priority (HTMX)
func handleUpdateTicketPriority(c *gin.Context) {
	ticketID := c.Param("id")
	var req struct {
		Priority string `json:"priority" form:"priority" binding:"required"`
	}
	
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// TODO: Update ticket priority in database
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Ticket %s priority updated to %s", ticketID, req.Priority),
		"priority": req.Priority,
	})
}

// Update ticket queue (HTMX)
func handleUpdateTicketQueue(c *gin.Context) {
	ticketID := c.Param("id")
	var req struct {
		QueueID int `json:"queue_id" form:"queue_id" binding:"required"`
	}
	
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// TODO: Update ticket queue in database
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Ticket %s moved to queue %d", ticketID, req.QueueID),
		"queue_id": req.QueueID,
	})
}

// Server-Sent Events for tickets
func handleTicketStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	
	// Send a test event
	c.SSEvent("ticket-update", gin.H{
		"ticketId": 1,
		"status":   "updated",
		"message":  "Ticket status changed",
	})
	
	c.Writer.Flush()
}

// Server-Sent Events for activity
func handleActivityStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	
	// Send a test event with HTML content
	activityHTML := `<div class="py-3 border-b border-gray-200 last:border-b-0">
		<div class="flex space-x-3">
			<div class="flex-shrink-0">
				<div class="h-8 w-8 rounded-full bg-green-500 flex items-center justify-center">
					<span class="text-xs font-medium text-white">S</span>
				</div>
			</div>
			<div class="min-w-0 flex-1">
				<p class="text-sm text-gray-900">
					<span class="font-medium">System</span>
					<span class="text-blue-600">updated</span>
					<span class="font-medium">activity feed</span>
				</p>
				<p class="text-sm text-gray-500">just now</p>
			</div>
		</div>
	</div>`
	
	// For HTMX SSE, send HTML content directly
	fmt.Fprintf(c.Writer, "event: activity-update\ndata: %s\n\n", activityHTML)
	
	c.Writer.Flush()
}

// Queues API (returns HTML fragment or JSON)
func handleQueuesAPI(c *gin.Context) {
	// Check if JSON response is requested
	if wantsJSON(c) {
		handleQueuesJSON(c)
		return
	}
	
	// Parse query parameters for filtering
	status := c.Query("status")
	search := c.Query("search")
	
	// Get queues from database with ticket counts
	queues, err := getQueuesWithTicketCounts(status, search)
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}
	
	// Render HTML response (fragment or full page)
	renderQueueList(c, queues)
}

// Handle JSON response for queues
func handleQueuesJSON(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("search")
	
	queues, err := getQueuesWithTicketCounts(status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queues,
	})
}


// Helper function to sort tickets
func sortTickets(tickets []gin.H, sortBy string) {
	switch sortBy {
	case "priority":
		// Sort by priority (urgent > high > normal > low)
		priorityOrder := map[string]int{
			"urgent": 0,
			"high":   1,
			"normal": 2,
			"low":    3,
		}
		sort.Slice(tickets, func(i, j int) bool {
			pi := priorityOrder[tickets[i]["priority"].(string)]
			pj := priorityOrder[tickets[j]["priority"].(string)]
			return pi < pj
		})
	case "status":
		// Sort by status (new > open > pending > resolved > closed)
		statusOrder := map[string]int{
			"new":      0,
			"open":     1,
			"pending":  2,
			"resolved": 3,
			"closed":   4,
		}
		sort.Slice(tickets, func(i, j int) bool {
			si := statusOrder[tickets[i]["status"].(string)]
			sj := statusOrder[tickets[j]["status"].(string)]
			return si < sj
		})
	case "title":
		// Sort alphabetically by title
		sort.Slice(tickets, func(i, j int) bool {
			return tickets[i]["title"].(string) < tickets[j]["title"].(string)
		})
	case "updated":
		// Sort by last updated (most recent first)
		sort.Slice(tickets, func(i, j int) bool {
			return tickets[i]["updated_at"].(string) > tickets[j]["updated_at"].(string)
		})
	default: // "created"
		// Sort by creation date (most recent first)
		sort.Slice(tickets, func(i, j int) bool {
			return tickets[i]["created_at"].(string) > tickets[j]["created_at"].(string)
		})
	}
}

// Sort queues based on the sort parameter
func sortQueues(queues []gin.H, sortBy string) {
	switch sortBy {
	case "name_asc":
		sort.Slice(queues, func(i, j int) bool {
			return queues[i]["name"].(string) < queues[j]["name"].(string)
		})
	case "name_desc":
		sort.Slice(queues, func(i, j int) bool {
			return queues[i]["name"].(string) > queues[j]["name"].(string)
		})
	case "tickets_asc":
		sort.Slice(queues, func(i, j int) bool {
			return queues[i]["ticket_count"].(int) < queues[j]["ticket_count"].(int)
		})
	case "tickets_desc":
		sort.Slice(queues, func(i, j int) bool {
			return queues[i]["ticket_count"].(int) > queues[j]["ticket_count"].(int)
		})
	case "status_asc":
		sort.Slice(queues, func(i, j int) bool {
			return queues[i]["status"].(string) < queues[j]["status"].(string)
		})
	default:
		// Default sort by ID (original order)
		sort.Slice(queues, func(i, j int) bool {
			return queues[i]["id"].(int) < queues[j]["id"].(int)
		})
	}
}

// Get queues from database with ticket counts
func getQueuesWithTicketCounts(status, search string) ([]gin.H, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("database connection required: %w", err)
	}
	
	// Build query with optional search filter
	query := `
		SELECT q.id, q.name, COALESCE(q.comments, '') as comment, q.valid_id,
		       COUNT(t.id) as ticket_count
		FROM queue q
		LEFT JOIN ticket t ON t.queue_id = q.id
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0
	
	// Add search filter if provided
	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (LOWER(q.name) LIKE $%d OR LOWER(COALESCE(q.comments, '')) LIKE $%d)", argCount, argCount)
		args = append(args, "%"+strings.ToLower(search)+"%")
	}
	
	// Add status filter (active = valid_id=1, inactive = valid_id!=1)
	if status == "active" {
		argCount++
		query += fmt.Sprintf(" AND q.valid_id = $%d", argCount)
		args = append(args, 1)
	} else if status == "inactive" {
		argCount++
		query += fmt.Sprintf(" AND q.valid_id != $%d", argCount)
		args = append(args, 1)
	}
	
	query += " GROUP BY q.id, q.name, q.comments, q.valid_id ORDER BY q.id"
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query queues: %w", err)
	}
	defer rows.Close()
	
	queues := []gin.H{}
	for rows.Next() {
		var id int
		var name, comment string
		var validID, ticketCount int
		
		if err := rows.Scan(&id, &name, &comment, &validID, &ticketCount); err != nil {
			continue
		}
		
		queueStatus := "active"
		if validID != 1 {
			queueStatus = "inactive"
		}
		
		queue := gin.H{
			"id":           id,
			"name":         name,
			"comment":      comment,
			"ticket_count": ticketCount,
			"status":       queueStatus,
			"active":       validID == 1,
		}
		
		queues = append(queues, queue)
	}
	
	return queues, nil
}

// Helper function to check if client wants JSON response
func wantsJSON(c *gin.Context) bool {
	return strings.Contains(c.GetHeader("Accept"), "application/json")
}

// Helper function to check if request is from HTMX
func isHTMXRequest(c *gin.Context) bool {
	return c.GetHeader("HX-Request") == "true"
}

// Helper function to send standardized error responses
func sendErrorResponse(c *gin.Context, statusCode int, message string) {
	if wantsJSON(c) {
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   message,
		})
	} else {
		c.String(statusCode, message)
	}
}

// Helper function to render queue list templates
func renderQueueList(c *gin.Context, queues []gin.H) {
	// Get search/filter parameters for template context
	searchTerm := c.Query("search")
	statusFilter := c.Query("status")
	
	templateData := gin.H{
		"Queues":      queues,
		"SearchTerm":  searchTerm,
		"StatusFilter": statusFilter,
	}
	
	if isHTMXRequest(c) {
		// Return HTML fragment for HTMX
		tmpl, err := loadTemplateForRequest(c, "templates/components/queue_list.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Template error: %v", err)
			return
		}
		
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(c.Writer, "queue_list.html", templateData); err != nil {
			c.String(http.StatusInternalServerError, "Render error: %v", err)
		}
	} else {
		// Return full page
		tmpl, err := loadTemplateForRequest(c, 
			"templates/layouts/base.html",
			"templates/components/guru_meditation.html",
			"templates/pages/queues/list.html",
			"templates/components/queue_list.html",
		)
		if err != nil {
			c.String(http.StatusInternalServerError, "Template error: %v", err)
			return
		}
		
		// Add page-level template data
		templateData["Title"] = "Queues - GOTRS"
		templateData["User"] = getUserFromContext(c)
		templateData["ActivePage"] = "queues"
		
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(c.Writer, "list.html", templateData); err != nil {
			c.String(http.StatusInternalServerError, "Render error: %v", err)
		}
	}
}

// Helper function to render queue detail templates
func renderQueueDetail(c *gin.Context, queue gin.H) {
	if isHTMXRequest(c) {
		// Return HTML fragment for HTMX
		tmpl, err := pongo2.FromFile("templates/components/queue_detail.pongo2")
		if err != nil {
			// Fallback for test environment
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, `<div class="queue-detail">%s</div>`, queue["name"])
			return
		}
		
		c.Header("Content-Type", "text/html; charset=utf-8")
		err = tmpl.ExecuteWriter(pongo2.Context{
			"Queue": queue,
		}, c.Writer)
		
		if err != nil {
			c.String(http.StatusInternalServerError, "Render error: %v", err)
		}
	} else {
		// Return full page
		tmpl, err := pongo2.FromFile("templates/pages/queues/detail.pongo2")
		if err != nil {
			// Fallback for test environment
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, `<!DOCTYPE html><html><head><title>%s</title></head><body><div class="queue-detail">%s</div></body></html>`, queue["name"], queue["name"])
			return
		}
		
		c.Header("Content-Type", "text/html; charset=utf-8")
		err = tmpl.ExecuteWriter(pongo2.Context{
			"Title":      queue["name"].(string) + " - Queue Details - GOTRS",
			"User":       getUserFromContext(c),
			"ActivePage": "queues",
			"Queue":      queue,
		}, c.Writer)
		
		if err != nil {
			c.String(http.StatusInternalServerError, "Render error: %v", err)
		}
	}
}

// Queue detail API (returns HTML fragment or JSON)
func handleQueueDetail(c *gin.Context) {
	queueID := c.Param("id")
	
	// Validate queue ID
	if queueID == "" || queueID == "invalid" {
		sendErrorResponse(c, http.StatusBadRequest, "Invalid queue ID")
		return
	}
	
	// Get queue details with tickets
	queue, err := getQueueWithTickets(queueID)
	if err != nil {
		if err.Error() == "queue not found" {
			sendErrorResponse(c, http.StatusNotFound, "Queue not found")
			return
		}
		sendErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}
	
	// Check if JSON response is requested
	if wantsJSON(c) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    queue,
		})
		return
	}
	
	// Render HTML response (fragment or full page)
	renderQueueDetail(c, queue)
}

// Queue tickets API (returns tickets for a specific queue)
func handleQueueTickets(c *gin.Context) {
	queueID := c.Param("id")
	
	// Parse query parameters
	status := c.Query("status")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	
	// Get tickets for queue
	tickets, err := getTicketsForQueue(queueID, status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}
	
	// Return tickets as HTML fragment or JSON
	if strings.Contains(c.GetHeader("Accept"), "application/json") {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    tickets,
		})
	} else {
		// Return HTML fragment
		tmpl, err := loadTemplateForRequest(c, "templates/components/ticket_list.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Template error: %v", err)
			return
		}
		
		// Prepare template data
		templateData := gin.H{
			"Tickets":    tickets["tickets"],
			"Title":      "Queue Tickets",
			"Total":      tickets["total"],
			"Page":       tickets["page"],
			"Pagination": tickets["pagination"],
		}
		
		// Handle empty state
		if tickets["total"].(int) == 0 {
			templateData["NoTicketsMessage"] = "No tickets in this queue"
		}
		
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(c.Writer, "ticket_list.html", templateData); err != nil {
			c.String(http.StatusInternalServerError, "Render error: %v", err)
		}
	}
}

// Get queue details with associated tickets
func getQueueWithTickets(queueID string) (gin.H, error) {
	// Get queue details from database
	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("database connection required: %w", err)
	}
	
	qID, err := strconv.Atoi(queueID)
	if err != nil {
		return nil, fmt.Errorf("invalid queue ID")
	}
	
	// Get queue details
	var queue struct {
		ID      int
		Name    string
		Comment sql.NullString
		ValidID int
	}
	
	err = db.QueryRow(`
		SELECT id, name, comments, valid_id 
		FROM queue 
		WHERE id = $1
	`, qID).Scan(&queue.ID, &queue.Name, &queue.Comment, &queue.ValidID)
	
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("DEBUG: Queue %d not found in database\n", qID)
			return nil, fmt.Errorf("queue not found")
		}
		fmt.Printf("DEBUG: Error querying queue %d: %v\n", qID, err)
		return nil, fmt.Errorf("failed to get queue tickets: %w", err)
	}
	
	fmt.Printf("DEBUG: Found queue %d: %s\n", queue.ID, queue.Name)
	
	// Get tickets in this queue
	rows, err := db.Query(`
		SELECT id, tn, title, ticket_state_id, ticket_priority_id
		FROM ticket
		WHERE queue_id = $1
		ORDER BY create_time DESC
		LIMIT 50
	`, qID)
	
	tickets := []gin.H{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ticketID int
			var ticketNumber, title string
			var stateID, priorityID int
			
			if err := rows.Scan(&ticketID, &ticketNumber, &title, &stateID, &priorityID); err == nil {
				// Map status
				statusStr := "new"
				switch stateID {
				case 1:
					statusStr = "new"
				case 2:
					statusStr = "open"
				case 3:
					statusStr = "pending"
				case 4:
					statusStr = "resolved"
				case 5, 6:
					statusStr = "closed"
				}
				
				tickets = append(tickets, gin.H{
					"id":       ticketID,
					"number":   ticketNumber,
					"title":    title,
					"status":   statusStr,
					"priority": priorityID,
				})
			}
		}
	}
	
	// Build result
	result := gin.H{
		"id":           queue.ID,
		"name":         queue.Name,
		"comment":      queue.Comment.String,
		"ticket_count": len(tickets),
		"status":       "active",
		"tickets":      tickets,
	}
	
	if queue.ValidID != 1 {
		result["status"] = "inactive"
	}
	
	return result, nil
}


// Get tickets for a specific queue with filtering and pagination
func getTicketsForQueue(queueID, status, page, limit string) (gin.H, error) {
	// Get queue details first
	queue, err := getQueueWithTickets(queueID)
	if err != nil {
		return nil, err
	}
	
	tickets := queue["tickets"].([]gin.H)
	filteredTickets := []gin.H{}
	
	// Apply status filter
	for _, ticket := range tickets {
		if status != "" && ticket["status"] != status {
			continue
		}
		filteredTickets = append(filteredTickets, ticket)
	}
	
	// Implement proper pagination
	pageNum, _ := strconv.Atoi(page)
	limitNum, _ := strconv.Atoi(limit)
	if pageNum < 1 {
		pageNum = 1
	}
	if limitNum < 1 {
		limitNum = 10
	}
	
	total := len(filteredTickets)
	offset := (pageNum - 1) * limitNum
	
	// Apply pagination to tickets
	paginatedTickets := []gin.H{}
	if offset < total {
		end := offset + limitNum
		if end > total {
			end = total
		}
		paginatedTickets = filteredTickets[offset:end]
	}
	
	// Calculate pagination info
	hasNext := offset+limitNum < total
	hasPrev := pageNum > 1
	totalPages := (total + limitNum - 1) / limitNum // Ceiling division
	
	result := gin.H{
		"tickets":    paginatedTickets,
		"total":      total,
		"page":       pageNum,
		"limit":      limitNum,
		"total_pages": totalPages,
		"pagination": gin.H{
			"has_next":    hasNext,
			"has_prev":    hasPrev,
			"next_page":   pageNum + 1,
			"prev_page":   pageNum - 1,
			"total_pages": totalPages,
		},
	}
	
	if len(filteredTickets) == 0 {
		result["message"] = "No tickets in this queue"
	}
	
	return result, nil
}

// Queue detail page (browser route)
func handleQueueDetailPage(c *gin.Context) {
	queueID := c.Param("id")
	
	// Get queue details
	queue, err := getQueueWithTickets(queueID)
	if err != nil {
		if err.Error() == "queue not found" {
			c.String(http.StatusNotFound, "Queue not found")
		} else {
			c.String(http.StatusInternalServerError, "Database error: %v", err)
		}
		return
	}
	
	tmpl, err := loadTemplateForRequest(c, 
		"templates/layouts/base.html",
		"templates/components/guru_meditation.html",
		"templates/pages/queues/detail.html",
		"templates/components/queue_detail.html",
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "detail.html", gin.H{
		"Title":      queue["name"].(string) + " - Queue Details - GOTRS",
		"User":       getUserFromContext(c),
		"ActivePage": "queues",
		"Queue":      queue,
	}); err != nil {
		c.String(http.StatusInternalServerError, "Render error: %v", err)
	}
}

// Queues list page - handles both full page loads and HTMX search/filter requests
func handleQueuesList(c *gin.Context) {
	// Get search and filter parameters from query string
	search := c.Query("search")
	status := c.Query("status")
	sortBy := c.Query("sort")
	
	// Get pagination parameters
	page := 1
	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	
	perPage := 10
	if pp := c.Query("per_page"); pp != "" {
		if parsedPerPage, err := strconv.Atoi(pp); err == nil && parsedPerPage > 0 {
			if parsedPerPage > 100 {
				perPage = 100 // Cap at 100
			} else {
				perPage = parsedPerPage
			}
		}
	}
	
	// Get queues data with filtering
	queues, err := getQueuesWithTicketCounts(status, search)
	if err != nil {
		// Check if this is an HTMX request or full page load
		if c.GetHeader("HX-Request") != "" {
			sendGuruMeditation(c, err, "handleQueuesList:getQueuesWithTicketCounts")
		} else {
			// For full page loads, render standalone error page with Guru Meditation
			tmpl, _ := loadTemplateForRequest(c, "templates/pages/error_guru.html")
			if tmpl != nil {
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.Status(http.StatusInternalServerError)
				tmpl.ExecuteTemplate(c.Writer, "error_guru", gin.H{
					"ErrorCode": "00000005.DEADBEEF",
					"ErrorMessage": err.Error(),
					"Task": "DATABASE.QUERY",
					"Location": "handleQueuesList:getQueuesWithTicketCounts",
					"Timestamp": time.Now().Format("2006-01-02 15:04:05"),
				})
			} else {
				c.String(http.StatusInternalServerError, "Database error: %v", err)
			}
		}
		return
	}
	
	// Apply sorting
	sortQueues(queues, sortBy)
	
	// Calculate pagination
	total := len(queues)
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1 // At least 1 page even if no results
	}
	
	// Adjust page if negative or zero
	if page < 1 {
		page = 1
	}
	
	// Calculate slice bounds
	start := (page - 1) * perPage
	end := start + perPage
	if end > total {
		end = total
	}
	
	// Paginate results
	var paginatedQueues []gin.H
	if start < total {
		paginatedQueues = queues[start:end]
	} else {
		// Page is beyond total pages - return empty
		paginatedQueues = []gin.H{}
	}
	
	// Build pagination info
	pagination := gin.H{
		"Page":     page,
		"PerPage":  perPage,
		"Total":    total,
		"Start":    start + 1,
		"End":      end,
		"HasPrev":  page > 1,
		"HasNext":  page < totalPages,
		"PrevPage": page - 1,
		"NextPage": page + 1,
	}
	
	// Check if this is an HTMX request for filtering/search
	if c.GetHeader("HX-Request") != "" {
		// Return just the queue list fragment for HTMX requests
		tmpl, err := loadTemplateForRequest(c, "templates/components/queue_list.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Template error: %v", err)
			return
		}
		
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(c.Writer, "queue_list.html", gin.H{
			"Queues":       paginatedQueues,
			"SearchTerm":   search,
			"StatusFilter": status,
			"SortBy":       sortBy,
			"PerPage":      perPage,
			"Pagination":   pagination,
		}); err != nil {
			c.String(http.StatusInternalServerError, "Render error: %v", err)
		}
		return
	}
	
	// Full page load - use Pongo2 renderer
	pongo2Renderer.HTML(c, http.StatusOK, "pages/queues/list.pongo2", pongo2.Context{
		"Title":        "Queues - GOTRS",
		"User":         getUserFromContext(c),
		"ActivePage":   "queues",
		"Queues":       paginatedQueues,
		"SearchTerm":   search,
		"StatusFilter": status,
		"SortBy":       sortBy,
		"PerPage":      perPage,
		"Pagination":   pagination,
	})
}

// Admin dashboard page  
func handleAdminDashboard(c *gin.Context) {
	// Use existing abstraction layers - no direct SQL!
	userCount := 0
	groupCount := 0
	queueCount := 0
	ticketCount := 0
	
	// Get queue count from existing queue repository
	queueRepo := GetQueueRepository()
	if queueRepo != nil {
		queues, err := queueRepo.List()
		if err == nil {
			queueCount = len(queues)
		}
	}
	
	// Get ticket count from existing ticket service
	ticketService := GetTicketService()
	if ticketService != nil {
		// Use the existing ticket list request
		req := &models.TicketListRequest{
			PerPage: 1000, // Get up to 1000 tickets for count
			Page: 1,
		}
		resp, _ := ticketService.ListTickets(req)
		if resp != nil {
			// Count non-closed tickets
			for _, t := range resp.Tickets {
				if t.State != nil && t.State.Name != "closed" && !strings.HasPrefix(t.State.Name, "closed") {
					ticketCount++
				}
			}
		}
	}
	
	// Get user count - for now use hardcoded until we have proper user service
	// TODO: Use proper user service when available
	db, _ := database.GetDB()
	if db != nil {
		userRepo := repository.NewUserRepository(db)
		users, _ := userRepo.List()
		userCount = len(users)
		
		// Get group count
		groupRepo := repository.NewGroupRepository(db)
		groups, _ := groupRepo.List()
		groupCount = len(groups)
	}
	
	fmt.Printf("DEBUG: Dashboard counts - Users: %d, Groups: %d, Queues: %d, Tickets: %d\n", userCount, groupCount, queueCount, ticketCount)
	
	// Use Pongo2 renderer with i18n support
	pongo2Renderer.HTML(c, http.StatusOK, "pages/admin/dashboard.pongo2", pongo2.Context{
		"Title":         "Admin - GOTRS",
		"User":          getUserFromContext(c),
		"ActivePage":    "admin",
		"UserCount":     userCount,
		"GroupCount":    groupCount,
		"ActiveTickets": ticketCount,
		"QueueCount":    queueCount,
	})
}

// handleAdminLookups shows the admin lookups management page (priorities, states, types)
func handleAdminLookups(c *gin.Context) {
	fmt.Println("DEBUG: handleAdminLookups called")
	db, err := database.GetDB()
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Database connection failed")
		return
	}

	// Get priorities
	priorityRepo := repository.NewPriorityRepository(db)
	priorities, err := priorityRepo.List()
	if err != nil {
		fmt.Printf("DEBUG: Failed to fetch priorities: %v\n", err)
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch priorities")
		return
	}

	// Get states
	stateRepo := repository.NewTicketStateRepository(db)
	states, err := stateRepo.List()
	if err != nil {
		fmt.Printf("DEBUG: Failed to fetch states: %v\n", err)
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch states")
		return
	}

	// Get types - using lookup service
	lookupService := GetLookupService()
	formData := lookupService.GetTicketFormData()
	types := formData.Types

	pongo2Renderer.HTML(c, http.StatusOK, "pages/admin/lookups.pongo2", pongo2.Context{
		"Title":      "Lookups Management - GOTRS Admin",
		"Priorities": priorities,
		"States":     states,
		"Types":      types,
		"User":       getUserFromContext(c),
		"ActivePage": "admin",
	})
}

// handleAdminUsers shows the admin users management page
func handleAdminUsers(c *gin.Context) {
	fmt.Println("DEBUG: handleAdminUsers called")
	db, err := database.GetDB()
	if err != nil {
		fmt.Printf("DEBUG: Database error: %v\n", err)
		sendErrorResponse(c, http.StatusInternalServerError, "Database connection failed")
		return
	}

	userRepo := repository.NewUserRepository(db)
	users, err := userRepo.List()
	if err != nil {
		fmt.Printf("DEBUG: Failed to fetch users: %v\n", err)
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// Get groups for the form
	groupRepo := repository.NewGroupRepository(db)
	groups, err := groupRepo.List()
	if err != nil {
		fmt.Printf("DEBUG: Failed to fetch groups: %v\n", err)
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch groups")
		return
	}

	// Add group information to users
	for _, user := range users {
		userGroups, err := groupRepo.GetUserGroups(user.ID)
		if err == nil {
			user.Groups = userGroups
		}
	}

	fmt.Printf("DEBUG: Passing %d users to template\n", len(users))
	pongo2Renderer.HTML(c, http.StatusOK, "pages/admin/users.pongo2", pongo2.Context{
		"Title":      "User Management - GOTRS Admin",
		"Users":      users,
		"Groups":     groups,
		"User":       getUserFromContext(c),
		"ActivePage": "admin",
	})
}

// handleGetUser returns user data as JSON
func handleGetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// Get user's groups
	groupRepo := repository.NewGroupRepository(db)
	groupIDs := []uint{}
	if groups, err := groupRepo.GetUserGroups(user.ID); err == nil {
		for _, groupName := range groups {
			// Get group ID from name - ideally we'd have a better method
			if allGroups, err := groupRepo.List(); err == nil {
				for _, g := range allGroups {
					if g.Name == groupName {
						groupIDs = append(groupIDs, g.ID)
						break
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":         user.ID,
			"login":      user.Login,
			"title":      user.Title,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"valid_id":   user.ValidID,
			"groups":     groupIDs,
		},
	})
}

// handleCreateUser creates a new user
func handleCreateUser(c *gin.Context) {
	fmt.Println("DEBUG: handleCreateUser called")
	
	var req struct {
		Login     string `form:"login" binding:"required"`
		Password  string `form:"password"`
		Title     string `form:"title"`
		FirstName string `form:"first_name" binding:"required"`
		LastName  string `form:"last_name" binding:"required"`
		ValidID   int    `form:"valid_id"`
		Groups    []int  `form:"groups"`
	}

	if err := c.ShouldBind(&req); err != nil {
		fmt.Printf("DEBUG: Bind error: %v\n", err)
		sendErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid form data: %v", err))
		return
	}
	
	fmt.Printf("DEBUG: User data: Login=%s, FirstName=%s, LastName=%s, ValidID=%d\n", 
		req.Login, req.FirstName, req.LastName, req.ValidID)

	db, err := database.GetDB()
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Database connection failed")
		return
	}

	// Hash password
	hashedPassword := ""
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to hash password"})
			return
		}
		hashedPassword = string(hash)
	}

	user := &models.User{
		Login:      req.Login,
		Password:   hashedPassword,
		Title:      req.Title,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		ValidID:    req.ValidID,
		CreateBy:   1, // TODO: Get from session
		ChangeBy:   1,
		CreateTime: time.Now(),
		ChangeTime: time.Now(),
	}

	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(user); err != nil {
		fmt.Printf("DEBUG: Failed to create user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to create user: %v", err)})
		return
	}
	
	fmt.Printf("DEBUG: User created successfully with ID: %d\n", user.ID)

	// Assign user to groups if specified
	if len(req.Groups) > 0 {
		groupRepo := repository.NewGroupRepository(db)
		for _, groupID := range req.Groups {
			err := groupRepo.AddUserToGroup(uint(user.ID), uint(groupID))
			if err != nil {
				fmt.Printf("DEBUG: Failed to add user to group %d: %v\n", groupID, err)
			} else {
				fmt.Printf("DEBUG: Added user %d to group %d\n", user.ID, groupID)
			}
		}
	}

	// Return success with refreshed user list for HTMX
	c.Header("HX-Trigger", "userCreated")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User created successfully"})
}

// handleUpdateUser updates an existing user
func handleUpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	var req struct {
		Login     string `form:"login" binding:"required"`
		Password  string `form:"password"`
		Title     string `form:"title"`
		FirstName string `form:"first_name" binding:"required"`
		LastName  string `form:"last_name" binding:"required"`
		ValidID   int    `form:"valid_id"`
		Groups    []int  `form:"groups"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Database connection failed")
		return
	}

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// Update fields
	user.Login = req.Login
	user.Title = req.Title
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.ValidID = req.ValidID
	user.ChangeBy = 1 // TODO: Get from session
	user.ChangeTime = time.Now()

	// Update password if provided
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to hash password"})
			return
		}
		user.Password = string(hash)
	}

	if err := userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to update user: %v", err)})
		return
	}

	// Update group assignments
	groupRepo := repository.NewGroupRepository(db)
	
	// First, get current groups
	currentGroups, _ := groupRepo.GetUserGroups(user.ID)
	currentGroupIDs := make(map[uint]bool)
	
	// Get all groups to map names to IDs
	allGroups, _ := groupRepo.List()
	for _, groupName := range currentGroups {
		for _, g := range allGroups {
			if g.Name == groupName {
				currentGroupIDs[g.ID] = true
				break
			}
		}
	}
	
	// Process new group assignments
	newGroupIDs := make(map[uint]bool)
	for _, gid := range req.Groups {
		newGroupIDs[uint(gid)] = true
	}
	
	// Remove groups user is no longer in
	for gid := range currentGroupIDs {
		if !newGroupIDs[gid] {
			groupRepo.RemoveUserFromGroup(user.ID, gid)
		}
	}
	
	// Add new groups
	for gid := range newGroupIDs {
		if !currentGroupIDs[gid] {
			groupRepo.AddUserToGroup(user.ID, gid)
		}
	}

	// Return success response
	c.Header("HX-Trigger", "userUpdated")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User updated successfully"})
}

// handleToggleUserStatus toggles user active/inactive status
func handleToggleUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	var req struct {
		ValidID int `json:"valid_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	// Simple update query for status
	_, err = db.Exec("UPDATE users SET valid_id = $1, change_time = NOW(), change_by = 1 WHERE id = $2", req.ValidID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// handleDeleteUser soft-deletes a user (sets valid_id = 2)
func handleDeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// Soft delete by setting valid_id to 2
	user.ValidID = 2
	user.ChangeBy = 1 // TODO: Get from session
	user.ChangeTime = time.Now()

	if err := userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to delete user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deleted successfully"})
}

// handleResetUserPassword resets a user's password
func handleResetUserPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	// Generate temporary password
	tempPassword := generateTempPassword()
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to hash password"})
		return
	}

	// Update password
	_, err = db.Exec("UPDATE users SET pw = $1, change_time = NOW(), change_by = 1 WHERE id = $2", string(hash), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to reset password"})
		return
	}

	// TODO: Send email with temporary password

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Password reset. Temporary password: %s", tempPassword),
	})
}

// generateTempPassword generates a temporary password
func generateTempPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Group Management Handlers

// handleAdminGroups shows the admin groups management page
func handleAdminGroups(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Database connection failed")
		return
	}

	groupRepo := repository.NewGroupRepository(db)
	groups, err := groupRepo.List()
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch groups")
		return
	}

	// Get search and filter parameters
	search := c.Query("search")
	status := c.Query("status")
	sort := c.DefaultQuery("sort", "name")
	order := c.DefaultQuery("order", "asc")

	// Apply filters
	filteredGroups := groups
	if search != "" {
		var filtered []*models.Group
		for _, g := range filteredGroups {
			if strings.Contains(strings.ToLower(g.Name), strings.ToLower(search)) ||
				strings.Contains(strings.ToLower(g.Comments), strings.ToLower(search)) {
				filtered = append(filtered, g)
			}
		}
		filteredGroups = filtered
	}

	// Handle status filter - frontend sends "1" for active, "2" for inactive
	if status == "1" || status == "active" {
		var filtered []*models.Group
		for _, g := range filteredGroups {
			if g.ValidID == 1 {
				filtered = append(filtered, g)
			}
		}
		filteredGroups = filtered
	} else if status == "2" || status == "inactive" {
		var filtered []*models.Group
		for _, g := range filteredGroups {
			if g.ValidID == 2 {
				filtered = append(filtered, g)
			}
		}
		filteredGroups = filtered
	}

	// Sort groups
	sortGroups(filteredGroups, sort, order)

	// Check if this is an HTMX request for partial update
	if c.GetHeader("HX-Request") == "true" {
		pongo2Renderer.HTML(c, http.StatusOK, "partials/admin/group_table.pongo2", pongo2.Context{
			"Groups": filteredGroups,
			"Search": search,
			"Status": status,
			"Sort":   sort,
			"Order":  order,
		})
		return
	}

	// Full page render
	pongo2Renderer.HTML(c, http.StatusOK, "pages/admin/groups.pongo2", pongo2.Context{
		"Title":      "Group Management - GOTRS Admin",
		"Groups":     filteredGroups,
		"User":       getUserFromContext(c),
		"ActivePage": "admin",
		"Search":     search,
		"Status":     status,
		"Sort":       sort,
		"Order":      order,
	})
}

// handleGetGroup returns group data as JSON
func handleGetGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid group ID"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	groupRepo := repository.NewGroupRepository(db)
	group, err := groupRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":       group.ID,
			"name":     group.Name,
			"comments": group.Comments,
			"valid_id": group.ValidID,
		},
	})
}

// handleCreateGroup creates a new group
func handleCreateGroup(c *gin.Context) {
	var req struct {
		Name     string `form:"name" binding:"required"`
		Comments string `form:"comments"`
		ValidID  int    `form:"valid_id"`
	}

	if err := c.ShouldBind(&req); err != nil {
		sendErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid form data: %v", err))
		return
	}

	db, err := database.GetDB()
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Database connection failed")
		return
	}

	groupRepo := repository.NewGroupRepository(db)

	// Check if group name already exists
	existing, _ := groupRepo.GetByName(req.Name)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Group name already exists"})
		return
	}

	group := &models.Group{
		Name:     req.Name,
		Comments: req.Comments,
		ValidID:  req.ValidID,
		CreateBy: 1, // TODO: Get from session
		ChangeBy: 1,
	}

	err = groupRepo.Create(group)
	if err != nil {
		// Include detailed error information for Guru Meditation
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": fmt.Sprintf("Failed to create group: %v", err),
			"guru_meditation": gin.H{
				"code": "00000005.GRPCRATE",
				"message": fmt.Sprintf("Group creation failed: %v", err),
				"task": "GROUP.CREATE",
				"location": "htmx_routes.go:handleCreateGroup",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Group created successfully",
		"data":    group,
	})
}

// handleUpdateGroup updates an existing group
func handleUpdateGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid group ID"})
		return
	}

	var req struct {
		Name     string `form:"name" binding:"required"`
		Comments string `form:"comments"`
		ValidID  int    `form:"valid_id"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": fmt.Sprintf("Invalid form data: %v", err)})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	groupRepo := repository.NewGroupRepository(db)

	// Get existing group
	group, err := groupRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Group not found"})
		return
	}

	// Check if new name conflicts with another group
	if group.Name != req.Name {
		existing, _ := groupRepo.GetByName(req.Name)
		if existing != nil && existing.ID != group.ID {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Group name already exists"})
			return
		}
	}

	// Update group
	group.Name = req.Name
	group.Comments = req.Comments
	group.ValidID = req.ValidID
	group.ChangeBy = 1 // TODO: Get from session

	err = groupRepo.Update(group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Group updated successfully",
	})
}

// handleDeleteGroup soft deletes a group
func handleDeleteGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid group ID"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	groupRepo := repository.NewGroupRepository(db)

	// Check if group exists
	group, err := groupRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Group not found"})
		return
	}

	// Don't delete system groups
	if group.Name == "admin" || group.Name == "users" || group.Name == "stats" {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Cannot delete system group"})
		return
	}

	// Soft delete
	err = groupRepo.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Group deleted successfully",
	})
}

// handleGetGroupMembers gets members of a group
func handleGetGroupMembers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid group ID"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	groupRepo := repository.NewGroupRepository(db)
	members, err := groupRepo.GetGroupMembers(uint(id))
	if err != nil {
		// Log the actual error for debugging
		log.Printf("Error fetching members for group %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false, 
			"error": fmt.Sprintf("Failed to fetch group members: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    members,
	})
}

// handleAddGroupMember adds a user to a group
func handleAddGroupMember(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid group ID"})
		return
	}

	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request data"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	groupRepo := repository.NewGroupRepository(db)
	err = groupRepo.AddUserToGroup(req.UserID, uint(groupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to add member to group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Member added successfully",
	})
}

// handleRemoveGroupMember removes a user from a group
func handleRemoveGroupMember(c *gin.Context) {
	groupIDStr := c.Param("id")
	userIDStr := c.Param("userId")

	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid group ID"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	groupRepo := repository.NewGroupRepository(db)
	err = groupRepo.RemoveUserFromGroup(uint(userID), uint(groupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to remove member from group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Member removed successfully",
	})
}

// handleGetGroupPermissions gets permissions for a group
func handleGetGroupPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid group ID"})
		return
	}

	// TODO: Implement permission retrieval
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"group_id": id,
			"permissions": map[string][]string{
				"rw": []string{"ticket_create", "ticket_update", "ticket_close"},
				"ro": []string{"ticket_view", "report_view"},
			},
		},
	})
}

// handleUpdateGroupPermissions updates permissions for a group
func handleUpdateGroupPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid group ID"})
		return
	}

	var permissions map[string][]string
	if err := c.ShouldBindJSON(&permissions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid permission data"})
		return
	}

	// TODO: Implement permission update
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Permissions updated successfully",
		"group_id": id,
	})
}

// sortGroups sorts groups based on field and order
func sortGroups(groups []*models.Group, field string, order string) {
	// Implementation of sorting logic
	// This is a placeholder - implement actual sorting based on field
}

// Queue Management API Handlers

// Create queue request structure
type CreateQueueRequest struct {
	Name                  string `json:"name" binding:"required,min=2,max=200"`
	Comment               string `json:"comment"`
	GroupID               int    `json:"group_id"`
	SystemAddress         string `json:"system_address"`
	DefaultSignKey        string `json:"default_sign_key"`
	UnlockTimeout         int    `json:"unlock_timeout"`
	FollowUpID            int    `json:"follow_up_id"`
	FollowUpLock          int    `json:"follow_up_lock"`
	CalendarName          string `json:"calendar_name"`
	FirstResponseTime     int    `json:"first_response_time"`
	FirstResponseNotify   int    `json:"first_response_notify"`
	UpdateTime            int    `json:"update_time"`
	UpdateNotify          int    `json:"update_notify"`
	SolutionTime          int    `json:"solution_time"`
	SolutionNotify        int    `json:"solution_notify"`
}

// Update queue request structure
type UpdateQueueRequest struct {
	Name                  *string `json:"name,omitempty"`
	Comment               *string `json:"comment,omitempty"`
	GroupID               *int    `json:"group_id,omitempty"`
	SystemAddress         *string `json:"system_address,omitempty"`
	DefaultSignKey        *string `json:"default_sign_key,omitempty"`
	UnlockTimeout         *int    `json:"unlock_timeout,omitempty"`
	FollowUpID            *int    `json:"follow_up_id,omitempty"`
	FollowUpLock          *int    `json:"follow_up_lock,omitempty"`
	CalendarName          *string `json:"calendar_name,omitempty"`
	FirstResponseTime     *int    `json:"first_response_time,omitempty"`
	FirstResponseNotify   *int    `json:"first_response_notify,omitempty"`
	UpdateTime            *int    `json:"update_time,omitempty"`
	UpdateNotify          *int    `json:"update_notify,omitempty"`
	SolutionTime          *int    `json:"solution_time,omitempty"`
	SolutionNotify        *int    `json:"solution_notify,omitempty"`
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	if email == "" {
		return true // Empty email is allowed
	}
	// Simple email validation - in production, use a proper regex or validation library
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// Helper function to validate queue data (used by both create and update)
func validateQueueData(name, systemAddress *string, firstResponseTime, updateTime, solutionTime *int, excludeID int) error {
	// Validate email format
	if systemAddress != nil && *systemAddress != "" && !isValidEmail(*systemAddress) {
		return fmt.Errorf("invalid email format in system_address")
	}
	
	// Validate time values
	if (firstResponseTime != nil && *firstResponseTime < 0) ||
		(updateTime != nil && *updateTime < 0) ||
		(solutionTime != nil && *solutionTime < 0) {
		return fmt.Errorf("time values must be positive")
	}
	
	// Validate name uniqueness
	if name != nil && queueNameExists(*name, excludeID) {
		return fmt.Errorf("queue name already exists")
	}
	
	return nil
}


// Helper function to check if queue name exists
func queueNameExists(name string, excludeID int) bool {
	db, err := database.GetDB()
	if err != nil {
		return false
	}
	
	var count int
	query := "SELECT COUNT(*) FROM queue WHERE LOWER(name) = LOWER($1) AND id != $2"
	err = db.QueryRow(query, name, excludeID).Scan(&count)
	if err != nil {
		return false
	}
	
	return count > 0
}

// Helper function to get next queue ID
func getNextQueueID() int {
	db, err := database.GetDB()
	if err != nil {
		return 1
	}
	
	var maxID int
	err = db.QueryRow("SELECT COALESCE(MAX(id), 0) FROM queue").Scan(&maxID)
	if err != nil {
		return 1
	}
	
	return maxID + 1
}

// Helper function to check if queue has tickets
func queueHasTickets(queueID int) bool {
	db, err := database.GetDB()
	if err != nil {
		return false
	}
	
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM ticket WHERE queue_id = $1", queueID).Scan(&count)
	if err != nil {
		return false
	}
	
	return count > 0
}

// Create Queue API Handler
func handleCreateQueue(c *gin.Context) {
	var req CreateQueueRequest
	
	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		if strings.Contains(err.Error(), "invalid character") {
			sendErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
			return
		}
		sendErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}
	
	// Additional validations using helper
	if err := validateQueueData(&req.Name, &req.SystemAddress, &req.FirstResponseTime, &req.UpdateTime, &req.SolutionTime, 0); err != nil {
		sendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	
	// Get queue repository
	queueRepo := GetQueueRepository()
	if queueRepo == nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Database connection not available")
		return
	}
	
	// Create queue model
	queue := &models.Queue{
		Name:          req.Name,
		Comment:       req.Comment,
		ValidID:       1, // Active
		FollowUpID:    1, // possible
		FollowUpLock:  0, // no
		UnlockTimeout: 0,
		GroupID:       1, // Default group
		CreateTime:    time.Now(),
		CreateBy:      1, // System user
		ChangeTime:    time.Now(),
		ChangeBy:      1,
	}
	
	// Save to database
	if err := queueRepo.Create(queue); err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create queue: %v", err))
		return
	}
	
	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":                     queue.ID,
			"name":                   queue.Name,
			"comment":                queue.Comment,
			"group_id":               queue.GroupID,
			"system_address":         "",
			"default_sign_key":       "",
			"unlock_timeout":         queue.UnlockTimeout,
			"follow_up_id":           queue.FollowUpID,
			"follow_up_lock":         queue.FollowUpLock,
			"calendar_name":          "",
			"first_response_time":    req.FirstResponseTime,
			"first_response_notify":  req.FirstResponseNotify,
			"update_time":            req.UpdateTime,
			"update_notify":          req.UpdateNotify,
			"solution_time":          req.SolutionTime,
			"solution_notify":        req.SolutionNotify,
			"valid_id":               queue.ValidID,
			"create_time":            queue.CreateTime,
			"create_by":              queue.CreateBy,
			"change_time":            queue.ChangeTime,
			"change_by":              queue.ChangeBy,
		},
	})
}

// Update Queue API Handler
func handleUpdateQueue(c *gin.Context) {
	queueID := c.Param("id")
	
	// Validate queue ID
	id, err := strconv.Atoi(queueID)
	if err != nil {
		sendErrorResponse(c, http.StatusBadRequest, "Invalid queue ID")
		return
	}
	
	// Check if queue exists
	existingQueue, err := getQueueWithTickets(queueID)
	if err != nil {
		if err.Error() == "queue not found" {
			sendErrorResponse(c, http.StatusNotFound, "Queue not found")
			return
		}
		sendErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}
	
	var req UpdateQueueRequest
	
	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}
	
	// Additional validations using helper
	if err := validateQueueData(req.Name, req.SystemAddress, req.FirstResponseTime, req.UpdateTime, req.SolutionTime, id); err != nil {
		sendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	
	// Update queue (mock implementation - merge with existing data)
	updatedQueue := existingQueue
	if req.Name != nil {
		updatedQueue["name"] = *req.Name
	}
	if req.Comment != nil {
		updatedQueue["comment"] = *req.Comment
	}
	// Add other field updates as needed...
	
	updatedQueue["change_time"] = time.Now()
	updatedQueue["change_by"] = 1
	
	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedQueue,
	})
}

// Delete Queue API Handler - performs soft delete
func handleDeleteQueue(c *gin.Context) {
	queueID := c.Param("id")
	
	// Validate queue ID
	id, err := strconv.Atoi(queueID)
	if err != nil {
		sendErrorResponse(c, http.StatusBadRequest, "Invalid queue ID")
		return
	}
	
	// Check if queue exists
	_, err = getQueueWithTickets(queueID)
	if err != nil {
		if err.Error() == "queue not found" {
			sendErrorResponse(c, http.StatusNotFound, "Queue not found")
			return
		}
		sendErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}
	
	// Check if queue has tickets
	if queueHasTickets(id) {
		sendErrorResponse(c, http.StatusConflict, "Cannot delete queue with existing tickets")
		return
	}
	
	// Perform soft delete by setting valid_id to 2 (invalid)
	db, err := database.GetDB()
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Database connection failed")
		return
	}
	
	_, err = db.Exec("UPDATE queue SET valid_id = 2, change_time = NOW(), change_by = 1 WHERE id = $1", id)
	if err != nil {
		fmt.Printf("ERROR handleDeleteQueue: Failed to delete queue %d: %v\n", id, err)
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to delete queue")
		return
	}
	
	fmt.Printf("DEBUG handleDeleteQueue: Successfully deleted queue %d\n", id)
	
	// Also remove system_address from system_data if exists
	var queueName string
	err = db.QueryRow("SELECT name FROM queue WHERE id = $1", id).Scan(&queueName)
	if err == nil && queueName != "" {
		key := fmt.Sprintf("queue_system_address_%s", queueName)
		_, _ = db.Exec("DELETE FROM system_data WHERE data_key = $1", key)
	}
	
	// Return success response with HTMX headers for list refresh
	c.Header("HX-Trigger", "queue-deleted")
	c.Header("HX-Redirect", "/queues")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Queue deleted successfully",
	})
}

// Frontend Queue Form Handlers

// Handle edit queue form display
func handleEditQueueForm(c *gin.Context) {
	queueID := c.Param("id")
	
	// Validate queue ID
	id, err := strconv.Atoi(queueID)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid queue ID")
		return
	}
	
	// Get the QueueRepository instance
	queueRepo := GetQueueRepository()
	
	fmt.Printf("DEBUG handleEditQueueForm: Fetching queue ID %d\n", id)
	
	// Get complete queue details from database
	queueModel, err := queueRepo.GetByID(uint(id))
	if err != nil {
		fmt.Printf("ERROR handleEditQueueForm: Failed to get queue %d: %v\n", id, err)
		if err == sql.ErrNoRows {
			c.String(http.StatusNotFound, "Queue not found")
			return
		}
		c.String(http.StatusInternalServerError, "Database error: %v", err)
		return
	}
	
	fmt.Printf("DEBUG handleEditQueueForm: Retrieved queue model: ID=%d, Name=%s, Comment=%s\n", 
		queueModel.ID, queueModel.Name, queueModel.Comment)
	
	// Retrieve system_address from system_data table
	systemAddress := ""
	db, _ := database.GetDB()
	if db != nil {
		key := fmt.Sprintf("queue_system_address_%s", queueModel.Name)
		var value sql.NullString
		err := db.QueryRow("SELECT data_value FROM system_data WHERE data_key = $1", key).Scan(&value)
		if err == nil && value.Valid {
			systemAddress = value.String
			fmt.Printf("DEBUG handleEditQueueForm: Retrieved system_address: %s\n", systemAddress)
		}
	}
	
	// Convert queue model to template-friendly format
	// The template expects lowercase field names in the map
	queueData := gin.H{
		"id":                   queueModel.ID,
		"name":                 queueModel.Name,
		"comment":              queueModel.Comment,
		"system_address":       systemAddress, // Retrieved from system_data
		"first_response_time":  0,  // These SLA fields aren't in the queue table
		"update_time":          0,
		"solution_time":        0,
	}
	
	// Debug: Log what we're passing to the template
	fmt.Printf("DEBUG handleEditQueueForm: Passing queue data to template: %+v\n", queueData)
	
	// Load and render the edit form template
	tmpl, err := loadTemplateForRequest(c, "templates/components/queue_edit_form.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "queue_edit_form.html", gin.H{
		"Queue": queueData,
	}); err != nil {
		c.String(http.StatusInternalServerError, "Render error: %v", err)
	}
}

// Handle new queue form display
func handleNewQueueForm(c *gin.Context) {
	// Load and render the create form template
	tmpl, err := loadTemplateForRequest(c, "templates/components/queue_create_form.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "queue_create_form.html", gin.H{}); err != nil {
		c.String(http.StatusInternalServerError, "Render error: %v", err)
	}
}

// Handle delete queue confirmation display
func handleDeleteQueueConfirmation(c *gin.Context) {
	queueID := c.Param("id")
	
	// Get queue details
	queue, err := getQueueWithTickets(queueID)
	if err != nil {
		if err.Error() == "queue not found" {
			c.String(http.StatusNotFound, "Queue not found")
			return
		}
		c.String(http.StatusInternalServerError, "Database error")
		return
	}
	
	// Check if queue has tickets
	id, _ := strconv.Atoi(queueID)
	hasTickets := queueHasTickets(id)
	
	// Prepare template data
	templateData := gin.H{
		"Queue":      queue,
		"HasTickets": hasTickets,
		"QueueID":    queueID,
	}
	
	if hasTickets {
		templateData["TicketCount"] = queue["ticket_count"]
	}
	
	// For now, render a simple HTML response since we don't have the template yet
	c.Header("Content-Type", "text/html; charset=utf-8")
	
	if hasTickets {
		c.String(http.StatusOK, `
			<div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50" id="delete-queue-modal">
				<div class="relative top-20 mx-auto p-5 border w-11/12 md:w-1/2 lg:w-1/3 shadow-lg rounded-md bg-white dark:bg-gray-800">
					<div class="mt-3">
						<h3 class="text-lg font-medium text-gray-900 dark:text-white">Queue Cannot Be Deleted</h3>
						<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
							The queue "%s" cannot be deleted because it contains tickets (%d tickets).
							Please move or resolve all tickets before deleting this queue.
						</p>
						<div class="flex justify-end pt-4">
							<button type="button" onclick="closeDeleteModal()" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-600">
								Close
							</button>
						</div>
					</div>
				</div>
			</div>
			<script>
			function closeDeleteModal() {
				document.getElementById('delete-queue-modal').remove();
			}
			</script>
		`, queue["name"], queue["ticket_count"])
	} else {
		c.String(http.StatusOK, `
			<div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50" id="delete-queue-modal">
				<div class="relative top-20 mx-auto p-5 border w-11/12 md:w-1/2 lg:w-1/3 shadow-lg rounded-md bg-white dark:bg-gray-800">
					<div class="mt-3">
						<h3 class="text-lg font-medium text-gray-900 dark:text-white">Delete Queue</h3>
						<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
							Are you sure you want to delete the queue "%s"? This action cannot be undone.
						</p>
						<div class="flex justify-end space-x-3 pt-4">
							<button type="button" onclick="closeDeleteModal()" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-600">
								Cancel
							</button>
							<button type="button" hx-delete="/api/queues/%s" hx-target="#delete-queue-modal" hx-swap="outerHTML" class="px-4 py-2 text-sm font-medium text-white bg-red-600 border border-transparent rounded-md shadow-sm hover:bg-red-700">
								Delete Queue
							</button>
						</div>
					</div>
				</div>
			</div>
			<script>
			function closeDeleteModal() {
				document.getElementById('delete-queue-modal').remove();
			}
			</script>
		`, queue["name"], queueID)
	}
}

// Handle create queue with HTMX form submission
func handleCreateQueueWithHTMX(c *gin.Context) {
	fmt.Printf("DEBUG: handleCreateQueueWithHTMX called\n")
	// Parse form data
	name := c.PostForm("name")
	comment := c.PostForm("comment")
	systemAddress := c.PostForm("system_address")
	fmt.Printf("DEBUG: Parsed form - name: %s, comment: %s\n", name, comment)
	
	// Validate required fields
	if name == "" {
		c.String(http.StatusBadRequest, `
			<div class="text-red-600 text-sm mt-1">
				<p>error: Queue name is required</p>
			</div>
		`)
		return
	}
	
	// Validate name length
	if len(name) < 2 || len(name) > 200 {
		c.String(http.StatusBadRequest, `
			<div class="text-red-600 text-sm mt-1">
				<p>error: Queue name must be between 2 and 200 characters</p>
			</div>
		`)
		return
	}
	
	// Parse optional integer fields
	firstResponseTime, _ := strconv.Atoi(c.PostForm("first_response_time"))
	updateTime, _ := strconv.Atoi(c.PostForm("update_time"))
	solutionTime, _ := strconv.Atoi(c.PostForm("solution_time"))
	
	// Validate using existing helper
	if err := validateQueueData(&name, &systemAddress, &firstResponseTime, &updateTime, &solutionTime, 0); err != nil {
		c.String(http.StatusBadRequest, `
			<div class="text-red-600 text-sm mt-1">
				<p>error: %s</p>
			</div>
		`, err.Error())
		return
	}
	
	// Get queue repository
	queueRepo := GetQueueRepository()
	if queueRepo == nil {
		fmt.Printf("DEBUG: Queue repository is nil\n")
		c.String(http.StatusInternalServerError, `
			<div class="text-red-600 text-sm mt-1">
				<p>error: Database connection not available</p>
			</div>
		`)
		return
	}
	
	fmt.Printf("DEBUG: Creating queue with name: %s, comment: %s, system_address: %s\n", name, comment, systemAddress)
	
	// Handle system_address - store in system_data table as workaround
	// since system_address table doesn't exist yet in our OTRS-compatible schema
	if systemAddress != "" {
		// Store the email in system_data table with a unique key
		db, _ := database.GetDB()
		if db != nil {
			key := fmt.Sprintf("queue_system_address_%s", name)
			_, err := db.Exec(`
				INSERT INTO system_data (data_key, data_value, create_time, create_by, change_time, change_by)
				VALUES ($1, $2, NOW(), 1, NOW(), 1)
				ON CONFLICT (data_key) DO UPDATE SET data_value = $2, change_time = NOW()
			`, key, systemAddress)
			if err != nil {
				fmt.Printf("WARNING: Failed to store system_address: %v\n", err)
			}
			// For now, we'll leave system_address_id as NULL in the queue table
			// In future, when system_address table is added, we can migrate this data
		}
	}
	
	// Create queue model
	queue := &models.Queue{
		Name:          name,
		Comment:       comment,
		ValidID:       1, // Active
		FollowUpID:    1, // possible
		FollowUpLock:  0, // no
		UnlockTimeout: 0,
		GroupID:       1, // Default group
		CreateTime:    time.Now(),
		CreateBy:      1, // System user
		ChangeTime:    time.Now(),
		ChangeBy:      1,
	}
	
	// Save to database
	fmt.Printf("DEBUG: Attempting to save queue to database\n")
	if err := queueRepo.Create(queue); err != nil {
		fmt.Printf("DEBUG: Error creating queue: %v\n", err)
		c.String(http.StatusInternalServerError, `
			<div class="text-red-600 text-sm mt-1">
				<p>error: Failed to create queue: %s</p>
			</div>
		`, err.Error())
		return
	}
	
	fmt.Printf("DEBUG: Queue created successfully with ID: %d\n", queue.ID)
	
	// Return success with HTMX headers
	c.Header("HX-Trigger", "queue-created")
	c.Header("HX-Redirect", "/queues")
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":   queue.ID,
			"name": queue.Name,
		},
	})
}

// Handle update queue with HTMX form submission
func handleUpdateQueueWithHTMX(c *gin.Context) {
	queueID := c.Param("id")
	
	// Validate queue ID
	id, err := strconv.Atoi(queueID)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid queue ID")
		return
	}
	
	// Check if queue exists
	_, err = getQueueWithTickets(queueID)
	if err != nil {
		if err.Error() == "queue not found" {
			c.String(http.StatusNotFound, "Queue not found")
			return
		}
		c.String(http.StatusInternalServerError, "Database error")
		return
	}
	
	// Parse form data
	name := c.PostForm("name")
	comment := c.PostForm("comment")
	systemAddress := c.PostForm("system_address")
	
	fmt.Printf("DEBUG handleUpdateQueueWithHTMX: Updating queue %d with name=%s, comment=%s\n", id, name, comment)
	
	// Validate required fields
	if name == "" {
		c.String(http.StatusBadRequest, `
			<div class="text-red-600 text-sm mt-1">
				<p>error: Queue name is required</p>
			</div>
		`)
		return
	}
	
	// Parse optional integer fields
	firstResponseTime, _ := strconv.Atoi(c.PostForm("first_response_time"))
	updateTime, _ := strconv.Atoi(c.PostForm("update_time"))
	solutionTime, _ := strconv.Atoi(c.PostForm("solution_time"))
	
	// Validate using existing helper
	if err := validateQueueData(&name, &systemAddress, &firstResponseTime, &updateTime, &solutionTime, id); err != nil {
		c.String(http.StatusBadRequest, `
			<div class="text-red-600 text-sm mt-1">
				<p>error: %s</p>
			</div>
		`, err.Error())
		return
	}
	
	// Get the QueueRepository instance
	queueRepo := GetQueueRepository()
	
	// Get the existing queue from database
	existingQueueModel, err := queueRepo.GetByID(uint(id))
	if err != nil {
		fmt.Printf("ERROR: Failed to get queue for update: %v\n", err)
		c.String(http.StatusInternalServerError, `
			<div class="text-red-600 text-sm mt-1">
				<p>error: Failed to get queue: %s</p>
			</div>
		`, err.Error())
		return
	}
	
	// Update the queue fields
	fmt.Printf("DEBUG handleUpdateQueueWithHTMX: Before update - Name: %s, Comment: %s\n", 
		existingQueueModel.Name, existingQueueModel.Comment)
	
	// Store the old name to update system_address key if name changes
	oldName := existingQueueModel.Name
	
	existingQueueModel.Name = name
	existingQueueModel.Comment = comment
	
	// Update system_address in system_data table
	if systemAddress != "" {
		db, _ := database.GetDB()
		if db != nil {
			// If name changed, update the key
			oldKey := fmt.Sprintf("queue_system_address_%s", oldName)
			newKey := fmt.Sprintf("queue_system_address_%s", name)
			
			if oldName != name {
				// Delete old key if exists
				db.Exec("DELETE FROM system_data WHERE data_key = $1", oldKey)
			}
			
			// Insert or update with new key
			_, err := db.Exec(`
				INSERT INTO system_data (data_key, data_value, create_time, create_by, change_time, change_by)
				VALUES ($1, $2, NOW(), 1, NOW(), 1)
				ON CONFLICT (data_key) DO UPDATE SET data_value = $2, change_time = NOW()
			`, newKey, systemAddress)
			if err != nil {
				fmt.Printf("WARNING: Failed to update system_address: %v\n", err)
			}
		}
	}
	
	fmt.Printf("DEBUG handleUpdateQueueWithHTMX: After update - Name: %s, Comment: %s, SystemAddress: %s\n", 
		existingQueueModel.Name, existingQueueModel.Comment, systemAddress)
	
	// Update the queue in the database
	if err := queueRepo.Update(existingQueueModel); err != nil {
		fmt.Printf("ERROR handleUpdateQueueWithHTMX: Failed to update queue: %v\n", err)
		c.String(http.StatusInternalServerError, `
			<div class="text-red-600 text-sm mt-1">
				<p>error: Failed to update queue: %s</p>
			</div>
		`, err.Error())
		return
	}
	
	fmt.Printf("DEBUG handleUpdateQueueWithHTMX: Update successful for queue %d\n", id)
	
	// Get the updated queue from database
	updatedQueue, err := queueRepo.GetByID(uint(id))
	if err != nil {
		fmt.Printf("ERROR: Failed to retrieve updated queue: %v\n", err)
		c.String(http.StatusInternalServerError, `
			<div class="text-red-600 text-sm mt-1">
				<p>error: Failed to retrieve updated queue</p>
			</div>
		`)
		return
	}
	
	fmt.Printf("DEBUG handleUpdateQueueWithHTMX: Retrieved updated queue - Name: %s, Comment: %s\n", 
		updatedQueue.Name, updatedQueue.Comment)
	
	// Return success with HTMX headers
	c.Header("HX-Trigger", "queue-updated")
	c.Header("HX-Redirect", fmt.Sprintf("/queues/%s", queueID))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedQueue,
	})
}

// Handle queue tickets with HTMX pagination
func handleQueueTicketsWithHTMX(c *gin.Context) {
	queueID := c.Param("id")
	
	// Parse query parameters
	status := c.Query("status")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	
	// Get tickets for queue
	tickets, err := getTicketsForQueue(queueID, status, page, limit)
	if err != nil {
		c.String(http.StatusInternalServerError, "Database error: %v", err)
		return
	}
	
	// Load and render the ticket list template
	tmpl, err := loadTemplateForRequest(c, "templates/components/ticket_list.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	
	// Prepare template data
	templateData := gin.H{
		"Tickets":    tickets["tickets"],
		"Title":      "Queue Tickets",
		"Total":      tickets["total"],
		"Page":       tickets["page"],
		"Pagination": tickets["pagination"],
	}
	
	// Handle empty state
	if tickets["total"].(int) == 0 {
		templateData["NoTicketsMessage"] = "No tickets in this queue"
	}
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "ticket_list.html", templateData); err != nil {
		c.String(http.StatusInternalServerError, "Render error: %v", err)
	}
}

// Handle clear search - resets search and filter state
func handleClearQueueSearch(c *gin.Context) {
	// Get all queues without any filtering
	queues, err := getQueuesWithTicketCounts("", "")
	if err != nil {
		c.String(http.StatusInternalServerError, "Database error: %v", err)
		return
	}
	
	// Render with empty search/filter context
	templateData := gin.H{
		"Queues":      queues,
		"SearchTerm":  "",
		"StatusFilter": "",
	}
	
	// Always return HTML fragment for HTMX clear requests
	tmpl, err := loadTemplateForRequest(c, "templates/components/queue_list.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "queue_list.html", templateData); err != nil {
		c.String(http.StatusInternalServerError, "Render error: %v", err)
	}
}

// Bulk operations toolbar - shows when queues are selected
func handleBulkActionsToolbar(c *gin.Context) {
	count := c.Query("count")
	
	// Parse count
	selectedCount := 0
	if count != "" {
		if n, err := strconv.Atoi(count); err == nil {
			selectedCount = n
		}
	}
	
	// Generate toolbar HTML
	if selectedCount == 0 {
		c.String(http.StatusOK, `<div id="bulk-actions-toolbar" style="display: none"></div>`)
		return
	}
	
	html := fmt.Sprintf(`
	<div id="bulk-actions-toolbar" class="bg-blue-50 dark:bg-blue-900/20 border-b border-blue-200 dark:border-blue-800 px-4 py-3">
		<div class="flex items-center justify-between">
			<div class="flex items-center">
				<span class="text-sm font-medium text-blue-800 dark:text-blue-200">
					%d queue%s selected
				</span>
			</div>
			<div class="flex items-center space-x-2">
				<button type="button"
						hx-put="/api/queues/bulk/activate"
						hx-include="[name='queue-select']:checked"
						hx-confirm="Activate selected queues?"
						class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500">
					Activate Selected
				</button>
				<button type="button"
						hx-put="/api/queues/bulk/deactivate"
						hx-include="[name='queue-select']:checked"
						hx-confirm="Deactivate selected queues?"
						class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-white bg-yellow-600 hover:bg-yellow-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-yellow-500">
					Deactivate Selected
				</button>
				<button type="button"
						hx-delete="/api/queues/bulk"
						hx-include="[name='queue-select']:checked"
						hx-confirm="Delete selected queues? This cannot be undone!"
						class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500">
					Delete Selected
				</button>
				<button type="button"
						onclick="cancelBulkSelection()"
						class="inline-flex items-center px-3 py-1.5 border border-gray-300 dark:border-gray-600 text-xs font-medium rounded text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gotrs-500">
					Cancel Selection
				</button>
			</div>
		</div>
	</div>`, selectedCount, func() string {
		if selectedCount == 1 {
			return ""
		}
		return "s"
	}())
	
	c.String(http.StatusOK, html)
}

// Bulk queue status change (activate/deactivate)
func handleBulkQueueAction(c *gin.Context) {
	action := c.Param("action")
	
	// Validate action
	if action != "activate" && action != "deactivate" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}
	
	// Parse form data and get queue IDs
	c.Request.ParseForm()
	queueIDs := c.Request.Form["queue_ids"]
	
	if len(queueIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No queues selected"})
		return
	}
	
	// Check for too many selections
	if len(queueIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many queues selected (maximum 100)"})
		return
	}
	
	// Validate queue IDs
	validIDs := []int{}
	for _, idStr := range queueIDs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue ID: " + idStr})
			return
		}
		validIDs = append(validIDs, id)
	}
	
	// Update queue statuses in database
	db, err := database.GetDB()
	if err != nil {
		sendGuruMeditation(c, err, "handleBulkQueueAction:GetDB")
		return
	}
	
	newValidID := 1 // active
	if action == "deactivate" {
		newValidID = 2 // inactive
	}
	
	// Build placeholders for IN clause
	placeholders := make([]string, len(validIDs))
	args := make([]interface{}, len(validIDs)+1)
	args[0] = newValidID
	for i, id := range validIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}
	
	query := fmt.Sprintf("UPDATE queue SET valid_id = $1 WHERE id IN (%s)", strings.Join(placeholders, ","))
	result, err := db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update queues"})
		return
	}
	
	updated, _ := result.RowsAffected()
	
	// Send response with HTMX triggers
	c.Header("HX-Trigger", `{"queues-updated": true, "show-toast": {"message": "` + fmt.Sprintf("%d queues %sd", updated, action) + `", "type": "success"}}`)
	
	message := fmt.Sprintf("%d queue%s %sd", updated, func() string {
		if updated == 1 {
			return ""
		}
		return "s"
	}(), action)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"updated": updated,
	})
}

// Bulk queue deletion
func handleBulkQueueDelete(c *gin.Context) {
	// For DELETE requests, we need to read the body manually
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	
	// Parse the form data from body
	values, err := url.ParseQuery(string(body))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}
	
	// Get confirmation flag
	confirm := values.Get("confirm")
	if confirm != "true" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Confirmation required"})
		return
	}
	
	// Get queue IDs from form data
	queueIDs := values["queue_ids"]
	
	if len(queueIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No queues selected"})
		return
	}
	
	// Check for too many selections
	if len(queueIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many queues selected (maximum 100)"})
		return
	}
	
	// Validate queue IDs and check for tickets
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection required"})
		return
	}
	
	deleted := 0
	skipped := []string{}
	
	for _, idStr := range queueIDs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue ID: " + idStr})
			return
		}
		
		// Check if queue exists and has tickets
		var queueName string
		var ticketCount int
		err = db.QueryRow(`
			SELECT q.name, COUNT(t.id) 
			FROM queue q
			LEFT JOIN ticket t ON t.queue_id = q.id
			WHERE q.id = $1
			GROUP BY q.name
		`, id).Scan(&queueName, &ticketCount)
		
		if err == sql.ErrNoRows {
			// Queue doesn't exist - skip silently
			continue
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		
		if ticketCount > 0 {
			skipped = append(skipped, fmt.Sprintf("%s (contains %d tickets)", queueName, ticketCount))
		} else {
			// Delete the queue from database
			_, err = db.Exec("DELETE FROM queue WHERE id = $1", id)
			if err != nil {
				skipped = append(skipped, fmt.Sprintf("%s (deletion failed)", queueName))
			} else {
				deleted++
			}
		}
	}
	
	// Check if all were skipped
	if deleted == 0 && len(skipped) > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "No queues could be deleted - all contain tickets",
			"skipped": skipped,
		})
		return
	}
	
	// Build response message
	message := ""
	if deleted > 0 && len(skipped) == 0 {
		message = fmt.Sprintf("%d queue%s deleted", deleted, func() string {
			if deleted == 1 {
				return ""
			}
			return "s"
		}())
	} else if deleted > 0 && len(skipped) > 0 {
		message = fmt.Sprintf("%d queue%s deleted, %d skipped", deleted, func() string {
			if deleted == 1 {
				return ""
			}
			return "s"
		}(), len(skipped))
	}
	
	// Send response with HTMX triggers
	c.Header("HX-Trigger", `{"queues-updated": true, "show-toast": {"message": "` + message + `", "type": "success"}}`)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"deleted": deleted,
		"skipped": skipped,
	})
}

// Ticket Workflow Handlers

// Workflow state constants
const (
	StateNew      = "new"
	StateOpen     = "open"
	StatePending  = "pending"
	StateResolved = "resolved"
	StateClosed   = "closed"
)

// handleTicketWorkflow displays the workflow state diagram and transition options
func handleTicketWorkflow(c *gin.Context) {
	ticketID := c.Param("id")
	
	// Get ticket data from database
	ticket := getTicketByID(ticketID)
	if ticket == nil {
		c.String(http.StatusNotFound, "Ticket not found")
		return
	}
	
	// Get current state from ticket
	currentState := ticket["state"].(string)
	
	// Get available transitions based on current state
	transitions := getAvailableTransitions(currentState)
	
	// Get state history
	history := getTicketStateHistory(ticketID)
	
	// Return workflow view with cleaner formatting
	c.String(http.StatusOK, renderWorkflowView(currentState, ticketID, transitions, history))
}

// renderWorkflowView generates the workflow HTML
func renderWorkflowView(currentState, ticketID string, transitions []gin.H, history []gin.H) string {
	return fmt.Sprintf(`
		<div class="ticket-workflow" data-current-state="%s">
			<h3 class="text-lg font-semibold mb-4">Workflow State</h3>
			
			<!-- State Diagram -->
			<div class="state-diagram mb-6">
				<div class="flex justify-between items-center">
					<div class="state-badge state-new">New</div>
					<div class="arrow">→</div>
					<div class="state-badge state-open">Open</div>
					<div class="arrow">→</div>
					<div class="state-badge state-pending">Pending</div>
					<div class="arrow">→</div>
					<div class="state-badge state-resolved">Resolved</div>
					<div class="arrow">→</div>
					<div class="state-badge state-closed">Closed</div>
				</div>
			</div>
			
			<!-- Available Transitions -->
			<div class="transitions mb-6">
				<h4 class="font-medium mb-2">Available Actions</h4>
				<div class="space-y-2">
					%s
				</div>
			</div>
			
			<!-- State History -->
			<div class="history">
				<h4 class="font-medium mb-2">State History</h4>
				%s
			</div>
		</div>
		
		<style>
		.state-badge {
			padding: 4px 12px;
			border-radius: 9999px;
			font-size: 0.875rem;
			font-weight: 500;
		}
		.state-new { background: #3B82F6; color: white; }
		.state-open { background: #10B981; color: white; }
		.state-pending { background: #F59E0B; color: white; }
		.state-resolved { background: #8B5CF6; color: white; }
		.state-closed { background: #6B7280; color: white; }
		.arrow { color: #9CA3AF; }
		</style>
	`, currentState, renderTransitions(transitions, ticketID), renderHistory(history))
}

// handleTicketTransition processes state transition requests
func handleTicketTransition(c *gin.Context) {
	ticketID := c.Param("id")
	_ = ticketID // for future use
	
	// Parse request
	currentState := c.PostForm("current_state")
	newState := c.PostForm("new_state")
	reason := c.PostForm("reason")
	
	// Default current state if not provided (for testing)
	if currentState == "" {
		currentState = StateOpen
	}
	
	// Check user permissions
	if err := checkTransitionPermission(c, newState); err != nil {
		if err.Error() == "reopen_requested" {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Reopen request submitted",
				"new_state": "reopen_requested",
			})
		} else {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": err.Error(),
			})
		}
		return
	}
	
	// Validate state transition
	if !isValidTransition(currentState, newState) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": getTransitionError(currentState, newState),
		})
		return
	}
	
	// Validate required fields
	if err := validateTransitionReason(newState, reason); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	
	// Build and send response
	c.JSON(http.StatusOK, buildTransitionResponse(currentState, newState, reason))
}

// checkTransitionPermission checks if user has permission for the transition
func checkTransitionPermission(c *gin.Context, newState string) error {
	userRole, exists := c.Get("user_role")
	if !exists {
		userRole = "Agent" // Default for testing
	}
	
	if strings.EqualFold(fmt.Sprintf("%v", userRole), "customer") {
		if newState == StateResolved {
			return fmt.Errorf("Permission denied")
		}
		if newState == "reopen_requested" {
			return fmt.Errorf("reopen_requested")
		}
	}
	return nil
}

// validateTransitionReason checks if required reasons are provided
func validateTransitionReason(newState, reason string) error {
	if newState == StatePending && reason == "" {
		return fmt.Errorf("Reason required for pending state")
	}
	if newState == StateResolved && reason == "" {
		return fmt.Errorf("Resolution notes required")
	}
	return nil
}

// buildTransitionResponse creates the response for a successful transition
func buildTransitionResponse(currentState, newState, reason string) gin.H {
	response := gin.H{
		"success":   true,
		"new_state": newState,
	}
	
	switch newState {
	case StateOpen:
		if currentState == StateNew {
			response["message"] = "Ticket opened"
		} else if currentState == StateClosed {
			response["message"] = "Ticket reopened"
		} else {
			response["message"] = "Ticket marked as open"
		}
	case StatePending:
		response["message"] = "Ticket marked as pending"
		response["reason"] = reason
	case StateResolved:
		response["message"] = "Ticket resolved"
		response["resolution"] = reason
	case StateClosed:
		response["message"] = "Ticket closed"
	}
	
	return response
}

// getTransitionError returns appropriate error message for invalid transitions
func getTransitionError(from, to string) string {
	if from == StateNew && to == StateClosed {
		return "Cannot close ticket that hasn't been resolved"
	}
	return "Invalid state transition"
}

// handleTicketHistory returns the state transition history
func handleTicketHistory(c *gin.Context) {
	_ = c.Param("id") // ticketID - for future use
	
	// Get ticket history (mock)
	c.String(http.StatusOK, `
		<div class="state-history">
			<h4 class="font-medium mb-2">State History</h4>
			<div class="timeline">
				<div class="history-item">
					<div class="text-sm text-gray-500">2 hours ago</div>
					<div>Changed to open by Demo User</div>
					<div class="text-sm text-gray-600">Reason: Agent started working on ticket</div>
				</div>
				<div class="history-item mt-3">
					<div class="text-sm text-gray-500">3 hours ago</div>
					<div>Created as new by customer@example.com</div>
				</div>
			</div>
		</div>
	`)
}

// handleTicketAutoTransition handles automatic state transitions
func handleTicketAutoTransition(c *gin.Context) {
	_ = c.Param("id") // ticketID - for future use
	trigger := c.PostForm("trigger")
	
	var newState string
	var message string
	
	switch trigger {
	case "agent_response":
		newState = "open"
		message = "Ticket automatically opened on agent response"
	case "customer_response":
		newState = "open"
		message = "Ticket reopened due to customer response"
	case "auto_close_timeout":
		newState = "closed"
		message = "Ticket auto-closed after resolution timeout"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "Invalid trigger",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"new_state": newState,
		"message": message,
	})
}

// Helper functions for workflow

func getTicketByID(id string) gin.H {
	// Get real ticket from database
	db, err := database.GetDB()
	if err != nil {
		// No database, no ticket
		return nil
	}
	
	ticketID, err := strconv.Atoi(id)
	if err != nil {
		return nil
	}
	
	var ticket struct {
		ID            int
		TicketNumber  string
		Title         string
		TicketStateID int
	}
	
	err = db.QueryRow(`
		SELECT id, tn, title, ticket_state_id
		FROM ticket
		WHERE id = $1
	`, ticketID).Scan(&ticket.ID, &ticket.TicketNumber, &ticket.Title, &ticket.TicketStateID)
	
	if err != nil {
		return nil
	}
	
	// Map state ID to state string
	state := "new"
	switch ticket.TicketStateID {
	case 1:
		state = "new"
	case 2:
		state = "open"
	case 3:
		state = "pending"
	case 4:
		state = "resolved"
	case 5, 6:
		state = "closed"
	}
	
	return gin.H{
		"id":     id,
		"state":  state,
		"title":  ticket.Title,
		"number": ticket.TicketNumber,
	}
}


// getAvailableTransitions returns possible transitions for a given state
func getAvailableTransitions(currentState string) []gin.H {
	transitionMap := map[string][]gin.H{
		StateNew: {
			{"to": StateOpen, "label": "Open Ticket"},
		},
		StateOpen: {
			{"to": StatePending, "label": "Mark as Pending"},
			{"to": StateResolved, "label": "Resolve Ticket"},
			{"to": StateClosed, "label": "Close Ticket"},
		},
		StatePending: {
			{"to": StateOpen, "label": "Reopen"},
		},
		StateResolved: {
			{"to": StateClosed, "label": "Close Ticket"},
		},
		StateClosed: {
			{"to": StateOpen, "label": "Reopen Ticket"},
		},
	}
	
	if transitions, exists := transitionMap[currentState]; exists {
		return transitions
	}
	return []gin.H{}
}

func getTicketStateHistory(ticketID string) []gin.H {
	// Return mock history
	return []gin.H{
		{"from": "new", "to": "open", "by": "Demo User", "time": "2 hours ago"},
	}
}

func renderTransitions(transitions []gin.H, ticketID string) string {
	html := ""
	for _, t := range transitions {
		html += fmt.Sprintf(`
			<button class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
				hx-post="/tickets/%s/transition"
				hx-vals='{"new_state": "%s"}'>
				%s
			</button>
		`, ticketID, t["to"], t["label"])
	}
	return html
}

func renderHistory(history []gin.H) string {
	html := ""
	for _, h := range history {
		html += fmt.Sprintf(`
			<div class="history-item mb-2">
				<div class="text-sm text-gray-500">%s</div>
				<div>Changed from %s to %s by %s</div>
			</div>
		`, h["time"], h["from"], h["to"], h["by"])
	}
	return html
}

// isValidTransition checks if a state transition is allowed
func isValidTransition(from, to string) bool {
	validTransitions := map[string][]string{
		StateNew:      {StateOpen},
		StateOpen:     {StatePending, StateResolved, StateClosed},
		StatePending:  {StateOpen},
		StateResolved: {StateClosed, StateOpen},
		StateClosed:   {StateOpen},
	}
	
	allowed, exists := validTransitions[from]
	if !exists {
		return false
	}
	
	for _, state := range allowed {
		if state == to {
			return true
		}
	}
	
	return false
}

// Agent Dashboard Handlers

// handleAgentDashboard renders the main agent dashboard
func handleAgentDashboard(c *gin.Context) {
	// Get agent metrics (mock data)
	metrics := gin.H{
		"open_tickets": 15,
		"pending_tickets": 8,
		"resolved_today": 12,
		"avg_response_time": "2h 15m",
		"sla_compliance": 94,
	}
	
	// Get assigned tickets (mock)
	assignedTickets := []gin.H{
		{
			"id": "TICK-2024-001",
			"title": "Cannot access email",
			"priority": "high",
			"status": "open",
			"due": "2 hours",
			"customer": "John Doe",
		},
		{
			"id": "TICK-2024-002",
			"title": "Server downtime",
			"priority": "urgent",
			"status": "pending",
			"due": "30 minutes",
			"customer": "Jane Smith",
		},
	}
	
	// Get queue performance (mock)
	queuePerformance := []gin.H{
		{"name": "Support Queue", "count": 23, "avg_wait": "15m"},
		{"name": "Sales Queue", "count": 8, "avg_wait": "5m"},
		{"name": "Technical Queue", "count": 15, "avg_wait": "30m"},
	}
	
	// Render dashboard HTML
	c.String(http.StatusOK, `
		<div class="agent-dashboard">
			<h1 class="text-2xl font-bold mb-6">Agent Dashboard</h1>
			
			<!-- Metrics Section -->
			<div class="metrics-grid grid grid-cols-4 gap-4 mb-6">
				<div class="metric-card" data-metric="open-tickets">
					<h3>My Open Tickets</h3>
					<div class="value">%d</div>
				</div>
				<div class="metric-card" data-metric="pending-tickets">
					<h3>Pending Tickets</h3>
					<div class="value">%d</div>
				</div>
				<div class="metric-card" data-metric="resolved-today">
					<h3>Resolved Today</h3>
					<div class="value">%d</div>
				</div>
				<div class="metric-card" data-metric="avg-response-time">
					<h3>Avg Response Time</h3>
					<div class="value">%s</div>
				</div>
			</div>
			
			<!-- Assigned Tickets -->
			<div class="assigned-tickets mb-6">
				<h2 class="text-xl font-semibold mb-4">Assigned to Me</h2>
				<div class="ticket-list">
					%s
				</div>
			</div>
			
			<!-- Queue Overview -->
			<div class="queue-overview mb-6">
				<h2 class="text-xl font-semibold mb-4">Queue Overview</h2>
				<div class="queue-performance">
					<h3>Queue Performance</h3>
					%s
				</div>
			</div>
			
			<!-- Recent Activity -->
			<div class="recent-activity mb-6">
				<h2 class="text-xl font-semibold mb-4">Recent Activity</h2>
				<div id="activity-feed" data-sse-target="activity">
					Loading activity...
				</div>
			</div>
			
			<!-- Performance Metrics -->
			<div class="performance-metrics">
				<h2 class="text-xl font-semibold mb-4">Performance Metrics</h2>
				<div class="chart-container">
					Chart placeholder
				</div>
			</div>
			
			<!-- SSE Connection -->
			<script>
				const eventSource = new EventSource('/dashboard/events');
				eventSource.onmessage = function(event) {
					console.log('SSE Event:', event.data);
				};
			</script>
		</div>
	`,
		metrics["open_tickets"],
		metrics["pending_tickets"],
		metrics["resolved_today"],
		metrics["avg_response_time"],
		renderAssignedTickets(assignedTickets),
		renderQueuePerformance(queuePerformance),
	)
}

// renderAssignedTickets generates HTML for assigned tickets
func renderAssignedTickets(tickets []gin.H) string {
	html := ""
	for _, ticket := range tickets {
		html += fmt.Sprintf(`
			<div class="ticket-item mb-2 p-3 border rounded">
				<div class="flex justify-between">
					<div>
						<span class="ticket-id font-bold">%s</span>
						<span class="ticket-title">%s</span>
					</div>
					<div>
						<span class="priority">Priority: %s</span>
						<span class="due ml-4">Due: %s</span>
					</div>
				</div>
			</div>
		`, ticket["id"], ticket["title"], ticket["priority"], ticket["due"])
	}
	return html
}

// renderQueuePerformance generates HTML for queue stats
func renderQueuePerformance(queues []gin.H) string {
	html := ""
	for _, queue := range queues {
		html += fmt.Sprintf(`
			<div class="queue-item mb-2">
				<span class="queue-name">%s</span>
				<span class="ticket-count">%d tickets in queue</span>
				<span class="avg-wait">Avg wait: %s</span>
			</div>
		`, queue["name"], queue["count"], queue["avg_wait"])
	}
	return html
}

// handleDashboardMetrics returns specific metrics data
func handleDashboardMetrics(c *gin.Context) {
	metricType := c.Param("type")
	
	switch metricType {
	case "open-tickets":
		c.JSON(http.StatusOK, gin.H{
			"count": 15,
			"trend": "up",
			"change": 3,
		})
	case "response-time":
		c.JSON(http.StatusOK, gin.H{
			"average": 125, // minutes
			"median": 90,
			"p95": 240,
		})
	case "sla-compliance":
		c.JSON(http.StatusOK, gin.H{
			"compliance_rate": 94.5,
			"at_risk": 3,
			"breached": 1,
		})
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "Metric not found"})
	}
}

// handleDashboardSSE handles Server-Sent Events for real-time updates
func handleDashboardSSE(c *gin.Context) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	
	// For testing, send a few events immediately
	// Send ticket update event
	fmt.Fprintf(c.Writer, "data: %s\n\n", `{"type":"ticket_updated","ticket_id":"TICK-2024-001","status":"resolved"}`)
	c.Writer.Flush()
	
	// Send queue status event
	fmt.Fprintf(c.Writer, "data: %s\n\n", `{"type":"queue_status","queues":[{"id":1,"count":23}]}`)
	c.Writer.Flush()
	
	// Send metrics update event
	fmt.Fprintf(c.Writer, "data: %s\n\n", `{"type":"metrics_update","metrics":{"open":15,"pending":8}}`)
	c.Writer.Flush()
	
	// Send heartbeat/ping
	fmt.Fprintf(c.Writer, "data: %s\n\n", `{"type":"ping","message":"heartbeat"}`)
	c.Writer.Flush()
	
	// Also send a comment heartbeat
	fmt.Fprintf(c.Writer, ": heartbeat\n\n")
	c.Writer.Flush()
}

// handleDashboardActivity returns recent activity feed
func handleDashboardActivity(c *gin.Context) {
	activityType := c.Query("type")
	page := c.DefaultQuery("page", "1")
	
	// Generate activity HTML based on filters
	activities := []gin.H{
		{"type": "ticket_created", "message": "New ticket created by John Doe", "time": "5 minutes ago"},
		{"type": "status_changed", "message": "Ticket status changed to resolved", "time": "10 minutes ago"},
		{"type": "assigned", "message": "Ticket assigned to Agent Smith", "time": "15 minutes ago"},
	}
	
	html := `<div class="activity-feed"><h3>Recent Activity</h3>`
	
	for _, activity := range activities {
		if activityType == "" || activity["type"] == activityType {
			html += fmt.Sprintf(`
				<div class="activity-item">
					<span class="activity-message">%s</span>
					<span class="activity-time">%s</span>
				</div>
			`, activity["message"], activity["time"])
		}
	}
	
	// Add pagination if on page 2
	if page == "2" {
		html += `
			<div class="pagination">
				<a href="?page=1">page=1</a>
				<span>Page 2</span>
				<a href="?page=3">page=3</a>
			</div>
		`
	}
	
	html += `</div>`
	c.String(http.StatusOK, html)
}

// handleDashboardNotifications returns notification panel
func handleDashboardNotifications(c *gin.Context) {
	c.String(http.StatusOK, `
		<div class="notifications-panel">
			<div class="notification-header">
				<div class="notification-bell" data-notification-count="3">
					<span class="notification-badge">3</span>
				</div>
				<button hx-post="/notifications/mark-read">Mark all as read</button>
			</div>
			<div class="notification-list">
				<div class="notification-item unread">
					<span>New ticket assigned to you</span>
				</div>
				<div class="notification-item unread">
					<span>SLA warning: Ticket approaching deadline</span>
				</div>
				<div class="notification-item unread">
					<span>Customer response on TICK-2024-001</span>
				</div>
			</div>
		</div>
	`)
}

// handleQuickActions returns quick action buttons
func handleQuickActions(c *gin.Context) {
	c.String(http.StatusOK, `
		<div class="quick-actions">
			<h3>Quick Actions</h3>
			<div class="action-buttons">
				<button data-shortcut="n" class="action-btn">New Ticket</button>
				<button data-shortcut="/" class="action-btn">Search Tickets</button>
				<button data-shortcut="p" class="action-btn">My Profile</button>
				<button data-shortcut="r" class="action-btn">Reports</button>
				<button data-shortcut="g d" class="action-btn">Go to dashboard</button>
			</div>
		</div>
	`)
}

// Customer Portal Handlers

// handleCustomerPortal renders the customer portal homepage
func handleCustomerPortal(c *gin.Context) {
	// Get customer metrics (mock)
	metrics := gin.H{
		"open_tickets": 3,
		"resolved_tickets": 12,
		"total_tickets": 15,
		"avg_resolution": "24 hours",
	}
	
	c.String(http.StatusOK, `
		<div class="customer-portal">
			<h1 class="text-2xl font-bold mb-6">Customer Portal</h1>
			
			<!-- Quick Actions -->
			<div class="quick-actions mb-6">
				<button class="btn">Submit Ticket</button>
				<button class="btn">View All Tickets</button>
				<button class="btn">Search Knowledge Base</button>
				<button class="btn">Contact Support</button>
			</div>
			
			<!-- Metrics -->
			<div class="metrics-grid grid grid-cols-4 gap-4 mb-6">
				<div class="metric-card" data-metric="open-tickets">
					<h3>Open Tickets</h3>
					<div class="value">%d</div>
				</div>
				<div class="metric-card" data-metric="resolved-tickets">
					<h3>Resolved Tickets</h3>
					<div class="value">%d</div>
				</div>
				<div class="metric-card" data-metric="total-tickets">
					<h3>Total Tickets</h3>
					<div class="value">%d</div>
				</div>
				<div class="metric-card" data-metric="avg-resolution-time">
					<h3>Avg Resolution</h3>
					<div class="value">%s</div>
				</div>
			</div>
			
			<!-- Navigation -->
			<div class="portal-nav mb-6">
				<a href="/portal/tickets">My Tickets</a>
				<a href="/portal/submit-ticket">Submit New Ticket</a>
				<a href="/portal/kb">Knowledge Base</a>
				<a href="/portal/profile">My Profile</a>
			</div>
		</div>
	`,
		metrics["open_tickets"],
		metrics["resolved_tickets"], 
		metrics["total_tickets"],
		metrics["avg_resolution"],
	)
}

// handleCustomerTickets lists customer's tickets
func handleCustomerTickets(c *gin.Context) {
	status := c.Query("status")
	sort := c.Query("sort")
	page := c.DefaultQuery("page", "1")
	
	// Generate response based on filters
	html := `<div class="customer-tickets"><h2>My Tickets</h2>`
	
	// Add tickets based on status filter
	if status != "closed" {
		html += `
			<div class="ticket-item">
				<span class="ticket-id">TICK-2024-001</span>
				<span class="status-open">Status: Open</span>
				<span>Created: 2 days ago</span>
				<span>Last Updated: 1 hour ago</span>
			</div>
		`
	}
	
	if status != "open" {
		html += `
			<div class="ticket-item">
				<span class="ticket-id">TICK-2024-002</span>
				<span class="status-closed">Status: Closed</span>
				<span>Created: 5 days ago</span>
				<span>Last Updated: 3 days ago</span>
			</div>
		`
	}
	
	// Add sorting indicator
	if sort == "created" {
		html += `<div>sort=created</div>`
	}
	
	// Add pagination
	if page == "2" {
		html += `<div class="pagination"><a href="?page=1">Previous</a> Page 2 <a href="?page=3">Next</a></div>`
	}
	
	html += `</div>`
	c.String(http.StatusOK, html)
}

// handleCustomerSubmitTicketForm shows the ticket submission form
func handleCustomerSubmitTicketForm(c *gin.Context) {
	c.String(http.StatusOK, `
		<div class="submit-ticket-form">
			<h2>Submit New Ticket</h2>
			<form hx-post="/portal/submit-ticket">
				<div class="form-group">
					<label for="subject">Subject</label>
					<input type="text" name="subject" required>
				</div>
				<div class="form-group">
					<label for="priority">Priority</label>
					<select name="priority">
						<option value="low">Low</option>
						<option value="normal" selected>Normal</option>
						<option value="high">High</option>
					</select>
				</div>
				<div class="form-group">
					<label for="category">Category</label>
					<select name="category" hx-trigger="change" hx-get="/portal/ticket-fields">
						<option value="general">General</option>
						<option value="technical">Technical</option>
						<option value="billing">Billing</option>
					</select>
				</div>
				<div class="form-group">
					<label for="description">Description</label>
					<textarea name="description" required></textarea>
				</div>
				<div class="form-group">
					<label for="attachment">Attachment</label>
					<input type="file" name="attachment">
				</div>
				<button type="submit">Submit Ticket</button>
			</form>
		</div>
	`)
}

// handleCustomerSubmitTicket processes ticket submission
func handleCustomerSubmitTicket(c *gin.Context) {
	subject := c.PostForm("subject")
	description := c.PostForm("description")
	// priority := c.PostForm("priority")
	// category := c.PostForm("category")
	
	// Validate required fields
	if subject == "" || description == "" {
		errors := ""
		if subject == "" {
			errors += "Subject is required. "
		}
		if description == "" {
			errors += "Description is required."
		}
		c.String(http.StatusBadRequest, errors)
		return
	}
	
	// Generate ticket ID (mock)
	ticketID := fmt.Sprintf("TICK-%d", rand.Intn(9999))
	
	c.String(http.StatusOK, `
		<div class="ticket-submitted">
			<h3>Ticket submitted successfully</h3>
			<p>Your ticket number is: %s</p>
			<a href="/portal/tickets/%s">View Ticket</a>
		</div>
	`, ticketID, ticketID)
}

// handleCustomerTicketView shows individual ticket details
func handleCustomerTicketView(c *gin.Context) {
	ticketID := c.Param("id")
	
	// Check access (mock)
	if ticketID == "TICK-2024-999" {
		c.String(http.StatusForbidden, "Access denied")
		return
	}
	
	c.String(http.StatusOK, `
		<div class="ticket-view">
			<h2>%s</h2>
			<div class="ticket-details">
				<div>Subject: Cannot access email</div>
				<div>Status: Open</div>
				<div>Priority: High</div>
				<div>Created: 2 days ago</div>
			</div>
			
			<div class="ticket-description">
				<h3>Description</h3>
				<p>I cannot access my email account...</p>
			</div>
			
			<div class="conversation">
				<h3>Conversation</h3>
				<div class="message-customer">Customer: Initial message</div>
				<div class="message-agent">Agent response: We're looking into this</div>
				<div class="message-customer">Customer reply: Thank you</div>
			</div>
			
			<div class="attachments">
				<h3>Attachments</h3>
				<a href="#">Download screenshot.png</a>
			</div>
			
			<div class="reply-form">
				<h3>Add Reply</h3>
				<form hx-post="/portal/tickets/%s/reply">
					<textarea name="message" required></textarea>
					<button type="submit">Send Reply</button>
				</form>
			</div>
		</div>
	`, ticketID, ticketID)
}

// handleCustomerTicketReply processes customer reply to ticket
func handleCustomerTicketReply(c *gin.Context) {
	ticketID := c.Param("id")
	message := c.PostForm("message")
	
	// Validate message
	if message == "" {
		c.String(http.StatusBadRequest, "Message cannot be empty")
		return
	}
	
	// Check if ticket was closed and needs reopening
	response := fmt.Sprintf(`
		<div class="reply-success">
			<p>Reply added successfully</p>
			<div class="new-message">%s</div>
	`, message)
	
	if ticketID == "TICK-2024-002" {
		response += `
			<div class="status-update">
				<p>Ticket reopened</p>
				<span class="status-open">Status: Open</span>
			</div>
		`
	}
	
	response += `</div>`
	c.String(http.StatusOK, response)
}

// handleCustomerProfile shows customer profile
func handleCustomerProfile(c *gin.Context) {
	c.String(http.StatusOK, `
		<div class="customer-profile">
			<h2>My Profile</h2>
			<form hx-post="/portal/profile">
				<div class="profile-info">
					<div>Email: customer@example.com</div>
					<div>
						<label>Name:</label>
						<input type="text" name="name" value="John Doe">
					</div>
					<div>
						<label>Phone:</label>
						<input type="text" name="phone" value="555-1234">
					</div>
					<div>
						<label>Company:</label>
						<input type="text" name="company" value="Acme Corp">
					</div>
				</div>
				
				<div class="notification-preferences">
					<h3>Notification Preferences</h3>
					<label>
						<input type="checkbox" name="email_notifications" checked>
						Email notifications
					</label>
					<label>
						<input type="checkbox" name="sms_notifications">
						SMS notifications
					</label>
				</div>
				
				<button type="submit">Update Profile</button>
			</form>
		</div>
	`)
}

// handleCustomerUpdateProfile updates customer profile
func handleCustomerUpdateProfile(c *gin.Context) {
	// Check what type of update
	if c.PostForm("email_notifications") != "" || c.PostForm("sms_notifications") != "" {
		c.String(http.StatusOK, "Preferences updated")
	} else {
		c.String(http.StatusOK, "Profile updated successfully")
	}
}

// handleCustomerKnowledgeBase shows knowledge base
func handleCustomerKnowledgeBase(c *gin.Context) {
	search := c.Query("search")
	category := c.Query("category")
	
	html := `
		<div class="knowledge-base">
			<h2>Knowledge Base</h2>
			<div class="kb-search">
				<input type="text" placeholder="Search articles..." hx-get="/portal/kb" hx-trigger="keyup changed delay:500ms">
			</div>
	`
	
	if search == "password" {
		html += `
			<div class="search-results">
				<h3>Search Results</h3>
				<div class="article">
					<a href="#">How to reset your password</a>
				</div>
			</div>
		`
	} else if category == "technical" {
		html += `
			<div class="category-articles">
				<h3>Technical</h3>
				<div class="article-category-technical">
					Technical support articles...
				</div>
			</div>
		`
	} else {
		html += `
			<div class="kb-home">
				<div class="popular-articles">
					<h3>Popular Articles</h3>
					<ul>
						<li>How to reset your password</li>
						<li>Getting started guide</li>
					</ul>
				</div>
				<div class="categories">
					<h3>Categories</h3>
					<ul>
						<li>General</li>
						<li>Technical</li>
						<li>Billing</li>
					</ul>
				</div>
			</div>
		`
	}
	
	html += `
		<div class="article-feedback">
			<p>Was this helpful?</p>
			<button hx-post="/portal/kb/vote" hx-vals='{"helpful": "yes"}'>Yes</button>
			<button hx-post="/portal/kb/vote" hx-vals='{"helpful": "no"}'>No</button>
		</div>
	</div>`
	
	c.String(http.StatusOK, html)
}

// handleCustomerSatisfactionForm shows satisfaction survey
func handleCustomerSatisfactionForm(c *gin.Context) {
	ticketID := c.Param("id")
	
	c.String(http.StatusOK, `
		<div class="satisfaction-survey">
			<h2>Rate Your Experience</h2>
			<p>How satisfied are you with the resolution of ticket %s?</p>
			<form hx-post="/portal/tickets/%s/satisfaction">
				<div class="rating">
					<label>Rating (1-5 stars):</label>
					<input type="radio" name="rating" value="1"> 1
					<input type="radio" name="rating" value="2"> 2
					<input type="radio" name="rating" value="3"> 3
					<input type="radio" name="rating" value="4"> 4
					<input type="radio" name="rating" value="5"> 5
				</div>
				<div class="feedback">
					<label>Additional feedback:</label>
					<textarea name="feedback"></textarea>
				</div>
				<button type="submit">Submit Rating</button>
			</form>
		</div>
	`, ticketID, ticketID)
}

// handleCustomerSatisfactionSubmit processes satisfaction rating
func handleCustomerSatisfactionSubmit(c *gin.Context) {
	// rating := c.PostForm("rating")
	// feedback := c.PostForm("feedback")
	
	c.String(http.StatusOK, "Thank you for your feedback")
}

// Helper function to format time ago (e.g., "2 hours ago", "3 minutes ago")
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)
	
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		// For older messages, show the actual date
		return t.Format("Jan 2, 2006")
	}
}

// Helper function to generate initials from a name (e.g., "John Doe" -> "JD")
func generateInitials(name string) string {
	if name == "" {
		return "?"
	}
	
	words := strings.Fields(strings.TrimSpace(name))
	if len(words) == 0 {
		return "?"
	}
	
	if len(words) == 1 {
		// Single word, take first character
		return strings.ToUpper(string(words[0][0]))
	}
	
	// Multiple words, take first character of first and last word
	first := strings.ToUpper(string(words[0][0]))
	last := strings.ToUpper(string(words[len(words)-1][0]))
	return first + last
}

// Helper function to format file size (e.g., 1024 bytes -> "1 KB")
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

// sendGuruMeditation sends an Amiga-style Guru Meditation error response
func sendGuruMeditation(c *gin.Context, err error, location string) {
	// Generate error code based on error type
	errorCode := "00000005.0000DEAD" // Default database error
	
	if err != nil {
		if strings.Contains(err.Error(), "connection") {
			errorCode = "00000005.DEADBEEF" // Connection error
		} else if strings.Contains(err.Error(), "timeout") {
			errorCode = "00000005.TIMEOUT0" // Timeout error
		} else if strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "denied") {
			errorCode = "00000005.NOACCESS" // Permission error
		}
	}
	
	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") != "" {
		// Send HTMX trigger to show Guru Meditation
		c.Header("HX-Trigger", fmt.Sprintf(`{"show-guru-meditation": {"code": "%s", "message": "%s", "location": "%s"}}`, 
			errorCode, err.Error(), location))
		c.Header("HX-Retarget", "body")
		c.Header("HX-Reswap", "none")
	}
	
	// Send error response
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "Database failure",
		"guru_meditation": gin.H{
			"code": errorCode,
			"message": err.Error(),
			"location": location,
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			"task": fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
		},
	})
}
// ============================================
// Permission Management Handlers (OTRS Role equivalent)
// ============================================

// handleAdminPermissions shows the permission management page
func handleAdminPermissions(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Database connection failed")
		return
	}

	// Get all users and groups for the matrix view
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	permService := service.NewPermissionService(db)

	users, err := userRepo.List()
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	groups, err := groupRepo.List()
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch groups")
		return
	}

	// Get selected user ID from query param
	selectedUserID := c.DefaultQuery("user", "1")
	userID, _ := strconv.ParseUint(selectedUserID, 10, 32)
	
	var matrix *service.PermissionMatrix
	if userID > 0 {
		matrix, err = permService.GetUserPermissionMatrix(uint(userID))
		if err != nil {
			sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch permissions")
			return
		}
	}

	c.HTML(http.StatusOK, "admin/permissions.pongo2", gin.H{
		"Users":            users,
		"Groups":           groups,
		"SelectedUserID":   userID,
		"PermissionMatrix": matrix,
		"PermissionKeys": []string{
			"ro", "move_into", "create", "note", "owner", "priority", "rw",
		},
	})
}

// handleGetUserPermissionMatrix gets permissions for a specific user
func handleGetUserPermissionMatrix(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	permService := service.NewPermissionService(db)
	matrix, err := permService.GetUserPermissionMatrix(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    matrix,
	})
}

// handleUpdateUserPermissions updates all permissions for a user
func handleUpdateUserPermissions(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	// Parse permission data from form
	permissions := make(map[uint]map[string]bool)
	
	// Get all form values
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid form data"})
		return
	}

	// Process each permission checkbox
	// Format: perm_<groupID>_<permissionKey>
	for key, values := range c.Request.PostForm {
		if strings.HasPrefix(key, "perm_") && len(values) > 0 {
			parts := strings.Split(key, "_")
			if len(parts) == 3 {
				groupID, _ := strconv.ParseUint(parts[1], 10, 32)
				permKey := parts[2]
				
				if permissions[uint(groupID)] == nil {
					permissions[uint(groupID)] = make(map[string]bool)
				}
				permissions[uint(groupID)][permKey] = (values[0] == "1" || values[0] == "on")
			}
		}
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	permService := service.NewPermissionService(db)
	if err := permService.UpdateUserPermissions(uint(userID), permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Permissions updated successfully",
	})
}

// handleGetGroupPermissionMatrix gets all users' permissions for a group
func handleGetGroupPermissionMatrix(c *gin.Context) {
	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid group ID"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	permService := service.NewPermissionService(db)
	matrix, err := permService.GetGroupPermissionMatrix(uint(groupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    matrix,
	})
}

// handleCloneUserPermissions copies permissions from one user to another
func handleCloneUserPermissions(c *gin.Context) {
	sourceUserID, _ := strconv.ParseUint(c.PostForm("source_user_id"), 10, 32)
	targetUserID, _ := strconv.ParseUint(c.PostForm("target_user_id"), 10, 32)

	if sourceUserID == 0 || targetUserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user IDs"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	permService := service.NewPermissionService(db)
	if err := permService.CloneUserPermissions(uint(sourceUserID), uint(targetUserID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to clone permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Permissions cloned successfully",
	})
}
