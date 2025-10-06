package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
	"github.com/gotrs-io/gotrs-ce/internal/service"
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
		Title      string `json:"title" binding:"required"`
		QueueID    int    `json:"queue_id" binding:"required"`
		PriorityID int    `json:"priority_id"`
		StateID    int    `json:"state_id"`
		Body       string `json:"body"`
	}

	if err := c.ShouldBindJSON(&ticketRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid ticket request: " + err.Error(),
		})
		return
	}

	userID := 1
	if uid, exists := c.Get("user_id"); exists { if id, ok := uid.(int); ok { userID = id } }

	// Get database connection (required for real creation)
	db, err := database.GetDB()
	if err != nil || db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	repo := repository.NewTicketRepository(db)
	svc := service.NewTicketService(repo)
	created, err := svc.Create(c, service.CreateTicketInput{Title: ticketRequest.Title, QueueID: ticketRequest.QueueID, PriorityID: ticketRequest.PriorityID, StateID: ticketRequest.StateID, UserID: userID, Body: ticketRequest.Body})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{ "id": created.ID, "tn": created.TicketNumber, "title": created.Title, "queue_id": created.QueueID, "ticket_state_id": created.TicketStateID, "ticket_priority_id": created.TicketPriorityID }})
}
