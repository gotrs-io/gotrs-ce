package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/models"
	"github.com/gotrs-io/gotrs-ce/internal/service"
)

// IncidentHandlers handles incident-related HTTP requests
type IncidentHandlers struct {
	incidentService service.IIncidentService
}

// NewIncidentHandlers creates a new incident handlers instance
func NewIncidentHandlers(incidentService service.IIncidentService) *IncidentHandlers {
	return &IncidentHandlers{
		incidentService: incidentService,
	}
}

// RegisterRoutes registers incident routes
func (h *IncidentHandlers) RegisterRoutes(router *gin.RouterGroup) {
	incidents := router.Group("/incidents")
	{
		// CRUD operations
		incidents.POST("", h.CreateIncident)
		incidents.GET("", h.ListIncidents)
		incidents.GET("/:id", h.GetIncident)
		incidents.PUT("/:id", h.UpdateIncident)
		incidents.DELETE("/:id", h.DeleteIncident)
		
		// Status and workflow operations
		incidents.POST("/:id/assign", h.AssignIncident)
		incidents.POST("/:id/status", h.UpdateStatus)
		incidents.POST("/:id/escalate", h.EscalateIncident)
		incidents.POST("/:id/resolve", h.ResolveIncident)
		incidents.POST("/:id/reopen", h.ReopenIncident)
		incidents.POST("/:id/close", h.CloseIncident)
		
		// Major incident operations
		incidents.POST("/:id/declare-major", h.DeclareMajorIncident)
		incidents.GET("/:id/war-room", h.GetWarRoom)
		incidents.POST("/:id/priority", h.UpdatePriority)
		
		// Relationship operations
		incidents.POST("/:id/link-ticket", h.LinkToTicket)
		incidents.POST("/:id/link-problem", h.LinkToProblem)
		incidents.POST("/:id/link-ci", h.LinkToCI)
		incidents.POST("/:id/link-service", h.LinkToService)
		incidents.GET("/:id/related", h.GetRelatedIncidents)
		incidents.POST("/:id/child", h.CreateChildIncident)
		
		// Communication operations
		incidents.POST("/:id/comments", h.AddComment)
		incidents.GET("/:id/comments", h.GetComments)
		incidents.POST("/:id/work-notes", h.AddWorkNote)
		incidents.POST("/:id/notify", h.NotifyStakeholders)
		
		// Attachment operations
		incidents.POST("/:id/attachments", h.AddAttachment)
		incidents.GET("/:id/attachments", h.GetAttachments)
		incidents.DELETE("/:id/attachments/:attachmentId", h.DeleteAttachment)
		
		// History and audit
		incidents.GET("/:id/history", h.GetHistory)
		incidents.GET("/:id/audit", h.GetAuditTrail)
		incidents.GET("/:id/timeline", h.GetTimeline)
		
		// SLA operations
		incidents.GET("/:id/sla", h.GetSLAStatus)
		incidents.POST("/:id/sla/recalculate", h.RecalculateSLA)
		
		// Reporting and analytics
		incidents.GET("/:id/report", h.GenerateReport)
		incidents.GET("/metrics", h.GetMetrics)
		incidents.GET("/sla-compliance", h.GetSLACompliance)
		incidents.GET("/trends", h.GetTrendAnalysis)
		incidents.GET("/dashboard", h.GetDashboardData)
		
		// Bulk operations
		incidents.POST("/bulk/assign", h.BulkAssign)
		incidents.POST("/bulk/update", h.BulkUpdate)
		incidents.POST("/bulk/close", h.BulkClose)
		
		// Search and filter
		incidents.POST("/search", h.SearchIncidents)
		incidents.GET("/filters", h.GetFilterOptions)
		incidents.GET("/export", h.ExportIncidents)
	}
}

