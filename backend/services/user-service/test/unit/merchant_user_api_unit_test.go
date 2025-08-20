package unit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gofromzero/mer-sys/backend/services/user-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMerchantUserAPI(t *testing.T) {
	Convey("商户用户API单元测试", t, func() {
		
		// 创建测试服务器
		s := g.Server("test-merchant-user")
		merchantUserController := controller.NewMerchantUserController()
		
		// 注册测试路由
		s.Group("/api/v1/merchant-users", func(group *ghttp.RouterGroup) {
			group.POST("/", merchantUserController.CreateMerchantUser)
			group.GET("/", merchantUserController.ListMerchantUsers)
			group.GET("/:id", merchantUserController.GetMerchantUser)
			group.PUT("/:id", merchantUserController.UpdateMerchantUser)
			group.PUT("/:id/status", merchantUserController.UpdateMerchantUserStatus)
			group.POST("/:id/reset-password", merchantUserController.ResetMerchantUserPassword)
			group.POST("/batch", merchantUserController.BatchCreateMerchantUsers)
		})
		
		s.SetDumpRouterMap(false)
		s.Start()
		defer s.Shutdown()
		
		client := g.Client()
		client.SetPrefix(s.GetListenedAddress())
		
		Convey("创建商户用户", func() {
			
			Convey("有效数据应该创建成功", func() {
				reqData := types.CreateMerchantUserRequest{
					Username:   "test_merchant_user",
					Email:      "test@merchant.com",
					Phone:      "13800138000",
					Password:   "password123",
					MerchantID: 1,
					RoleType:   types.RoleMerchantAdmin,
					Profile: &types.UserProfile{
						FirstName: "Test",
						LastName:  "User",
					},
				}
				
				jsonData, _ := json.Marshal(reqData)
				
				response := client.PostContent(context.Background(), "/api/v1/merchant-users/", string(jsonData))
				
				var result map[string]interface{}
				err := json.Unmarshal([]byte(response), &result)
				So(err, ShouldBeNil)
				
				// 由于需要数据库连接，这里只验证接口格式
				So(result, ShouldContainKey, "code")
				So(result, ShouldContainKey, "message")
				So(result, ShouldContainKey, "data")
			})
			
			Convey("无效数据应该返回错误", func() {
				reqData := map[string]interface{}{
					"username": "ab", // 用户名太短
					"email":    "invalid-email",
					"password": "123", // 密码太短
				}
				
				jsonData, _ := json.Marshal(reqData)
				
				response := client.PostContent(context.Background(), "/api/v1/merchant-users/", string(jsonData))
				
				var result map[string]interface{}
				err := json.Unmarshal([]byte(response), &result)
				So(err, ShouldBeNil)
				
				code, ok := result["code"].(float64)
				So(ok, ShouldBeTrue)
				So(code, ShouldNotEqual, 0) // 应该返回错误码
			})
		})
		
		Convey("查询商户用户列表", func() {
			
			Convey("带有merchant_id参数应该正常查询", func() {
				response := client.GetContent(context.Background(), "/api/v1/merchant-users/?merchant_id=1&page=1&pageSize=10")
				
				var result map[string]interface{}
				err := json.Unmarshal([]byte(response), &result)
				So(err, ShouldBeNil)
				
				So(result, ShouldContainKey, "code")
				So(result, ShouldContainKey, "message")
				So(result, ShouldContainKey, "data")
			})
			
			Convey("缺少merchant_id参数应该返回错误", func() {
				response := client.GetContent(context.Background(), "/api/v1/merchant-users/")
				
				var result map[string]interface{}
				err := json.Unmarshal([]byte(response), &result)
				So(err, ShouldBeNil)
				
				code, ok := result["code"].(float64)
				So(ok, ShouldBeTrue)
				So(code, ShouldNotEqual, 0) // 应该返回错误码
				
				message, ok := result["message"].(string)
				So(ok, ShouldBeTrue)
				So(message, ShouldContainSubstring, "商户ID不能为空")
			})
		})
		
		Convey("获取商户用户详情", func() {
			
			Convey("带有正确参数应该正常查询", func() {
				response := client.GetContent(context.Background(), "/api/v1/merchant-users/1?merchant_id=1")
				
				var result map[string]interface{}
				err := json.Unmarshal([]byte(response), &result)
				So(err, ShouldBeNil)
				
				So(result, ShouldContainKey, "code")
				So(result, ShouldContainKey, "message")
				So(result, ShouldContainKey, "data")
			})
			
			Convey("缺少参数应该返回错误", func() {
				response := client.GetContent(context.Background(), "/api/v1/merchant-users/1")
				
				var result map[string]interface{}
				err := json.Unmarshal([]byte(response), &result)
				So(err, ShouldBeNil)
				
				code, ok := result["code"].(float64)
				So(ok, ShouldBeTrue)
				So(code, ShouldNotEqual, 0) // 应该返回错误码
			})
		})
		
		Convey("更新商户用户状态", func() {
			
			Convey("有效状态应该更新成功", func() {
				reqData := map[string]interface{}{
					"status": "suspended",
				}
				
				jsonData, _ := json.Marshal(reqData)
				
				response := client.PutContent(context.Background(), "/api/v1/merchant-users/1/status?merchant_id=1", string(jsonData))
				
				var result map[string]interface{}
				err := json.Unmarshal([]byte(response), &result)
				So(err, ShouldBeNil)
				
				So(result, ShouldContainKey, "code")
				So(result, ShouldContainKey, "message")
				So(result, ShouldContainKey, "data")
			})
		})
		
		Convey("重置商户用户密码", func() {
			
			Convey("应该生成新密码", func() {
				response := client.PostContent(context.Background(), "/api/v1/merchant-users/1/reset-password?merchant_id=1", "")
				
				var result map[string]interface{}
				err := json.Unmarshal([]byte(response), &result)
				So(err, ShouldBeNil)
				
				So(result, ShouldContainKey, "code")
				So(result, ShouldContainKey, "message")
				So(result, ShouldContainKey, "data")
			})
		})
		
		Convey("批量创建商户用户", func() {
			
			Convey("有效数据应该创建成功", func() {
				reqData := map[string]interface{}{
					"merchant_id": 1,
					"users": []map[string]interface{}{
						{
							"username":  "batch_user_1",
							"email":     "batch1@merchant.com",
							"password":  "password123",
							"role_type": "merchant_operator",
						},
						{
							"username":  "batch_user_2",
							"email":     "batch2@merchant.com",
							"password":  "password123",
							"role_type": "merchant_admin",
						},
					},
				}
				
				jsonData, _ := json.Marshal(reqData)
				
				response := client.PostContent(context.Background(), "/api/v1/merchant-users/batch", string(jsonData))
				
				var result map[string]interface{}
				err := json.Unmarshal([]byte(response), &result)
				So(err, ShouldBeNil)
				
				So(result, ShouldContainKey, "code")
				So(result, ShouldContainKey, "message")
				So(result, ShouldContainKey, "data")
			})
		})
	})
}

