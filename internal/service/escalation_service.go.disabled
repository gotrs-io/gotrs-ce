package service

import (
	"fmt"
	"time"

	"github.com/gotrs-io/gotrs-ce/internal/models"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
)

// EscalationService handles ticket escalation logic
type EscalationService struct {
	escalationRepo repository.EscalationRepository
	ticketRepo     repository.TicketRepository
	userRepo       repository.UserRepository
	businessHours  map[int]*models.BusinessCalendar // Cache of business calendars by queue ID
	notifyService  *NotificationService
	auditService   *AuditService
}

// NewEscalationService creates a new escalation service
func NewEscalationService(
	escalationRepo repository.EscalationRepository,
	ticketRepo repository.TicketRepository,
	userRepo repository.UserRepository,
	notifyService *NotificationService,
	auditService *AuditService,
) *EscalationService {
	return &EscalationService{
		escalationRepo: escalationRepo,
		ticketRepo:     ticketRepo,
		userRepo:       userRepo,
		businessHours:  make(map[int]*models.BusinessCalendar),
		notifyService:  notifyService,
		auditService:   auditService,
	}
}

// InitializeBusinessHours loads business hours configurations
func (s *EscalationService) InitializeBusinessHours() error {
	configs, err := s.escalationRepo.GetAllBusinessHoursConfigs()
	if err != nil {
		return fmt.Errorf("failed to load business hours configs: %w", err)
	}

	for _, config := range configs {
		calendar, err := models.NewBusinessCalendar(config)
		if err != nil {
			return fmt.Errorf("failed to create calendar for config %d: %w", config.ID, err)
		}
		
		// Map calendar to queues
		// In a real implementation, this would be based on queue configuration
		if config.IsDefault {
			s.businessHours[0] = calendar // Default calendar
		}
	}

	return nil
}

// EscalateTicket manually escalates a ticket
func (s *EscalationService) EscalateTicket(ticketID int, targetLevel models.EscalationLevel, reason string, escalatedBy int) error {
	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return fmt.Errorf("ticket not found: %w", err)
	}

	// Get current escalation level
	currentLevel := s.getCurrentEscalationLevel(ticket)
	
	// Validate escalation
	if !currentLevel.CanEscalateTo(targetLevel) {
		return fmt.Errorf("invalid escalation from level %d to %d", currentLevel, targetLevel)
	}

	// Get escalation path for the queue
	path, err := s.escalationRepo.GetEscalationPath(ticket.QueueID)
	if err != nil {
		return fmt.Errorf("no escalation path defined for queue: %w", err)
	}

	// Get target level configuration
	levelConfig := path.GetLevelConfig(targetLevel)
	if levelConfig == nil {
		return fmt.Errorf("no configuration for level %d", targetLevel)
	}

	// Perform escalation
	history := &models.EscalationHistory{
		TicketID:         ticketID,
		FromLevel:        currentLevel,
		ToLevel:          targetLevel,
		Trigger:          models.EscalationTriggerManual,
		TriggerDetails:   reason,
		PreviousAssignee: &ticket.UserID,
		EscalatedBy:      &escalatedBy,
		EscalatedAt:      time.Now(),
		Notes:            reason,
		AutoEscalated:    false,
	}

	// Assign to new user/group based on configuration
	if levelConfig.AssignToUserID != nil {
		ticket.UserID = *levelConfig.AssignToUserID
		history.NewAssignee = levelConfig.AssignToUserID
	} else if levelConfig.AssignToGroupID != nil {
		// Assign to group using assignment method
		newAssignee := s.assignToGroup(*levelConfig.AssignToGroupID, levelConfig.AssignmentMethod)
		if newAssignee > 0 {
			ticket.UserID = newAssignee
			history.NewAssignee = &newAssignee
		}
	}

	// Update ticket priority if needed
	s.updateTicketPriorityForEscalation(ticket, targetLevel)

	// Save changes
	if err := s.ticketRepo.Update(ticket); err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	// Record escalation history
	if err := s.escalationRepo.CreateEscalationHistory(history); err != nil {
		return fmt.Errorf("failed to record escalation history: %w", err)
	}

	// Send notifications
	s.sendEscalationNotifications(ticket, history, levelConfig)

	// Audit log
	s.auditService.LogAction("ticket.escalated", map[string]interface{}{
		"ticket_id":   ticketID,
		"from_level":  currentLevel,
		"to_level":    targetLevel,
		"escalated_by": escalatedBy,
		"reason":      reason,
	})

	return nil
}

