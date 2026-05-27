package handlers

import (
	"net/http"

	"github.com/bxf1/ERP/backend/pkg/permission/models"
	"github.com/bxf1/ERP/backend/pkg/permission/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleHandler struct {
	db    *gorm.DB
	rbac  *services.RBACService
}

func NewRoleHandler(db *gorm.DB, rbac *services.RBACService) *RoleHandler {
	return &RoleHandler{db: db, rbac: rbac}
}

// ListRoles returns all roles.
func (h *RoleHandler) ListRoles(c *gin.Context) {
	var roles []models.Role
	if err := h.db.Preload("Permissions").Order("sort_order ASC").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

// GetRole returns a single role by ID.
func (h *RoleHandler) GetRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	var role models.Role
	if err := h.db.Preload("Permissions").Preload("DataScopes").First(&role, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": role})
}

// CreateRole creates a new role.
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := models.Role{
		Name:        input.Name,
		Code:        input.Code,
		Description: input.Description,
		SortOrder:   input.SortOrder,
		Status:      "active",
	}
	if err := h.db.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": role})
}

// UpdateRole updates a role.
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	var role models.Role
	if err := h.db.First(&role, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent modifying code and is_system for system roles
	delete(input, "code")
	delete(input, "id")
	if role.IsSystem {
		delete(input, "is_system")
	}

	if err := h.db.Model(&role).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.rbac.InvalidateRoleCache(c.Request.Context(), role.Code)
	c.JSON(http.StatusOK, gin.H{"data": role})
}

// DeleteRole deletes a role (soft delete for non-system roles).
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	var role models.Role
	if err := h.db.First(&role, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete system role"})
		return
	}

	if err := h.db.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.rbac.InvalidateRoleCache(c.Request.Context(), role.Code)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// AssignPermissions assigns permissions to a role.
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	var input struct {
		PermissionIDs []uuid.UUID `json:"permission_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if err := h.db.First(&role, "id = ?", roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// Replace all permissions for this role
	if err := h.db.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, permID := range input.PermissionIDs {
		rp := models.RolePermission{RoleID: roleID, PermissionID: permID}
		if err := h.db.Create(&rp).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	_ = h.rbac.InvalidateRoleCache(c.Request.Context(), role.Code)
	c.JSON(http.StatusOK, gin.H{"message": "permissions assigned"})
}

// SetDataScope sets the data scope for a role on a target model.
func (h *RoleHandler) SetDataScope(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	var input struct {
		TargetModel string `json:"target_model" binding:"required"`
		ScopeType   string `json:"scope_type" binding:"required"`
		ScopeRule   string `json:"scope_rule"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scope := models.DataScope{
		RoleID:      roleID,
		TargetModel: input.TargetModel,
		ScopeType:   input.ScopeType,
		ScopeRule:   input.ScopeRule,
	}

	if err := h.db.Where("role_id = ? AND target_model = ?", roleID, input.TargetModel).
		Assign(scope).FirstOrCreate(&scope).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if err := h.db.First(&role, "id = ?", roleID).Error; err == nil {
		_ = h.rbac.InvalidateRoleCache(c.Request.Context(), role.Code)
	}

	c.JSON(http.StatusOK, gin.H{"data": scope})
}
