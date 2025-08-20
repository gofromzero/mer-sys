package types

import (
	"fmt"
	"time"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusPending     UserStatus = "pending"
	UserStatusActive      UserStatus = "active"
	UserStatusSuspended   UserStatus = "suspended"
	UserStatusDeactivated UserStatus = "deactivated"
)

// User represents the user entity with multi-tenant support
type User struct {
	ID           uint64       `json:"id" db:"id"`
	UUID         string       `json:"uuid" db:"uuid"`
	Username     string       `json:"username" db:"username"`
	Email        string       `json:"email" db:"email"`
	Phone        string       `json:"phone,omitempty" db:"phone"`
	PasswordHash string       `json:"-" db:"password_hash"` // 不在JSON中输出
	TenantID     uint64       `json:"tenant_id" db:"tenant_id"`
	MerchantID   *uint64      `json:"merchant_id,omitempty" db:"merchant_id"` // 关联商户ID，可为空
	Status       UserStatus   `json:"status" db:"status"`
	Profile      *UserProfile `json:"profile,omitempty" db:"profile"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time   `json:"last_login_at,omitempty" db:"last_login_at"`
}

// UserProfile represents additional user profile information
type UserProfile struct {
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	Department  string `json:"department,omitempty"`
	Position    string `json:"position,omitempty"`
	Description string `json:"description,omitempty"`
}

// UserInfo represents user information for API responses (excludes sensitive data)
type UserInfo struct {
	ID          uint64       `json:"id"`
	UUID        string       `json:"uuid"`
	Username    string       `json:"username"`
	Email       string       `json:"email"`
	Phone       string       `json:"phone,omitempty"`
	Status      UserStatus   `json:"status"`
	TenantID    uint64       `json:"tenant_id"`
	MerchantID  *uint64      `json:"merchant_id,omitempty"`
	Roles       []RoleType   `json:"roles,omitempty"`
	Profile     *UserProfile `json:"profile,omitempty"`
	CreatedAt   time.Time    `json:"created_at,omitempty"`
	UpdatedAt   time.Time    `json:"updated_at,omitempty"`
	LastLoginAt *time.Time   `json:"last_login_at,omitempty"`
}

// UserRole represents user role assignments
type UserRole struct {
	ID         uint64   `json:"id" db:"id"`
	UserID     uint64   `json:"user_id" db:"user_id"`
	RoleType   string   `json:"role_type" db:"role_type"`
	ResourceID *uint64  `json:"resource_id,omitempty" db:"resource_id"`
	Permissions []string `json:"permissions"`
}

// CreateMerchantUserRequest 创建商户用户请求
type CreateMerchantUserRequest struct {
	Username   string    `json:"username" validate:"required,min=3,max=50"`
	Email      string    `json:"email" validate:"required,email"`
	Phone      string    `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
	Password   string    `json:"password" validate:"required,min=8,max=128"`
	MerchantID uint64    `json:"merchant_id" validate:"required"`
	RoleType   RoleType  `json:"role_type" validate:"required,oneof=merchant_admin merchant_operator"`
	Profile    *UserProfile `json:"profile,omitempty"`
}

// UpdateMerchantUserRequest 更新商户用户请求
type UpdateMerchantUserRequest struct {
	Username string       `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    string       `json:"email,omitempty" validate:"omitempty,email"`
	Phone    string       `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
	RoleType RoleType     `json:"role_type,omitempty" validate:"omitempty,oneof=merchant_admin merchant_operator"`
	Status   UserStatus   `json:"status,omitempty" validate:"omitempty,oneof=pending active suspended deactivated"`
	Profile  *UserProfile `json:"profile,omitempty"`
}

// ValidateMerchantUser 验证商户用户数据
func (u *User) ValidateMerchantUser() error {
	// 商户用户必须有merchant_id
	if u.MerchantID == nil {
		return fmt.Errorf("商户用户必须关联商户")
	}

	// 验证用户名长度
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return fmt.Errorf("用户名长度必须在3-50字符之间")
	}

	// 验证邮箱格式
	if u.Email == "" {
		return fmt.Errorf("邮箱不能为空")
	}

	// 验证手机号格式（如果提供）
	if u.Phone != "" && (len(u.Phone) < 10 || len(u.Phone) > 20) {
		return fmt.Errorf("手机号长度必须在10-20字符之间")
	}

	return nil
}

// IsMerchantUser 检查是否为商户用户
func (u *User) IsMerchantUser() bool {
	return u.MerchantID != nil
}

// HasMerchantRole 检查用户是否具有指定的商户角色
func (u *User) HasMerchantRole(roleType RoleType) bool {
	return u.IsMerchantUser() && 
		   (roleType == RoleMerchantAdmin || roleType == RoleMerchantOperator)
}