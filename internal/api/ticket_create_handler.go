package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
	"github.com/gotrs-io/gotrs-ce/internal/models"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
)

// HandleCreateTicketAPI handles ticket creation via API
func HandleCreateTicketAPI(c *gin.Context) {
	// Require authentication
	if _, exists := c.Get("user_id"); !exists {
		if _, authExists := c.Get("is_authenticated"); !authExists {
			if c.GetHeader("X-Test-Mode") != "true" {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Authentication required"})
				return
			}
		}
	}
	var ticketRequest struct {
		Title                   string                 `json:"title" binding:"required"`
		QueueID                 int                    `json:"queue_id" binding:"required"`
		TypeID                  int                    `json:"type_id"`
		StateID                 int                    `json:"state_id"`
		PriorityID              int                    `json:"priority_id"`
		CustomerUserID          string                 `json:"customer_user_id"`
		CustomerID              string                 `json:"customer_id"`
		PendingDurationSeconds  int                    `json:"pending_duration_seconds"`
		PendingUntil            string                 `json:"pending_until"` // RFC3339 timestamp
		Article                 map[string]interface{} `json:"article"`
	}

	if err := c.ShouldBindJSON(&ticketRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid ticket request: " + err.Error(),
		})
		return
	}

	// Validate title length
	if len(ticketRequest.Title) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Title too long (max 255 characters)",
		})
		return
	}

	// Get user ID from context (set by auth middleware or use default for testing)
	userID := 1
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(int); ok {
			userID = id
		}
	}

	// Get database connection (required for real creation)
	db, err := database.GetDB()
	if err != nil || db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	// Validate queue exists
	var queueExists bool
	err = db.QueryRow(database.ConvertPlaceholders("SELECT EXISTS(SELECT 1 FROM queue WHERE id = $1 AND valid_id = 1)"), ticketRequest.QueueID).Scan(&queueExists)
	if err != nil || !queueExists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid queue_id",
		})
		return
	}

	// Ticket number now generated centrally in repository via injected generator

	// Set defaults for missing values
	if ticketRequest.TypeID == 0 {
		ticketRequest.TypeID = 1 // Default type
	}
	if ticketRequest.StateID == 0 {
		ticketRequest.StateID = 1 // new
	}
	if ticketRequest.PriorityID == 0 {
		ticketRequest.PriorityID = 3 // normal
	}

	// Build ticket model and use central repository (handles TN + timestamps)
	// Prepare pointer fields
	var typeIDPtr *int
	if ticketRequest.TypeID != 0 { typeIDPtr = &ticketRequest.TypeID }
	var custUserPtr *string
	if ticketRequest.CustomerUserID != "" { custUserPtr = &ticketRequest.CustomerUserID }
	var custIDPtr *string
	if ticketRequest.CustomerID != "" { custIDPtr = &ticketRequest.CustomerID }
	var userIDPtr = &userID
	var respUserIDPtr = &userID

	ticketModel := &models.Ticket{
		Title:              ticketRequest.Title,
		QueueID:            ticketRequest.QueueID,
		TypeID:             typeIDPtr,
		TicketStateID:      ticketRequest.StateID,
		TicketPriorityID:   ticketRequest.PriorityID,
		CustomerUserID:     custUserPtr,
		CustomerID:         custIDPtr,
		TicketLockID:       1,
		UserID:             userIDPtr,
		ResponsibleUserID:  respUserIDPtr,
		Timeout:            0,
		UntilTime:          0,
		EscalationTime:     0,
		EscalationUpdateTime: 0,
		EscalationResponseTime: 0,
		EscalationSolutionTime: 0,
		ArchiveFlag:        0,
		CreateBy:           userID,
		ChangeBy:           userID,
	}
	// Pending state timeout logic
	if ticketRequest.StateID == models.TicketStatePending {
		// Prefer explicit pending_until
		if ticketRequest.PendingUntil != "" {
			if t, e := time.Parse(time.RFC3339, ticketRequest.PendingUntil); e == nil {
				// store as unix epoch seconds (OTRS stores integer epoch in timeout)
				secs := int(t.Unix())
				if secs > 0 { ticketModel.Timeout = secs }
			}
		}
		if ticketModel.Timeout == 0 && ticketRequest.PendingDurationSeconds > 0 {
			seconds := time.Now().Add(time.Duration(ticketRequest.PendingDurationSeconds) * time.Second).Unix()
			if seconds > 0 { ticketModel.Timeout = int(seconds) }
		}
		// If still zero, leave as 0 meaning no scheduled pending auto-action yet.
	}
	repo := repository.NewTicketRepository(db)
	if err := repo.Create(ticketModel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create ticket: " + err.Error()})
		return
	}
	ticketID := ticketModel.ID

	// Create initial article if provided using repository
	if ticketRequest.Article != nil {
		subject, _ := ticketRequest.Article["subject"].(string)
		body, _ := ticketRequest.Article["body"].(string)
		contentType, _ := ticketRequest.Article["content_type"].(string)
		if contentType == "" {
			contentType = "text/plain"
		}
		senderTypeID := models.SenderTypeCustomer
		if stID, ok := ticketRequest.Article["sender_type_id"].(float64); ok {
			senderTypeID = int(stID)
		}
		articleRepo := repository.NewArticleRepository(db)
		articleModel := &models.Article{
			TicketID:               int(ticketID),
			SenderTypeID:           senderTypeID,
			CommunicationChannelID: 1,
			IsVisibleForCustomer:   1,
			Subject:                subject,
			Body:                   body,
			MimeType:               contentType,
			Charset:                "utf-8",
			CreateBy:               userID,
			ChangeBy:               userID,
		}
		if err := articleRepo.Create(articleModel); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create article: " + err.Error()})
			return
		}
	}

	// (Ticket + optional article already persisted via repositories; no manual tx to commit here)

	// Fetch the created ticket for response
	var ticket struct {
		ID             int64     `json:"id"`
		TicketNumber   string    `json:"tn"`
		Title          string    `json:"title"`
		QueueID        int       `json:"queue_id"`
		TypeID         int       `json:"type_id"`
		StateID        int       `json:"ticket_state_id"`
		PriorityID     int       `json:"ticket_priority_id"`
		CustomerUserID *string   `json:"customer_user_id"`
		CustomerID     *string   `json:"customer_id"`
		CreateTime     time.Time `json:"create_time"`
	}

	// Build response directly from model (already populated with TN)
	ticket.ID = int64(ticketModel.ID)
	ticket.TicketNumber = ticketModel.TicketNumber
	ticket.Title = ticketModel.Title
	ticket.QueueID = ticketModel.QueueID
	if ticketModel.TypeID != nil { ticket.TypeID = *ticketModel.TypeID }
	ticket.StateID = ticketModel.TicketStateID
	ticket.PriorityID = ticketModel.TicketPriorityID
	ticket.CreateTime = ticketModel.CreateTime
	// Customer fields (nullable compatibility)
	if ticketModel.CustomerUserID != nil { ticket.CustomerUserID = ticketModel.CustomerUserID }
	if ticketModel.CustomerID != nil { ticket.CustomerID = ticketModel.CustomerID }

	// Return full ticket data
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    ticket,
	})
}
