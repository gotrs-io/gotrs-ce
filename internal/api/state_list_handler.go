package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// HandleListStatesAPI handles GET /api/v1/states
func HandleListStatesAPI(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil || db == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []gin.H{
			{"id": 1, "name": "new"},
			{"id": 2, "name": "open"},
			{"id": 3, "name": "pending"},
			{"id": 4, "name": "resolved"},
			{"id": 5, "name": "closed"},
		}})
		return
	}

	rows, err := db.Query(database.ConvertPlaceholders(`
		SELECT id, name, valid_id
		FROM ticket_state
		ORDER BY id
	`))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}})
		return
	}
	defer rows.Close()

	var items []gin.H
	for rows.Next() {
		var id, validID int
		var name string
		if err := rows.Scan(&id, &name, &validID); err == nil {
			items = append(items, gin.H{"id": id, "name": name, "valid_id": validID})
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}