// ProcessEscalationPolicies processes all active escalation policies
func (s *EscalationService) ProcessEscalationPolicies() error {
	policies, err := s.escalationRepo.GetActiveEscalationPolicies()
	if err != nil {
		return fmt.Errorf("failed to get active policies: %w", err)
	}

	for _, policy := range policies {
		if err := s.processPolicy(policy); err != nil {
			// Log error but continue processing other policies
			fmt.Printf("Error processing policy %d: %v\n", policy.ID, err)
		}
	}

	return nil
}

// processPolicy processes a single escalation policy
func (s *EscalationService) processPolicy(policy *models.EscalationPolicy) error {
	// Get tickets that match policy criteria
	tickets, err := s.getTicketsForPolicy(policy)
	if err != nil {
		return err
	}

	for _, ticket := range tickets {
		// Check each rule in the policy
		for _, rule := range policy.Rules {
			if s.shouldEscalate(ticket, rule, policy) {
				if err := s.autoEscalate(ticket, rule, policy); err != nil {
					fmt.Printf("Failed to auto-escalate ticket %d: %v\n", ticket.ID, err)
				}
				break // Only apply first matching rule
			}
		}
	}

	return nil
}

// shouldEscalate checks if a ticket should be escalated based on a rule
func (s *EscalationService) shouldEscalate(ticket *models.Ticket, rule models.EscalationRule, policy *models.EscalationPolicy) bool {
	// Skip if already assigned and rule says to skip
	if rule.SkipIfAssigned && ticket.UserID > 0 {
		return false
	}

	// Build context for condition evaluation
	context := s.buildTicketContext(ticket)

	// Check trigger conditions
	switch rule.Trigger {
	case models.EscalationTriggerTimeElapsed:
		return s.checkTimeElapsed(ticket, rule)
	case models.EscalationTriggerSLAWarning:
		return s.checkSLAWarning(ticket)
	case models.EscalationTriggerSLABreach:
		return s.checkSLABreach(ticket)
	case models.EscalationTriggerPriority:
		return s.checkPriorityTrigger(ticket, rule)
	case models.EscalationTriggerNoResponse:
		return s.checkNoResponse(ticket, rule)
	case models.EscalationTriggerReopenCount:
		return s.checkReopenCount(ticket, rule)
	default:
		return false
	}

	// Evaluate additional conditions
	return rule.ShouldTrigger(context)
}

// checkTimeElapsed checks if enough time has elapsed for escalation
func (s *EscalationService) checkTimeElapsed(ticket *models.Ticket, rule models.EscalationRule) bool {
	elapsed := time.Since(ticket.CreateTime)
	
	if rule.BusinessHours {
		// Calculate business hours elapsed
		calendar := s.getBusinessCalendar(ticket.QueueID)
		if calendar != nil {
			businessHours := calendar.GetBusinessHoursBetween(ticket.CreateTime, time.Now())
			elapsed = time.Duration(businessHours * float64(time.Hour))
		}
	}

	threshold := time.Duration(rule.TimeThreshold) * time.Minute
	return elapsed >= threshold
}

// checkSLAWarning checks if ticket is approaching SLA breach
func (s *EscalationService) checkSLAWarning(ticket *models.Ticket) bool {
	// Check if ticket has active SLA
	// Implementation would check SLA status
	return false
}

// checkSLABreach checks if ticket has breached SLA
func (s *EscalationService) checkSLABreach(ticket *models.Ticket) bool {
	// Check if ticket has breached SLA
	// Implementation would check SLA status
	return false
}