// CreateIncident creates a new incident
// @Summary Create a new incident
// @Description Create a new incident record
// @Tags Incidents
// @Accept json
// @Produce json
// @Param incident body models.Incident true "Incident data"
// @Success 201 {object} models.Incident
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/incidents [post]
func (h *IncidentHandlers) CreateIncident(c *gin.Context) {
	var req struct {
		Title               string                   `json:"title" binding:"required"`
		Description         string                   `json:"description"`
		Severity            models.IncidentSeverity  `json:"severity"`
		Category            models.IncidentCategory  `json:"category"`
		SubCategory         string                   `json:"sub_category"`
		Impact              int                      `json:"impact"`
		Urgency             int                      `json:"urgency"`
		AffectedUserID      *uint                    `json:"affected_user_id"`
		ConfigurationItemID *uint                    `json:"configuration_item_id"`
		ServiceID           *uint                    `json:"service_id"`
		ServiceImpact       string                   `json:"service_impact"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get user ID from context
	userID := getUserID(c)
	
	incident := &models.Incident{
		Title:               req.Title,
		Description:         req.Description,
		Severity:            req.Severity,
		Category:            req.Category,
		SubCategory:         req.SubCategory,
		Impact:              req.Impact,
		Urgency:             req.Urgency,
		AffectedUserID:      req.AffectedUserID,
		ConfigurationItemID: req.ConfigurationItemID,
		ServiceID:           req.ServiceID,
		ServiceImpact:       req.ServiceImpact,
	}
	
	createdIncident, err := h.incidentService.CreateIncident(incident, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, createdIncident)
}

// ListIncidents lists incidents with filtering and pagination
// @Summary List incidents
// @Description Get a paginated list of incidents with optional filtering
// @Tags Incidents
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Param status query string false "Filter by status"
// @Param severity query string false "Filter by severity"
// @Param category query string false "Filter by category"
// @Param search query string false "Search in title and description"
// @Success 200 {object} models.IncidentListResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/incidents [get]
func (h *IncidentHandlers) ListIncidents(c *gin.Context) {
	req := &models.IncidentListRequest{
		Page:    getIntParam(c, "page", 1),
		PerPage: getIntParam(c, "per_page", 20),
	}
	
	// Parse query parameters
	if status := c.Query("status"); status != "" {
		req.Status = models.IncidentStatus(status)
	}
	if severity := c.Query("severity"); severity != "" {
		req.Severity = models.IncidentSeverity(severity)
	}
	if category := c.Query("category"); category != "" {
		req.Category = models.IncidentCategory(category)
	}
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		if id, err := strconv.ParseUint(assignedTo, 10, 32); err == nil {
			req.AssignedToID = uint(id)
		}
	}
	req.Search = c.Query("search")
	req.SortBy = c.Query("sort_by")
	req.SortOrder = c.Query("sort_order")
	
	// Parse date filters
	if fromDate := c.Query("from_date"); fromDate != "" {
		if t, err := time.Parse("2006-01-02", fromDate); err == nil {
			req.FromDate = &t
		}
	}
	if toDate := c.Query("to_date"); toDate != "" {
		if t, err := time.Parse("2006-01-02", toDate); err == nil {
			req.ToDate = &t
		}
	}
	
	// Parse boolean filters
	if isMajor := c.Query("is_major_incident"); isMajor != "" {
		major := isMajor == "true"
		req.IsMajorIncident = &major
	}
	
	response, err := h.incidentService.ListIncidents(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetIncident gets an incident by ID
// @Summary Get incident by ID
// @Description Get detailed information about a specific incident
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Success 200 {object} models.Incident
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/incidents/{id} [get]
func (h *IncidentHandlers) GetIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	incident, err := h.incidentService.GetIncident(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
		return
	}
	
	c.JSON(http.StatusOK, incident)
}

// UpdateIncident updates an incident
// @Summary Update an incident
// @Description Update an existing incident
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param incident body models.Incident true "Updated incident data"
// @Success 200 {object} models.Incident
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/incidents/{id} [put]
func (h *IncidentHandlers) UpdateIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var incident models.Incident
	if err := c.ShouldBindJSON(&incident); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	incident.ID = uint(id)
	userID := getUserID(c)
	
	if err := h.incidentService.UpdateIncident(&incident, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Get updated incident
	updatedIncident, err := h.incidentService.GetIncident(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, updatedIncident)
}

// DeleteIncident deletes an incident
// @Summary Delete an incident
// @Description Delete an incident by ID
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Success 204 "No content"
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/incidents/{id} [delete]
func (h *IncidentHandlers) DeleteIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	if err := h.incidentService.DeleteIncident(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Status(http.StatusNoContent)
}

// AssignIncident assigns an incident to a user
// @Summary Assign an incident
// @Description Assign an incident to a user and/or group
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param assignment body object true "Assignment data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/incidents/{id}/assign [post]
func (h *IncidentHandlers) AssignIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		AssigneeID uint  `json:"assignee_id" binding:"required"`
		GroupID    *uint `json:"group_id"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	if err := h.incidentService.AssignIncident(uint(id), req.AssigneeID, req.GroupID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "incident assigned successfully"})
}

