package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/constants"
	"github.com/gotrs-io/gotrs-ce/internal/core"
	"github.com/gotrs-io/gotrs-ce/internal/database"
	"github.com/gotrs-io/gotrs-ce/internal/models"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
)

// HandleAgentCreateTicket creates a new ticket from the agent interface
func HandleAgentCreateTicket(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get agent user info from context
		userID := c.GetUint("user_id")
		if userID == 0 {
			// Fallback for now if auth middleware not applied (development/testing)
			userID = 1
		}
		// username := c.GetString("username")

		// Get form data
		title := c.PostForm("subject") // Agent form uses 'subject' field name
		message := c.PostForm("body")
		queueID := c.PostForm("queue_id")
		priorityID := c.PostForm("priority")
		typeID := c.PostForm("type_id")
		stateID := c.PostForm("state_id")
		customerUserID := c.PostForm("customer_user_id")
		customerEmail := c.PostForm("customer_email")
		// customerName := c.PostForm("customer_name")
		customerID := c.PostForm("customer_id")

		// Validate required fields
		if title == "" || message == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Subject and message are required"})
			return
		}
		// Customer selection is optional: allow creating tickets without a customer user/email/id
		// If all customer fields are empty, we'll insert NULLs for customer_id and customer_user_id

		// Set defaults for agent-created tickets
		if queueID == "" {
			queueID = "1" // Default queue
		}
		if priorityID == "" {
			priorityID = "3" // Normal priority
		}
		if typeID == "" {
			typeID = "1" // Default type
		}
		if stateID == "" {
			stateID = "1" // New state
		}

		// Map textual priority codes (form values) to numeric IDs
		switch priorityID {
		case "very_low":
			priorityID = "1"
		case "low":
			priorityID = "2"
		case "normal":
			priorityID = "3"
		case "high":
			priorityID = "4"
		case "very_high":
			priorityID = "5"
		}

		// Manual TN generation removed; repository handles ticket number based on configured generator

		// Get database connection
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database connection failed"})
			return
		}

		// Parse numeric IDs
		queueIDInt, err := strconv.Atoi(queueID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue ID"})
			return
		}
		priorityIDInt, err := strconv.Atoi(priorityID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority ID"})
			return
		}
		typeIDInt, err := strconv.Atoi(typeID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type ID"})
			return
		}
		stateIDInt, err := strconv.Atoi(stateID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state ID"})
			return
		}

		// Handle customer assignment
		var customerIDValue sql.NullString
		var customerUserIDValue sql.NullString

		if customerID != "" {
			customerIDValue = sql.NullString{String: customerID, Valid: true}
		}
		// Handle customer user selection
		if customerUserID != "" {
			// Get customer info from the selected customer user
			var foundCustomerID sql.NullString
			err := db.QueryRow(database.ConvertPlaceholders(`
				SELECT customer_id
				FROM customer_user
				WHERE login = $1 AND valid_id = 1
			`), customerUserID).Scan(&foundCustomerID)

			if err == nil {
				customerUserIDValue = sql.NullString{String: customerUserID, Valid: true}
				if foundCustomerID.Valid {
					customerIDValue = foundCustomerID
				}
			} else {
				log.Printf("Error finding customer user %s: %v", customerUserID, err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer user selected"})
				return
			}
		}
		// Fallback to email lookup if customer_user_id is not provided (for backward compatibility)
		if customerEmail != "" && !customerUserIDValue.Valid {
			// Try to find existing customer user by email
			var foundCustomerID, foundCustomerUserID sql.NullString
			err := db.QueryRow(database.ConvertPlaceholders(`
				SELECT customer_id, login
				FROM customer_user
				WHERE email = $1 AND valid_id = 1
			`), customerEmail).Scan(&foundCustomerID, &foundCustomerUserID)

			if err == nil && foundCustomerUserID.Valid {
				// Found existing customer user
				customerUserIDValue = foundCustomerUserID
				customerIDValue = foundCustomerID
			} else {
				// Create new customer user - use email as login for now
				customerUserIDValue = sql.NullString{String: customerEmail, Valid: true}
			}
		}

		// Build ticket model and use repository (central generator + logging)
		var typePtr *int
		if typeIDInt != 0 { typePtr = &typeIDInt }
		var custIDPtr *string
		if customerIDValue.Valid { v := customerIDValue.String; custIDPtr = &v }
		var custUserPtr *string
		if customerUserIDValue.Valid { v := customerUserIDValue.String; custUserPtr = &v }
		var userIDInt = int(userID)
		ticketModel := &models.Ticket{
			Title:            title,
			QueueID:          queueIDInt,
			TicketLockID:     1,
			TypeID:           typePtr,
			UserID:           &userIDInt,
			ResponsibleUserID:&userIDInt,
			TicketPriorityID: priorityIDInt,
			TicketStateID:    stateIDInt,
			CustomerID:       custIDPtr,
			CustomerUserID:   custUserPtr,
			CreateBy:         userIDInt,
			ChangeBy:         userIDInt,
		}
		tRepo := repository.NewTicketRepository(db)
		if err := tRepo.Create(ticketModel); err != nil {
			log.Printf("Error creating ticket via repository: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
			return
		}
		ticketID := ticketModel.ID

		// Determine interaction / article type
		interaction := c.PostForm("interaction_type")
		// Resolve article type + visibility
		{
			articleRepo := repository.NewArticleRepository(db)
			intent := core.ArticleIntent{Interaction: constants.InteractionType(interaction), SenderTypeID: constants.ArticleSenderAgent}
			resolved, derr := core.DetermineArticleType(intent)
			if derr != nil {
				log.Printf("Article type resolution failed: %v", derr)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interaction type"})
				return
			}
			if verr := core.ValidateArticleCombination(resolved.ArticleTypeID, resolved.ArticleSenderTypeID, resolved.CustomerVisible); verr != nil {
				log.Printf("Article combination invalid: %v", verr)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article combination"})
				return
			}
			visibility := 0
			if resolved.CustomerVisible {
				visibility = 1
			}
			articleModel := &models.Article{
				TicketID:               ticketID,
				ArticleTypeID:          resolved.ArticleTypeID,
				SenderTypeID:           resolved.ArticleSenderTypeID,
				CommunicationChannelID: core.MapCommunicationChannel(resolved.ArticleTypeID),
				IsVisibleForCustomer:   visibility,
				Subject:                title,
				Body:                   message,
				MimeType:               detectTicketContentType(message),
				Charset:                "utf-8",
				CreateBy:               int(userID),
				ChangeBy:               int(userID),
			}
			if err := articleRepo.Create(articleModel); err != nil {
				log.Printf("Error creating initial article via repository: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create initial article"})
				return
			}
		}

		// Redirect to ticket view using repository-assigned ticket number
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/tickets/%s", ticketModel.TicketNumber))
	}
}

// detectTicketContentType determines the MIME type based on content analysis
func detectTicketContentType(content string) string {
	// Check for HTML tags
	if strings.Contains(content, "<") && strings.Contains(content, ">") {
		// Look for common HTML tags
		htmlTags := []string{"<p>", "<br", "<div>", "<span>", "<strong>", "<em>", "<b>", "<i>", "<h1>", "<h2>", "<h3>", "<ul>", "<ol>", "<li>"}
		for _, tag := range htmlTags {
			if strings.Contains(content, tag) {
				return "text/html"
			}
		}
	}
	
	// Check for markdown syntax
	if strings.Contains(content, "#") || strings.Contains(content, "**") || strings.Contains(content, "*") || strings.Contains(content, "`") {
		return "text/markdown"
	}
	
	// Default to plain text
	return "text/plain"
}
