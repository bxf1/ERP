package handlers

import (
	"net/http"
	"time"

	"github.com/bxf1/ERP/backend/pkg/permission/models"
	"github.com/bxf1/ERP/backend/pkg/permission/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRoleHandler struct {
	db    *gorm.DB
	rbac  *services.RBACService
}

func NewUserRoleHandler(db *gorm.DB, rbac *services.RBACService) *UserRoleHandler {
	return &UserRoleHandler{db: db, rbac: rbac}
}

// GetUserRoles returns all roles assigned to a user.
func (h *UserRoleHandler) GetUserRoles(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var userRoles []models.UserRole
	if err := h.db.Preload("Role").Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userRoles})
}

// GetUserPermissions returns all effective permissions for a user.
func (h *UserRoleHandler) GetUserPermissions(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	perms, err := h.rbac.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	roles, _ := h.rbac.GetUserRoles(c.Request.Context(), userID)

	if perms == nil {
		perms = []string{}
	}
	if roles == nil {
		roles = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user_id":     userID,
			"permissions": perms,
			"roles":       roles,
		},
	})
}

// AssignRole assigns a role to a user.
func (h *UserRoleHandler) AssignRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var input struct {
		RoleID        uuid.UUID `json:"role_id" binding:"required"`
		EffectiveFrom string    `json:"effective_from"`
		EffectiveTo   string    `json:"effective_to"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	effectiveFrom := time.Now()
	if input.EffectiveFrom != "" {
		effectiveFrom, err = time.Parse(time.RFC3339, input.EffectiveFrom)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid effective_from format"})
			return
		}
	}

	var effectiveTo *time.Time
	if input.EffectiveTo != "" {
		t, err := time.Parse(time.RFC3339, input.EffectiveTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid effective_to format"})
			return
		}
		effectiveTo = &t
	}

	if err := h.rbac.AssignRoleToUser(c.Request.Context(), userID, input.RoleID, effectiveFrom, effectiveTo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role assigned"})
}

// RemoveRole removes a role from a user.
func (h *UserRoleHandler) RemoveRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	if err := h.rbac.RemoveRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role removed"})
}

// ListAllRoles returns all available roles (for admin UI).
func (h *UserRoleHandler) ListAllRoles(c *gin.Context) {
	var roles []models.Role
	if err := h.db.Where("status = 'active'").Order("sort_order ASC").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}
