package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// Queue represents a ticket queue
type Queue struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	GroupID          int     `json:"group_id"`
	SystemAddressID  *int    `json:"system_address_id,omitempty"`
	SalutationID     *int    `json:"salutation_id,omitempty"`
	SignatureID      *int    `json:"signature_id,omitempty"`
	UnlockTimeout    int     `json:"unlock_timeout"`
	FollowUpID       int     `json:"follow_up_id"`
	FollowUpLock     int     `json:"follow_up_lock"`
	Comments         *string `json:"comments,omitempty"`
	ValidID          int     `json:"valid_id"`
	GroupName        *string `json:"group_name,omitempty"`
}

// QueueDetails includes additional statistics
type QueueDetails struct {
	Queue
	TicketCount     int     `json:"ticket_count"`
	OpenTickets     int     `json:"open_tickets"`
	AgentCount      int     `json:"agent_count"`
	AvgResponseTime *string `json:"avg_response_time,omitempty"`
}

// handleGetQueuesAPI returns all queues for API
func handleGetQueuesAPI(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}
	
	query := `
		SELECT 
			q.id, q.name, q.group_id, q.system_address_id, q.salutation_id,
			q.signature_id, q.unlock_timeout, q.follow_up_id, q.follow_up_lock,
			q.comments, q.valid_id, g.name as group_name
		FROM queue q
		LEFT JOIN groups g ON q.group_id = g.id
		WHERE q.valid_id = 1
		ORDER BY q.name
	`
	
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch queues",
		})
		return
	}
	defer rows.Close()
	
	var queues []Queue
	for rows.Next() {
		var q Queue
		var systemAddressID, salutationID, signatureID sql.NullInt32
		var comments, groupName sql.NullString
		
		err := rows.Scan(
			&q.ID, &q.Name, &q.GroupID, &systemAddressID, &salutationID,
			&signatureID, &q.UnlockTimeout, &q.FollowUpID, &q.FollowUpLock,
			&comments, &q.ValidID, &groupName,
		)
		if err != nil {
			continue
		}
		
		if systemAddressID.Valid {
			val := int(systemAddressID.Int32)
			q.SystemAddressID = &val
		}
		if salutationID.Valid {
			val := int(salutationID.Int32)
			q.SalutationID = &val
		}
		if signatureID.Valid {
			val := int(signatureID.Int32)
			q.SignatureID = &val
		}
		if comments.Valid {
			q.Comments = &comments.String
		}
		if groupName.Valid {
			q.GroupName = &groupName.String
		}
		
		queues = append(queues, q)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queues,
	})
}

