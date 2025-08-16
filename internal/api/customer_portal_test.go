package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCustomerPortal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "should show customer portal homepage",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Customer Portal")
				assert.Contains(t, body, "My Tickets")
				assert.Contains(t, body, "Submit New Ticket")
				assert.Contains(t, body, "Knowledge Base")
				assert.Contains(t, body, "My Profile")
			},
		},
		{
			name:           "should display customer ticket summary",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, `data-metric="open-tickets"`)
				assert.Contains(t, body, `data-metric="resolved-tickets"`)
				assert.Contains(t, body, `data-metric="total-tickets"`)
				assert.Contains(t, body, `data-metric="avg-resolution-time"`)
			},
		},
		{
			name:           "should have quick actions for customers",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Submit Ticket")
				assert.Contains(t, body, "View All Tickets")
				assert.Contains(t, body, "Search Knowledge Base")
				assert.Contains(t, body, "Contact Support")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/portal", handleCustomerPortal)

			req, _ := http.NewRequest("GET", "/portal", nil)
			req.Header.Set("HX-Request", "true")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestCustomerTicketList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "should list customer's tickets",
			query:          "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "My Tickets")
				assert.Contains(t, body, "TICK-2024")
				assert.Contains(t, body, "Status:")
				assert.Contains(t, body, "Created:")
				assert.Contains(t, body, "Last Updated:")
			},
		},
		{
			name:           "should filter by status",
			query:          "?status=open",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, `class="status-open"`)
				assert.NotContains(t, body, `class="status-closed"`)
			},
		},
		{
			name:           "should sort tickets",
			query:          "?sort=created",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "sort=created")
			},
		},
		{
			name:           "should paginate tickets",
			query:          "?page=2",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Previous")
				assert.Contains(t, body, "Next")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/portal/tickets", handleCustomerTickets)

			req, _ := http.NewRequest("GET", "/portal/tickets"+tt.query, nil)
			req.Header.Set("HX-Request", "true")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestCustomerTicketSubmission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		formData       string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "should show ticket submission form",
			method:         "GET",
			formData:       "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Submit New Ticket")
				assert.Contains(t, body, `name="subject"`)
				assert.Contains(t, body, `name="priority"`)
				assert.Contains(t, body, `name="category"`)
				assert.Contains(t, body, `name="description"`)
				assert.Contains(t, body, `type="file"`) // Attachment
			},
		},
		{
			name:           "should validate required fields",
			method:         "POST",
			formData:       "subject=&description=",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Subject is required")
				assert.Contains(t, body, "Description is required")
			},
		},
		{
			name:           "should submit ticket successfully",
			method:         "POST",
			formData:       "subject=Cannot+login&priority=normal&category=technical&description=I+cannot+access+my+account",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Ticket submitted successfully")
				assert.Contains(t, body, "TICK-")
				assert.Contains(t, body, "View Ticket")
			},
		},
		{
			name:           "should show category-specific fields",
			method:         "GET",
			formData:       "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				// Dynamic fields based on category
				assert.Contains(t, body, `hx-trigger="change"`)
				assert.Contains(t, body, `hx-get="/portal/ticket-fields"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/portal/submit-ticket", handleCustomerSubmitTicketForm)
			router.POST("/portal/submit-ticket", handleCustomerSubmitTicket)

			var req *http.Request
			if tt.method == "POST" {
				req, _ = http.NewRequest("POST", "/portal/submit-ticket", strings.NewReader(tt.formData))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req, _ = http.NewRequest("GET", "/portal/submit-ticket", nil)
			}
			req.Header.Set("HX-Request", "true")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestCustomerTicketView(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		ticketID       string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "should show ticket details",
			ticketID:       "TICK-2024-001",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "TICK-2024-001")
				assert.Contains(t, body, "Subject:")
				assert.Contains(t, body, "Status:")
				assert.Contains(t, body, "Priority:")
				assert.Contains(t, body, "Created:")
				assert.Contains(t, body, "Description")
			},
		},
		{
			name:           "should show conversation history",
			ticketID:       "TICK-2024-001",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Conversation")
				assert.Contains(t, body, "Agent response")
				assert.Contains(t, body, "Customer reply")
				assert.Contains(t, body, `class="message-agent"`)
				assert.Contains(t, body, `class="message-customer"`)
			},
		},
		{
			name:           "should allow customer to reply",
			ticketID:       "TICK-2024-001",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Add Reply")
				assert.Contains(t, body, `name="message"`)
				assert.Contains(t, body, `hx-post="/portal/tickets/TICK-2024-001/reply"`)
			},
		},
		{
			name:           "should show attachments",
			ticketID:       "TICK-2024-001",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Attachments")
				assert.Contains(t, body, "Download")
			},
		},
		{
			name:           "should not show ticket from another customer",
			ticketID:       "TICK-2024-999",
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Access denied")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/portal/tickets/:id", handleCustomerTicketView)

			req, _ := http.NewRequest("GET", "/portal/tickets/"+tt.ticketID, nil)
			req.Header.Set("HX-Request", "true")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestCustomerTicketReply(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		ticketID       string
		formData       string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "should add reply to ticket",
			ticketID:       "TICK-2024-001",
			formData:       "message=Thank+you+for+the+update",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Reply added successfully")
				assert.Contains(t, body, "Thank you for the update")
			},
		},
		{
			name:           "should validate empty reply",
			ticketID:       "TICK-2024-001",
			formData:       "message=",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Message cannot be empty")
			},
		},
		{
			name:           "should reopen closed ticket on reply",
			ticketID:       "TICK-2024-002",
			formData:       "message=I+still+have+this+issue",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Ticket reopened")
				assert.Contains(t, body, `class="status-open"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/portal/tickets/:id/reply", handleCustomerTicketReply)

			req, _ := http.NewRequest("POST", "/portal/tickets/"+tt.ticketID+"/reply",
				strings.NewReader(tt.formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("HX-Request", "true")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestCustomerProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		formData       string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "should show customer profile",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "My Profile")
				assert.Contains(t, body, "Email:")
				assert.Contains(t, body, "Name:")
				assert.Contains(t, body, "Phone:")
				assert.Contains(t, body, "Company:")
				assert.Contains(t, body, "Notification Preferences")
			},
		},
		{
			name:           "should update profile",
			method:         "POST",
			formData:       "name=John+Doe&phone=555-1234&company=Acme+Corp",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Profile updated successfully")
			},
		},
		{
			name:           "should update notification preferences",
			method:         "POST",
			formData:       "email_notifications=true&sms_notifications=false",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Preferences updated")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/portal/profile", handleCustomerProfile)
			router.POST("/portal/profile", handleCustomerUpdateProfile)

			var req *http.Request
			if tt.method == "POST" {
				req, _ = http.NewRequest("POST", "/portal/profile", strings.NewReader(tt.formData))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req, _ = http.NewRequest("GET", "/portal/profile", nil)
			}
			req.Header.Set("HX-Request", "true")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestCustomerKnowledgeBase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "should show knowledge base home",
			query:          "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Knowledge Base")
				assert.Contains(t, body, "Popular Articles")
				assert.Contains(t, body, "Categories")
				assert.Contains(t, body, "Search")
			},
		},
		{
			name:           "should search articles",
			query:          "?search=password",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Search Results")
				assert.Contains(t, body, "password")
				assert.Contains(t, body, "How to reset your password")
			},
		},
		{
			name:           "should filter by category",
			query:          "?category=technical",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Technical")
				assert.Contains(t, body, "article-category-technical")
			},
		},
		{
			name:           "should show article helpful voting",
			query:          "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Was this helpful?")
				assert.Contains(t, body, `hx-post="/portal/kb/vote"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/portal/kb", handleCustomerKnowledgeBase)

			req, _ := http.NewRequest("GET", "/portal/kb"+tt.query, nil)
			req.Header.Set("HX-Request", "true")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestCustomerSatisfaction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		ticketID       string
		formData       string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "should show satisfaction survey for resolved ticket",
			ticketID:       "TICK-2024-001",
			formData:       "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Rate Your Experience")
				assert.Contains(t, body, `name="rating"`)
				assert.Contains(t, body, `name="feedback"`)
				assert.Contains(t, body, "1-5 stars")
			},
		},
		{
			name:           "should submit satisfaction rating",
			ticketID:       "TICK-2024-001",
			formData:       "rating=5&feedback=Excellent+service",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Thank you for your feedback")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/portal/tickets/:id/satisfaction", handleCustomerSatisfactionForm)
			router.POST("/portal/tickets/:id/satisfaction", handleCustomerSatisfactionSubmit)

			var req *http.Request
			if tt.formData != "" {
				req, _ = http.NewRequest("POST", "/portal/tickets/"+tt.ticketID+"/satisfaction",
					strings.NewReader(tt.formData))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req, _ = http.NewRequest("GET", "/portal/tickets/"+tt.ticketID+"/satisfaction", nil)
			}
			req.Header.Set("HX-Request", "true")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}