package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/os/gtime"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证业务逻辑服务
type AuthService struct {
	userRepo *repository.UserRepository
}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
	}
}

// ValidateUserCredentials 验证用户凭证
func (s *AuthService) ValidateUserCredentials(ctx context.Context, username, password string, tenantID uint64) (*types.User, *types.UserPermissions, error) {
	// 通过用户名和租户ID查找用户
	user, err := s.userRepo.FindByUsernameAndTenant(ctx, username, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("用户不存在: %v", err)
	}

	// 验证密码
	if !s.verifyPassword(password, user.PasswordHash) {
		return nil, nil, fmt.Errorf("密码错误")
	}

	// 检查用户状态
	if user.Status != types.UserStatusActive {
		return nil, nil, fmt.Errorf("用户账户状态异常: %s", user.Status)
	}

	// 获取用户权限信息
	userPermissions, err := s.getUserPermissions(ctx, user.ID, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("获取用户权限失败: %v", err)
	}

	return user, userPermissions, nil
}

// verifyPassword 验证密码
func (s *AuthService) verifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// HashPassword 加密密码
func (s *AuthService) HashPassword(password string) (string, error) {
	// 使用默认的cost（通常是10-12）
	const cost = 12
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %v", err)
	}
	return string(hashedBytes), nil
}

// getUserPermissions 获取用户权限信息
func (s *AuthService) getUserPermissions(ctx context.Context, userID, tenantID uint64) (*types.UserPermissions, error) {
	// 获取用户角色
	userRoles, err := s.userRepo.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("获取用户角色失败: %v", err)
	}

	// 如果用户没有角色，分配默认客户角色
	if len(userRoles) == 0 {
		userRoles = []types.RoleType{types.RoleCustomer}
	}

	// 获取角色对应的权限
	permissions := s.calculateUserPermissions(userRoles)

	return &types.UserPermissions{
		UserID:      userID,
		TenantID:    tenantID,
		Roles:       userRoles,
		Permissions: permissions,
	}, nil
}

// calculateUserPermissions 计算用户权限
func (s *AuthService) calculateUserPermissions(roles []types.RoleType) []types.Permission {
	// 获取默认角色配置
	defaultRoles := types.GetDefaultRoles()

	// 收集所有权限，去重
	permissionSet := make(map[types.Permission]bool)

	for _, roleType := range roles {
		if role, exists := defaultRoles[roleType]; exists {
			for _, permission := range role.Permissions {
				permissionSet[permission] = true
			}
		}
	}

	// 转换为数组
	var permissions []types.Permission
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}

	return permissions
}

// UpdateLastLoginTime 更新用户最后登录时间
func (s *AuthService) UpdateLastLoginTime(ctx context.Context, userID uint64) error {
	now := gtime.Now()
	return s.userRepo.UpdateLastLoginTime(ctx, userID, now)
}

// GetUserByID 根据ID获取用户信息
func (s *AuthService) GetUserByID(ctx context.Context, userID, tenantID uint64) (*types.User, error) {
	return s.userRepo.FindByIDAndTenant(ctx, userID, tenantID)
}

// CreateUser 创建用户（用于注册等场景）
func (s *AuthService) CreateUser(ctx context.Context, user *types.User, password string) error {
	// 加密密码
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// 设置密码哈希
	user.PasswordHash = hashedPassword

	// 生成UUID
	if user.UUID == "" {
		user.UUID = s.generateUserUUID(user.Username, user.TenantID)
	}

	// 设置默认状态
	if user.Status == "" {
		user.Status = types.UserStatusPending // 新用户默认待激活
	}

	// 设置创建时间
	now := gtime.Now().Time
	user.CreatedAt = now
	user.UpdatedAt = now

	// 保存用户
	return s.userRepo.Create(ctx, user)
}

// generateUserUUID 生成用户UUID
func (s *AuthService) generateUserUUID(username string, tenantID uint64) string {
	data := fmt.Sprintf("%s-%d-%d", username, tenantID, time.Now().UnixNano())
	return gmd5.MustEncryptString(data)
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	// 获取用户当前密码哈希
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %v", err)
	}

	// 验证旧密码
	if !s.verifyPassword(oldPassword, user.PasswordHash) {
		return fmt.Errorf("原密码错误")
	}

	// 加密新密码
	newPasswordHash, err := s.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("新密码加密失败: %v", err)
	}

	// 更新密码
	return s.userRepo.UpdatePassword(ctx, userID, newPasswordHash)
}

// ResetPassword 重置密码（管理员功能）
func (s *AuthService) ResetPassword(ctx context.Context, userID uint64, newPassword string) error {
	// 加密新密码
	newPasswordHash, err := s.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// 更新密码
	return s.userRepo.UpdatePassword(ctx, userID, newPasswordHash)
}

// DeactivateUser 停用用户
func (s *AuthService) DeactivateUser(ctx context.Context, userID uint64) error {
	return s.userRepo.UpdateStatus(ctx, userID, types.UserStatusDeactivated)
}

// ActivateUser 激活用户
func (s *AuthService) ActivateUser(ctx context.Context, userID uint64) error {
	return s.userRepo.UpdateStatus(ctx, userID, types.UserStatusActive)
}

// SuspendUser 暂停用户
func (s *AuthService) SuspendUser(ctx context.Context, userID uint64) error {
	return s.userRepo.UpdateStatus(ctx, userID, types.UserStatusSuspended)
}

// ValidatePassword 验证密码强度
func (s *AuthService) ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("密码长度不能少于6位")
	}

	if len(password) > 128 {
		return fmt.Errorf("密码长度不能超过128位")
	}

	// 这里可以添加更复杂的密码强度验证逻辑
	// 例如：必须包含大小写字母、数字、特殊字符等

	return nil
}

// CheckUsernameExists 检查用户名是否已存在
func (s *AuthService) CheckUsernameExists(ctx context.Context, username string, tenantID uint64) (bool, error) {
	user, err := s.userRepo.FindByUsernameAndTenant(ctx, username, tenantID)
	if err != nil {
		// 如果是"记录不存在"错误，说明用户名可用
		if err.Error() == "record not found" || err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, err
	}

	return user != nil, nil
}

// CheckEmailExists 检查邮箱是否已存在
func (s *AuthService) CheckEmailExists(ctx context.Context, email string, tenantID uint64) (bool, error) {
	user, err := s.userRepo.FindByEmailAndTenant(ctx, email, tenantID)
	if err != nil {
		// 如果是"记录不存在"错误，说明邮箱可用
		if err.Error() == "record not found" || err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, err
	}

	return user != nil, nil
}
