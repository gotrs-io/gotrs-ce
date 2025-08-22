package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// State represents a ticket state
type State struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	TypeID   int     `json:"type_id"`
	Comments *string `json:"comments,omitempty"`
	ValidID  int     `json:"valid_id"`
}

// handleGetStates returns all active states
func handleGetStates(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}
	
	rows, err := db.Query(`
		SELECT id, name, type_id, comments, valid_id 
		FROM ticket_state 
		WHERE valid_id = 1 
		ORDER BY id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch states",
		})
		return
	}
	defer rows.Close()
	
	var states []State
	for rows.Next() {
		var s State
		var comments sql.NullString
		
		err := rows.Scan(&s.ID, &s.Name, &s.TypeID, &comments, &s.ValidID)
		if err != nil {
			continue
		}
		
		if comments.Valid {
			s.Comments = &comments.String
		}
		
		states = append(states, s)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    states,
	})
}

// handleCreateState creates a new state
func handleCreateState(c *gin.Context) {
	var input struct {
		Name     string  `json:"name" binding:"required"`
		TypeID   int     `json:"type_id" binding:"required"`
		Comments *string `json:"comments"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Name and type_id are required",
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
		INSERT INTO ticket_state (name, type_id, comments, valid_id, create_by, change_by) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id
	`
	
	err = db.QueryRow(query, input.Name, input.TypeID, input.Comments, 1, 1, 1).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create state",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": State{
			ID:       id,
			Name:     input.Name,
			TypeID:   input.TypeID,
			Comments: input.Comments,
			ValidID:  1,
		},
	})
}

// handleUpdateState updates an existing state
func handleUpdateState(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid state ID",
		})
		return
	}
	
	var input struct {
		Name     string  `json:"name"`
		TypeID   *int    `json:"type_id"`
		Comments *string `json:"comments"`
		ValidID  *int    `json:"valid_id"`
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
	query := `UPDATE ticket_state SET change_by = $1, change_time = CURRENT_TIMESTAMP`
	args := []interface{}{1} // change_by = 1
	argCount := 2
	
	if input.Name != "" {
		query += `, name = $` + strconv.Itoa(argCount)
		args = append(args, input.Name)
		argCount++
	}
	
	if input.TypeID != nil {
		query += `, type_id = $` + strconv.Itoa(argCount)
		args = append(args, *input.TypeID)
		argCount++
	}
	
	if input.Comments != nil {
		query += `, comments = $` + strconv.Itoa(argCount)
		args = append(args, *input.Comments)
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
			"error":   "Failed to update state",
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "State not found",
		})
		return
	}
	
	// Return the updated state
	response := State{
		ID:       id,
		Name:     input.Name,
		Comments: input.Comments,
	}
	
	if input.TypeID != nil {
		response.TypeID = *input.TypeID
	}
	
	if input.ValidID != nil {
		response.ValidID = *input.ValidID
	} else {
		response.ValidID = 1
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// handleDeleteState soft deletes a state
func handleDeleteState(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid state ID",
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
		UPDATE ticket_state 
		SET valid_id = 2, change_by = $1, change_time = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, 1, id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete state",
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "State not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "State deleted successfully",
	})
}