// handleCreateQueue creates a new queue
func handleCreateQueue(c *gin.Context) {
	var input struct {
		Name            string  `json:"name" binding:"required"`
		GroupID         int     `json:"group_id" binding:"required"`
		SystemAddressID *int    `json:"system_address_id"`
		SalutationID    *int    `json:"salutation_id"`
		SignatureID     *int    `json:"signature_id"`
		UnlockTimeout   int     `json:"unlock_timeout"`
		FollowUpID      int     `json:"follow_up_id"`
		FollowUpLock    int     `json:"follow_up_lock"`
		Comments        *string `json:"comments"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Name and group_id are required",
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
	
	// Set defaults
	if input.FollowUpID == 0 {
		input.FollowUpID = 1
	}
	
	var id int
	query := `
		INSERT INTO queue (
			name, group_id, system_address_id, salutation_id, signature_id,
			unlock_timeout, follow_up_id, follow_up_lock, comments, valid_id,
			create_by, change_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	
	err = db.QueryRow(
		query,
		input.Name, input.GroupID, input.SystemAddressID, input.SalutationID,
		input.SignatureID, input.UnlockTimeout, input.FollowUpID, input.FollowUpLock,
		input.Comments, 1, 1, 1,
	).Scan(&id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create queue",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": Queue{
			ID:              id,
			Name:            input.Name,
			GroupID:         input.GroupID,
			SystemAddressID: input.SystemAddressID,
			SalutationID:    input.SalutationID,
			SignatureID:     input.SignatureID,
			UnlockTimeout:   input.UnlockTimeout,
			FollowUpID:      input.FollowUpID,
			FollowUpLock:    input.FollowUpLock,
			Comments:        input.Comments,
			ValidID:         1,
		},
	})
}

// handleGetQueue returns a single queue
func handleGetQueue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid queue ID",
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
	
	var q Queue
	var systemAddressID, salutationID, signatureID sql.NullInt32
	var comments sql.NullString
	
	query := `
		SELECT 
			id, name, group_id, system_address_id, salutation_id,
			signature_id, unlock_timeout, follow_up_id, follow_up_lock,
			comments, valid_id
		FROM queue
		WHERE id = $1
	`
	
	err = db.QueryRow(query, id).Scan(
		&q.ID, &q.Name, &q.GroupID, &systemAddressID, &salutationID,
		&signatureID, &q.UnlockTimeout, &q.FollowUpID, &q.FollowUpLock,
		&comments, &q.ValidID,
	)
	
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Queue not found",
		})
		return
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch queue",
		})
		return
	}
	
	if systemAddressID.Valid {
		val := int(systemAddressID.Int32)
		q.SystemAddressID = &val
	}
	if salutationID.Valid {
		val := int(salutationID.Int32)
		q.SalutationID = &val
	}
	if signatureID.Valid {
		val := int(signatureID.Int32)
		q.SignatureID = &val
	}
	if comments.Valid {
		q.Comments = &comments.String
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    q,
	})
}

// handleUpdateQueue updates an existing queue
func handleUpdateQueue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid queue ID",
		})
		return
	}
	
	var input struct {
		Name            string  `json:"name"`
		GroupID         *int    `json:"group_id"`
		SystemAddressID *int    `json:"system_address_id"`
		SalutationID    *int    `json:"salutation_id"`
		SignatureID     *int    `json:"signature_id"`
		UnlockTimeout   *int    `json:"unlock_timeout"`
		FollowUpID      *int    `json:"follow_up_id"`
		FollowUpLock    *int    `json:"follow_up_lock"`
		Comments        *string `json:"comments"`
		ValidID         *int    `json:"valid_id"`
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
	
	// Build update query dynamically
	query := `UPDATE queue SET change_by = $1, change_time = CURRENT_TIMESTAMP`
	args := []interface{}{1} // change_by = 1
	argCount := 2
	
	responseQueue := Queue{ID: id}
	
	if input.Name != "" {
		query += `, name = $` + strconv.Itoa(argCount)
		args = append(args, input.Name)
		responseQueue.Name = input.Name
		argCount++
	}
	
	if input.GroupID != nil {
		query += `, group_id = $` + strconv.Itoa(argCount)
		args = append(args, *input.GroupID)
		responseQueue.GroupID = *input.GroupID
		argCount++
	}
	
	if input.SystemAddressID != nil {
		query += `, system_address_id = $` + strconv.Itoa(argCount)
		args = append(args, *input.SystemAddressID)
		responseQueue.SystemAddressID = input.SystemAddressID
		argCount++
	}
	
	if input.SalutationID != nil {
		query += `, salutation_id = $` + strconv.Itoa(argCount)
		args = append(args, *input.SalutationID)
		responseQueue.SalutationID = input.SalutationID
		argCount++
	}
	
	if input.SignatureID != nil {
		query += `, signature_id = $` + strconv.Itoa(argCount)
		args = append(args, *input.SignatureID)
		responseQueue.SignatureID = input.SignatureID
		argCount++
	}
	
	if input.UnlockTimeout != nil {
		query += `, unlock_timeout = $` + strconv.Itoa(argCount)
		args = append(args, *input.UnlockTimeout)
		responseQueue.UnlockTimeout = *input.UnlockTimeout
		argCount++
	}
	
	if input.FollowUpID != nil {
		query += `, follow_up_id = $` + strconv.Itoa(argCount)
		args = append(args, *input.FollowUpID)
		responseQueue.FollowUpID = *input.FollowUpID
		argCount++
	}
	
	if input.FollowUpLock != nil {
		query += `, follow_up_lock = $` + strconv.Itoa(argCount)
		args = append(args, *input.FollowUpLock)
		responseQueue.FollowUpLock = *input.FollowUpLock
		argCount++
	}
	
	if input.Comments != nil {
		query += `, comments = $` + strconv.Itoa(argCount)
		args = append(args, *input.Comments)
		responseQueue.Comments = input.Comments
		argCount++
	}
	
	if input.ValidID != nil {
		query += `, valid_id = $` + strconv.Itoa(argCount)
		args = append(args, *input.ValidID)
		responseQueue.ValidID = *input.ValidID
		argCount++
	}
	
	query += ` WHERE id = $` + strconv.Itoa(argCount)
	args = append(args, id)
	
	result, err := db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update queue",
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Queue not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseQueue,
	})
}