// checkPriorityTrigger checks if priority matches escalation criteria
func (s *EscalationService) checkPriorityTrigger(ticket *models.Ticket, rule models.EscalationRule) bool {
	// Check ticket priority against rule conditions
	return ticket.TicketPriorityID >= 4 // High or Urgent priority
}

// checkNoResponse checks if ticket has no response within threshold
func (s *EscalationService) checkNoResponse(ticket *models.Ticket, rule models.EscalationRule) bool {
	// Check last response time
	// Implementation would check ticket messages
	return false
}

// checkReopenCount checks if ticket reopen count exceeds threshold
func (s *EscalationService) checkReopenCount(ticket *models.Ticket, rule models.EscalationRule) bool {
	// Get ticket reopen count
	// Implementation would check ticket history
	return false
}

// autoEscalate automatically escalates a ticket based on a rule
func (s *EscalationService) autoEscalate(ticket *models.Ticket, rule models.EscalationRule, policy *models.EscalationPolicy) error {
	if !rule.AutoEscalate {
		// Create escalation request for approval
		return s.createEscalationRequest(ticket, rule, policy)
	}

	// Get current escalation level
	currentLevel := s.getCurrentEscalationLevel(ticket)
	targetLevel := rule.Level

	// Record escalation
	history := &models.EscalationHistory{
		TicketID:       ticket.ID,
		PolicyID:       policy.ID,
		RuleID:         rule.ID,
		FromLevel:      currentLevel,
		ToLevel:        targetLevel,
		Trigger:        rule.Trigger,
		TriggerDetails: fmt.Sprintf("Policy: %s, Rule: %d", policy.Name, rule.ID),
		EscalatedAt:    time.Now(),
		AutoEscalated:  true,
	}

	// Execute escalation actions
	for _, action := range rule.Actions {
		if err := s.executeEscalationAction(ticket, action, history); err != nil {
			return err
		}
	}

	// Save escalation history
	if err := s.escalationRepo.CreateEscalationHistory(history); err != nil {
		return fmt.Errorf("failed to record escalation history: %w", err)
	}

	// Send notifications
	if policy.Notifications.NotifyCustomer {
		s.notifyCustomer(ticket, history, policy.Notifications.CustomerTemplate)
	}
	if policy.Notifications.NotifyAssignee {
		s.notifyAssignee(ticket, history, policy.Notifications.InternalTemplate)
	}
	if policy.Notifications.NotifyManager {
		s.notifyManager(ticket, history, policy.Notifications.InternalTemplate)
	}

	return nil
}

// executeEscalationAction executes a single escalation action
func (s *EscalationService) executeEscalationAction(ticket *models.Ticket, action models.EscalationAction, history *models.EscalationHistory) error {
	switch action.Type {
	case "assign_to_user":
		if config, ok := action.Config.(map[string]interface{}); ok {
			if userID, ok := config["user_id"].(int); ok {
				ticket.UserID = userID
				history.NewAssignee = &userID
				return s.ticketRepo.Update(ticket)
			}
		}
	case "assign_to_group":
		if config, ok := action.Config.(map[string]interface{}); ok {
			if groupID, ok := config["group_id"].(int); ok {
				method := "round_robin"
				if m, ok := config["method"].(string); ok {
					method = m
				}
				newAssignee := s.assignToGroup(groupID, method)
				if newAssignee > 0 {
					ticket.UserID = newAssignee
					history.NewAssignee = &newAssignee
					return s.ticketRepo.Update(ticket)
				}
			}
		}
	case "change_priority":
		if config, ok := action.Config.(map[string]interface{}); ok {
			if priority, ok := config["priority"].(int); ok {
				ticket.TicketPriorityID = priority
				return s.ticketRepo.Update(ticket)
			}
		}
	case "notify":
		// Send notification
		return nil
	}
	
	return nil
}

// Helper methods

// getCurrentEscalationLevel gets the current escalation level of a ticket
func (s *EscalationService) getCurrentEscalationLevel(ticket *models.Ticket) models.EscalationLevel {
	// Get from escalation history
	history, err := s.escalationRepo.GetLatestEscalationHistory(ticket.ID)
	if err == nil && history != nil {
		return history.ToLevel
	}
	return models.EscalationLevel1
}

