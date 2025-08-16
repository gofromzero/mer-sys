package types

import "time"

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusExpired   TenantStatus = "expired"
)

// Tenant represents the tenant entity for multi-tenant architecture
type Tenant struct {
	ID        uint64       `json:"id" db:"id"`
	Name      string       `json:"name" db:"name"`
	Code      string       `json:"code" db:"code"`
	Status    TenantStatus `json:"status" db:"status"`
	Config    string       `json:"config" db:"config"` // JSON string
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

// TenantConfig represents tenant configuration (will be JSON marshaled)
type TenantConfig struct {
	MaxUsers     int               `json:"max_users"`
	MaxMerchants int               `json:"max_merchants"`
	Features     []string          `json:"features"`
	Settings     map[string]string `json:"settings"`
}