package api

import (
	"github.com/gin-gonic/gin"
)

// Simplified router for HTMX demo
func NewSimpleRouter() *gin.Engine {
	r := gin.Default()
	SetupHTMXRoutes(r)
	return r
}