// getBusinessCalendar gets the business calendar for a queue
func (s *EscalationService) getBusinessCalendar(queueID int) *models.BusinessCalendar {
	if calendar, exists := s.businessHours[queueID]; exists {
		return calendar
	}
	// Return default calendar
	return s.businessHours[0]
}

// assignToGroup assigns a ticket to a group member
func (s *EscalationService) assignToGroup(groupID int, method string) int {
	// Implementation would handle different assignment methods
	switch method {
	case "round_robin":
		// Get next available agent in rotation
		return 0
	case "least_loaded":
		// Get agent with fewest tickets
		return 0
	case "skills_based":
		// Match skills to ticket requirements
		return 0
	default:
		return 0
	}
}

// updateTicketPriorityForEscalation updates ticket priority based on escalation level
func (s *EscalationService) updateTicketPriorityForEscalation(ticket *models.Ticket, level models.EscalationLevel) {
	// Increase priority based on escalation level
	if level >= models.EscalationLevel3 && ticket.TicketPriorityID < 4 {
		ticket.TicketPriorityID = 4 // High priority
	}
	if level >= models.EscalationLevel4 && ticket.TicketPriorityID < 5 {
		ticket.TicketPriorityID = 5 // Urgent priority
	}
}

// buildTicketContext builds context for condition evaluation
func (s *EscalationService) buildTicketContext(ticket *models.Ticket) map[string]interface{} {
	return map[string]interface{}{
		"ticket_id":     ticket.ID,
		"priority":      ticket.TicketPriorityID,
		"status":        ticket.TicketStateID,
		"queue_id":      ticket.QueueID,
		"created_at":    ticket.CreateTime,
		"updated_at":    ticket.ChangeTime,
		"customer_id":   ticket.CustomerID,
		"assigned_to":   ticket.UserID,
	}
}

// getTicketsForPolicy gets tickets that match policy criteria
func (s *EscalationService) getTicketsForPolicy(policy *models.EscalationPolicy) ([]*models.Ticket, error) {
	// Implementation would query tickets based on policy criteria
	return []*models.Ticket{}, nil
}

// createEscalationRequest creates an escalation request for approval
func (s *EscalationService) createEscalationRequest(ticket *models.Ticket, rule models.EscalationRule, policy *models.EscalationPolicy) error {
	request := &models.EscalationRequest{
		TicketID:    ticket.ID,
		RequestedAt: time.Now(),
		TargetLevel: rule.Level,
		Reason:      fmt.Sprintf("Auto-escalation triggered by policy: %s", policy.Name),
		Priority:    "normal",
		Status:      "pending",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	
	return s.escalationRepo.CreateEscalationRequest(request)
}

// Notification methods

// sendEscalationNotifications sends escalation notifications
func (s *EscalationService) sendEscalationNotifications(ticket *models.Ticket, history *models.EscalationHistory, config *models.EscalationPathLevel) {
	// Notify configured users
	for _, userID := range config.NotifyUsers {
		s.notifyService.SendNotification(userID, fmt.Sprintf(
			"Ticket #%d has been escalated to %s",
			ticket.ID,
			history.ToLevel.GetLevelName(),
		))
	}
	
	// Notify configured groups
	for _, groupID := range config.NotifyGroups {
		// Implementation would notify group members
	}
}

// notifyCustomer sends escalation notification to customer
func (s *EscalationService) notifyCustomer(ticket *models.Ticket, history *models.EscalationHistory, templateID int) {
	// Implementation would send customer notification
}

// notifyAssignee sends escalation notification to new assignee
func (s *EscalationService) notifyAssignee(ticket *models.Ticket, history *models.EscalationHistory, templateID int) {
	if history.NewAssignee != nil {
		s.notifyService.SendNotification(*history.NewAssignee, fmt.Sprintf(
			"You have been assigned escalated ticket #%d",
			ticket.ID,
		))
	}
}

// notifyManager sends escalation notification to manager
func (s *EscalationService) notifyManager(ticket *models.Ticket, history *models.EscalationHistory, templateID int) {
	// Implementation would notify manager
}