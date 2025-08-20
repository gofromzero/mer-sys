package types

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUserMerchantExtensions(t *testing.T) {
	Convey("商户用户扩展功能测试", t, func() {
		
		Convey("商户用户数据验证", func() {
			merchantID := uint64(123)
			
			Convey("有效的商户用户应该通过验证", func() {
				user := &User{
					ID:           1,
					UUID:         "test-uuid",
					Username:     "merchant_user",
					Email:        "test@example.com",
					Phone:        "1234567890",
					TenantID:     1,
					MerchantID:   &merchantID,
					Status:       UserStatusActive,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				
				err := user.ValidateMerchantUser()
				So(err, ShouldBeNil)
			})
			
			Convey("没有merchant_id的用户应该验证失败", func() {
				user := &User{
					Username: "test_user",
					Email:    "test@example.com",
					TenantID: 1,
				}
				
				err := user.ValidateMerchantUser()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "商户用户必须关联商户")
			})
			
			Convey("用户名长度不合法应该验证失败", func() {
				user := &User{
					Username:   "ab", // 太短
					Email:      "test@example.com",
					MerchantID: &merchantID,
				}
				
				err := user.ValidateMerchantUser()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "用户名长度必须在3-50字符之间")
			})
			
			Convey("邮箱为空应该验证失败", func() {
				user := &User{
					Username:   "testuser",
					Email:      "",
					MerchantID: &merchantID,
				}
				
				err := user.ValidateMerchantUser()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "邮箱不能为空")
			})
		})
		
		Convey("商户用户角色检查", func() {
			merchantID := uint64(123)
			
			Convey("IsMerchantUser应该正确判断商户用户", func() {
				merchantUser := &User{MerchantID: &merchantID}
				regularUser := &User{MerchantID: nil}
				
				So(merchantUser.IsMerchantUser(), ShouldBeTrue)
				So(regularUser.IsMerchantUser(), ShouldBeFalse)
			})
			
			Convey("HasMerchantRole应该正确检查商户角色", func() {
				merchantUser := &User{MerchantID: &merchantID}
				regularUser := &User{MerchantID: nil}
				
				So(merchantUser.HasMerchantRole(RoleMerchantAdmin), ShouldBeTrue)
				So(merchantUser.HasMerchantRole(RoleMerchantOperator), ShouldBeTrue)
				So(merchantUser.HasMerchantRole(RoleTenantAdmin), ShouldBeFalse)
				
				So(regularUser.HasMerchantRole(RoleMerchantAdmin), ShouldBeFalse)
			})
		})
	})
}

func TestMerchantRoles(t *testing.T) {
	Convey("商户角色权限测试", t, func() {
		
		Convey("商户管理员角色应该拥有完整权限", func() {
			roles := GetDefaultRoles()
			merchantAdminRole := roles[RoleMerchantAdmin]
			
			So(merchantAdminRole.Name, ShouldEqual, "商户管理员")
			So(merchantAdminRole.HasPermission(PermissionMerchantProductView), ShouldBeTrue)
			So(merchantAdminRole.HasPermission(PermissionMerchantProductCreate), ShouldBeTrue)
			So(merchantAdminRole.HasPermission(PermissionMerchantProductEdit), ShouldBeTrue)
			So(merchantAdminRole.HasPermission(PermissionMerchantProductDelete), ShouldBeTrue)
			So(merchantAdminRole.HasPermission(PermissionMerchantUserManage), ShouldBeTrue)
		})
		
		Convey("商户操作员角色应该有限制的权限", func() {
			roles := GetDefaultRoles()
			merchantOperatorRole := roles[RoleMerchantOperator]
			
			So(merchantOperatorRole.Name, ShouldEqual, "商户操作员")
			So(merchantOperatorRole.HasPermission(PermissionMerchantProductView), ShouldBeTrue)
			So(merchantOperatorRole.HasPermission(PermissionMerchantProductCreate), ShouldBeTrue)
			So(merchantOperatorRole.HasPermission(PermissionMerchantProductEdit), ShouldBeTrue)
			So(merchantOperatorRole.HasPermission(PermissionMerchantProductDelete), ShouldBeFalse) // 没有删除权限
			So(merchantOperatorRole.HasPermission(PermissionMerchantUserManage), ShouldBeFalse)    // 没有用户管理权限
		})
	})
}

func TestUserPermissionsWithMerchant(t *testing.T) {
	Convey("包含商户信息的用户权限测试", t, func() {
		merchantID := uint64(456)
		userPermissions := UserPermissions{
			UserID:      1,
			TenantID:    1,
			MerchantID:  &merchantID,
			Roles:       []RoleType{RoleMerchantAdmin},
			Permissions: []Permission{PermissionMerchantUserManage, PermissionMerchantProductView},
		}
		
		Convey("应该正确检查权限", func() {
			So(userPermissions.HasPermission(PermissionMerchantUserManage), ShouldBeTrue)
			So(userPermissions.HasPermission(PermissionMerchantProductView), ShouldBeTrue)
			So(userPermissions.HasPermission(PermissionTenantManage), ShouldBeFalse)
		})
		
		Convey("应该正确检查角色", func() {
			So(userPermissions.HasRole(RoleMerchantAdmin), ShouldBeTrue)
			So(userPermissions.HasRole(RoleMerchantOperator), ShouldBeFalse)
			So(userPermissions.HasRole(RoleTenantAdmin), ShouldBeFalse)
		})
	})
}