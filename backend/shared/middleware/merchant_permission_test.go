package middleware

import (
	"context"
	"testing"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/net/ghttp"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMerchantPermissionUtilities(t *testing.T) {
	Convey("商户权限工具函数测试", t, func() {
		
		Convey("ExtractMerchantIDFromRequest 函数应该存在", func() {
			// 这里只测试函数存在性，因为需要ghttp.Request对象
			var _ func(*ghttp.Request) (uint64, error) = ExtractMerchantIDFromRequest
		})
		
		Convey("上下文工具函数", func() {
			ctx := context.Background()
			merchantID := uint64(123)
			roles := []types.RoleType{types.RoleMerchantAdmin}
			user := &types.User{
				ID:         1,
				Username:   "test_user",
				MerchantID: &merchantID,
			}
			
			Convey("GetMerchantIDFromContext 应该正确获取商户ID", func() {
				// 设置商户ID到上下文
				ctx = context.WithValue(ctx, "merchant_id", merchantID)
				
				id, ok := GetMerchantIDFromContext(ctx)
				So(ok, ShouldBeTrue)
				So(id, ShouldEqual, merchantID)
				
				// 测试空上下文
				emptyCtx := context.Background()
				_, ok = GetMerchantIDFromContext(emptyCtx)
				So(ok, ShouldBeFalse)
			})
			
			Convey("GetMerchantUserFromContext 应该正确获取商户用户", func() {
				// 设置商户用户到上下文
				ctx = context.WithValue(ctx, "merchant_user", user)
				
				retrievedUser, ok := GetMerchantUserFromContext(ctx)
				So(ok, ShouldBeTrue)
				So(retrievedUser, ShouldNotBeNil)
				So(retrievedUser.ID, ShouldEqual, user.ID)
				So(retrievedUser.Username, ShouldEqual, user.Username)
				
				// 测试空上下文
				emptyCtx := context.Background()
				_, ok = GetMerchantUserFromContext(emptyCtx)
				So(ok, ShouldBeFalse)
			})
			
			Convey("GetMerchantRolesFromContext 应该正确获取商户角色", func() {
				// 设置商户角色到上下文
				ctx = context.WithValue(ctx, "merchant_roles", roles)
				
				retrievedRoles, ok := GetMerchantRolesFromContext(ctx)
				So(ok, ShouldBeTrue)
				So(len(retrievedRoles), ShouldEqual, 1)
				So(retrievedRoles[0], ShouldEqual, types.RoleMerchantAdmin)
				
				// 测试空上下文
				emptyCtx := context.Background()
				_, ok = GetMerchantRolesFromContext(emptyCtx)
				So(ok, ShouldBeFalse)
			})
		})
		
		Convey("角色检查函数", func() {
			
			Convey("IsMerchantAdmin 应该正确检查商户管理员角色", func() {
				// 设置管理员角色
				ctx := context.WithValue(context.Background(), "merchant_roles", []types.RoleType{types.RoleMerchantAdmin})
				So(IsMerchantAdmin(ctx), ShouldBeTrue)
				
				// 设置操作员角色
				ctx = context.WithValue(context.Background(), "merchant_roles", []types.RoleType{types.RoleMerchantOperator})
				So(IsMerchantAdmin(ctx), ShouldBeFalse)
				
				// 空上下文
				So(IsMerchantAdmin(context.Background()), ShouldBeFalse)
			})
			
			Convey("IsMerchantOperator 应该正确检查商户操作员角色", func() {
				// 设置操作员角色
				ctx := context.WithValue(context.Background(), "merchant_roles", []types.RoleType{types.RoleMerchantOperator})
				So(IsMerchantOperator(ctx), ShouldBeTrue)
				
				// 设置管理员角色
				ctx = context.WithValue(context.Background(), "merchant_roles", []types.RoleType{types.RoleMerchantAdmin})
				So(IsMerchantOperator(ctx), ShouldBeFalse)
				
				// 空上下文
				So(IsMerchantOperator(context.Background()), ShouldBeFalse)
			})
		})
		
		Convey("权限检查函数", func() {
			
			Convey("HasMerchantPermission 应该正确检查商户权限", func() {
				// 设置管理员角色（拥有所有权限）
				ctx := context.WithValue(context.Background(), "merchant_roles", []types.RoleType{types.RoleMerchantAdmin})
				
				So(HasMerchantPermission(ctx, types.PermissionMerchantProductView), ShouldBeTrue)
				So(HasMerchantPermission(ctx, types.PermissionMerchantProductCreate), ShouldBeTrue)
				So(HasMerchantPermission(ctx, types.PermissionMerchantProductDelete), ShouldBeTrue)
				So(HasMerchantPermission(ctx, types.PermissionMerchantUserManage), ShouldBeTrue)
				
				// 设置操作员角色（有限权限）
				ctx = context.WithValue(context.Background(), "merchant_roles", []types.RoleType{types.RoleMerchantOperator})
				
				So(HasMerchantPermission(ctx, types.PermissionMerchantProductView), ShouldBeTrue)
				So(HasMerchantPermission(ctx, types.PermissionMerchantProductCreate), ShouldBeTrue)
				So(HasMerchantPermission(ctx, types.PermissionMerchantProductDelete), ShouldBeFalse) // 操作员没有删除权限
				So(HasMerchantPermission(ctx, types.PermissionMerchantUserManage), ShouldBeFalse)    // 操作员没有用户管理权限
				
				// 空上下文
				So(HasMerchantPermission(context.Background(), types.PermissionMerchantProductView), ShouldBeFalse)
			})
		})
	})
}

