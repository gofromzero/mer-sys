package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofromzero/mer-sys/backend/shared/audit"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// TenantSecurityService 租户安全服务
type TenantSecurityService struct {
	tenantRepo ITenantService
}

// NewTenantSecurityService 创建租户安全服务实例
func NewTenantSecurityService(tenantRepo ITenantService) *TenantSecurityService {
	return &TenantSecurityService{
		tenantRepo: tenantRepo,
	}
}

// ValidateResourceAccess 验证用户对租户资源的访问权限
func (s *TenantSecurityService) ValidateResourceAccess(ctx context.Context, userTenantID, targetTenantID uint64, operation string) error {
	// 获取当前用户的租户ID
	ctxTenantID := s.getTenantIDFromContext(ctx)
	
	// 检查上下文中的租户ID是否与用户租户ID一致
	if ctxTenantID != userTenantID {
		audit.LogSecurityViolation(ctx, userTenantID, "tenant_id_mismatch", 
			fmt.Sprintf("用户声称的租户ID与上下文不匹配: user_tenant=%d, ctx_tenant=%d", userTenantID, ctxTenantID), 
			map[string]interface{}{
				"user_tenant_id": userTenantID,
				"ctx_tenant_id":  ctxTenantID,
				"operation":      operation,
			})
		return errors.New("租户ID验证失败")
	}

	// 检查是否为跨租户访问
	if targetTenantID != 0 && targetTenantID != userTenantID {
		// 检查用户是否有系统级别的租户管理权限
		if !s.hasSystemTenantManagePermission(ctx) {
			audit.LogCrossTenantAttempt(ctx, userTenantID, targetTenantID, "tenant", operation, map[string]interface{}{
				"reason": "用户尝试访问其他租户资源但没有系统级权限",
			})
			return errors.New("无权访问其他租户的资源")
		}
	}

	return nil
}

// ValidateTenantOperation 验证租户操作权限
func (s *TenantSecurityService) ValidateTenantOperation(ctx context.Context, tenantID uint64, operation string, requireSystemLevel bool) error {
	userTenantID := s.getTenantIDFromContext(ctx)
	
	// 如果操作需要系统级权限
	if requireSystemLevel {
		if !s.hasSystemTenantManagePermission(ctx) {
			audit.LogSecurityViolation(ctx, userTenantID, "insufficient_privileges", 
				fmt.Sprintf("用户尝试执行需要系统级权限的操作: %s", operation),
				map[string]interface{}{
					"operation":   operation,
					"tenant_id":   tenantID,
					"user_tenant": userTenantID,
				})
			return errors.New("权限不足：需要系统级租户管理权限")
		}
	}

	// 验证资源访问权限
	return s.ValidateResourceAccess(ctx, userTenantID, tenantID, operation)
}

// ValidateTenantStatusChange 验证租户状态变更权限
func (s *TenantSecurityService) ValidateTenantStatusChange(ctx context.Context, tenantID uint64, oldStatus, newStatus types.TenantStatus, reason string) error {
	// 状态变更需要系统级权限
	if err := s.ValidateTenantOperation(ctx, tenantID, "status_change", true); err != nil {
		return err
	}

	// 验证状态转换的合法性
	if !s.isValidStatusTransition(oldStatus, newStatus) {
		audit.LogSecurityViolation(ctx, s.getTenantIDFromContext(ctx), "invalid_status_transition",
			fmt.Sprintf("尝试进行无效的状态转换: %s -> %s", oldStatus, newStatus),
			map[string]interface{}{
				"tenant_id":  tenantID,
				"old_status": oldStatus,
				"new_status": newStatus,
				"reason":     reason,
			})
		return fmt.Errorf("无效的状态转换: %s -> %s", oldStatus, newStatus)
	}

	// 记录状态变更审计
	audit.LogTenantAccess(ctx, tenantID, "tenant", "status_validation_passed", map[string]interface{}{
		"old_status": oldStatus,
		"new_status": newStatus,
		"reason":     reason,
	})

	return nil
}