// UpdateStatus updates the status of an incident
// @Summary Update incident status
// @Description Update the status of an incident
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param status body object true "Status data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/incidents/{id}/status [post]
func (h *IncidentHandlers) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		Status models.IncidentStatus `json:"status" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	if err := h.incidentService.UpdateIncidentStatus(uint(id), req.Status, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}

// EscalateIncident escalates an incident
// @Summary Escalate an incident
// @Description Escalate an incident to a higher level
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param escalation body object true "Escalation data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/incidents/{id}/escalate [post]
func (h *IncidentHandlers) EscalateIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		Level  int    `json:"level" binding:"required,min=1,max=5"`
		Reason string `json:"reason" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	if err := h.incidentService.EscalateIncident(uint(id), req.Level, req.Reason, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "incident escalated successfully"})
}

// ResolveIncident resolves an incident
// @Summary Resolve an incident
// @Description Mark an incident as resolved with resolution details
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param resolution body object true "Resolution data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/incidents/{id}/resolve [post]
func (h *IncidentHandlers) ResolveIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		ResolutionCode  string `json:"resolution_code" binding:"required"`
		ResolutionNotes string `json:"resolution_notes" binding:"required"`
		RootCause       string `json:"root_cause"`
		WorkaroundProvided bool `json:"workaround_provided"`
		WorkaroundDetails  string `json:"workaround_details"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	if err := h.incidentService.ResolveIncident(uint(id), req.ResolutionCode, req.ResolutionNotes, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "incident resolved successfully"})
}

// ReopenIncident reopens a closed incident
// @Summary Reopen an incident
// @Description Reopen a closed or resolved incident
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param reopen body object true "Reopen data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/incidents/{id}/reopen [post]
func (h *IncidentHandlers) ReopenIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	if err := h.incidentService.ReopenIncident(uint(id), req.Reason, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "incident reopened successfully"})
}

// CloseIncident closes an incident
// @Summary Close an incident
// @Description Close a resolved incident
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/incidents/{id}/close [post]
func (h *IncidentHandlers) CloseIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	userID := getUserID(c)
	
	if err := h.incidentService.UpdateIncidentStatus(uint(id), models.IncidentStatusClosed, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "incident closed successfully"})
}

// DeclareMajorIncident declares an incident as a major incident
// @Summary Declare major incident
// @Description Declare an incident as a major incident
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/incidents/{id}/declare-major [post]
func (h *IncidentHandlers) DeclareMajorIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	userID := getUserID(c)
	
	if err := h.incidentService.DeclareMajorIncident(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "major incident declared successfully"})
}

// GetWarRoom gets war room information for a major incident
// @Summary Get war room
// @Description Get war room information for a major incident
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Success 200 {object} service.WarRoom
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/incidents/{id}/war-room [get]
func (h *IncidentHandlers) GetWarRoom(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	warRoom, err := h.incidentService.CreateWarRoom(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, warRoom)
}

// UpdatePriority updates the priority of an incident
// @Summary Update incident priority
// @Description Update the priority of an incident based on impact and urgency
// @Tags Incidents
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param priority body object true "Priority data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/incidents/{id}/priority [post]
func (h *IncidentHandlers) UpdatePriority(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		Impact  int `json:"impact" binding:"required,min=1,max=5"`
		Urgency int `json:"urgency" binding:"required,min=1,max=5"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	if err := h.incidentService.UpdateIncidentPriority(uint(id), req.Impact, req.Urgency, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "priority updated successfully"})
}

