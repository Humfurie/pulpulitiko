package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type RoleHandler struct {
	roleService *services.RoleService
}

func NewRoleHandler(roleService *services.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// List returns all roles
func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	roles, err := h.roleService.ListRoles(ctx, includeDeleted)
	if err != nil {
		WriteInternalError(w, "Failed to list roles")
		return
	}

	WriteSuccess(w, roles)
}

// GetByID returns a role by ID
func (h *RoleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid role ID")
		return
	}

	role, err := h.roleService.GetRoleByID(ctx, id)
	if err != nil {
		WriteNotFound(w, "Role not found")
		return
	}

	WriteSuccess(w, role)
}

// Create creates a new role
func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	if req.Name == "" || req.Slug == "" {
		WriteBadRequest(w, "Name and slug are required")
		return
	}

	role, err := h.roleService.CreateRole(ctx, &req)
	if err != nil {
		WriteInternalError(w, "Failed to create role: "+err.Error())
		return
	}

	WriteCreated(w, role)
}

// Update updates an existing role
func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid role ID")
		return
	}

	var req models.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	role, err := h.roleService.UpdateRole(ctx, id, &req)
	if err != nil {
		WriteInternalError(w, "Failed to update role: "+err.Error())
		return
	}

	WriteSuccess(w, role)
}

// Delete soft deletes a role
func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid role ID")
		return
	}

	if err := h.roleService.DeleteRole(ctx, id); err != nil {
		WriteInternalError(w, "Failed to delete role: "+err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "Role deleted successfully"})
}

// Restore restores a soft-deleted role
func (h *RoleHandler) Restore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid role ID")
		return
	}

	if err := h.roleService.RestoreRole(ctx, id); err != nil {
		WriteInternalError(w, "Failed to restore role: "+err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "Role restored successfully"})
}

// ListPermissions returns all permissions grouped by category
func (h *RoleHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	permissions, err := h.roleService.ListPermissionsGrouped(ctx)
	if err != nil {
		WriteInternalError(w, "Failed to list permissions")
		return
	}

	WriteSuccess(w, permissions)
}