// ValidateTenantConfigChange 验证租户配置变更权限
func (s *TenantSecurityService) ValidateTenantConfigChange(ctx context.Context, tenantID uint64, oldConfig, newConfig *types.TenantConfig) error {
	// 配置变更需要系统级权限
	if err := s.ValidateTenantOperation(ctx, tenantID, "config_change", true); err != nil {
		return err
	}

	// 验证配置变更的合理性
	if newConfig.MaxUsers <= 0 || newConfig.MaxMerchants <= 0 {
		audit.LogSecurityViolation(ctx, s.getTenantIDFromContext(ctx), "invalid_config_values",
			"尝试设置无效的配置值",
			map[string]interface{}{
				"tenant_id":      tenantID,
				"max_users":      newConfig.MaxUsers,
				"max_merchants":  newConfig.MaxMerchants,
			})
		return errors.New("配置值无效：用户数和商户数必须大于0")
	}

	// 检查是否有危险的配置变更
	if s.isDangerousConfigChange(oldConfig, newConfig) {
		audit.LogSecurityViolation(ctx, s.getTenantIDFromContext(ctx), "dangerous_config_change",
			"检测到危险的配置变更",
			map[string]interface{}{
				"tenant_id":  tenantID,
				"old_config": oldConfig,
				"new_config": newConfig,
			})
		return errors.New("拒绝危险的配置变更")
	}

	return nil
}

// ValidateSensitiveOperation 验证敏感操作
func (s *TenantSecurityService) ValidateSensitiveOperation(ctx context.Context, operation string, params map[string]interface{}) error {
	userTenantID := s.getTenantIDFromContext(ctx)
	
	// 敏感操作需要二次确认
	if !s.hasSystemTenantManagePermission(ctx) {
		audit.LogSecurityViolation(ctx, userTenantID, "unauthorized_sensitive_operation",
			fmt.Sprintf("用户尝试执行敏感操作但权限不足: %s", operation),
			map[string]interface{}{
				"operation": operation,
				"params":    params,
			})
		return errors.New("权限不足：无法执行敏感操作")
	}

	// 记录敏感操作审计
	audit.LogTenantAccess(ctx, userTenantID, "tenant", "sensitive_operation", map[string]interface{}{
		"operation": operation,
		"params":    params,
	})

	return nil
}

// 辅助方法

// getTenantIDFromContext 从上下文获取租户ID
func (s *TenantSecurityService) getTenantIDFromContext(ctx context.Context) uint64 {
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(uint64); ok {
			return id
		}
	}
	return 0
}

// hasSystemTenantManagePermission 检查用户是否具有系统级租户管理权限
func (s *TenantSecurityService) hasSystemTenantManagePermission(ctx context.Context) bool {
	// 从上下文获取用户权限
	if permissions := ctx.Value("user_permissions"); permissions != nil {
		if userPerms, ok := permissions.(types.UserPermissions); ok {
			return userPerms.HasPermission(types.PermissionTenantManage)
		}
	}
	return false
}

// isValidStatusTransition 检查状态转换是否有效
func (s *TenantSecurityService) isValidStatusTransition(oldStatus, newStatus types.TenantStatus) bool {
	// 定义有效的状态转换
	validTransitions := map[types.TenantStatus][]types.TenantStatus{
		types.TenantStatusActive: {
			types.TenantStatusSuspended,
			types.TenantStatusExpired,
		},
		types.TenantStatusSuspended: {
			types.TenantStatusActive,
			types.TenantStatusExpired,
		},
		types.TenantStatusExpired: {
			types.TenantStatusActive,
			types.TenantStatusSuspended,
		},
	}

	validNext, exists := validTransitions[oldStatus]
	if !exists {
		return false
	}

	for _, status := range validNext {
		if status == newStatus {
			return true
		}
	}

	return false
}

// isDangerousConfigChange 检查是否为危险的配置变更
func (s *TenantSecurityService) isDangerousConfigChange(oldConfig, newConfig *types.TenantConfig) bool {
	// 检查是否有大幅度的配置缩减
	if newConfig.MaxUsers < oldConfig.MaxUsers/2 || newConfig.MaxMerchants < oldConfig.MaxMerchants/2 {
		return true
	}

	// 检查是否移除了关键功能
	oldFeatures := make(map[string]bool)
	for _, feature := range oldConfig.Features {
		oldFeatures[feature] = true
	}

	criticalFeatures := []string{"basic"}
	for _, feature := range criticalFeatures {
		if oldFeatures[feature] {
			newHasFeature := false
			for _, newFeature := range newConfig.Features {
				if newFeature == feature {
					newHasFeature = true
					break
				}
			}
			if !newHasFeature {
				return true // 移除了关键功能
			}
		}
	}

	return false
}