// LinkToTicket links an incident to a ticket
func (h *IncidentHandlers) LinkToTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		TicketID uint `json:"ticket_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.incidentService.LinkToTicket(uint(id), req.TicketID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "linked to ticket successfully"})
}

// LinkToProblem links an incident to a problem
func (h *IncidentHandlers) LinkToProblem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		ProblemID uint `json:"problem_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.incidentService.LinkToProblem(uint(id), req.ProblemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "linked to problem successfully"})
}

// LinkToCI links an incident to a configuration item
func (h *IncidentHandlers) LinkToCI(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		CIID uint `json:"ci_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.incidentService.LinkToCI(uint(id), req.CIID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "linked to CI successfully"})
}

// LinkToService links an incident to a service
func (h *IncidentHandlers) LinkToService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		ServiceID uint `json:"service_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.incidentService.LinkToService(uint(id), req.ServiceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "linked to service successfully"})
}

// GetRelatedIncidents gets related incidents
func (h *IncidentHandlers) GetRelatedIncidents(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	incidents, err := h.incidentService.GetRelatedIncidents(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, incidents)
}

// CreateChildIncident creates a child incident
func (h *IncidentHandlers) CreateChildIncident(c *gin.Context) {
	parentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent incident ID"})
		return
	}
	
	var incident models.Incident
	if err := c.ShouldBindJSON(&incident); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	childIncident, err := h.incidentService.CreateChildIncident(uint(parentID), &incident, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, childIncident)
}

// AddComment adds a comment to an incident
func (h *IncidentHandlers) AddComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		Comment  string `json:"comment" binding:"required"`
		IsPublic bool   `json:"is_public"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	if err := h.incidentService.AddComment(uint(id), req.Comment, req.IsPublic, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "comment added successfully"})
}

// GetComments gets comments for an incident
func (h *IncidentHandlers) GetComments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	comments, err := h.incidentService.GetComments(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, comments)
}

// AddWorkNote adds a work note to an incident
func (h *IncidentHandlers) AddWorkNote(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		Note string `json:"note" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	if err := h.incidentService.AddWorkNote(uint(id), req.Note, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "work note added successfully"})
}

