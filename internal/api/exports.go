package api

// Exported handlers for use by the routing system

// Core handlers
var (
	HandleLoginPage      = handleLoginPage
	HandleLogout         = handleLogout
	HandleDashboard      = handleDashboard
)

// Auth API handlers are directly exported from auth_handlers.go

// Admin handlers
var (
	HandleAdminDashboard = handleAdminDashboard
	HandleAdminUsers     = handleAdminUsers
	// HandleAdminUserGet is already exported in admin_users_handlers.go
	HandleAdminGroups    = handleAdminGroups
	HandleGetGroup       = handleGetGroup
	HandleCreateGroup    = handleCreateGroup
	HandleUpdateGroup    = handleUpdateGroup
	HandleDeleteGroup    = handleDeleteGroup
	HandleGroupMembers   = handleGetGroupMembers
	HandleAddUserToGroup = handleAddUserToGroup
	HandleRemoveUserFromGroup = handleRemoveUserFromGroup
	HandleAdminQueues    = handleAdminQueues
	HandleAdminPriorities = handleAdminPriorities
	HandleAdminPermissions = handleAdminPermissions  // Renamed from roles
	HandleAdminStates    = handleAdminStates
	HandleAdminTypes     = handleAdminTypes
	HandleAdminServices  = handleAdminServices
	HandleAdminSLA       = handleAdminSLA
	HandleAdminLookups   = handleAdminLookups
)

// GetPongo2Renderer returns the pongo2 renderer for template rendering
func GetPongo2Renderer() *Pongo2Renderer {
	return pongo2Renderer
}