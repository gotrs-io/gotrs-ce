package api

import "github.com/gin-gonic/gin"

var (
	HandleAdminSettings  gin.HandlerFunc = handleAdminSettings
	HandleAdminTemplates gin.HandlerFunc = handleAdminTemplates
	HandleAdminReports   gin.HandlerFunc = handleAdminReports
	HandleAdminLogs      gin.HandlerFunc = handleAdminLogs
	HandleAdminBackup    gin.HandlerFunc = handleAdminBackup
)
