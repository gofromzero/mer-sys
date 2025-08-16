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
	ID         uint64     `json:"id" db:"id"`
	UUID       string     `json:"uuid" db:"uuid"`
	Username   string     `json:"username" db:"username"`
	Email      string     `json:"email" db:"email"`
	Phone      string     `json:"phone,omitempty" db:"phone"`
	TenantID   uint64     `json:"tenant_id" db:"tenant_id"`
	Status     UserStatus `json:"status" db:"status"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

// UserRole represents user role assignments
type UserRole struct {
	ID         uint64   `json:"id" db:"id"`
	UserID     uint64   `json:"user_id" db:"user_id"`
	RoleType   string   `json:"role_type" db:"role_type"`
	ResourceID *uint64  `json:"resource_id,omitempty" db:"resource_id"`
	Permissions []string `json:"permissions"`
}