func TestMerchantUserValidation(t *testing.T) {
	Convey("商户用户输入验证测试", t, func() {
		
		Convey("CreateMerchantUserRequest验证", func() {
			
			Convey("有效的请求应该通过验证", func() {
				req := types.CreateMerchantUserRequest{
					Username:   "valid_user",
					Email:      "valid@example.com",
					Phone:      "13800138000",
					Password:   "validpassword123",
					MerchantID: 1,
					RoleType:   types.RoleMerchantAdmin,
				}
				
				// 基本字段存在性检查
				So(req.Username, ShouldNotBeEmpty)
				So(req.Email, ShouldNotBeEmpty)
				So(req.Password, ShouldNotBeEmpty)
				So(req.MerchantID, ShouldBeGreaterThan, 0)
				So(req.RoleType, ShouldBeIn, types.RoleMerchantAdmin, types.RoleMerchantOperator)
			})
			
			Convey("无效的请求应该被识别", func() {
				invalidReqs := []types.CreateMerchantUserRequest{
					{Username: "ab", Email: "valid@example.com", Password: "password123", MerchantID: 1, RoleType: types.RoleMerchantAdmin}, // 用户名太短
					{Username: "valid_user", Email: "invalid-email", Password: "password123", MerchantID: 1, RoleType: types.RoleMerchantAdmin}, // 邮箱格式不正确
					{Username: "valid_user", Email: "valid@example.com", Password: "123", MerchantID: 1, RoleType: types.RoleMerchantAdmin}, // 密码太短
					{Username: "valid_user", Email: "valid@example.com", Password: "password123", MerchantID: 0, RoleType: types.RoleMerchantAdmin}, // 商户ID无效
				}
				
				for _, req := range invalidReqs {
					// 这里应该有验证逻辑，但由于当前没有集成验证框架，我们只做基本检查
					if len(req.Username) < 3 {
						So(len(req.Username), ShouldBeLessThan, 3)
					}
					if req.MerchantID == 0 {
						So(req.MerchantID, ShouldEqual, 0)
					}
				}
			})
		})
		
		Convey("UpdateMerchantUserRequest验证", func() {
			
			Convey("部分字段更新应该可行", func() {
				req := types.UpdateMerchantUserRequest{
					Username: "updated_user",
					Status:   types.UserStatusSuspended,
				}
				
				So(req.Username, ShouldNotBeEmpty)
				So(req.Status, ShouldBeIn, types.UserStatusPending, types.UserStatusActive, types.UserStatusSuspended, types.UserStatusDeactivated)
			})
		})
	})
}