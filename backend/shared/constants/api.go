package constants

// API response codes
const (
	SuccessCode    = 0
	BadRequestCode = 400
	UnauthorizedCode = 401
	ForbiddenCode    = 403
	NotFoundCode     = 404
	InternalErrorCode = 500
)

// API paths
const (
	APIPrefix      = "/api/v1"
	HealthEndpoint = "/health"
	AuthPrefix     = "/auth"
	UsersPrefix    = "/users"
	TenantsPrefix  = "/tenants"
)

// HTTP headers
const (
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"
	HeaderXTenantID     = "X-Tenant-ID"
	HeaderXUserID       = "X-User-ID"
)