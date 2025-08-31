package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// HandleGetQueueAPI handles GET /api/v1/queues/:id
func HandleGetQueueAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = userID // Will use for permission checks later

	// Parse queue ID
	queueID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue ID"})
		return
	}

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Query queue with optional stats
	includeStats := c.Query("include_stats") == "true"

	// Get queue details
	var queue struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Description  string `json:"description"`
		ValidID      int    `json:"valid_id"`
		GroupAccess  []int  `json:"group_access"`
		TicketCounts map[string]int `json:"ticket_counts,omitempty"`
	}

	query := database.ConvertPlaceholders(`
		SELECT q.id, q.name, q.description, q.valid_id
		FROM queues q
		WHERE q.id = $1
	`)

	err = db.QueryRow(query, queueID).Scan(&queue.ID, &queue.Name, &queue.Description, &queue.ValidID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Queue not found"})
		return
	}

	// Get group access
	groupQuery := database.ConvertPlaceholders(`
		SELECT group_id FROM queue_groups 
		WHERE queue_id = $1
	`)
	
	rows, err := db.Query(groupQuery, queueID)
	if err == nil {
		defer rows.Close()
		queue.GroupAccess = []int{}
		for rows.Next() {
			var groupID int
			if err := rows.Scan(&groupID); err == nil {
				queue.GroupAccess = append(queue.GroupAccess, groupID)
			}
		}
	}

	// Get ticket statistics if requested
	if includeStats {
		queue.TicketCounts = make(map[string]int)
		
		statsQuery := database.ConvertPlaceholders(`
			SELECT s.name, COUNT(t.id) as count
			FROM tickets t
			JOIN ticket_state s ON t.ticket_state_id = s.id
			WHERE t.queue_id = $1
			GROUP BY s.id, s.name
		`)
		
		rows, err := db.Query(statsQuery, queueID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var stateName string
				var count int
				if err := rows.Scan(&stateName, &count); err == nil {
					queue.TicketCounts[stateName] = count
				}
			}
		}
	}

	c.JSON(http.StatusOK, queue)
}