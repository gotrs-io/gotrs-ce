package shared

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// sendToastResponse sends either an HTMX toast notification or redirects with success message
func SendToastResponse(c *gin.Context, success bool, message, redirectPath string) {
	if c.GetHeader("HX-Request") == "true" {
		// Return HTML partial with toast notification
		if success {
			html := fmt.Sprintf(`
				<div id="toast" class="fixed top-4 right-4 bg-green-500 text-white px-6 py-3 rounded-lg shadow-lg z-50">
					<div class="flex items-center">
						<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
						</svg>
						%s
					</div>
				</div>
				<script>
					setTimeout(function() {
						var toast = document.getElementById('toast');
						if (toast) toast.remove();
					}, 3000);
				</script>
			`, message)
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		} else {
			html := fmt.Sprintf(`
				<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded" role="alert">
					%s
				</div>
			`, message)
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(html))
		}
	} else {
		// For regular form submissions, redirect back with success message
		if success && redirectPath != "" {
			c.Redirect(http.StatusFound, redirectPath+"?success=1")
		} else {
			c.JSON(http.StatusOK, gin.H{"success": success, "message": message})
		}
	}
}