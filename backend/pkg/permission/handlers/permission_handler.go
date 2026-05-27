package handlers

import (
	"net/http"

	"github.com/bxf1/ERP/backend/pkg/permission/models"
	"github.com/bxf1/ERP/backend/pkg/permission/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PermissionHandler struct {
	db   *gorm.DB
	rbac *services.RBACService
}

func NewPermissionHandler(db *gorm.DB, rbac *services.RBACService) *PermissionHandler {
	return &PermissionHandler{db: db, rbac: rbac}
}

// ListPermissions returns all permissions as a tree.
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	var permissions []models.Permission
	if err := h.db.Order("sort_order ASC").Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tree := h.buildTree(permissions)
	c.JSON(http.StatusOK, gin.H{"data": tree})
}

// ListPermissionsFlat returns all permissions as a flat list.
func (h *PermissionHandler) ListPermissionsFlat(c *gin.Context) {
	var permissions []models.Permission
	if err := h.db.Order("resource ASC, sort_order ASC").Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": permissions})
}

// CreatePermission creates a new permission.
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var input struct {
		Name        string     `json:"name" binding:"required"`
		Code        string     `json:"code" binding:"required"`
		Resource    string     `json:"resource" binding:"required"`
		Action      string     `json:"action" binding:"required"`
		Description string     `json:"description"`
		ParentID    *uuid.UUID `json:"parent_id"`
		SortOrder   int        `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	perm := models.Permission{
		Name:        input.Name,
		Code:        input.Code,
		Resource:    input.Resource,
		Action:      input.Action,
		Description: input.Description,
		ParentID:    input.ParentID,
		SortOrder:   input.SortOrder,
	}
	if err := h.db.Create(&perm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": perm})
}

// UpdatePermission updates a permission.
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission id"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	delete(input, "id")
	delete(input, "code")

	if err := h.db.Model(&models.Permission{}).Where("id = ?", id).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.rbac.InvalidateRoleCache(c.Request.Context(), "*")
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeletePermission deletes a permission.
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission id"})
		return
	}

	if err := h.db.Where("id = ?", id).Delete(&models.Permission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.rbac.InvalidateRoleCache(c.Request.Context(), "*")
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// buildTree converts flat permission list to tree structure.
func (h *PermissionHandler) buildTree(permissions []models.Permission) []*models.Permission {
	byID := make(map[uuid.UUID]*models.Permission)
	var roots []*models.Permission

	for i := range permissions {
		p := &permissions[i]
		p.Children = []*models.Permission{}
		byID[p.ID] = p
	}

	for _, p := range byID {
		if p.ParentID != nil {
			if parent, ok := byID[*p.ParentID]; ok {
				parent.Children = append(parent.Children, p)
				continue
			}
		}
		roots = append(roots, p)
	}

	return roots
}
