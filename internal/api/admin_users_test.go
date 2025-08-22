package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAdminUsersPageLoad(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Setup routes (we'll need to set up the actual routes)
	// This will fail initially - that's the point of TDD
	setupHTMXRoutesWithAuth(router, nil, nil, nil)
	
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		checkContent   []string
		shouldNotHave  []string
	}{
		{
			name:           "Admin users page should load",
			url:            "/admin/users",
			expectedStatus: http.StatusOK,
			checkContent: []string{
				"<table",                    // Should have a table
				"users",                     // Should mention users
				"class=\"",                  // Should have HTML classes
			},
			shouldNotHave: []string{
				"Template error",            // No template errors
				"Guru Meditation",           // No error screens
				"500",                       // No error codes
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest("GET", tt.url, nil)
			
			// Add authentication cookie (using demo session)
			req.AddCookie(&http.Cookie{
				Name:  "access_token",
				Value: "demo_session_test",
			})
			
			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Check status
			assert.Equal(t, tt.expectedStatus, w.Code, 
				"Expected status %d but got %d", tt.expectedStatus, w.Code)
			
			// Check content exists
			body := w.Body.String()
			for _, content := range tt.checkContent {
				assert.Contains(t, body, content, 
					"Response should contain '%s'", content)
			}
			
			// Check unwanted content doesn't exist
			for _, content := range tt.shouldNotHave {
				assert.NotContains(t, body, content, 
					"Response should NOT contain '%s'", content)
			}
		})
	}
}

func TestAPIUsersEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Setup routes
	setupHTMXRoutesWithAuth(router, nil, nil, nil)
	
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "API users endpoint should return JSON",
			url:            "/api/users",
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
		{
			name:           "API single user endpoint should return JSON",
			url:            "/api/users/1",
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest("GET", tt.url, nil)
			
			// Add authentication
			req.AddCookie(&http.Cookie{
				Name:  "access_token",
				Value: "demo_session_test",
			})
			
			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Check status
			assert.Equal(t, tt.expectedStatus, w.Code,
				"Expected status %d but got %d for %s", 
				tt.expectedStatus, w.Code, tt.url)
			
			// Check content type
			contentType := w.Header().Get("Content-Type")
			assert.Contains(t, contentType, tt.expectedType,
				"Expected content type %s but got %s", 
				tt.expectedType, contentType)
			
			// If JSON, check it's valid
			if tt.expectedType == "application/json" {
				assert.NotContains(t, w.Body.String(), "<!DOCTYPE",
					"JSON endpoint should not return HTML")
			}
		})
	}
}

func TestUserCRUDOperations(t *testing.T) {
	// This will test actual CRUD operations
	t.Run("Should create a user", func(t *testing.T) {
		// TODO: Implement
		t.Skip("Implement after basic endpoints work")
	})
	
	t.Run("Should update a user", func(t *testing.T) {
		// TODO: Implement
		t.Skip("Implement after basic endpoints work")
	})
	
	t.Run("Should delete/deactivate a user", func(t *testing.T) {
		// TODO: Implement
		t.Skip("Implement after basic endpoints work")
	})
}