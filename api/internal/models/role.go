package models

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a single permission in the system
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

// Role represents a user role with associated permissions
type Role struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Slug        string       `json:"slug"`
	Description *string      `json:"description,omitempty"`
	IsSystem    bool         `json:"is_system"`
	Permissions []Permission `json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty"`
}

// RoleWithPermissionCount is used for listing roles with permission counts
type RoleWithPermissionCount struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Slug            string     `json:"slug"`
	Description     *string    `json:"description,omitempty"`
	IsSystem        bool       `json:"is_system"`
	PermissionCount int        `json:"permission_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// CreateRoleRequest is used to create a new role
type CreateRoleRequest struct {
	Name          string   `json:"name" validate:"required,min=2,max=50"`
	Slug          string   `json:"slug" validate:"required,min=2,max=50"`
	Description   *string  `json:"description,omitempty"`
	PermissionIDs []string `json:"permission_ids,omitempty"`
}

// UpdateRoleRequest is used to update an existing role
type UpdateRoleRequest struct {
	Name          *string  `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Slug          *string  `json:"slug,omitempty" validate:"omitempty,min=2,max=50"`
	Description   *string  `json:"description,omitempty"`
	PermissionIDs []string `json:"permission_ids,omitempty"`
}

// PermissionCategory groups permissions by category for display
type PermissionCategory struct {
	Category    string       `json:"category"`
	Permissions []Permission `json:"permissions"`
}

// Common permission slugs for middleware checks
const (
	PermViewArticles    = "view_articles"
	PermCreateArticles  = "create_articles"
	PermEditArticles    = "edit_articles"
	PermDeleteArticles  = "delete_articles"
	PermPublishArticles = "publish_articles"

	PermViewCategories   = "view_categories"
	PermCreateCategories = "create_categories"
	PermEditCategories   = "edit_categories"
	PermDeleteCategories = "delete_categories"

	PermViewTags   = "view_tags"
	PermCreateTags = "create_tags"
	PermEditTags   = "edit_tags"
	PermDeleteTags = "delete_tags"

	PermViewUsers   = "view_users"
	PermCreateUsers = "create_users"
	PermEditUsers   = "edit_users"
	PermDeleteUsers = "delete_users"

	PermViewRoles   = "view_roles"
	PermCreateRoles = "create_roles"
	PermEditRoles   = "edit_roles"
	PermDeleteRoles = "delete_roles"

	PermViewMetrics = "view_metrics"
	PermUploadFiles = "upload_files"
)
