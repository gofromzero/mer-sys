package middleware

import (
	"strconv"
	"strings"

	"github.com/gofromzero/mer-sys/backend/shared/constants"
	"github.com/gofromzero/mer-sys/backend/shared/utils"
	"github.com/gogf/gf/v2/net/ghttp"
)

// TenantIsolation ensures tenant data isolation
func TenantIsolation() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		// Skip for public paths
		if isPublicPath(r.URL.Path) {
			r.Middleware.Next()
			return
		}

		// Extract tenant ID from header or JWT token (simplified for now)
		tenantHeader := r.Header.Get(constants.HeaderXTenantID)
		if tenantHeader == "" {
			utils.ErrorResponse(r, constants.UnauthorizedCode, "Tenant ID required")
			return
		}

		tenantID, err := strconv.ParseUint(tenantHeader, 10, 64)
		if err != nil {
			utils.ErrorResponse(r, constants.BadRequestCode, "Invalid tenant ID")
			return
		}

		// Set tenant context
		r.SetCtxVar("tenant_id", tenantID)
		r.Middleware.Next()
	}
}

// isPublicPath checks if the path is public (no tenant isolation required)
func isPublicPath(path string) bool {
	publicPaths := []string{
		constants.APIPrefix + constants.HealthEndpoint,
		constants.APIPrefix + constants.AuthPrefix + "/login",
		constants.APIPrefix + constants.AuthPrefix + "/register",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}
