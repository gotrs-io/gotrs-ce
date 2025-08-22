package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// Priority represents a ticket priority
type Priority struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Color   *string `json:"color,omitempty"`
	ValidID int     `json:"valid_id"`
}

// handleGetPriorities returns all active priorities
func handleGetPriorities(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}
	
	rows, err := db.Query(`
		SELECT id, name, color, valid_id 
		FROM ticket_priority 
		WHERE valid_id = 1 
		ORDER BY id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch priorities",
		})
		return
	}
	defer rows.Close()
	
	var priorities []Priority
	for rows.Next() {
		var p Priority
		var color sql.NullString
		
		err := rows.Scan(&p.ID, &p.Name, &color, &p.ValidID)
		if err != nil {
			continue
		}
		
		if color.Valid {
			p.Color = &color.String
		}
		
		priorities = append(priorities, p)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    priorities,
	})
}

// handleGetPriority returns a single priority by ID
func handleGetPriority(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid priority ID",
		})
		return
	}
	
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}
	
	var p Priority
	var color sql.NullString
	
	err = db.QueryRow(`
		SELECT id, name, color, valid_id 
		FROM ticket_priority 
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &color, &p.ValidID)
	
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Priority not found",
		})
		return
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch priority",
		})
		return
	}
	
	if color.Valid {
		p.Color = &color.String
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    p,
	})
}

// handleCreatePriority creates a new priority
func handleCreatePriority(c *gin.Context) {
	var input struct {
		Name  string  `json:"name" binding:"required"`
		Color *string `json:"color"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Name is required",
		})
		return
	}
	
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}
	
	var id int
	query := `
		INSERT INTO ticket_priority (name, color, valid_id, create_by, change_by) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id
	`
	
	err = db.QueryRow(query, input.Name, input.Color, 1, 1, 1).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create priority",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": Priority{
			ID:      id,
			Name:    input.Name,
			Color:   input.Color,
			ValidID: 1,
		},
	})
}

// handleUpdatePriority updates an existing priority
func handleUpdatePriority(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid priority ID",
		})
		return
	}
	
	var input struct {
		Name    string  `json:"name"`
		Color   *string `json:"color"`
		ValidID *int    `json:"valid_id"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}
	
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}
	
	// Build update query dynamically based on provided fields
	query := `UPDATE ticket_priority SET change_by = $1, change_time = CURRENT_TIMESTAMP`
	args := []interface{}{1} // change_by = 1
	argCount := 2
	
	if input.Name != "" {
		query += `, name = $` + strconv.Itoa(argCount)
		args = append(args, input.Name)
		argCount++
	}
	
	if input.Color != nil {
		query += `, color = $` + strconv.Itoa(argCount)
		args = append(args, *input.Color)
		argCount++
	}
	
	if input.ValidID != nil {
		query += `, valid_id = $` + strconv.Itoa(argCount)
		args = append(args, *input.ValidID)
		argCount++
	}
	
	query += ` WHERE id = $` + strconv.Itoa(argCount)
	args = append(args, id)
	
	result, err := db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update priority",
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Priority not found",
		})
		return
	}
	
	// Return the updated priority
	validID := 1
	if input.ValidID != nil {
		validID = *input.ValidID
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": Priority{
			ID:      id,
			Name:    input.Name,
			Color:   input.Color,
			ValidID: validID,
		},
	})
}

// handleDeletePriority soft deletes a priority
func handleDeletePriority(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid priority ID",
		})
		return
	}
	
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}
	
	// Soft delete by setting valid_id = 2
	result, err := db.Exec(`
		UPDATE ticket_priority 
		SET valid_id = 2, change_by = $1, change_time = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, 1, id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete priority",
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Priority not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Priority deleted successfully",
	})
}