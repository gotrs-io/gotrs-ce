package routing

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/middleware"
	"github.com/gotrs-io/gotrs-ce/internal/shared"
)

// RegisterExistingHandlers registers common handlers and middleware with the registry
func RegisterExistingHandlers(registry *HandlerRegistry) error {
	// Register basic handlers
	handlers := map[string]gin.HandlerFunc{
		// Health checks
		"handleHealthCheck": func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "healthy",
				"version": "1.0.0",
			})
		},
		
		"handleDetailedHealthCheck": func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
				"components": gin.H{
					"database": "unknown", // Would check actual DB
					"cache":    "healthy",
					"queue":    "healthy",
				},
			})
		},
		
		"handleMetrics": func(c *gin.Context) {
			// Placeholder for Prometheus metrics
			c.String(http.StatusOK, "# HELP gotrs_up GOTRS is up\n# TYPE gotrs_up gauge\ngotrs_up 1\n")
		},
		
		// Authentication handlers - redirect to traditional routes for now
		"handleLoginForm": func(c *gin.Context) {
			// For now, redirect to working login page
			c.Redirect(http.StatusFound, "/login")
		},
		"handleLogin": func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Login handler (needs proper implementation)",
				"method": c.Request.Method,
			})
		},
		"handleLogout": func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Logout successful",
			})
		},
		"handleTokenRefresh": createPlaceholderHandler("Token Refresh"),
		"handleVerifyAuth": createPlaceholderHandler("Verify Auth"),
		"handlePasswordResetForm": createPlaceholderHandler("Password Reset Form"),
		"handlePasswordReset": createPlaceholderHandler("Password Reset"),
		"handlePasswordChange": createPlaceholderHandler("Password Change"),
		
		// Admin handlers
		"handleAdminCustomerCompanies": createPlaceholderHandler("Admin Customer Companies"),
		"handleAdminNewCustomerCompany": createPlaceholderHandler("New Customer Company"),
		"handleAdminCreateCustomerCompany": createPlaceholderHandler("Create Customer Company"),
		"handleAdminEditCustomerCompany": createPlaceholderHandler("Edit Customer Company"),
		"handleAdminUpdateCustomerCompany": createPlaceholderHandler("Update Customer Company"),
		"handleAdminDeleteCustomerCompany": createPlaceholderHandler("Delete Customer Company"),
		"handleAdminCustomerCompanyUsers": createPlaceholderHandler("Customer Company Users"),
		"handleAdminCustomerCompanyTickets": createPlaceholderHandler("Customer Company Tickets"),
		"handleAdminCustomerCompanyServices": createPlaceholderHandler("Customer Company Services"),
		"handleAdminUpdateCustomerCompanyServices": createPlaceholderHandler("Update Customer Company Services"),
		"handleAdminCustomerPortalSettings": createPlaceholderHandler("Customer Portal Settings"),
		"handleAdminUpdateCustomerPortalSettings": createPlaceholderHandler("Update Customer Portal Settings"),
		"handleAdminUploadCustomerPortalLogo": createPlaceholderHandler("Upload Customer Portal Logo"),
		
		// Agent handlers
		"handleAgentDashboard": createPlaceholderHandler("Agent Dashboard"),
		"handleAgentTickets": createPlaceholderHandler("Agent Tickets"),
		"handleAgentTicketView": createPlaceholderHandler("Agent Ticket View"),
		"handleAgentTicketUpdate": createPlaceholderHandler("Agent Ticket Update"),
		"handleAgentAddNote": createPlaceholderHandler("Agent Add Note"),
		"handleAgentTicketReply": createPlaceholderHandler("Agent Ticket Reply"),
		"handleAgentAssignTicket": createPlaceholderHandler("Agent Assign Ticket"),
		"handleAgentCustomers": createPlaceholderHandler("Agent Customers"),
		"handleAgentCustomerView": createPlaceholderHandler("Agent Customer View"),
		"handleAgentCustomerTickets": createPlaceholderHandler("Agent Customer Tickets"),
		"handleAgentQueues": createPlaceholderHandler("Agent Queues"),
		"handleAgentQueueTickets": createPlaceholderHandler("Agent Queue Tickets"),
		"handleAgentStatistics": createPlaceholderHandler("Agent Statistics"),
		"handleAgentTicketStats": createPlaceholderHandler("Agent Ticket Stats"),
		"handleAgentSearch": createPlaceholderHandler("Agent Search"),
		"handleAgentMergeTickets": createPlaceholderHandler("Agent Merge Tickets"),
		"handleAgentSplitTicket": createPlaceholderHandler("Agent Split Ticket"),
		
		// Customer handlers
		"handleCustomerDashboard": createPlaceholderHandler("Customer Dashboard"),
		"handleCustomerTickets": createPlaceholderHandler("Customer Tickets"),
		"handleCustomerNewTicket": createPlaceholderHandler("Customer New Ticket"),
		"handleCustomerCreateTicket": createPlaceholderHandler("Customer Create Ticket"),
		"handleCustomerTicketView": createPlaceholderHandler("Customer Ticket View"),
		"handleCustomerTicketReply": createPlaceholderHandler("Customer Ticket Reply"),
		"handleCustomerCloseTicket": createPlaceholderHandler("Customer Close Ticket"),
		"handleCustomerProfile": createPlaceholderHandler("Customer Profile"),
		"handleCustomerUpdateProfile": createPlaceholderHandler("Customer Update Profile"),
		"handleCustomerPasswordForm": createPlaceholderHandler("Customer Password Form"),
		"handleCustomerChangePassword": createPlaceholderHandler("Customer Change Password"),
		"handleCustomerKnowledgeBase": createPlaceholderHandler("Customer Knowledge Base"),
		"handleCustomerKBSearch": createPlaceholderHandler("Customer KB Search"),
		"handleCustomerKBArticle": createPlaceholderHandler("Customer KB Article"),
		"handleCustomerCompanyInfo": createPlaceholderHandler("Customer Company Info"),
		"handleCustomerCompanyUsers": createPlaceholderHandler("Customer Company Users"),
	}
	
	// Register middleware
	middlewares := map[string]gin.HandlerFunc{
		"auth": func(c *gin.Context) {
			// Use shared JWT manager for authentication
			jwtManager := shared.GetJWTManager()
			middleware.SessionMiddleware(jwtManager)(c)
		},
		
		"admin": func(c *gin.Context) {
			role, exists := c.Get("user_role")
			if !exists || role != "Admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
				c.Abort()
				return
			}
			c.Next()
		},
		
		"agent": func(c *gin.Context) {
			role, exists := c.Get("user_role")
			if !exists || (role != "Agent" && role != "Admin") {
				c.JSON(http.StatusForbidden, gin.H{"error": "Agent access required"})
				c.Abort()
				return
			}
			c.Next()
		},
		
		"customer": func(c *gin.Context) {
			isCustomer, _ := c.Get("is_customer")
			if !isCustomer.(bool) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Customer access required"})
				c.Abort()
				return
			}
			c.Next()
		},
		
		"audit": func(c *gin.Context) {
			// Simple audit logging
			c.Next()
		},
		
		"cors": func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			
			c.Next()
		},
		
		"rateLimit": func(c *gin.Context) {
			// Placeholder rate limiting
			c.Next()
		},
		
		"i18n": func(c *gin.Context) {
			// Placeholder i18n middleware
			// TODO: Use actual i18n middleware
			c.Next()
		},
	}
	
	// Register all handlers
	if err := registry.RegisterBatch(handlers); err != nil {
		return err
	}
	
	// Register all middleware
	return registry.RegisterMiddlewareBatch(middlewares)
}

// createPlaceholderHandler creates a placeholder handler that shows which handler should be implemented
func createPlaceholderHandler(handlerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":     "Handler placeholder",
			"handler":     handlerName,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"description": "This is a placeholder - real handler needs to be implemented",
		})
	}
}