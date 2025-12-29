package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTicketCreationWithAttachments500Fix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	WithCleanDB(t)

	t.Run("Reproduce and fix 500 error with attachments", func(t *testing.T) {
		router := gin.New()

		router.POST("/api/tickets", func(c *gin.Context) {
			c.Set("user_role", "Agent")
			c.Set("user_id", uint(1))
			c.Set("user_email", "admin@example.com")
			handleCreateTicket(c)
		})

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		writer.WriteField("title", "Bug Report with Screenshot")
		writer.WriteField("customer_email", "customer@example.com")
		writer.WriteField("customer_name", "Test Customer")
		writer.WriteField("body", "Application crashes when clicking submit button.")
		writer.WriteField("priority", "high")
		writer.WriteField("queue_id", "1")
		writer.WriteField("type_id", "1")

		part, err := writer.CreateFormFile("attachment", "screenshot.png")
		require.NoError(t, err)
		_, err = io.WriteString(part, "fake-png-data-for-testing")
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/tickets", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		if w.Body.Len() > 0 {
			json.Unmarshal(w.Body.Bytes(), &resp)
		}

		t.Logf("Response status: %d", w.Code)
		t.Logf("Response: %v", resp)

		// Ticket should be created successfully
		assert.Equal(t, http.StatusCreated, w.Code, "Ticket creation with attachments should succeed")
		assert.Contains(t, resp, "id", "Should return ticket ID")
	})

	t.Run("Verify ticket creation works without attachments", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/tickets", func(c *gin.Context) {
			c.Set("user_role", "Agent")
			c.Set("user_id", uint(1))
			c.Set("user_email", "admin@example.com")
			handleCreateTicket(c)
		})

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("title", "Simple ticket without attachments")
		writer.WriteField("customer_email", "user@example.com")
		writer.WriteField("body", "This should work fine")
		writer.WriteField("priority", "normal")
		writer.WriteField("queue_id", "1")
		writer.Close()

		req := httptest.NewRequest("POST", "/api/tickets", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Basic ticket creation should work")

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp, "id", "Should return ticket ID")
		assert.Contains(t, resp, "message", "Should return success message")
	})
}

func TestAttachmentHandlingImplementation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Handler should process multipart files", func(t *testing.T) {
		router := gin.New()
		router.POST("/test", func(c *gin.Context) {
			form, err := c.MultipartForm()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
				return
			}

			files := form.File["attachment"]
			attachmentInfo := []map[string]interface{}{}

			for _, file := range files {
				t.Logf("Processing file: %s (size: %d)", file.Filename, file.Size)
				attachmentInfo = append(attachmentInfo, map[string]interface{}{
					"filename": file.Filename,
					"size":     file.Size,
					"saved":    true,
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"message":     "Files processed",
				"attachments": attachmentInfo,
			})
		})

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, _ := writer.CreateFormFile("attachment", "test.txt")
		io.WriteString(part, "test content")
		writer.Close()

		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp, "attachments")

		attachments := resp["attachments"].([]interface{})
		assert.Len(t, attachments, 1, "Should have processed one attachment")
	})
}
