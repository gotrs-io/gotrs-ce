package api

import (
	"fmt"
	"html/template"
	"net/http"
	
	"github.com/gin-gonic/gin"
)

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// sendGuruMeditation sends a detailed error response (similar to VirtualBox's Guru Meditation)
func sendGuruMeditation(c *gin.Context, err error, message string) {
	// Log the full error for debugging
	if err != nil {
		fmt.Printf("Guru Meditation: %s - Error: %v\n", message, err)
	}
	
	// Send a user-friendly error response
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": message,
		"details": err.Error(),
		"status": "error",
	})
}

// getPriorityID converts a priority string to its database ID
func getPriorityID(priority string) int {
	switch priority {
	case "low":
		return 1
	case "normal", "medium":
		return 2
	case "high":
		return 3
	case "critical", "very-high":
		return 4
	default:
		return 2 // Default to normal priority
	}
}

// loadTemplate loads and parses HTML template files
func loadTemplate(files ...string) (*template.Template, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no template files provided")
	}
	
	return template.ParseFiles(files...)
}