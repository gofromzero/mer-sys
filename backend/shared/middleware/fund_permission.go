package middleware

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
)

// FundPermissions 资金操作权限常量
const (
	FundPermissionDeposit  = "fund:deposit"
	FundPermissionAllocate = "fund:allocate"
	FundPermissionView     = "fund:view"
	FundPermissionFreeze   = "fund:freeze"
)

// FundPermissionMiddleware 资金权限检查中间件
func FundPermissionMiddleware(requiredPermission string) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		// 检查用户是否已认证
		userID := r.GetCtxVar("user_id")
		if userID == nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    401,
				"message": "未认证用户",
			})
			return
		}

		// 获取用户权限
		permissions := r.GetCtxVar("permissions")
		if permissions == nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    403,
				"message": "无权限信息",
			})
			return
		}

		// 检查是否有所需权限
		userPermissions := gconv.Strings(permissions)
		hasPermission := false
		for _, perm := range userPermissions {
			if perm == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			r.Response.WriteJsonExit(g.Map{
				"code":    403,
				"message": "权限不足，需要权限: " + requiredPermission,
			})
			return
		}

		r.Middleware.Next()
	}
}

// RequireFundDeposit 需要充值权限
func RequireFundDeposit(r *ghttp.Request) {
	FundPermissionMiddleware(FundPermissionDeposit)(r)
}

// RequireFundAllocate 需要分配权限
func RequireFundAllocate(r *ghttp.Request) {
	FundPermissionMiddleware(FundPermissionAllocate)(r)
}

// RequireFundView 需要查看权限
func RequireFundView(r *ghttp.Request) {
	FundPermissionMiddleware(FundPermissionView)(r)
}

// RequireFundFreeze 需要冻结权限
func RequireFundFreeze(r *ghttp.Request) {
	FundPermissionMiddleware(FundPermissionFreeze)(r)
}