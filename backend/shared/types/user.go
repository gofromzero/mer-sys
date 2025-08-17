package types

import "time"

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