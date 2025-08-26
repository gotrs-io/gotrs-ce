package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// Customer handler exports that get database from connection pool
var (
	HandleCustomerDashboard = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerDashboard(db)(c)
	}
	
	HandleCustomerTickets = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerTickets(db)(c)
	}
	
	HandleCustomerNewTicket = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerNewTicket(db)(c)
	}
	
	HandleCustomerCreateTicket = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerCreateTicket(db)(c)
	}
	
	HandleCustomerTicketView = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerTicketView(db)(c)
	}
	
	HandleCustomerTicketReply = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerTicketReply(db)(c)
	}
	
	HandleCustomerCloseTicket = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerCloseTicket(db)(c)
	}
	
	HandleCustomerProfile = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerProfile(db)(c)
	}
	
	HandleCustomerUpdateProfile = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerUpdateProfile(db)(c)
	}
	
	HandleCustomerPasswordForm = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerPasswordForm(db)(c)
	}
	
	HandleCustomerChangePassword = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerChangePassword(db)(c)
	}
	
	HandleCustomerKnowledgeBase = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerKnowledgeBase(db)(c)
	}
	
	HandleCustomerKBSearch = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerKBSearch(db)(c)
	}
	
	HandleCustomerKBArticle = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerKBArticle(db)(c)
	}
	
	HandleCustomerCompanyInfo = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerCompanyInfo(db)(c)
	}
	
	HandleCustomerCompanyUsers = func(c *gin.Context) {
		db, _ := database.GetDB()
		handleCustomerCompanyUsers(db)(c)
	}
)