// NotifyStakeholders sends notifications to stakeholders
func (h *IncidentHandlers) NotifyStakeholders(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	var req struct {
		Message string `json:"message" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.incidentService.NotifyStakeholders(uint(id), req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "stakeholders notified successfully"})
}

// AddAttachment adds an attachment to an incident
func (h *IncidentHandlers) AddAttachment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	// Handle file upload
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file upload failed"})
		return
	}
	defer file.Close()
	
	// Save file (simplified - in production, use proper storage service)
	filePath := "/tmp/" + header.Filename
	
	userID := getUserID(c)
	
	if err := h.incidentService.AddAttachment(
		uint(id),
		header.Filename,
		filePath,
		header.Size,
		header.Header.Get("Content-Type"),
		userID,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "attachment added successfully"})
}

// GetAttachments gets attachments for an incident
func (h *IncidentHandlers) GetAttachments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	attachments, err := h.incidentService.GetAttachments(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, attachments)
}

// DeleteAttachment deletes an attachment
func (h *IncidentHandlers) DeleteAttachment(c *gin.Context) {
	attachmentID, err := strconv.ParseUint(c.Param("attachmentId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attachment ID"})
		return
	}
	
	if err := h.incidentService.DeleteAttachment(uint(attachmentID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Status(http.StatusNoContent)
}

// GetHistory gets the history of an incident
func (h *IncidentHandlers) GetHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	history, err := h.incidentService.GetHistory(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, history)
}

// GetAuditTrail gets the audit trail of an incident
func (h *IncidentHandlers) GetAuditTrail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	audit, err := h.incidentService.GetAuditTrail(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, audit)
}

// GetTimeline gets the timeline of an incident
func (h *IncidentHandlers) GetTimeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	// Get history and convert to timeline
	history, err := h.incidentService.GetHistory(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	timeline := make([]gin.H, 0, len(history))
	for _, h := range history {
		timeline = append(timeline, gin.H{
			"timestamp": h.CreatedAt,
			"event":     h.ChangeType,
			"field":     h.FieldName,
			"old_value": h.OldValue,
			"new_value": h.NewValue,
			"user":      h.ChangedBy,
		})
	}
	
	c.JSON(http.StatusOK, timeline)
}

// GetSLAStatus gets the SLA status of an incident
func (h *IncidentHandlers) GetSLAStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	status, err := h.incidentService.CheckSLACompliance(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, status)
}

// RecalculateSLA recalculates SLA for an incident
func (h *IncidentHandlers) RecalculateSLA(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	if err := h.incidentService.UpdateSLATimes(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "SLA recalculated successfully"})
}

// GenerateReport generates a report for an incident
func (h *IncidentHandlers) GenerateReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident ID"})
		return
	}
	
	report, err := h.incidentService.GenerateIncidentReport(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, report)
}

// GetMetrics gets incident metrics
func (h *IncidentHandlers) GetMetrics(c *gin.Context) {
	// Parse date range
	from, _ := time.Parse("2006-01-02", c.DefaultQuery("from", time.Now().AddDate(0, -1, 0).Format("2006-01-02")))
	to, _ := time.Parse("2006-01-02", c.DefaultQuery("to", time.Now().Format("2006-01-02")))
	
	metrics, err := h.incidentService.GetIncidentMetrics(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, metrics)
}

// GetSLACompliance gets SLA compliance metrics
func (h *IncidentHandlers) GetSLACompliance(c *gin.Context) {
	// Parse date range
	from, _ := time.Parse("2006-01-02", c.DefaultQuery("from", time.Now().AddDate(0, -1, 0).Format("2006-01-02")))
	to, _ := time.Parse("2006-01-02", c.DefaultQuery("to", time.Now().Format("2006-01-02")))
	
	compliance, err := h.incidentService.GetSLACompliance(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, compliance)
}

// GetTrendAnalysis gets trend analysis
func (h *IncidentHandlers) GetTrendAnalysis(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	
	trends, err := h.incidentService.GetTrendAnalysis(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, trends)
}

// GetDashboardData gets dashboard data
func (h *IncidentHandlers) GetDashboardData(c *gin.Context) {
	// Get various metrics for dashboard
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	metrics, _ := h.incidentService.GetIncidentMetrics(startOfMonth, now)
	slaCompliance, _ := h.incidentService.GetSLACompliance(startOfMonth, now)
	trends, _ := h.incidentService.GetTrendAnalysis("week")
	
	dashboard := gin.H{
		"metrics":        metrics,
		"sla_compliance": slaCompliance,
		"trends":         trends,
		"last_updated":   time.Now(),
	}
	
	c.JSON(http.StatusOK, dashboard)
}

// BulkAssign bulk assigns incidents
func (h *IncidentHandlers) BulkAssign(c *gin.Context) {
	var req struct {
		IncidentIDs []uint `json:"incident_ids" binding:"required"`
		AssigneeID  uint   `json:"assignee_id" binding:"required"`
		GroupID     *uint  `json:"group_id"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	var errors []string
	for _, id := range req.IncidentIDs {
		if err := h.incidentService.AssignIncident(id, req.AssigneeID, req.GroupID, userID); err != nil {
			errors = append(errors, err.Error())
		}
	}
	
	if len(errors) > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"message": "some assignments failed",
			"errors":  errors,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "all incidents assigned successfully"})
}

