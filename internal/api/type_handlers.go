package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// Type represents a ticket type
type Type struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Comments *string `json:"comments,omitempty"`
	ValidID  int     `json:"valid_id"`
}

// handleGetTypes returns all active types (overrides the existing one)
func handleGetTypes(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}
	
	rows, err := db.Query(`
		SELECT id, name, comments, valid_id 
		FROM ticket_type 
		WHERE valid_id = 1 
		ORDER BY id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch types",
		})
		return
	}
	defer rows.Close()
	
	var types []Type
	for rows.Next() {
		var t Type
		var comments sql.NullString
		
		err := rows.Scan(&t.ID, &t.Name, &comments, &t.ValidID)
		if err != nil {
			continue
		}
		
		if comments.Valid {
			t.Comments = &comments.String
		}
		
		types = append(types, t)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    types,
	})
}

// handleCreateType creates a new type
func handleCreateType(c *gin.Context) {
	var input struct {
		Name     string  `json:"name" binding:"required"`
		Comments *string `json:"comments"`
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
		INSERT INTO ticket_type (name, comments, valid_id, create_by, change_by) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id
	`
	
	err = db.QueryRow(query, input.Name, input.Comments, 1, 1, 1).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create type",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": Type{
			ID:       id,
			Name:     input.Name,
			Comments: input.Comments,
			ValidID:  1,
		},
	})
}

// handleUpdateType updates an existing type
func handleUpdateType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid type ID",
		})
		return
	}
	
	var input struct {
		Name     string  `json:"name"`
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
	query := `UPDATE ticket_type SET change_by = $1, change_time = CURRENT_TIMESTAMP`
	args := []interface{}{1} // change_by = 1
	argCount := 2
	
	if input.Name != "" {
		query += `, name = $` + strconv.Itoa(argCount)
		args = append(args, input.Name)
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
			"error":   "Failed to update type",
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Type not found",
		})
		return
	}
	
	// Return the updated type
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": Type{
			ID:       id,
			Name:     input.Name,
			Comments: input.Comments,
			ValidID:  1,
		},
	})
}

// handleDeleteType soft deletes a type
func handleDeleteType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid type ID",
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
		UPDATE ticket_type 
		SET valid_id = 2, change_by = $1, change_time = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, 1, id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete type",
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Type not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Type deleted successfully",
	})
}