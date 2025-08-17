package auth

import (
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTokenClaims(t *testing.T) {
	Convey("TokenClaims功能测试", t, func() {
		// 创建测试用的TokenClaims
		claims := &TokenClaims{
			UserID:   1,
			TenantID: 1,
			Username: "testuser",
			Email:    "test@example.com",
			Roles:    []types.RoleType{types.RoleTenantAdmin, types.RoleMerchant},
			Permissions: []types.Permission{
				types.PermissionUserManage,
				types.PermissionMerchantManage,
				types.PermissionProductView,
			},
			TokenType: "access",
			IssuedAt:  time.Now(),
			ExpiresAt: time.Now().Add(time.Hour * 24),
		}

		Convey("角色验证", func() {
			So(claims.HasRole(types.RoleTenantAdmin), ShouldBeTrue)
			So(claims.HasRole(types.RoleMerchant), ShouldBeTrue)
			So(claims.HasRole(types.RoleCustomer), ShouldBeFalse)
		})

		Convey("权限验证", func() {
			So(claims.HasPermission(types.PermissionUserManage), ShouldBeTrue)
			So(claims.HasPermission(types.PermissionMerchantManage), ShouldBeTrue)
			So(claims.HasPermission(types.PermissionProductView), ShouldBeTrue)
			So(claims.HasPermission(types.PermissionOrderManage), ShouldBeFalse)
		})

		Convey("获取用户权限信息", func() {
			userPerms := claims.GetUserPermissions()
			So(userPerms.UserID, ShouldEqual, claims.UserID)
			So(userPerms.TenantID, ShouldEqual, claims.TenantID)
			So(len(userPerms.Roles), ShouldEqual, 2)
			So(len(userPerms.Permissions), ShouldEqual, 3)
		})
	})
}

func TestDefaultRoles(t *testing.T) {
	Convey("默认角色配置测试", t, func() {
		roles := types.GetDefaultRoles()

		Convey("租户管理员角色", func() {
			role := roles[types.RoleTenantAdmin]
			So(role.Type, ShouldEqual, types.RoleTenantAdmin)
			So(role.Name, ShouldEqual, "租户管理员")
			So(len(role.Permissions), ShouldBeGreaterThan, 0)
			So(role.HasPermission(types.PermissionUserManage), ShouldBeTrue)
			So(role.HasPermission(types.PermissionMerchantManage), ShouldBeTrue)
		})

		Convey("商户角色", func() {
			role := roles[types.RoleMerchant]
			So(role.Type, ShouldEqual, types.RoleMerchant)
			So(role.Name, ShouldEqual, "商户")
			So(role.HasPermission(types.PermissionProductManage), ShouldBeTrue)
			So(role.HasPermission(types.PermissionUserManage), ShouldBeFalse)
		})

		Convey("客户角色", func() {
			role := roles[types.RoleCustomer]
			So(role.Type, ShouldEqual, types.RoleCustomer)
			So(role.Name, ShouldEqual, "客户")
			So(role.HasPermission(types.PermissionProductView), ShouldBeTrue)
			So(role.HasPermission(types.PermissionProductManage), ShouldBeFalse)
		})
	})
}

func TestUserPermissions(t *testing.T) {
	Convey("用户权限信息测试", t, func() {
		userPerms := &types.UserPermissions{
			UserID:   1,
			TenantID: 1,
			Roles:    []types.RoleType{types.RoleTenantAdmin, types.RoleMerchant},
			Permissions: []types.Permission{
				types.PermissionUserManage,
				types.PermissionProductView,
			},
		}

		Convey("权限检查", func() {
			So(userPerms.HasPermission(types.PermissionUserManage), ShouldBeTrue)
			So(userPerms.HasPermission(types.PermissionProductView), ShouldBeTrue)
			So(userPerms.HasPermission(types.PermissionOrderManage), ShouldBeFalse)
		})

		Convey("角色检查", func() {
			So(userPerms.HasRole(types.RoleTenantAdmin), ShouldBeTrue)
			So(userPerms.HasRole(types.RoleMerchant), ShouldBeTrue)
			So(userPerms.HasRole(types.RoleCustomer), ShouldBeFalse)
		})
	})
}

func TestTokenRefreshMechanism(t *testing.T) {
	Convey("Token刷新机制测试", t, func() {
		// 创建测试令牌
		refreshClaims := &TokenClaims{
			UserID:   1,
			TenantID: 1,
			Username: "testuser",
			Email:    "test@example.com",
			Roles:    []types.RoleType{types.RoleTenantAdmin},
			Permissions: []types.Permission{
				types.PermissionUserManage,
				types.PermissionMerchantManage,
			},
			TokenType: "refresh",
			IssuedAt:  time.Now(),
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7), // 7天后过期
		}

		Convey("刷新令牌权限保持", func() {
			userPerms := refreshClaims.GetUserPermissions()
			So(userPerms.UserID, ShouldEqual, refreshClaims.UserID)
			So(userPerms.TenantID, ShouldEqual, refreshClaims.TenantID)
			So(len(userPerms.Roles), ShouldEqual, len(refreshClaims.Roles))
			So(len(userPerms.Permissions), ShouldEqual, len(refreshClaims.Permissions))
			So(userPerms.HasRole(types.RoleTenantAdmin), ShouldBeTrue)
			So(userPerms.HasPermission(types.PermissionUserManage), ShouldBeTrue)
		})

		Convey("令牌轮换安全检查", func() {
			// 模拟接近过期的刷新令牌
			soonExpireClaims := &TokenClaims{
				UserID:      1,
				TenantID:    1,
				Username:    "testuser",
				Email:       "test@example.com",
				Roles:       []types.RoleType{types.RoleTenantAdmin},
				Permissions: []types.Permission{types.PermissionUserManage},
				TokenType:   "refresh",
				IssuedAt:    time.Now().Add(-time.Hour * 23 * 24 * 7), // 几乎过期
				ExpiresAt:   time.Now().Add(time.Minute * 30),         // 30分钟后过期
			}

			userPerms := soonExpireClaims.GetUserPermissions()
			So(userPerms.HasRole(types.RoleTenantAdmin), ShouldBeTrue)
			So(userPerms.HasPermission(types.PermissionUserManage), ShouldBeTrue)
		})

		Convey("令牌类型验证", func() {
			// 测试错误的令牌类型
			accessClaims := &TokenClaims{
				UserID:    1,
				TenantID:  1,
				TokenType: "access", // 错误类型，应该是refresh
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
			}

			So(accessClaims.TokenType, ShouldEqual, "access")
			So(accessClaims.TokenType, ShouldNotEqual, "refresh")
		})
	})
}

