package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "github.com/gotrs-io/gotrs-ce/internal/api"
)

func TestUpdateTicket_BasicFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.PUT("/api/v1/tickets/:id", func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("is_authenticated", true)
		HandleUpdateTicketAPI(c)
	})

	tests := []struct {
		name       string
		ticketID   string
		payload    map[string]interface{}
		wantStatus int
		checkBody  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:     "update title",
			ticketID: "1",
			payload: map[string]interface{}{
				"title": "Updated ticket title",
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, "Updated ticket title", data["title"])
			},
		},
		{
			name:     "update priority",
			ticketID: "1",
			payload: map[string]interface{}{
				"priority_id": 5,
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, float64(5), data["priority_id"])
			},
		},
		{
			name:     "update queue",
			ticketID: "1",
			payload: map[string]interface{}{
				"queue_id": 2,
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, float64(2), data["queue_id"])
			},
		},
		{
			name:     "update state",
			ticketID: "1",
			payload: map[string]interface{}{
				"state_id": 2, // closed successful
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, float64(2), data["state_id"])
			},
		},
		{
			name:     "update type",
			ticketID: "1",
			payload: map[string]interface{}{
				"type_id": 2,
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, float64(2), data["type_id"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.payload)
			require.NoError(t, err)
			
			req := httptest.NewRequest("PUT", "/api/v1/tickets/"+tt.ticketID, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.wantStatus, w.Code)
			
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.checkBody != nil {
				tt.checkBody(t, response)
			}
		})
	}
}

func TestUpdateTicket_Assignment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.PUT("/api/v1/tickets/:id", func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("is_authenticated", true)
		HandleUpdateTicketAPI(c)
	})

	tests := []struct {
		name       string
		ticketID   string
		payload    map[string]interface{}
		wantStatus int
		checkBody  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:     "assign to user",
			ticketID: "1",
			payload: map[string]interface{}{
				"responsible_user_id": 5,
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, float64(5), data["responsible_user_id"])
			},
		},
		{
			name:     "change owner",
			ticketID: "1",
			payload: map[string]interface{}{
				"user_id": 3,
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, float64(3), data["user_id"])
			},
		},
		{
			name:     "unassign ticket",
			ticketID: "1",
			payload: map[string]interface{}{
				"responsible_user_id": nil,
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Nil(t, data["responsible_user_id"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.payload)
			require.NoError(t, err)
			
			req := httptest.NewRequest("PUT", "/api/v1/tickets/"+tt.ticketID, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.wantStatus, w.Code)
			
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.checkBody != nil {
				tt.checkBody(t, response)
			}
		})
	}
}

func TestUpdateTicket_CustomerInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.PUT("/api/v1/tickets/:id", func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("is_authenticated", true)
		HandleUpdateTicketAPI(c)
	})

	tests := []struct {
		name       string
		ticketID   string
		payload    map[string]interface{}
		wantStatus int
		checkBody  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:     "update customer user",
			ticketID: "1",
			payload: map[string]interface{}{
				"customer_user_id": "newcustomer@example.com",
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, "newcustomer@example.com", data["customer_user_id"])
			},
		},
		{
			name:     "update customer company",
			ticketID: "1",
			payload: map[string]interface{}{
				"customer_id": "company123",
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, "company123", data["customer_id"])
			},
		},
		{
			name:     "clear customer info",
			ticketID: "1",
			payload: map[string]interface{}{
				"customer_user_id": "",
				"customer_id":      "",
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, "", data["customer_user_id"])
				assert.Equal(t, "", data["customer_id"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.payload)
			require.NoError(t, err)
			
			req := httptest.NewRequest("PUT", "/api/v1/tickets/"+tt.ticketID, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.wantStatus, w.Code)
			
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.checkBody != nil {
				tt.checkBody(t, response)
			}
		})
	}
}

func TestUpdateTicket_MultipleFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.PUT("/api/v1/tickets/:id", func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("is_authenticated", true)
		HandleUpdateTicketAPI(c)
	})

	payload := map[string]interface{}{
		"title":               "Completely updated ticket",
		"priority_id":         4,
		"queue_id":            3,
		"state_id":            4, // open
		"responsible_user_id": 2,
		"customer_user_id":    "updated@customer.com",
	}
	
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)
	
	req := httptest.NewRequest("PUT", "/api/v1/tickets/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Completely updated ticket", data["title"])
	assert.Equal(t, float64(4), data["priority_id"])
	assert.Equal(t, float64(3), data["queue_id"])
	assert.Equal(t, float64(4), data["state_id"])
	assert.Equal(t, float64(2), data["responsible_user_id"])
	assert.Equal(t, "updated@customer.com", data["customer_user_id"])
}

