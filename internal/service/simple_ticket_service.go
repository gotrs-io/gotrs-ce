package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/gotrs-io/gotrs-ce/internal/models"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
)

// SimpleTicketService handles ticket operations without database transactions
// This is for development/testing with in-memory repository
type SimpleTicketService struct {
	ticketRepo repository.ITicketRepository
	messages   map[uint][]*SimpleTicketMessage // In-memory message storage by ticket ID
	messagesMu sync.RWMutex                   // Mutex for thread-safe access to messages
	nextMsgID  uint                          // Auto-incrementing message ID
}

// NewSimpleTicketService creates a new simple ticket service
func NewSimpleTicketService(repo repository.ITicketRepository) *SimpleTicketService {
	return &SimpleTicketService{
		ticketRepo: repo,
		messages:   make(map[uint][]*SimpleTicketMessage),
		nextMsgID:  1,
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
	ID          uint                    `json:"id"`
	TicketID    uint                    `json:"ticket_id"`
	Body        string                  `json:"body"`
	Subject     string                  `json:"subject"`
	CreatedBy   uint                    `json:"created_by"`
	AuthorName  string                  `json:"author_name"`
	AuthorEmail string                  `json:"author_email"`
	AuthorType  string                  `json:"author_type"` // "Customer", "Agent", "System"
	IsPublic    bool                    `json:"is_public"`
	IsInternal  bool                    `json:"is_internal"`
	CreatedAt   time.Time               `json:"created_at"`
	Attachments []*SimpleAttachment     `json:"attachments,omitempty"`
}

// SimpleAttachment is a simplified attachment model
type SimpleAttachment struct {
	ID          uint      `json:"id"`
	MessageID   uint      `json:"message_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	URL         string    `json:"url"` // Download URL for the attachment
	CreatedAt   time.Time `json:"created_at"`
}

// AddMessage adds a message to a ticket
func (s *SimpleTicketService) AddMessage(ticketID uint, message *SimpleTicketMessage) error {
	// Validate the ticket exists
	_, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return fmt.Errorf("ticket not found: %w", err)
	}
	
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}
	
	if message.Body == "" {
		return fmt.Errorf("message body is required")
	}
	
	// Lock for thread-safe access
	s.messagesMu.Lock()
	defer s.messagesMu.Unlock()
	
	// Set message metadata
	message.ID = s.nextMsgID
	message.TicketID = ticketID
	message.CreatedAt = time.Now()
	s.nextMsgID++
	
	// Set default author info if not provided
	if message.AuthorName == "" {
		if message.AuthorType == "Customer" {
			message.AuthorName = "Customer"
		} else {
			message.AuthorName = "System User"
		}
	}
	
	if message.AuthorType == "" {
		message.AuthorType = "System"
	}
	
	// Store the message
	if s.messages[ticketID] == nil {
		s.messages[ticketID] = make([]*SimpleTicketMessage, 0)
	}
	s.messages[ticketID] = append(s.messages[ticketID], message)
	
	return nil
}

// GetMessages retrieves all messages for a ticket
func (s *SimpleTicketService) GetMessages(ticketID uint) ([]*SimpleTicketMessage, error) {
	// Validate the ticket exists
	_, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}
	
	s.messagesMu.RLock()
	defer s.messagesMu.RUnlock()
	
	messages := s.messages[ticketID]
	if messages == nil {
		return make([]*SimpleTicketMessage, 0), nil
	}
	
	// Return a copy to prevent external modification
	result := make([]*SimpleTicketMessage, len(messages))
	copy(result, messages)
	return result, nil
}