func TestMerchantPermissionMiddleware(t *testing.T) {
	Convey("商户权限中间件结构测试", t, func() {
		
		Convey("MerchantPermissionMiddleware 应该可以实例化", func() {
			// 这里只测试结构体存在性，避免数据库依赖
			middleware := &MerchantPermissionMiddleware{}
			So(middleware, ShouldNotBeNil)
		})
		
		Convey("中间件方法应该存在", func() {
			middleware := &MerchantPermissionMiddleware{}
			
			// 验证方法存在性
			var _ func(*MerchantPermissionMiddleware, ...types.Permission) ghttp.HandlerFunc = (*MerchantPermissionMiddleware).RequireMerchantPermission
			var _ func(*MerchantPermissionMiddleware, ...types.RoleType) ghttp.HandlerFunc = (*MerchantPermissionMiddleware).RequireMerchantRole
			var _ func(*MerchantPermissionMiddleware, []types.RoleType, []types.Permission) bool = (*MerchantPermissionMiddleware).checkUserMerchantPermissions
			
			So(middleware, ShouldNotBeNil)
		})
	})
}

func TestMerchantPermissionLogic(t *testing.T) {
	Convey("商户权限逻辑测试", t, func() {
		middleware := &MerchantPermissionMiddleware{}
		
		Convey("checkUserMerchantPermissions 权限检查逻辑", func() {
			
			Convey("管理员角色应该拥有所有权限", func() {
				userRoles := []types.RoleType{types.RoleMerchantAdmin}
				requiredPermissions := []types.Permission{
					types.PermissionMerchantProductView,
					types.PermissionMerchantProductCreate,
					types.PermissionMerchantProductEdit,
					types.PermissionMerchantProductDelete,
					types.PermissionMerchantUserManage,
				}
				
				hasPermission := middleware.checkUserMerchantPermissions(userRoles, requiredPermissions)
				So(hasPermission, ShouldBeTrue)
			})
			
			Convey("操作员角色应该有限制权限", func() {
				userRoles := []types.RoleType{types.RoleMerchantOperator}
				
				// 操作员应该有的权限
				allowedPermissions := []types.Permission{
					types.PermissionMerchantProductView,
					types.PermissionMerchantProductCreate,
					types.PermissionMerchantProductEdit,
				}
				hasPermission := middleware.checkUserMerchantPermissions(userRoles, allowedPermissions)
				So(hasPermission, ShouldBeTrue)
				
				// 操作员不应该有的权限
				deniedPermissions := []types.Permission{
					types.PermissionMerchantProductDelete,
					types.PermissionMerchantUserManage,
				}
				hasPermission = middleware.checkUserMerchantPermissions(userRoles, deniedPermissions)
				So(hasPermission, ShouldBeFalse)
			})
			
			Convey("无角色用户应该没有权限", func() {
				userRoles := []types.RoleType{}
				requiredPermissions := []types.Permission{types.PermissionMerchantProductView}
				
				hasPermission := middleware.checkUserMerchantPermissions(userRoles, requiredPermissions)
				So(hasPermission, ShouldBeFalse)
			})
			
			Convey("非商户角色应该没有商户权限", func() {
				userRoles := []types.RoleType{types.RoleTenantAdmin}
				requiredPermissions := []types.Permission{types.PermissionMerchantProductView}
				
				hasPermission := middleware.checkUserMerchantPermissions(userRoles, requiredPermissions)
				So(hasPermission, ShouldBeFalse)
			})
		})
	})
}