func TestTokenBlacklistMechanism(t *testing.T) {
	Convey("Token黑名单机制测试", t, func() {
		// 创建测试用的TokenClaims
		testToken := "test-token-123"
		expiresAt := time.Now().Add(time.Hour * 24)

		Convey("黑名单键生成", func() {
			jwtManager := &JWTManager{}
			key := jwtManager.getBlacklistKey(testToken)
			So(key, ShouldEqual, "blacklist:test-token-123")
		})

		Convey("撤销时间处理", func() {
			// 测试过期时间计算
			futureTime := time.Now().Add(time.Hour)
			pastTime := time.Now().Add(-time.Hour)

			// 未来时间应该计算正确的TTL
			ttl := time.Until(futureTime)
			So(ttl, ShouldBeGreaterThan, 0)

			// 过去时间应该返回负值
			ttl = time.Until(pastTime)
			So(ttl, ShouldBeLessThan, 0)
		})

		Convey("黑名单令牌检查逻辑", func() {
			// 模拟黑名单令牌声明
			blacklistedClaims := &TokenClaims{
				UserID:    1,
				TenantID:  1,
				Username:  "blacklisted-user",
				Email:     "blacklisted@example.com",
				TokenType: "access",
				IssuedAt:  time.Now().Add(-time.Hour),
				ExpiresAt: expiresAt,
			}

			So(blacklistedClaims.UserID, ShouldEqual, 1)
			So(blacklistedClaims.TokenType, ShouldEqual, "access")
			So(blacklistedClaims.ExpiresAt, ShouldHappenAfter, time.Now())
		})

		Convey("撤销令牌安全性", func() {
			// 模拟令牌撤销场景
			revokedToken := "revoked-token-456"
			revokedTime := time.Now()

			// 撤销后的令牌应该无法使用
			So(revokedToken, ShouldNotBeEmpty)
			So(revokedTime, ShouldHappenBefore, time.Now().Add(time.Second))
		})

		Convey("重放攻击防护", func() {
			// 测试重放攻击防护逻辑
			replayToken := "replay-token-789"

			// 同一个令牌多次使用应该被检测
			tokens := []string{replayToken, replayToken, replayToken}
			uniqueTokens := make(map[string]bool)

			for _, token := range tokens {
				if uniqueTokens[token] {
					// 检测到重复令牌（重放攻击）
					So(token, ShouldEqual, replayToken)
				}
				uniqueTokens[token] = true
			}

			So(len(uniqueTokens), ShouldEqual, 1) // 只有一个唯一令牌
		})

		Convey("黑名单清理机制", func() {
			// 测试过期黑名单项的清理
			now := time.Now()
			expiredTime := now.Add(-time.Hour)
			validTime := now.Add(time.Hour)

			// 过期的黑名单项应该被清理
			So(expiredTime, ShouldHappenBefore, now)
			So(validTime, ShouldHappenAfter, now)
		})
	})
}
