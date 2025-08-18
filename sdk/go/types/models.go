package types

import (
	"time"
)

// Ticket represents a support ticket
type Ticket struct {
	ID           uint                   `json:"id"`
	TicketNumber string                 `json:"ticket_number"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Status       string                 `json:"status"`
	Priority     string                 `json:"priority"`
	Type         string                 `json:"type"`
	QueueID      uint                   `json:"queue_id"`
	CustomerID   uint                   `json:"customer_id"`
	AssignedTo   *uint                  `json:"assigned_to,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	ClosedAt     *time.Time             `json:"closed_at,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	Customer     *User                  `json:"customer,omitempty"`
	AssignedUser *User                  `json:"assigned_user,omitempty"`
	Queue        *Queue                 `json:"queue,omitempty"`
	Messages     []TicketMessage        `json:"messages,omitempty"`
	Attachments  []Attachment           `json:"attachments,omitempty"`
}

// TicketMessage represents a message in a ticket
type TicketMessage struct {
	ID           uint                   `json:"id"`
	TicketID     uint                   `json:"ticket_id"`
	Content      string                 `json:"content"`
	MessageType  string                 `json:"message_type"`
	IsInternal   bool                   `json:"is_internal"`
	AuthorID     uint                   `json:"author_id"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Author       *User                  `json:"author,omitempty"`
	Attachments  []Attachment           `json:"attachments,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// User represents a user in the system
type User struct {
	ID          uint      `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Login       string    `json:"login"`
	Title       string    `json:"title"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at"`
}

// Queue represents a ticket queue
type Queue struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID          uint      `json:"id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	TicketID    uint      `json:"ticket_id"`
	MessageID   *uint     `json:"message_id,omitempty"`
	UploadedBy  uint      `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// Group represents a user group
type Group struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalTickets      int            `json:"total_tickets"`
	OpenTickets       int            `json:"open_tickets"`
	ClosedTickets     int            `json:"closed_tickets"`
	PendingTickets    int            `json:"pending_tickets"`
	OverdueTickets    int            `json:"overdue_tickets"`
	UnassignedTickets int            `json:"unassigned_tickets"`
	MyTickets         int            `json:"my_tickets"`
	TicketsByStatus   map[string]int `json:"tickets_by_status"`
	TicketsByPriority map[string]int `json:"tickets_by_priority"`
	TicketsByQueue    map[string]int `json:"tickets_by_queue"`
}

