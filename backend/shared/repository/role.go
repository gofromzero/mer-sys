package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// UserRole 用户角色数据库实体
type UserRole struct {
	ID           uint64              `json:"id" db:"id"`
	UserID       uint64              `json:"user_id" db:"user_id"`
	TenantID     uint64              `json:"tenant_id" db:"tenant_id"`
	RoleType     types.RoleType      `json:"role_type" db:"role_type"`
	ResourceID   *uint64             `json:"resource_id" db:"resource_id"`
	ResourceType *string             `json:"resource_type" db:"resource_type"`
	GrantedBy    uint64              `json:"granted_by" db:"granted_by"`
	ExpiresAt    *gtime.Time         `json:"expires_at" db:"expires_at"`
	Status       string              `json:"status" db:"status"`
	CreatedAt    *gtime.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    *gtime.Time         `json:"updated_at" db:"updated_at"`
}

// UserPermissionsCache 用户权限缓存数据库实体
type UserPermissionsCache struct {
	ID              uint64      `json:"id" db:"id"`
	UserID          uint64      `json:"user_id" db:"user_id"`
	TenantID        uint64      `json:"tenant_id" db:"tenant_id"`
	PermissionsJSON string      `json:"permissions_json" db:"permissions_json"`
	RolesJSON       string      `json:"roles_json" db:"roles_json"`
	Version         int         `json:"version" db:"version"`
	ExpiresAt       *gtime.Time `json:"expires_at" db:"expires_at"`
	CreatedAt       *gtime.Time `json:"created_at" db:"created_at"`
	UpdatedAt       *gtime.Time `json:"updated_at" db:"updated_at"`
}

// RoleRepository 角色管理仓储接口
type RoleRepository interface {
	// 角色分配
	AssignRole(ctx context.Context, userID, tenantID uint64, roleType types.RoleType, grantedBy uint64) error
	RevokeRole(ctx context.Context, userID, tenantID uint64, roleType types.RoleType) error
	
	// 角色查询
	GetUserRoles(ctx context.Context, userID, tenantID uint64) ([]types.RoleType, error)
	GetUserPermissions(ctx context.Context, userID, tenantID uint64) (*types.UserPermissions, error)
	
	// 权限验证
	HasPermission(ctx context.Context, userID, tenantID uint64, permission types.Permission) (bool, error)
	HasRole(ctx context.Context, userID, tenantID uint64, role types.RoleType) (bool, error)
	
	// 缓存管理
	RefreshPermissionsCache(ctx context.Context, userID, tenantID uint64) error
	ClearPermissionsCache(ctx context.Context, userID, tenantID uint64) error
	
	// 批量操作
	GetRoleUsers(ctx context.Context, tenantID uint64, roleType types.RoleType) ([]uint64, error)
	GetTenantRoleStats(ctx context.Context, tenantID uint64) (map[types.RoleType]int, error)
}

// roleRepository 角色管理仓储实现
type roleRepository struct {
	*BaseRepository
}

