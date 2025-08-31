package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// HandleListPrioritiesAPI handles GET /api/v1/priorities
func HandleListPrioritiesAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = userID // Will use for permission checks later

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Build query based on filters
	query := database.ConvertPlaceholders(`
		SELECT id, name, valid_id
		FROM ticket_priority
		WHERE 1=1
	`)
	args := []interface{}{}
	paramCount := 0

	// Filter by valid status if specified
	if validFilter := c.Query("valid"); validFilter == "true" {
		paramCount++
		query += database.ConvertPlaceholders(` AND valid_id = $` + strconv.Itoa(paramCount))
		args = append(args, 1)
	}

	query += ` ORDER BY id`

	// Execute query
	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch priorities"})
		return
	}
	defer rows.Close()

	// Collect results
	priorities := []gin.H{}
	for rows.Next() {
		var priority struct {
			ID      int
			Name    string
			ValidID int
		}
		if err := rows.Scan(&priority.ID, &priority.Name, &priority.ValidID); err != nil {
			continue
		}
		priorities = append(priorities, gin.H{
			"id":       priority.ID,
			"name":     priority.Name,
			"valid_id": priority.ValidID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"priorities": priorities,
		"total":      len(priorities),
	})
}