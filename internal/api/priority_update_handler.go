package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// HandleUpdatePriorityAPI handles PUT /api/v1/priorities/:id
func HandleUpdatePriorityAPI(c *gin.Context) {
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

	var req struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	// Update priority
	updateQuery := database.ConvertPlaceholders(`
		UPDATE ticket_priority 
		SET name = $1, change_time = NOW(), change_by = $2
		WHERE id = $3
	`)

	result, err := db.Exec(updateQuery, req.Name, userID, priorityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update priority"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update priority"})
		return
	}

	// Return updated priority
	response := gin.H{
		"id":       priorityID,
		"name":     req.Name,
		"valid_id": 1,
	}

	c.JSON(http.StatusOK, response)
}