// NewRoleRepository 创建角色管理仓储实例
func NewRoleRepository() RoleRepository {
	return &roleRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// WithTenant 创建带租户上下文的context
func (r *roleRepository) WithTenant(ctx context.Context, tenantID uint64) context.Context {
	return context.WithValue(ctx, "tenant_id", tenantID)
}

// AssignRole 为用户分配角色
func (r *roleRepository) AssignRole(ctx context.Context, userID, tenantID uint64, roleType types.RoleType, grantedBy uint64) error {
	tenantCtx := r.WithTenant(ctx, tenantID)
	
	// 检查是否已经存在该角色
	count, err := g.DB().Model("user_roles").Ctx(tenantCtx).
		Where("user_id = ? AND tenant_id = ? AND role_type = ? AND status = 'active'", userID, tenantID, roleType).
		Count()
	if err != nil {
		return fmt.Errorf("检查角色分配失败: %w", err)
	}
	
	if count > 0 {
		return fmt.Errorf("用户已拥有角色 %s", roleType)
	}
	
	// 插入新角色
	_, err = g.DB().Model("user_roles").Ctx(tenantCtx).Insert(g.Map{
		"user_id":    userID,
		"tenant_id":  tenantID,
		"role_type":  roleType,
		"granted_by": grantedBy,
		"status":     "active",
		"created_at": gtime.Now(),
		"updated_at": gtime.Now(),
	})
	
	if err != nil {
		return fmt.Errorf("角色分配失败: %w", err)
	}
	
	// 清除权限缓存
	_ = r.ClearPermissionsCache(ctx, userID, tenantID)
	
	return nil
}

// RevokeRole 撤销用户角色
func (r *roleRepository) RevokeRole(ctx context.Context, userID, tenantID uint64, roleType types.RoleType) error {
	tenantCtx := r.WithTenant(ctx, tenantID)
	
	_, err := g.DB().Model("user_roles").Ctx(tenantCtx).
		Where("user_id = ? AND tenant_id = ? AND role_type = ?", userID, tenantID, roleType).
		Update(g.Map{
			"status":     "suspended",
			"updated_at": gtime.Now(),
		})
	
	if err != nil {
		return fmt.Errorf("角色撤销失败: %w", err)
	}
	
	// 清除权限缓存
	_ = r.ClearPermissionsCache(ctx, userID, tenantID)
	
	return nil
}

// GetUserRoles 获取用户角色列表
func (r *roleRepository) GetUserRoles(ctx context.Context, userID, tenantID uint64) ([]types.RoleType, error) {
	tenantCtx := r.WithTenant(ctx, tenantID)
	
	var roles []string
	err := g.DB().Model("user_roles").Ctx(tenantCtx).
		Fields("role_type").
		Where("user_id = ? AND tenant_id = ? AND status = 'active'", userID, tenantID).
		Where("expires_at IS NULL OR expires_at > ?", gtime.Now()).
		Scan(&roles)
	
	if err != nil {
		return nil, fmt.Errorf("获取用户角色失败: %w", err)
	}
	
	result := make([]types.RoleType, len(roles))
	for i, role := range roles {
		result[i] = types.RoleType(role)
	}
	
	return result, nil
}

// GetUserPermissions 获取用户完整权限信息
func (r *roleRepository) GetUserPermissions(ctx context.Context, userID, tenantID uint64) (*types.UserPermissions, error) {
	// 首先尝试从缓存获取
	cached, err := r.getPermissionsFromCache(ctx, userID, tenantID)
	if err == nil && cached != nil {
		return cached, nil
	}
	
	// 缓存未命中，从数据库计算权限
	roles, err := r.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("获取用户角色失败: %w", err)
	}
	
	// 计算所有权限
	allPermissions := make([]types.Permission, 0)
	defaultRoles := types.GetDefaultRoles()
	
	for _, roleType := range roles {
		if role, exists := defaultRoles[roleType]; exists {
			allPermissions = append(allPermissions, role.Permissions...)
		}
	}
	
	// 去重权限
	uniquePermissions := r.deduplicatePermissions(allPermissions)
	
	userPermissions := &types.UserPermissions{
		UserID:      userID,
		TenantID:    tenantID,
		Roles:       roles,
		Permissions: uniquePermissions,
	}
	
	// 更新缓存
	_ = r.updatePermissionsCache(ctx, userPermissions)
	
	return userPermissions, nil
}

// HasPermission 检查用户是否拥有指定权限
func (r *roleRepository) HasPermission(ctx context.Context, userID, tenantID uint64, permission types.Permission) (bool, error) {
	userPermissions, err := r.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}
	
	return userPermissions.HasPermission(permission), nil
}

// HasRole 检查用户是否拥有指定角色
func (r *roleRepository) HasRole(ctx context.Context, userID, tenantID uint64, role types.RoleType) (bool, error) {
	userPermissions, err := r.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}
	
	return userPermissions.HasRole(role), nil
}

// RefreshPermissionsCache 刷新用户权限缓存
func (r *roleRepository) RefreshPermissionsCache(ctx context.Context, userID, tenantID uint64) error {
	// 清除现有缓存
	err := r.ClearPermissionsCache(ctx, userID, tenantID)
	if err != nil {
		return err
	}
	
	// 重新计算权限
	_, err = r.GetUserPermissions(ctx, userID, tenantID)
	return err
}

// ClearPermissionsCache 清除用户权限缓存
func (r *roleRepository) ClearPermissionsCache(ctx context.Context, userID, tenantID uint64) error {
	tenantCtx := r.WithTenant(ctx, tenantID)
	
	_, err := g.DB().Model("user_permissions_cache").Ctx(tenantCtx).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Delete()
	
	return err
}

