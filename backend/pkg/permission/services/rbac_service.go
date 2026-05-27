package services

import (
	"context"
	"fmt"
	"time"

	"github.com/bxf1/ERP/backend/pkg/permission/cache"
	"github.com/bxf1/ERP/backend/pkg/permission/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RBACService struct {
	db    *gorm.DB
	cache *cache.PermissionCache
}

func NewRBACService(db *gorm.DB, c *cache.PermissionCache) *RBACService {
	return &RBACService{db: db, cache: c}
}

// CheckPermission verifies if a user has a specific permission code.
// Permission codes follow the pattern: resource:action (e.g., "user:create").
// Wildcard "*" grants all permissions.
func (s *RBACService) CheckPermission(ctx context.Context, userID uuid.UUID, permCode string) (bool, error) {
	perms, err := s.getUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, p := range perms.Permissions {
		if p == "*" || p == permCode {
			return true, nil
		}
	}
	return false, nil
}

// CheckPermissions verifies if a user has ALL of the specified permission codes.
func (s *RBACService) CheckPermissions(ctx context.Context, userID uuid.UUID, permCodes []string) (bool, error) {
	perms, err := s.getUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	permSet := make(map[string]bool, len(perms.Permissions))
	hasWildcard := false
	for _, p := range perms.Permissions {
		permSet[p] = true
		if p == "*" {
			hasWildcard = true
		}
	}
	if hasWildcard {
		return true, nil
	}
	for _, code := range permCodes {
		if !permSet[code] {
			return false, nil
		}
	}
	return true, nil
}

// CheckAnyPermission verifies if a user has ANY of the specified permission codes.
func (s *RBACService) CheckAnyPermission(ctx context.Context, userID uuid.UUID, permCodes []string) (bool, error) {
	perms, err := s.getUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	permSet := make(map[string]bool, len(perms.Permissions))
	for _, p := range perms.Permissions {
		permSet[p] = true
		if p == "*" {
			return true, nil
		}
	}
	for _, code := range permCodes {
		if permSet[code] {
			return true, nil
		}
	}
	return false, nil
}

// GetUserPermissions returns all permission codes for a user.
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	p, err := s.getUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	return p.Permissions, nil
}

// GetUserRoles returns all role codes for a user.
func (s *RBACService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	p, err := s.getUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	return p.Roles, nil
}

// getUserPermissions loads from cache or DB.
func (s *RBACService) getUserPermissions(ctx context.Context, userID uuid.UUID) (*cache.CachedUserPermissions, error) {
	uid := userID.String()

	if s.cache != nil {
		cached, err := s.cache.GetUserPermissions(ctx, uid)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	perms, err := s.loadUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetUserPermissions(ctx, uid, perms)
	}

	return perms, nil
}

// loadUserPermissions queries DB for user's effective permissions.
func (s *RBACService) loadUserPermissions(ctx context.Context, userID uuid.UUID) (*cache.CachedUserPermissions, error) {
	now := time.Now()

	var userRoles []models.UserRole
	err := s.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("user_id = ?", userID).
		Where("effective_from <= ?", now).
		Where("effective_to IS NULL OR effective_to > ?", now).
		Find(&userRoles).Error
	if err != nil {
		return nil, fmt.Errorf("load user roles: %w", err)
	}

	permSet := make(map[string]bool)
	roleSet := make(map[string]bool)

	for _, ur := range userRoles {
		if ur.Role == nil || ur.Role.Status != "active" {
			continue
		}
		roleSet[ur.Role.Code] = true
		for _, p := range ur.Role.Permissions {
			permSet[p.Code] = true
		}
	}

	perms := make([]string, 0, len(permSet))
	for p := range permSet {
		perms = append(perms, p)
	}
	roles := make([]string, 0, len(roleSet))
	for r := range roleSet {
		roles = append(roles, r)
	}

	return &cache.CachedUserPermissions{
		Permissions: perms,
		Roles:       roles,
	}, nil
}

// InvalidateUserCache clears permission cache for a user.
func (s *RBACService) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.InvalidateUser(ctx, userID.String())
}

// InvalidateRoleCache clears permission cache for all users with a given role.
func (s *RBACService) InvalidateRoleCache(ctx context.Context, roleCode string) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.InvalidateRole(ctx, roleCode)
}

// AssignRoleToUser assigns a role to a user with optional time range.
func (s *RBACService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, effectiveFrom time.Time, effectiveTo *time.Time) error {
	ur := models.UserRole{
		UserID:        userID,
		RoleID:        roleID,
		EffectiveFrom: effectiveFrom,
		EffectiveTo:   effectiveTo,
	}
	if err := s.db.WithContext(ctx).Create(&ur).Error; err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
	_ = s.InvalidateUserCache(ctx, userID)
	return nil
}

// RemoveRoleFromUser removes a role assignment from a user.
func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{}).Error; err != nil {
		return fmt.Errorf("remove role: %w", err)
	}
	_ = s.InvalidateUserCache(ctx, userID)
	return nil
}
