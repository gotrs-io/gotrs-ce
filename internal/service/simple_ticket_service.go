package service

import (
	"fmt"

	"github.com/gotrs-io/gotrs-ce/internal/models"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
)

// SimpleTicketService handles ticket operations without database transactions
// This is for development/testing with in-memory repository
type SimpleTicketService struct {
	ticketRepo repository.ITicketRepository
}

// NewSimpleTicketService creates a new simple ticket service
func NewSimpleTicketService(repo repository.ITicketRepository) *SimpleTicketService {
	return &SimpleTicketService{
		ticketRepo: repo,
	}
}

// CreateTicket creates a new ticket
func (s *SimpleTicketService) CreateTicket(ticket *models.Ticket) error {
	if ticket == nil {
		return fmt.Errorf("ticket cannot be nil")
	}
	
	// Validate required fields
	if ticket.Title == "" {
		return fmt.Errorf("ticket title is required")
	}
	
	// Create the ticket
	return s.ticketRepo.Create(ticket)
}

// GetTicket retrieves a ticket by ID
func (s *SimpleTicketService) GetTicket(ticketID uint) (*models.Ticket, error) {
	return s.ticketRepo.GetByID(ticketID)
}

// UpdateTicket updates an existing ticket
func (s *SimpleTicketService) UpdateTicket(ticket *models.Ticket) error {
	if ticket == nil {
		return fmt.Errorf("ticket cannot be nil")
	}
	
	return s.ticketRepo.Update(ticket)
}

// DeleteTicket deletes a ticket
func (s *SimpleTicketService) DeleteTicket(ticketID uint) error {
	return s.ticketRepo.Delete(ticketID)
}

// ListTickets returns a paginated list of tickets
func (s *SimpleTicketService) ListTickets(req *models.TicketListRequest) (*models.TicketListResponse, error) {
	if req == nil {
		req = &models.TicketListRequest{
			Page:    1,
			PerPage: 20,
		}
	}
	
	return s.ticketRepo.List(req)
}

// SimpleTicketMessage is a simplified message model
type SimpleTicketMessage struct {
	TicketID  uint   `json:"ticket_id"`
	Body      string `json:"body"`
	CreatedBy uint   `json:"created_by"`
	IsPublic  bool   `json:"is_public"`
}

// AddMessage adds a message to a ticket (simplified - not fully implemented)
func (s *SimpleTicketService) AddMessage(ticketID uint, message *SimpleTicketMessage) error {
	// For now, just validate the ticket exists
	_, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return fmt.Errorf("ticket not found: %w", err)
	}
	
	// TODO: Implement message storage
	return nil
}