// BulkUpdate bulk updates incidents
func (h *IncidentHandlers) BulkUpdate(c *gin.Context) {
	var req struct {
		IncidentIDs []uint                  `json:"incident_ids" binding:"required"`
		Status      models.IncidentStatus   `json:"status,omitempty"`
		Severity    models.IncidentSeverity `json:"severity,omitempty"`
		Category    models.IncidentCategory `json:"category,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	var errors []string
	for _, id := range req.IncidentIDs {
		incident, err := h.incidentService.GetIncident(id)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		
		if req.Status != "" {
			incident.Status = req.Status
		}
		if req.Severity != "" {
			incident.Severity = req.Severity
		}
		if req.Category != "" {
			incident.Category = req.Category
		}
		
		if err := h.incidentService.UpdateIncident(incident, userID); err != nil {
			errors = append(errors, err.Error())
		}
	}
	
	if len(errors) > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"message": "some updates failed",
			"errors":  errors,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "all incidents updated successfully"})
}

// BulkClose bulk closes incidents
func (h *IncidentHandlers) BulkClose(c *gin.Context) {
	var req struct {
		IncidentIDs []uint `json:"incident_ids" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID := getUserID(c)
	
	var errors []string
	for _, id := range req.IncidentIDs {
		if err := h.incidentService.UpdateIncidentStatus(id, models.IncidentStatusClosed, userID); err != nil {
			errors = append(errors, err.Error())
		}
	}
	
	if len(errors) > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"message": "some closures failed",
			"errors":  errors,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "all incidents closed successfully"})
}

// SearchIncidents searches incidents
func (h *IncidentHandlers) SearchIncidents(c *gin.Context) {
	var req models.IncidentListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	response, err := h.incidentService.ListIncidents(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetFilterOptions gets available filter options
func (h *IncidentHandlers) GetFilterOptions(c *gin.Context) {
	options := gin.H{
		"statuses": []string{
			string(models.IncidentStatusNew),
			string(models.IncidentStatusAssigned),
			string(models.IncidentStatusInProgress),
			string(models.IncidentStatusPending),
			string(models.IncidentStatusResolved),
			string(models.IncidentStatusClosed),
		},
		"severities": []string{
			string(models.SeverityCritical),
			string(models.SeverityHigh),
			string(models.SeverityMedium),
			string(models.SeverityLow),
		},
		"categories": []string{
			string(models.CategoryHardware),
			string(models.CategorySoftware),
			string(models.CategoryNetwork),
			string(models.CategorySecurity),
			string(models.CategoryAccess),
			string(models.CategoryPerformance),
			string(models.CategoryOther),
		},
		"priorities": []int{1, 2, 3, 4, 5},
	}
	
	c.JSON(http.StatusOK, options)
}

// ExportIncidents exports incidents to CSV/Excel
func (h *IncidentHandlers) ExportIncidents(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	
	// Get incidents
	req := &models.IncidentListRequest{
		Page:    1,
		PerPage: 10000, // Export all
	}
	
	response, err := h.incidentService.ListIncidents(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	if format == "csv" {
		// Generate CSV
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=incidents.csv")
		
		// Write CSV header
		c.Writer.WriteString("ID,Number,Title,Status,Severity,Priority,Created,Resolved\n")
		
		// Write data
		for _, incident := range response.Incidents {
			resolvedStr := ""
			if incident.ResolvedAt != nil {
				resolvedStr = incident.ResolvedAt.Format("2006-01-02 15:04:05")
			}
			c.Writer.WriteString(
				strconv.FormatUint(uint64(incident.ID), 10) + "," +
					incident.IncidentNumber + "," +
					incident.Title + "," +
					string(incident.Status) + "," +
					string(incident.Severity) + "," +
					strconv.Itoa(incident.Priority) + "," +
					incident.CreatedAt.Format("2006-01-02 15:04:05") + "," +
					resolvedStr + "\n",
			)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported format"})
	}
}

// Helper functions

func getUserID(c *gin.Context) uint {
	// Get user ID from context (set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 1 // Default for testing
}

func getIntParam(c *gin.Context, param string, defaultValue int) int {
	if val := c.Query(param); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}