// SearchResult represents search results
type SearchResult struct {
	TotalCount int      `json:"total_count"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	Tickets    []Ticket `json:"tickets"`
}

// InternalNote represents an internal note
type InternalNote struct {
	ID          uint      `json:"id"`
	TicketID    uint      `json:"ticket_id"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	IsImportant bool      `json:"is_important"`
	IsPinned    bool      `json:"is_pinned"`
	Tags        []string  `json:"tags"`
	AuthorID    uint      `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"author_email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	EditedAt    time.Time `json:"edited_at"`
	EditedBy    uint      `json:"edited_by"`
}

// NoteTemplate represents a note template
type NoteTemplate struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Content     string   `json:"content"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	IsImportant bool     `json:"is_important"`
	CreatedBy   uint     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LDAPUser represents a user from LDAP
type LDAPUser struct {
	DN          string            `json:"dn"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	FirstName   string            `json:"first_name"`
	LastName    string            `json:"last_name"`
	DisplayName string            `json:"display_name"`
	Phone       string            `json:"phone"`
	Department  string            `json:"department"`
	Title       string            `json:"title"`
	Manager     string            `json:"manager"`
	Groups      []string          `json:"groups"`
	Attributes  map[string]string `json:"attributes"`
	ObjectGUID  string            `json:"object_guid"`
	ObjectSID   string            `json:"object_sid"`
	LastLogin   time.Time         `json:"last_login"`
	IsActive    bool              `json:"is_active"`
}

// LDAPSyncResult represents the result of an LDAP sync operation
type LDAPSyncResult struct {
	UsersFound    int           `json:"users_found"`
	UsersCreated  int           `json:"users_created"`
	UsersUpdated  int           `json:"users_updated"`
	UsersDisabled int           `json:"users_disabled"`
	GroupsFound   int           `json:"groups_found"`
	GroupsCreated int           `json:"groups_created"`
	GroupsUpdated int           `json:"groups_updated"`
	Errors        []string      `json:"errors"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Duration      time.Duration `json:"duration"`
	DryRun        bool          `json:"dry_run"`
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Events      []string  `json:"events"`
	Secret      string    `json:"secret,omitempty"`
	IsActive    bool      `json:"is_active"`
	RetryCount  int       `json:"retry_count"`
	Timeout     int       `json:"timeout"`
	Headers     map[string]string `json:"headers,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastFiredAt *time.Time `json:"last_fired_at,omitempty"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID          uint      `json:"id"`
	WebhookID   uint      `json:"webhook_id"`
	Event       string    `json:"event"`
	Payload     string    `json:"payload"`
	StatusCode  int       `json:"status_code"`
	Response    string    `json:"response"`
	Success     bool      `json:"success"`
	Attempt     int       `json:"attempt"`
	DeliveredAt time.Time `json:"delivered_at"`
}

// Request/Response types for API operations

// TicketCreateRequest represents a request to create a ticket
type TicketCreateRequest struct {
	Title        string                 `json:"title" binding:"required"`
	Description  string                 `json:"description" binding:"required"`
	Priority     string                 `json:"priority"`
	Type         string                 `json:"type"`
	QueueID      uint                   `json:"queue_id"`
	CustomerID   uint                   `json:"customer_id"`
	AssignedTo   *uint                  `json:"assigned_to,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// TicketUpdateRequest represents a request to update a ticket
type TicketUpdateRequest struct {
	Title        *string                `json:"title,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Status       *string                `json:"status,omitempty"`
	Priority     *string                `json:"priority,omitempty"`
	Type         *string                `json:"type,omitempty"`
	QueueID      *uint                  `json:"queue_id,omitempty"`
	AssignedTo   *uint                  `json:"assigned_to,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// TicketListOptions represents options for listing tickets
type TicketListOptions struct {
	Page         int      `json:"page,omitempty"`
	PageSize     int      `json:"page_size,omitempty"`
	Status       []string `json:"status,omitempty"`
	Priority     []string `json:"priority,omitempty"`
	QueueID      []uint   `json:"queue_id,omitempty"`
	AssignedTo   *uint    `json:"assigned_to,omitempty"`
	CustomerID   *uint    `json:"customer_id,omitempty"`
	Search       string   `json:"search,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	CreatedAfter *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	SortBy       string   `json:"sort_by,omitempty"`
	SortOrder    string   `json:"sort_order,omitempty"`
}

// TicketListResponse represents a response from listing tickets
type TicketListResponse struct {
	Tickets    []Ticket `json:"tickets"`
	TotalCount int      `json:"total_count"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

// MessageCreateRequest represents a request to create a message
type MessageCreateRequest struct {
	Content      string                 `json:"content" binding:"required"`
	MessageType  string                 `json:"message_type"`
	IsInternal   bool                   `json:"is_internal"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// UserCreateRequest represents a request to create a user
type UserCreateRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Login     string `json:"login" binding:"required"`
	Title     string `json:"title"`
	Role      string `json:"role"`
	Password  string `json:"password" binding:"required,min=8"`
}

// UserUpdateRequest represents a request to update a user
type UserUpdateRequest struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Title     *string `json:"title,omitempty"`
	Role      *string `json:"role,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// AuthLoginRequest represents a login request
type AuthLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthLoginResponse represents a login response
type AuthLoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         User      `json:"user"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}