// handleDeleteQueue soft deletes a queue
func handleDeleteQueue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid queue ID",
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
	
	// Check if queue has tickets
	var ticketCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM ticket WHERE queue_id = $1`, id).Scan(&ticketCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to check queue tickets",
		})
		return
	}
	
	if ticketCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Cannot delete queue with existing tickets",
		})
		return
	}
	
	// Soft delete by setting valid_id = 2
	result, err := db.Exec(`
		UPDATE queue 
		SET valid_id = 2, change_by = $1, change_time = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, 1, id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete queue",
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Queue not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Queue deleted successfully",
	})
}

// handleGetQueueDetails returns detailed queue information with statistics
func handleGetQueueDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid queue ID",
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
	
	// Get queue details
	var details QueueDetails
	var systemAddressID, salutationID, signatureID sql.NullInt32
	var comments, groupName sql.NullString
	
	query := `
		SELECT 
			q.id, q.name, q.group_id, q.system_address_id, q.salutation_id,
			q.signature_id, q.unlock_timeout, q.follow_up_id, q.follow_up_lock,
			q.comments, q.valid_id, g.name as group_name
		FROM queue q
		LEFT JOIN groups g ON q.group_id = g.id
		WHERE q.id = $1
	`
	
	err = db.QueryRow(query, id).Scan(
		&details.ID, &details.Name, &details.GroupID, &systemAddressID, &salutationID,
		&signatureID, &details.UnlockTimeout, &details.FollowUpID, &details.FollowUpLock,
		&comments, &details.ValidID, &groupName,
	)
	
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Queue not found",
		})
		return
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch queue details",
		})
		return
	}
	
	if systemAddressID.Valid {
		val := int(systemAddressID.Int32)
		details.SystemAddressID = &val
	}
	if salutationID.Valid {
		val := int(salutationID.Int32)
		details.SalutationID = &val
	}
	if signatureID.Valid {
		val := int(signatureID.Int32)
		details.SignatureID = &val
	}
	if comments.Valid {
		details.Comments = &comments.String
	}
	if groupName.Valid {
		details.GroupName = &groupName.String
	}
	
	// Get ticket count
	db.QueryRow(`SELECT COUNT(*) FROM ticket WHERE queue_id = $1`, id).Scan(&details.TicketCount)
	
	// Get open tickets count
	db.QueryRow(`
		SELECT COUNT(*) FROM ticket 
		WHERE queue_id = $1 AND ticket_state_id IN (1, 2, 3)
	`, id).Scan(&details.OpenTickets)
	
	// Get agent count
	db.QueryRow(`
		SELECT COUNT(DISTINCT user_id) 
		FROM user_groups 
		WHERE group_id = $1
	`, details.GroupID).Scan(&details.AgentCount)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    details,
	})
}