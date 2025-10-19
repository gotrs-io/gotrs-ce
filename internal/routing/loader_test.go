package routing

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoadYAMLRoutesPreservesPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	yaml := `apiVersion: v1
kind: RouteGroup
metadata:
  name: api-notifications
  description: "Test notification routes"
  namespace: default
  enabled: true
spec:
  prefix: /api
  middleware:
    - auth
  routes:
    - path: /notifications/pending
      method: GET
      handler: handlePendingReminderFeed
`
	file := filepath.Join(dir, "api-notifications.yaml")
	if err := os.WriteFile(file, []byte(yaml), 0o644); err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	router := gin.New()
	registry := NewHandlerRegistry()
	if err := registry.RegisterMiddleware("auth", func(c *gin.Context) { c.Next() }); err != nil {
		t.Fatalf("register middleware failed: %v", err)
	}
	if err := registry.Register("handlePendingReminderFeed", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	}); err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	if err := LoadYAMLRoutes(router, dir, registry); err != nil {
		t.Fatalf("LoadYAMLRoutes failed: %v", err)
	}

	t.Run("prefixed route", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/notifications/pending", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", w.Code)
		}
	})

	t.Run("no stray root route", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/notifications/pending", nil)
		router.ServeHTTP(w, req)
		if w.Code == http.StatusNoContent {
			t.Fatalf("unexpected handler without prefix: status %d", w.Code)
		}
	})
}
