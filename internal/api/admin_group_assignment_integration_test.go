package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This test focuses on the EXACT scenario the user reported:
// 1. User has groups in database (we confirmed robbie has admin, OBC)
// 2. But UI/API might not be displaying or updating them correctly
func TestRealGroupAssignmentIssue(t *testing.T) {
	// Manual database check first - this should pass based on our earlier query
	t.Run("Database verification: Robbie should have groups", func(t *testing.T) {
		db, err := database.GetDB()
		if err != nil {
			t.Skip("Database not available")
		}

		// This is the EXACT query we ran earlier that showed admin, OBC
		var groups string
		err = db.QueryRow(`
			SELECT string_agg(g.name, ', ') as groups 
			FROM users u 
			LEFT JOIN group_user gu ON u.id = gu.user_id 
			LEFT JOIN groups g ON gu.group_id = g.id 
			WHERE u.login = 'robbie' 
			GROUP BY u.id, u.login, u.first_name, u.last_name`).Scan(&groups)

		if err != nil {
			t.Logf("Could not find robbie user or query failed: %v", err)
			t.Skip("Robbie user not found - create him first through UI")
		}

		assert.NotEmpty(t, groups, "Robbie should have groups in database")
		assert.Contains(t, groups, "admin", "Robbie should have admin group")
		t.Logf("SUCCESS: Database shows Robbie has groups: %s", groups)
	})

	t.Run("API GET verification: Does HandleAdminUserGet return the groups?", func(t *testing.T) {
		db, err := database.GetDB()
		if err != nil {
			t.Skip("Database not available")
		}

		// Get Robbie's user ID
		var robbieID int
		err = db.QueryRow("SELECT id FROM users WHERE login = 'robbie'").Scan(&robbieID)
		if err != nil {
			t.Skip("Robbie user not found")
		}

		// Test the HandleAdminUserGet function directly
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/admin/users/:id", HandleAdminUserGet)

		// Make request
		req, _ := http.NewRequest("GET", "/admin/users/"+strconv.Itoa(robbieID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code, "API should return 200")
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check if response contains groups
		if data, ok := response["data"].(map[string]interface{}); ok {
			if groups, hasGroups := data["groups"]; hasGroups {
				t.Logf("SUCCESS: API returns groups: %v", groups)
			} else {
				t.Error("FAILED: API does not include groups in response")
			}
		} else {
			t.Error("FAILED: API response malformed")
		}
	})

	t.Run("Form submission test: Can we update Robbie's groups via HandleAdminUserUpdate?", func(t *testing.T) {
		db, err := database.GetDB()
		if err != nil {
			t.Skip("Database not available")
		}

		// Get Robbie's user ID and current data
		var robbieID int
		var firstName, lastName, login string
		err = db.QueryRow("SELECT id, login, first_name, last_name FROM users WHERE login = 'robbie'").Scan(&robbieID, &login, &firstName, &lastName)
		if err != nil {
			t.Skip("Robbie user not found")
		}

		// Test the HandleAdminUserUpdate function
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.PUT("/admin/users/:id", HandleAdminUserUpdate)

		// Create form data that adds a test group
		formData := url.Values{}
		formData.Set("login", login)
		formData.Set("first_name", firstName)
		formData.Set("last_name", lastName)
		formData.Set("valid_id", "1")
		formData.Add("groups", "admin")    // Keep existing group
		formData.Add("groups", "OBC")      // Keep existing group
		formData.Add("groups", "users")    // Add another group if it exists

		// Make request
		req, _ := http.NewRequest("PUT", "/admin/users/"+strconv.Itoa(robbieID),
			strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check response
		t.Logf("Update response status: %d", w.Code)
		t.Logf("Update response body: %s", w.Body.String())

		// Now check if the database was actually updated
		var newGroups string
		err = db.QueryRow(`
			SELECT string_agg(g.name, ', ') as groups 
			FROM users u 
			LEFT JOIN group_user gu ON u.id = gu.user_id 
			LEFT JOIN groups g ON gu.group_id = g.id 
			WHERE u.login = 'robbie' 
			GROUP BY u.id, u.login, u.first_name, u.last_name`).Scan(&newGroups)

		if err == nil {
			t.Logf("Groups after update: %s", newGroups)
		} else {
			t.Logf("Could not query groups after update: %v", err)
		}
	})
}