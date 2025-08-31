package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// HandleDeletePriorityAPI handles DELETE /api/v1/priorities/:id
func HandleDeletePriorityAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse priority ID
	priorityID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority ID"})
		return
	}

	// Protect system priorities (IDs 1-5 are typically system priorities)
	if priorityID <= 5 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete system priority"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Check if priority exists
	var count int
	checkQuery := database.ConvertPlaceholders(`
		SELECT 1 FROM ticket_priority
		WHERE id = $1 AND valid_id = 1
	`)
	db.QueryRow(checkQuery, priorityID).Scan(&count)
	if count != 1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Priority not found"})
		return
	}

	// Check if priority is used by any tickets
	var ticketCount int
	ticketQuery := database.ConvertPlaceholders(`
		SELECT COUNT(*) FROM tickets 
		WHERE ticket_priority_id = $1
	`)
	db.QueryRow(ticketQuery, priorityID).Scan(&ticketCount)
	if ticketCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Priority is in use",
			"message": "Cannot delete priority that is assigned to tickets",
			"ticket_count": ticketCount,
		})
		return
	}

	// Soft delete priority (OTRS style - set valid_id = 2)
	deleteQuery := database.ConvertPlaceholders(`
		UPDATE ticket_priority 
		SET valid_id = 2, change_time = NOW(), change_by = $1
		WHERE id = $2
	`)
	
	result, err := db.Exec(deleteQuery, userID, priorityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete priority"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete priority"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Priority deleted successfully",
		"id": priorityID,
	})
}