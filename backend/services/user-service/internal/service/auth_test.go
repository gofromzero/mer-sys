package service

import (
	"testing"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthService(t *testing.T) {
	Convey("认证服务测试", t, func() {
		// 创建认证服务实例
		authService := NewAuthService()

		Convey("服务创建", func() {
			So(authService, ShouldNotBeNil)
			So(authService.userRepo, ShouldNotBeNil)
		})
	})
}

func TestPasswordHashing(t *testing.T) {
	Convey("密码加密测试", t, func() {
		authService := NewAuthService()
		password := "test123456"

		Convey("密码加密", func() {
			hashedPassword, err := authService.HashPassword(password)
			So(err, ShouldBeNil)
			So(hashedPassword, ShouldNotBeEmpty)
			So(hashedPassword, ShouldNotEqual, password)
			So(len(hashedPassword), ShouldBeGreaterThan, 20) // bcrypt hash should be much longer
		})

		Convey("密码验证", func() {
			hashedPassword, err := authService.HashPassword(password)
			So(err, ShouldBeNil)

			// 正确密码验证
			isValid := authService.verifyPassword(password, hashedPassword)
			So(isValid, ShouldBeTrue)

			// 错误密码验证
			isValid = authService.verifyPassword("wrongpassword", hashedPassword)
			So(isValid, ShouldBeFalse)
		})
	})
}

func TestPasswordValidation(t *testing.T) {
	Convey("密码强度验证测试", t, func() {
		authService := NewAuthService()

		Convey("有效密码", func() {
			err := authService.ValidatePassword("test123456")
			So(err, ShouldBeNil)
		})

		Convey("过短密码", func() {
			err := authService.ValidatePassword("12345")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "密码长度不能少于6位")
		})

		Convey("过长密码", func() {
			longPassword := make([]byte, 130)
			for i := range longPassword {
				longPassword[i] = 'a'
			}
			err := authService.ValidatePassword(string(longPassword))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "密码长度不能超过128位")
		})
	})
}

func TestUserPermissionsCalculation(t *testing.T) {
	Convey("用户权限计算测试", t, func() {
		authService := NewAuthService()

		Convey("租户管理员权限", func() {
			roles := []types.RoleType{types.RoleTenantAdmin}
			permissions := authService.calculateUserPermissions(roles)

			So(len(permissions), ShouldBeGreaterThan, 10) // 租户管理员有很多权限

			// 检查是否包含特定权限
			hasUserManage := false
			hasMerchantManage := false
			for _, perm := range permissions {
				if perm == types.PermissionUserManage {
					hasUserManage = true
				}
				if perm == types.PermissionMerchantManage {
					hasMerchantManage = true
				}
			}
			So(hasUserManage, ShouldBeTrue)
			So(hasMerchantManage, ShouldBeTrue)
		})

		Convey("商户权限", func() {
			roles := []types.RoleType{types.RoleMerchant}
			permissions := authService.calculateUserPermissions(roles)

			So(len(permissions), ShouldBeGreaterThan, 3)

			// 检查是否包含商品管理权限
			hasProductManage := false
			hasUserManage := false
			for _, perm := range permissions {
				if perm == types.PermissionProductManage {
					hasProductManage = true
				}
				if perm == types.PermissionUserManage {
					hasUserManage = true
				}
			}
			So(hasProductManage, ShouldBeTrue)
			So(hasUserManage, ShouldBeFalse) // 商户不应有用户管理权限
		})

		Convey("客户权限", func() {
			roles := []types.RoleType{types.RoleCustomer}
			permissions := authService.calculateUserPermissions(roles)

			So(len(permissions), ShouldBeGreaterThan, 0)

			// 检查基本权限
			hasProductView := false
			hasOrderCreate := false
			hasUserManage := false
			for _, perm := range permissions {
				if perm == types.PermissionProductView {
					hasProductView = true
				}
				if perm == types.PermissionOrderCreate {
					hasOrderCreate = true
				}
				if perm == types.PermissionUserManage {
					hasUserManage = true
				}
			}
			So(hasProductView, ShouldBeTrue)
			So(hasOrderCreate, ShouldBeTrue)
			So(hasUserManage, ShouldBeFalse) // 客户不应有管理权限
		})

		Convey("多角色权限合并", func() {
			roles := []types.RoleType{types.RoleMerchant, types.RoleCustomer}
			permissions := authService.calculateUserPermissions(roles)

			// 权限应该是两个角色的合集
			So(len(permissions), ShouldBeGreaterThan, 3)

			// 应该包含商户权限
			hasProductManage := false
			// 应该包含客户权限
			hasProductView := false
			for _, perm := range permissions {
				if perm == types.PermissionProductManage {
					hasProductManage = true
				}
				if perm == types.PermissionProductView {
					hasProductView = true
				}
			}
			So(hasProductManage, ShouldBeTrue)
			So(hasProductView, ShouldBeTrue)
		})
	})
}

func TestGenerateUserUUID(t *testing.T) {
	Convey("用户UUID生成测试", t, func() {
		authService := NewAuthService()

		Convey("UUID生成", func() {
			uuid1 := authService.generateUserUUID("user1", 1)
			uuid2 := authService.generateUserUUID("user2", 1)
			uuid3 := authService.generateUserUUID("user1", 2)

			So(uuid1, ShouldNotBeEmpty)
			So(uuid2, ShouldNotBeEmpty)
			So(uuid3, ShouldNotBeEmpty)

			// 不同用户或租户应该生成不同的UUID
			So(uuid1, ShouldNotEqual, uuid2)
			So(uuid1, ShouldNotEqual, uuid3)
			So(uuid2, ShouldNotEqual, uuid3)
		})
	})
}