// GetRoleUsers 获取指定角色的所有用户
func (r *roleRepository) GetRoleUsers(ctx context.Context, tenantID uint64, roleType types.RoleType) ([]uint64, error) {
	tenantCtx := r.WithTenant(ctx, tenantID)
	
	var userIDs []uint64
	err := g.DB().Model("user_roles").Ctx(tenantCtx).
		Fields("user_id").
		Where("tenant_id = ? AND role_type = ? AND status = 'active'", tenantID, roleType).
		Where("expires_at IS NULL OR expires_at > ?", gtime.Now()).
		Scan(&userIDs)
	
	if err != nil {
		return nil, fmt.Errorf("获取角色用户失败: %w", err)
	}
	
	return userIDs, nil
}

// GetTenantRoleStats 获取租户角色统计信息
func (r *roleRepository) GetTenantRoleStats(ctx context.Context, tenantID uint64) (map[types.RoleType]int, error) {
	tenantCtx := r.WithTenant(ctx, tenantID)
	
	type RoleCount struct {
		RoleType string `db:"role_type"`
		Count    int    `db:"count"`
	}
	
	var stats []RoleCount
	err := g.DB().Model("user_roles").Ctx(tenantCtx).
		Fields("role_type, COUNT(*) as count").
		Where("tenant_id = ? AND status = 'active'", tenantID).
		Where("expires_at IS NULL OR expires_at > ?", gtime.Now()).
		Group("role_type").
		Scan(&stats)
	
	if err != nil {
		return nil, fmt.Errorf("获取角色统计失败: %w", err)
	}
	
	result := make(map[types.RoleType]int)
	for _, stat := range stats {
		result[types.RoleType(stat.RoleType)] = stat.Count
	}
	
	return result, nil
}

// getPermissionsFromCache 从缓存获取权限信息
func (r *roleRepository) getPermissionsFromCache(ctx context.Context, userID, tenantID uint64) (*types.UserPermissions, error) {
	tenantCtx := r.WithTenant(ctx, tenantID)
	
	var cache UserPermissionsCache
	err := g.DB().Model("user_permissions_cache").Ctx(tenantCtx).
		Where("user_id = ? AND tenant_id = ? AND expires_at > ?", userID, tenantID, gtime.Now()).
		Scan(&cache)
	
	if err != nil {
		return nil, err
	}
	
	// 解析权限JSON
	var permissions []types.Permission
	if err := json.Unmarshal([]byte(cache.PermissionsJSON), &permissions); err != nil {
		return nil, fmt.Errorf("解析权限缓存失败: %w", err)
	}
	
	// 解析角色JSON
	var roles []types.RoleType
	if err := json.Unmarshal([]byte(cache.RolesJSON), &roles); err != nil {
		return nil, fmt.Errorf("解析角色缓存失败: %w", err)
	}
	
	return &types.UserPermissions{
		UserID:      userID,
		TenantID:    tenantID,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

// updatePermissionsCache 更新权限缓存
func (r *roleRepository) updatePermissionsCache(ctx context.Context, userPermissions *types.UserPermissions) error {
	tenantCtx := r.WithTenant(ctx, userPermissions.TenantID)
	
	// 序列化权限和角色
	permissionsJSON, err := json.Marshal(userPermissions.Permissions)
	if err != nil {
		return fmt.Errorf("序列化权限失败: %w", err)
	}
	
	rolesJSON, err := json.Marshal(userPermissions.Roles)
	if err != nil {
		return fmt.Errorf("序列化角色失败: %w", err)
	}
	
	// 缓存过期时间（1小时）
	expiresAt := gtime.Now().Add(time.Hour)
	
	// 更新或插入缓存
	_, err = g.DB().Model("user_permissions_cache").Ctx(tenantCtx).
		Replace(g.Map{
			"user_id":          userPermissions.UserID,
			"tenant_id":        userPermissions.TenantID,
			"permissions_json": string(permissionsJSON),
			"roles_json":       string(rolesJSON),
			"version":          1,
			"expires_at":       expiresAt,
			"created_at":       gtime.Now(),
			"updated_at":       gtime.Now(),
		})
	
	return err
}

// deduplicatePermissions 去重权限列表
func (r *roleRepository) deduplicatePermissions(permissions []types.Permission) []types.Permission {
	seen := make(map[types.Permission]bool)
	result := make([]types.Permission, 0)
	
	for _, permission := range permissions {
		if !seen[permission] {
			seen[permission] = true
			result = append(result, permission)
		}
	}
	
	return result
}