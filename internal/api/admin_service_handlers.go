package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

// Service represents a service in the system
type Service struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Comments   sql.NullString `json:"comments"`
	ValidID    int            `json:"valid_id"`
	CreateTime time.Time      `json:"create_time"`
	CreateBy   int            `json:"create_by"`
	ChangeTime time.Time      `json:"change_time"`
	ChangeBy   int            `json:"change_by"`
	
	// Additional fields for display
	TicketCount int    `json:"ticket_count"`
	Validity    string `json:"validity"`
}

// handleAdminServices handles the admin services page
func handleAdminServices(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get all services (including invalid ones for admin)
		query := `
			SELECT 
				s.id, s.name, s.comments, s.valid_id,
				s.create_time, s.create_by, s.change_time, s.change_by,
				COUNT(DISTINCT t.id) as ticket_count
			FROM service s
			LEFT JOIN ticket t ON t.service_id = s.id
			GROUP BY s.id, s.name, s.comments, s.valid_id, 
				s.create_time, s.create_by, s.change_time, s.change_by
			ORDER BY s.name`

		rows, err := db.Query(query)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.pongo2", pongo2.Context{
				"Title":   "Database Error",
				"Message": fmt.Sprintf("Failed to fetch services: %v", err),
			})
			return
		}
		defer rows.Close()

		var services []Service
		for rows.Next() {
			var s Service
			err := rows.Scan(
				&s.ID, &s.Name, &s.Comments, &s.ValidID,
				&s.CreateTime, &s.CreateBy, &s.ChangeTime, &s.ChangeBy,
				&s.TicketCount,
			)
			if err != nil {
				continue
			}

			// Set validity display text
			switch s.ValidID {
			case 1:
				s.Validity = "valid"
			case 2:
				s.Validity = "invalid"
			case 3:
				s.Validity = "invalid-temporarily"
			default:
				s.Validity = "unknown"
			}

			services = append(services, s)
		}

		// Render the template
		if err := pongo2Renderer.HTML(c, http.StatusOK, "pages/admin/services.pongo2", pongo2.Context{
			"Title":       "Service Management",
			"ActivePage":  "admin",
			"Services":    services,
			"User":        getUserFromContext(c),
			"CurrentTime": time.Now(),
		}); err != nil {
			c.HTML(http.StatusInternalServerError, "error.pongo2", pongo2.Context{
				"Title":   "Template Error",
				"Message": fmt.Sprintf("Failed to render template: %v", err),
			})
		}
	}
}

// handleCreateService handles service creation
func handleCreateService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name     string `json:"name" form:"name" binding:"required"`
			Comments string `json:"comments" form:"comments"`
			ValidID  int    `json:"valid_id" form:"valid_id"`
		}

		if err := c.ShouldBind(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Default to valid if not specified
		if input.ValidID == 0 {
			input.ValidID = 1
		}

		userID := getUserIDFromContext(c)
		if userID == 0 {
			userID = 1 // Default to admin user
		}

		// Insert the new service
		var id int
		query := `
			INSERT INTO service (name, comments, valid_id, create_by, change_by, create_time, change_time)
			VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			RETURNING id`

		var comments sql.NullString
		if input.Comments != "" {
			comments = sql.NullString{String: input.Comments, Valid: true}
		}

		err := db.QueryRow(query, input.Name, comments, input.ValidID, userID, userID).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create service: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Service created successfully",
			"id":      id,
			"name":    input.Name,
			"valid_id": input.ValidID,
		})
	}
}

// handleUpdateService handles service updates
func handleUpdateService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
			return
		}

		var input struct {
			Name     string `json:"name" form:"name" binding:"required"`
			Comments string `json:"comments" form:"comments"`
			ValidID  int    `json:"valid_id" form:"valid_id"`
		}

		if err := c.ShouldBind(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Default to valid if not specified
		if input.ValidID == 0 {
			input.ValidID = 1
		}

		userID := getUserIDFromContext(c)
		if userID == 0 {
			userID = 1
		}

		// Update the service
		query := `
			UPDATE service
			SET name = $1, comments = $2, valid_id = $3, 
				change_by = $4, change_time = CURRENT_TIMESTAMP
			WHERE id = $5`

		var comments sql.NullString
		if input.Comments != "" {
			comments = sql.NullString{String: input.Comments, Valid: true}
		}

		result, err := db.Exec(query, input.Name, comments, input.ValidID, userID, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update service: %v", err)})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Service updated successfully",
			"id":      id,
			"name":    input.Name,
			"valid_id": input.ValidID,
		})
	}
}

// handleGetService handles fetching a single service
func handleGetService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
			return
		}

		var s Service
		query := `
			SELECT id, name, comments, valid_id, 
				create_time, create_by, change_time, change_by
			FROM service
			WHERE id = $1`

		err = db.QueryRow(query, id).Scan(
			&s.ID, &s.Name, &s.Comments, &s.ValidID,
			&s.CreateTime, &s.CreateBy, &s.ChangeTime, &s.ChangeBy,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", err)})
			}
			return
		}

		c.JSON(http.StatusOK, s)
	}
}

// handleDeleteService handles service soft deletion (invalidation)
func handleDeleteService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
			return
		}

		userID := getUserIDFromContext(c)
		if userID == 0 {
			userID = 1
		}

		// Soft delete by setting valid_id to 2 (invalid)
		query := `
			UPDATE service
			SET valid_id = 2, change_by = $1, change_time = CURRENT_TIMESTAMP
			WHERE id = $2`

		result, err := db.Exec(query, userID, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to invalidate service: %v", err)})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Service invalidated successfully",
		})
	}
}