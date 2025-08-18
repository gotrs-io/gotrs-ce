package models

import "time"

// Group represents a user group in the system
type Group struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"` // ldap, local, external
	DN          string            `json:"dn,omitempty"` // LDAP Distinguished Name
	Members     []string          `json:"members,omitempty"` // User IDs
	Permissions []string          `json:"permissions,omitempty"`
	IsActive    bool              `json:"is_active"`
	IsSystem    bool              `json:"is_system"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// GroupMembership represents a user's membership in a group
type GroupMembership struct {
	GroupID   string    `json:"group_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role,omitempty"` // member, admin, etc.
	AddedBy   string    `json:"added_by,omitempty"`
	AddedAt   time.Time `json:"added_at"`
}

// Common group types
const (
	GroupTypeLocal    = "local"
	GroupTypeLDAP     = "ldap"
	GroupTypeExternal = "external"
)

// Common group roles
const (
	GroupRoleMember = "member"
	GroupRoleAdmin  = "admin"
	GroupRoleOwner  = "owner"
)