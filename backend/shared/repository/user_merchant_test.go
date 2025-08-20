package repository

import (
	"context"
	"testing"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMerchantUserRepository(t *testing.T) {
	Convey("商户用户Repository测试", t, func() {
		
		Convey("商户用户数据验证", func() {
			
			Convey("ValidateMerchantUser应该正确验证商户用户", func() {
				merchantID := uint64(123)
				
				// 有效的商户用户
				validUser := &types.User{
					Username:   "test_merchant_user",
					Email:      "test@merchant.com",
					Phone:      "13800138000",
					MerchantID: &merchantID,
				}
				
				err := validUser.ValidateMerchantUser()
				So(err, ShouldBeNil)
				
				// 无效的商户用户 - 缺少merchant_id
				invalidUser1 := &types.User{
					Username: "test_user",
					Email:    "test@example.com",
				}
				
				err = invalidUser1.ValidateMerchantUser()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "商户用户必须关联商户")
				
				// 无效的商户用户 - 用户名太短
				invalidUser2 := &types.User{
					Username:   "ab",
					Email:      "test@example.com",
					MerchantID: &merchantID,
				}
				
				err = invalidUser2.ValidateMerchantUser()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "用户名长度必须在3-50字符之间")
				
				// 无效的商户用户 - 邮箱为空
				invalidUser3 := &types.User{
					Username:   "testuser",
					Email:      "",
					MerchantID: &merchantID,
				}
				
				err = invalidUser3.ValidateMerchantUser()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "邮箱不能为空")
			})
		})
		
		Convey("商户用户辅助方法", func() {
			merchantID := uint64(123)
			
			Convey("IsMerchantUser应该正确判断", func() {
				merchantUser := &types.User{MerchantID: &merchantID}
				regularUser := &types.User{MerchantID: nil}
				
				So(merchantUser.IsMerchantUser(), ShouldBeTrue)
				So(regularUser.IsMerchantUser(), ShouldBeFalse)
			})
			
			Convey("HasMerchantRole应该正确检查商户角色", func() {
				merchantUser := &types.User{MerchantID: &merchantID}
				regularUser := &types.User{MerchantID: nil}
				
				So(merchantUser.HasMerchantRole(types.RoleMerchantAdmin), ShouldBeTrue)
				So(merchantUser.HasMerchantRole(types.RoleMerchantOperator), ShouldBeTrue)
				So(merchantUser.HasMerchantRole(types.RoleTenantAdmin), ShouldBeFalse)
				
				So(regularUser.HasMerchantRole(types.RoleMerchantAdmin), ShouldBeFalse)
			})
		})
		
		Convey("创建商户用户请求验证", func() {
			
			Convey("有效的CreateMerchantUserRequest", func() {
				req := types.CreateMerchantUserRequest{
					Username:   "merchant_user",
					Email:      "user@merchant.com",
					Phone:      "13800138000",
					Password:   "securepassword123",
					MerchantID: 1,
					RoleType:   types.RoleMerchantAdmin,
				}
				
				So(req.Username, ShouldNotBeEmpty)
				So(req.Email, ShouldNotBeEmpty)
				So(req.Password, ShouldNotBeEmpty)
				So(req.MerchantID, ShouldBeGreaterThan, 0)
				So(req.RoleType, ShouldBeIn, types.RoleMerchantAdmin, types.RoleMerchantOperator)
			})
			
			Convey("有效的UpdateMerchantUserRequest", func() {
				req := types.UpdateMerchantUserRequest{
					Username: "updated_user",
					Email:    "updated@merchant.com",
					Status:   types.UserStatusSuspended,
				}
				
				So(req.Username, ShouldNotBeEmpty)
				So(req.Email, ShouldNotBeEmpty)
				So(req.Status, ShouldBeIn, types.UserStatusPending, types.UserStatusActive, types.UserStatusSuspended, types.UserStatusDeactivated)
			})
		})
	})
}

func TestMerchantUserRepositoryMethods(t *testing.T) {
	Convey("商户用户Repository方法结构测试", t, func() {
		
		Convey("UserRepository类型应该有商户用户相关方法", func() {
			// 直接测试类型方法存在性，不需要实例化
			repo := &UserRepository{}
			
			// 验证方法存在性（通过类型断言）
			var _ func(*UserRepository, context.Context, *types.User, types.RoleType) error = (*UserRepository).CreateMerchantUser
			var _ func(*UserRepository, context.Context, uint64, int, int, string) ([]*types.User, int, error) = (*UserRepository).FindMerchantUsers
			var _ func(*UserRepository, context.Context, uint64, uint64) (*types.User, error) = (*UserRepository).FindMerchantUserByID
			var _ func(*UserRepository, context.Context, uint64, uint64, types.UserStatus) error = (*UserRepository).UpdateMerchantUserStatus
			var _ func(*UserRepository, context.Context, uint64, uint64, interface{}) error = (*UserRepository).UpdateMerchantUser
			var _ func(*UserRepository, context.Context, uint64, uint64, string) error = (*UserRepository).ResetMerchantUserPassword
			var _ func(*UserRepository, context.Context, string, interface{}, uint64) (bool, error) = (*UserRepository).MerchantUserExists
			var _ func(*UserRepository, context.Context, uint64, uint64) ([]types.RoleType, error) = (*UserRepository).GetMerchantUserRoles
			var _ func(*UserRepository, context.Context, uint64, uint64, types.RoleType) error = (*UserRepository).AssignMerchantUserRole
			
			So(repo, ShouldNotBeNil) // 确保类型存在
		})
	})
}

// 模拟测试数据的辅助函数
func createTestMerchantUser() *types.User {
	merchantID := uint64(1)
	return &types.User{
		ID:         1,
		UUID:       "test-uuid-12345",
		Username:   "test_merchant_user",
		Email:      "test@merchant.com",
		Phone:      "13800138000",
		TenantID:   1,
		MerchantID: &merchantID,
		Status:     types.UserStatusActive,
	}
}

func createTestCreateMerchantUserRequest() types.CreateMerchantUserRequest {
	return types.CreateMerchantUserRequest{
		Username:   "new_merchant_user",
		Email:      "new@merchant.com",
		Phone:      "13900139000",
		Password:   "password123",
		MerchantID: 1,
		RoleType:   types.RoleMerchantAdmin,
		Profile: &types.UserProfile{
			FirstName: "New",
			LastName:  "User",
		},
	}
}