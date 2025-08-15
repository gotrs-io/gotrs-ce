package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := NewRouter(db, "test-secret")
	
	assert.NotNil(t, router)
	assert.NotNil(t, router.engine)
	// Internal fields are private, just check that router is created
}

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := NewRouter(db, "test-secret")
	router.SetupRoutes()

	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
	}{
		{
			name:       "Health check endpoint",
			method:     "GET",
			path:       "/health",
			statusCode: http.StatusOK,
		},
		{
			name:       "API v1 status endpoint",
			method:     "GET",
			path:       "/api/v1/status",
			statusCode: http.StatusOK,
		},
		{
			name:       "Auth login endpoint exists",
			method:     "POST",
			path:       "/api/v1/auth/login",
			statusCode: http.StatusBadRequest, // Will fail due to no body
		},
		{
			name:       "Auth refresh endpoint exists",
			method:     "POST",
			path:       "/api/v1/auth/refresh",
			statusCode: http.StatusBadRequest, // Will fail due to no body
		},
		{
			name:       "Auth logout endpoint exists",
			method:     "POST",
			path:       "/api/v1/auth/logout",
			statusCode: http.StatusOK,
		},
		{
			name:       "Protected route requires auth",
			method:     "GET",
			path:       "/api/v1/auth/me",
			statusCode: http.StatusUnauthorized, // No token
		},
		{
			name:       "Tickets endpoint requires auth",
			method:     "GET",
			path:       "/api/v1/tickets",
			statusCode: http.StatusUnauthorized, // No token
		},
		{
			name:       "Users endpoint requires auth",
			method:     "GET",
			path:       "/api/v1/users",
			statusCode: http.StatusUnauthorized, // No token
		},
		{
			name:       "Non-existent route returns 404",
			method:     "GET",
			path:       "/api/v1/nonexistent",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			
			router.engine.ServeHTTP(w, req)
			
			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestProtectedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := NewRouter(db, "test-secret")
	router.SetupRoutes()

	// Create a valid JWT token
	jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
	token, err := jwtManager.GenerateToken(1, "test@example.com", "Agent", 1)
	require.NoError(t, err)

	tests := []struct {
		name       string
		method     string
		path       string
		token      string
		statusCode int
	}{
		{
			name:       "Auth me endpoint with valid token",
			method:     "GET",
			path:       "/api/v1/auth/me",
			token:      token,
			statusCode: http.StatusOK,
		},
		{
			name:       "Auth me endpoint without token",
			method:     "GET",
			path:       "/api/v1/auth/me",
			token:      "",
			statusCode: http.StatusUnauthorized,
		},
		{
			name:       "Auth me endpoint with invalid token",
			method:     "GET",
			path:       "/api/v1/auth/me",
			token:      "invalid.token.here",
			statusCode: http.StatusUnauthorized,
		},
		{
			name:       "Tickets endpoint with valid token",
			method:     "GET",
			path:       "/api/v1/tickets",
			token:      token,
			statusCode: http.StatusOK,
		},
		{
			name:       "Users endpoint with valid token",
			method:     "GET",
			path:       "/api/v1/users",
			token:      token,
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()
			
			router.engine.ServeHTTP(w, req)
			
			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestRoleBasedAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := NewRouter(db, "test-secret")
	router.SetupRoutes()

	jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)

	// Create tokens for different roles
	adminToken, err := jwtManager.GenerateToken(1, "admin@example.com", "Admin", 1)
	require.NoError(t, err)
	
	agentToken, err := jwtManager.GenerateToken(2, "agent@example.com", "Agent", 1)
	require.NoError(t, err)
	
	customerToken, err := jwtManager.GenerateToken(3, "customer@example.com", "Customer", 1)
	require.NoError(t, err)

	tests := []struct {
		name       string
		method     string
		path       string
		token      string
		role       string
		statusCode int
	}{
		{
			name:       "Admin can access users endpoint",
			method:     "GET",
			path:       "/api/v1/users",
			token:      adminToken,
			role:       "Admin",
			statusCode: http.StatusOK,
		},
		{
			name:       "Agent can access users endpoint",
			method:     "GET",
			path:       "/api/v1/users",
			token:      agentToken,
			role:       "Agent",
			statusCode: http.StatusOK,
		},
		{
			name:       "Customer cannot access users endpoint",
			method:     "GET",
			path:       "/api/v1/users",
			token:      customerToken,
			role:       "Customer",
			statusCode: http.StatusForbidden,
		},
		{
			name:       "Admin can access admin endpoints",
			method:     "GET",
			path:       "/api/v1/admin/dashboard",
			token:      adminToken,
			role:       "Admin",
			statusCode: http.StatusOK,
		},
		{
			name:       "Agent cannot access admin endpoints",
			method:     "GET",
			path:       "/api/v1/admin/dashboard",
			token:      agentToken,
			role:       "Agent",
			statusCode: http.StatusForbidden,
		},
		{
			name:       "Customer can access their tickets",
			method:     "GET",
			path:       "/api/v1/tickets",
			token:      customerToken,
			role:       "Customer",
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			w := httptest.NewRecorder()
			
			router.engine.ServeHTTP(w, req)
			
			assert.Equal(t, tt.statusCode, w.Code, "Role %s accessing %s", tt.role, tt.path)
		})
	}
}

func TestGetEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := NewRouter(db, "test-secret")
	engine := router.GetEngine()
	
	assert.NotNil(t, engine)
	assert.IsType(t, &gin.Engine{}, engine)
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := NewRouter(db, "test-secret")
	router.SetupRoutes()
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.engine.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestCORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := NewRouter(db, "test-secret")
	router.SetupRoutes()

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	
	router.engine.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func TestRouteGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := NewRouter(db, "test-secret")
	router.SetupRoutes()

	// Get all routes
	routes := router.engine.Routes()
	
	// Check that routes are properly grouped
	apiV1Routes := 0
	authRoutes := 0
	ticketRoutes := 0
	userRoutes := 0
	adminRoutes := 0
	
	for _, route := range routes {
		if len(route.Path) > 7 && route.Path[:7] == "/api/v1" {
			apiV1Routes++
			
			if len(route.Path) > 12 && route.Path[:12] == "/api/v1/auth" {
				authRoutes++
			} else if len(route.Path) > 15 && route.Path[:15] == "/api/v1/tickets" {
				ticketRoutes++
			} else if len(route.Path) > 13 && route.Path[:13] == "/api/v1/users" {
				userRoutes++
			} else if len(route.Path) > 13 && route.Path[:13] == "/api/v1/admin" {
				adminRoutes++
			}
		}
	}
	
	assert.Greater(t, apiV1Routes, 0, "Should have API v1 routes")
	assert.Greater(t, authRoutes, 0, "Should have auth routes")
	assert.Greater(t, ticketRoutes, 0, "Should have ticket routes")
	assert.Greater(t, userRoutes, 0, "Should have user routes")
	assert.Greater(t, adminRoutes, 0, "Should have admin routes")
}

func BenchmarkRouting(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	
	db, _, _ := sqlmock.New()
	defer db.Close()
	
	router := NewRouter(db, "test-secret")
	router.SetupRoutes()
	
	req := httptest.NewRequest("GET", "/health", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.engine.ServeHTTP(w, req)
	}
}

func BenchmarkProtectedRoute(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	
	db, _, _ := sqlmock.New()
	defer db.Close()
	
	router := NewRouter(db, "test-secret")
	router.SetupRoutes()
	
	jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
	token, _ := jwtManager.GenerateToken(1, "test@example.com", "Agent", 1)
	
	req := httptest.NewRequest("GET", "/api/v1/tickets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.engine.ServeHTTP(w, req)
	}
}