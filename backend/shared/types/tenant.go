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
	ID               uint64       `json:"id" db:"id"`
	Name             string       `json:"name" db:"name"`
	Code             string       `json:"code" db:"code"`
	Status           TenantStatus `json:"status" db:"status"`
	Config           string       `json:"config" db:"config"` // JSON string
	BusinessType     string       `json:"business_type" db:"business_type"`
	ContactPerson    string       `json:"contact_person" db:"contact_person"`
	ContactEmail     string       `json:"contact_email" db:"contact_email"`
	ContactPhone     string       `json:"contact_phone" db:"contact_phone"`
	Address          string       `json:"address" db:"address"`
	RegistrationTime *time.Time   `json:"registration_time" db:"registration_time"`
	ActivationTime   *time.Time   `json:"activation_time" db:"activation_time"`
	CreatedAt        time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at" db:"updated_at"`
}

// TenantConfig represents tenant configuration (will be JSON marshaled)
type TenantConfig struct {
	MaxUsers     int               `json:"max_users"`
	MaxMerchants int               `json:"max_merchants"`
	Features     []string          `json:"features"`
	Settings     map[string]string `json:"settings"`
}

// CreateTenantRequest represents the request payload for tenant registration
type CreateTenantRequest struct {
	Name          string `json:"name" v:"required|length:2,100#租户名称不能为空|租户名称长度为2-100个字符"`
	Code          string `json:"code" v:"required|length:2,50|regex:^[a-zA-Z0-9_-]+$#租户代码不能为空|租户代码长度为2-50个字符|租户代码只能包含字母、数字、下划线和连字符"`
	BusinessType  string `json:"business_type" v:"required#业务类型不能为空"`
	ContactPerson string `json:"contact_person" v:"required#联系人不能为空"`
	ContactEmail  string `json:"contact_email" v:"required|email#联系邮箱不能为空|邮箱格式不正确"`
	ContactPhone  string `json:"contact_phone" v:"phone#联系电话格式不正确"`
	Address       string `json:"address"`
}

// UpdateTenantRequest represents the request payload for tenant updates
type UpdateTenantRequest struct {
	Name          string `json:"name" v:"length:2,100#租户名称长度为2-100个字符"`
	BusinessType  string `json:"business_type"`
	ContactPerson string `json:"contact_person"`
	ContactEmail  string `json:"contact_email" v:"email#邮箱格式不正确"`
	ContactPhone  string `json:"contact_phone" v:"phone#联系电话格式不正确"`
	Address       string `json:"address"`
}

// UpdateTenantStatusRequest represents the request payload for tenant status changes
type UpdateTenantStatusRequest struct {
	Status TenantStatus `json:"status" v:"required|in:active,suspended,expired#状态不能为空|状态值无效"`
	Reason string       `json:"reason" v:"required#状态变更原因不能为空"`
}

// ListTenantsRequest represents the request for listing tenants with filters
type ListTenantsRequest struct {
	Page         int          `json:"page" v:"min:1#页码必须大于0"`
	PageSize     int          `json:"page_size" v:"min:1|max:100#每页数量必须在1-100之间"`
	Status       TenantStatus `json:"status"`
	BusinessType string       `json:"business_type"`
	Search       string       `json:"search"` // Search in name, code, contact_person, contact_email
}

// TenantResponse represents the response format for tenant data
type TenantResponse struct {
	ID               uint64     `json:"id"`
	Name             string     `json:"name"`
	Code             string     `json:"code"`
	Status           string     `json:"status"`
	BusinessType     string     `json:"business_type"`
	ContactPerson    string     `json:"contact_person"`
	ContactEmail     string     `json:"contact_email"`
	ContactPhone     string     `json:"contact_phone"`
	Address          string     `json:"address"`
	RegistrationTime *time.Time `json:"registration_time"`
	ActivationTime   *time.Time `json:"activation_time"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ListTenantsResponse represents the response for tenant list
type ListTenantsResponse struct {
	Total   int              `json:"total"`
	Page    int              `json:"page"`
	Size    int              `json:"size"`
	Tenants []TenantResponse `json:"tenants"`
}