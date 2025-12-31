package auth_models

import (
	"common/types"
	"time"
)

// Role
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Create Role
type RegisterRolePayload struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=200"`
}

type RegisterRoleResponse struct {
	Message string `json:"message"`
	Role    Role   `json:"data"`
}

// Update Role
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
}

type UpdateRoleResponse struct {
	Message string `json:"message"`
}

// Get Roles
type GetRolesResponse struct {
	Data []Role `json:"data"`
}

type GetRoleResponse struct {
	Data Role `json:"data"`
}

// Delete Role
type DeleteRoleResponse struct {
	Message string `json:"message"`
}

// Assign Role
type AssignRoleRequest struct {
	RoleID string `json:"roleId" binding:"required"`
}

type UserRole struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	RoleID    string    `json:"roleId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserRoleRegisterPayload struct {
	UserID string `json:"userId" binding:"required"`
	RoleID string `json:"roleId" binding:"required"`
}

type UserRoleRegisterResponse struct {
	Message  string   `json:"message"`
	UserRole UserRole `json:"data"`
}

type RolePermission struct {
	ID           string    `json:"id"`
	RoleID       string    `json:"roleId"`
	PermissionID string    `json:"permissionId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type RolePermissionRegisterPayload struct {
	RoleID       string `json:"roleId" binding:"required"`
	PermissionID string `json:"permissionId" binding:"required"`
}

type RolePermissionRegisterResponse struct {
	Message        string         `json:"message"`
	RolePermission RolePermission `json:"data"`
}

// ----- Permissions -----

type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type RegisterPermissionPayload struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=200"`
}

type RegisterPermissionResponse struct {
	Message    string     `json:"message"`
	Permission Permission `json:"data"`
}

// Check Permission
type CheckPermissionRequest struct {
	UserID    string `json:"userId" binding:"required"`
	CompanyID string `json:"companyId" binding:"required"`
	Resource  string `json:"resource" binding:"required"`
	Action    string `json:"action" binding:"required"`
}

type CheckPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message,omitempty"`
}

// Get All Available Permissions (for UI)
type GetAvailablePermissionsResponse struct {
	Data []AvailablePermission `json:"data"`
}

type AvailablePermission struct {
	Resource    string   `json:"resource"`
	Description string   `json:"description"`
	Actions     []Action `json:"actions"`
}

type Action struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Permission Template (for quick role setup)
type PermissionTemplate struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Category    string             `json:"category"`
	Permissions []types.Permission `json:"permissions"`
}

type GetPermissionTemplatesResponse struct {
	Data []PermissionTemplate `json:"data"`
}

// Apply Permission Template
type ApplyPermissionTemplateRequest struct {
	TemplateID string `json:"templateId" binding:"required"`
}

type ApplyPermissionTemplateResponse struct {
	Message     string             `json:"message"`
	Permissions []types.Permission `json:"permissions"`
}

// Bulk Permission Assignment
type AssignPermissionsToRoleRequest struct {
	PermissionIDs []string `json:"permissionIds" binding:"required"`
}

type AssignPermissionsToRoleResponse struct {
	Message string `json:"message"`
}
