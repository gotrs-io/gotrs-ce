package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// HandleDeleteQueueAPI handles DELETE /api/v1/queues/:id
func HandleDeleteQueueAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}

	// Parse queue ID
	queueID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid queue ID"})
		return
	}

	// Protect system queues (IDs 1-3 are typically system queues)
	if queueID <= 3 {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Cannot delete system queue"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	// Check if queue has tickets
	var ticketCount int
	ticketQuery := database.ConvertPlaceholders(`SELECT COUNT(*) FROM ticket WHERE queue_id = $1`)
	if err := db.QueryRow(ticketQuery, queueID).Scan(&ticketCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to check queue tickets"})
		return
	}

	if ticketCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "Cannot delete queue with existing tickets"})
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Remove group associations
	deleteGroupsQuery := database.ConvertPlaceholders(`
        DELETE FROM queue_group WHERE queue_id = $1
    `)
	if _, err := tx.Exec(deleteGroupsQuery, queueID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to remove group associations"})
		return
	}

	// Soft delete queue (OTRS style - set valid_id = 2)
	deleteQuery := database.ConvertPlaceholders(`
	        UPDATE queue 
	        SET valid_id = 2, change_time = NOW(), change_by = $2
	        WHERE id = $1
	    `)

	args := []interface{}{queueID, userID}
	if database.IsMySQL() {
		args = []interface{}{userID, queueID}
	}

	result, err := tx.Exec(deleteQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete queue"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete queue"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Queue not found"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Queue deleted successfully"})
}
