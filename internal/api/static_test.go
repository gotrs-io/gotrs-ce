package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestStaticFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewSimpleRouter()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Favicon ICO should be served",
			path:           "/static/favicon.ico",
			expectedStatus: http.StatusOK,
			description:    "Legacy favicon format",
		},
		{
			name:           "Favicon SVG should be served",
			path:           "/static/favicon.svg",
			expectedStatus: http.StatusOK,
			description:    "Modern favicon format",
		},
		{
			name:           "Root favicon.ico should be served",
			path:           "/favicon.ico",
			expectedStatus: http.StatusOK,
			description:    "Browser default location",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, 
				"Path %s should return %d, got %d", 
				tt.path, tt.expectedStatus, w.Code)
		})
	}
}