func TestUpdateTicket_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.PUT("/api/v1/tickets/:id", func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("is_authenticated", true)
		HandleUpdateTicketAPI(c)
	})

	tests := []struct {
		name       string
		ticketID   string
		payload    interface{}
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid ticket ID",
			ticketID:   "abc",
			payload:    map[string]interface{}{"title": "Test"},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid ticket ID",
		},
		{
			name:       "non-existent ticket",
			ticketID:   "999999",
			payload:    map[string]interface{}{"title": "Test"},
			wantStatus: http.StatusNotFound,
			wantError:  "Ticket not found",
		},
		{
			name:       "invalid JSON",
			ticketID:   "1",
			payload:    "not json",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request",
		},
		{
			name:     "empty payload",
			ticketID: "1",
			payload:  map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
			wantError:  "No fields to update",
		},
		{
			name:     "invalid queue_id",
			ticketID: "1",
			payload: map[string]interface{}{
				"queue_id": 99999,
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid queue_id",
		},
		{
			name:     "invalid state_id",
			ticketID: "1",
			payload: map[string]interface{}{
				"state_id": 99999,
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid state_id",
		},
		{
			name:     "invalid priority_id",
			ticketID: "1",
			payload: map[string]interface{}{
				"priority_id": 99999,
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid priority_id",
		},
		{
			name:     "title too long",
			ticketID: "1",
			payload: map[string]interface{}{
				"title": string(make([]byte, 256)),
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "Title too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonData []byte
			var err error
			
			if str, ok := tt.payload.(string); ok {
				jsonData = []byte(str)
			} else {
				jsonData, err = json.Marshal(tt.payload)
				require.NoError(t, err)
			}
			
			req := httptest.NewRequest("PUT", "/api/v1/tickets/"+tt.ticketID, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.wantStatus, w.Code)
			
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			assert.False(t, response["success"].(bool))
			assert.Contains(t, response["error"].(string), tt.wantError)
		})
	}
}

func TestUpdateTicket_Permissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name       string
		setupAuth  func(c *gin.Context)
		ticketID   string
		payload    map[string]interface{}
		wantStatus int
	}{
		{
			name: "authenticated user can update",
			setupAuth: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("is_authenticated", true)
			},
			ticketID: "1",
			payload: map[string]interface{}{
				"title": "Updated by authenticated user",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "unauthenticated user cannot update",
			setupAuth: func(c *gin.Context) {
				// No auth set
			},
			ticketID: "1",
			payload: map[string]interface{}{
				"title": "Should fail",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "customer can only update own tickets",
			setupAuth: func(c *gin.Context) {
				c.Set("user_id", 100)
				c.Set("is_customer", true)
				c.Set("customer_email", "customer@example.com")
				c.Set("is_authenticated", true)
			},
			ticketID: "1",
			payload: map[string]interface{}{
				"title": "Customer update",
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			apiRouter := NewAPIRouter(nil, nil, nil)
			
			router.PUT("/api/v1/tickets/:id", func(c *gin.Context) {
				tt.setupAuth(c)
				apiRouter.HandleUpdateTicket(c)
			})
			
			jsonData, err := json.Marshal(tt.payload)
			require.NoError(t, err)
			
			req := httptest.NewRequest("PUT", "/api/v1/tickets/"+tt.ticketID, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUpdateTicket_LockStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.PUT("/api/v1/tickets/:id", func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("is_authenticated", true)
		HandleUpdateTicketAPI(c)
	})

	tests := []struct {
		name       string
		ticketID   string
		payload    map[string]interface{}
		wantStatus int
		checkBody  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:     "lock ticket",
			ticketID: "1",
			payload: map[string]interface{}{
				"ticket_lock_id": 2, // locked
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, float64(2), data["ticket_lock_id"])
			},
		},
		{
			name:     "unlock ticket",
			ticketID: "1",
			payload: map[string]interface{}{
				"ticket_lock_id": 1, // unlocked
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Equal(t, float64(1), data["ticket_lock_id"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.payload)
			require.NoError(t, err)
			
			req := httptest.NewRequest("PUT", "/api/v1/tickets/"+tt.ticketID, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.wantStatus, w.Code)
			
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.checkBody != nil {
				tt.checkBody(t, response)
			}
		})
	}
}

func TestUpdateTicket_AuditFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiRouter := NewAPIRouter(nil, nil, nil)
	
	router.PUT("/api/v1/tickets/:id", func(c *gin.Context) {
		c.Set("user_id", 5)
		c.Set("is_authenticated", true)
		apiRouter.HandleUpdateTicket(c)
	})

	payload := map[string]interface{}{
		"title": "Check audit fields",
	}
	
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)
	
	req := httptest.NewRequest("PUT", "/api/v1/tickets/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	
	// Check that change_by is updated to current user
	assert.Equal(t, float64(5), data["change_by"])
	
	// Check that change_time is updated (should be recent)
	assert.NotNil(t, data["change_time"])
}