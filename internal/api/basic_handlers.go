package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// HandleRedirect handles redirect routes
func HandleRedirect(c *gin.Context) {
	routeConfig, exists := c.Get("route_config")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	
	config := routeConfig.(map[string]interface{})
	redirectTo := "/login" // default
	if to, ok := config["redirect_to"].(string); ok {
		redirectTo = to
	}
	
	redirectCode := http.StatusFound // Default 302
	if code, ok := config["redirect_code"].(float64); ok {
		redirectCode = int(code)
	}
	
	c.Redirect(redirectCode, redirectTo)
}

// HandleTemplate handles template rendering routes
func HandleTemplate(c *gin.Context) {
	routeConfig, exists := c.Get("route_config")
	if !exists {
		c.String(http.StatusInternalServerError, "No route config")
		return
	}
	
	config := routeConfig.(map[string]interface{})
	template := "pages/login.pongo2" // default
	if tmpl, ok := config["template"].(string); ok {
		template = tmpl
	}
	
	data := gin.H{}
	if configData, ok := config["data"].(map[string]interface{}); ok {
		for k, v := range configData {
			data[k] = v
		}
	}
	
	// Add user context if authenticated
	if user, exists := c.Get("user"); exists {
		data["User"] = user
	}
	
	// Use the pongo2 renderer
	GetPongo2Renderer().HTML(c, http.StatusOK, template, data)
}