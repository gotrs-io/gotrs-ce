package shared

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/i18n"
	"github.com/gotrs-io/gotrs-ce/internal/middleware"
)

// TemplateRenderer handles template rendering with pongo2
type TemplateRenderer struct {
	templateSet *pongo2.TemplateSet
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(templateDir string) (*TemplateRenderer, error) {
	if templateDir == "" {
		return nil, fmt.Errorf("template directory is required")
	}

	// Check if directory exists
	if _, err := os.Stat(templateDir); err != nil {
		return nil, fmt.Errorf("template directory not found: %v", err)
	}

	// Normalize path
	abs, _ := filepath.Abs(templateDir)

	// Create template set
	templateSet := pongo2.NewSet("gotrs", pongo2.MustNewLocalFileSystemLoader(abs))

	return &TemplateRenderer{
		templateSet: templateSet,
	}, nil
}

// HTML renders a template
func (r *TemplateRenderer) HTML(c *gin.Context, code int, name string, data interface{}) {
	// Convert gin.H to pongo2.Context
	var ctx pongo2.Context
	switch v := data.(type) {
	case pongo2.Context:
		ctx = v
	case gin.H:
		ctx = pongo2.Context(v)
	default:
		ctx = pongo2.Context{"data": data}
	}

	// Language helpers injected via middleware detection
	lang := middleware.GetLanguage(c)
	i18nInst := i18n.GetInstance()
	ctx["t"] = func(key string, args ...interface{}) string {
		return translateWithFallback(i18nInst, lang, key, args...)
	}
	ctx["getLang"] = func() string { return lang }
	ctx["getDirection"] = func() string { return string(i18n.GetDirection(lang)) }
	ctx["isRTL"] = func() bool { return i18n.IsRTL(lang) }

	// Get the template (fallback for tests when templates missing)
	if r == nil || r.templateSet == nil {
		// Minimal safe fallback for tests: render a tiny stub
		c.String(code, "GOTRS")
		return
	}
	tmpl, err := r.templateSet.FromFile(name)
	if err != nil {
		// Fallback for missing templates
		c.String(code, "Template not found: %s", name)
		return
	}

	// Render template
	err = tmpl.ExecuteWriter(ctx, c.Writer)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template execution error: %v", err)
	}
}

// translateWithFallback provides a fallback translation function
func translateWithFallback(i18nInst *i18n.I18n, lang, key string, args ...interface{}) string {
	if i18nInst == nil {
		return key
	}
	return i18nInst.Translate(lang, key, args...)
}

// GetGlobalRenderer returns the global template renderer instance
func GetGlobalRenderer() *TemplateRenderer {
	return globalTemplateRenderer
}

// SetGlobalRenderer sets the global template renderer instance
func SetGlobalRenderer(renderer *TemplateRenderer) {
	globalTemplateRenderer = renderer
}

var globalTemplateRenderer *TemplateRenderer