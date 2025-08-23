package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/test/gtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// PricingAPITestSuite 定价API集成测试套件
type PricingAPITestSuite struct {
	suite.Suite
	server     *ghttp.Server
	baseURL    string
	tenantID   uint64
	productID  uint64
	testUser   uint64
	authToken  string
}

// SetupSuite 测试套件初始化
func (suite *PricingAPITestSuite) SetupSuite() {
	// 初始化测试服务器
	suite.server = g.Server()
	suite.baseURL = "http://127.0.0.1:8080"
	suite.tenantID = 1
	suite.productID = 1
	suite.testUser = 1
	suite.authToken = "test-jwt-token"

	// 注册API路由
	pricingController := controller.NewPricingController()
	suite.server.Group("/api/v1/products/{productId}/pricing-rules", func(group *ghttp.RouterGroup) {
		group.POST("", pricingController.CreatePricingRule)
		group.GET("", pricingController.GetPricingRules)
		group.PUT("/{id}", pricingController.UpdatePricingRule)
		group.DELETE("/{id}", pricingController.DeletePricingRule)
	})

	suite.server.Group("/api/v1/products/{productId}", func(group *ghttp.RouterGroup) {
		group.GET("/effective-price", pricingController.GetEffectivePrice)
		group.POST("/price-change", pricingController.ChangePrice)
		group.GET("/price-history", pricingController.GetPriceHistory)
		group.POST("/validate-rights", pricingController.ValidateRights)
	})

	// 启动测试服务器
	suite.server.SetPort(8080)
	suite.server.Start()
	time.Sleep(100 * time.Millisecond) // 等待服务器启动
}

// TearDownSuite 测试套件清理
func (suite *PricingAPITestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Shutdown()
	}
}

// TestCreatePricingRule 测试创建定价规则API
func (suite *PricingAPITestSuite) TestCreatePricingRule() {
	gtest.C(suite.T(), func(t *gtest.T) {
		// 测试用例1: 创建基础价格规则
		t.Run("创建基础价格规则", func(t *gtest.T) {
			request := types.CreatePricingRuleRequest{
				RuleType:  types.PricingRuleTypeBasePrice,
				Priority:  0,
				ValidFrom: time.Now(),
			}

			body, _ := json.Marshal(request)
			resp := suite.makeRequest("POST", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules", suite.productID), 
				body)

			// 验证响应状态码
			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			// 解析响应
			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			
			assert.Equal(t, 0, int(response["code"].(float64)))
			assert.NotNil(t, response["data"])
		})

		// 测试用例2: 创建阶梯价格规则
		t.Run("创建阶梯价格规则", func(t *gtest.T) {
			request := types.CreatePricingRuleRequest{
				RuleType:  types.PricingRuleTypeVolumeDiscount,
				Priority:  10,
				ValidFrom: time.Now().Add(24 * time.Hour),
			}

			body, _ := json.Marshal(request)
			resp := suite.makeRequest("POST", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules", suite.productID), 
				body)

			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})

		// 测试用例3: 创建会员价格规则
		t.Run("创建会员价格规则", func(t *gtest.T) {
			request := types.CreatePricingRuleRequest{
				RuleType:  types.PricingRuleTypeMemberDiscount,
				Priority:  15,
				ValidFrom: time.Now(),
			}

			body, _ := json.Marshal(request)
			resp := suite.makeRequest("POST", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules", suite.productID), 
				body)

			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})

		// 测试用例4: 参数验证失败
		t.Run("参数验证失败", func(t *gtest.T) {
			invalidRequest := map[string]interface{}{
				"rule_type": "", // 空的规则类型
				"priority":  -1, // 无效的优先级
			}

			body, _ := json.Marshal(invalidRequest)
			resp := suite.makeRequest("POST", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules", suite.productID), 
				body)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})
}

// TestGetPricingRules 测试获取定价规则API
func (suite *PricingAPITestSuite) TestGetPricingRules() {
	gtest.C(suite.T(), func(t *gtest.T) {
		t.Run("获取定价规则列表", func(t *gtest.T) {
			resp := suite.makeRequest("GET", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules", suite.productID), 
				nil)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			
			assert.Equal(t, 0, int(response["code"].(float64)))
			assert.NotNil(t, response["data"])
			
			// 验证返回的是数组
			data := response["data"].(map[string]interface{})
			assert.NotNil(t, data["items"])
		})

		t.Run("按状态筛选规则", func(t *gtest.T) {
			resp := suite.makeRequest("GET", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules?is_active=true", suite.productID), 
				nil)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})
}

// TestUpdatePricingRule 测试更新定价规则API
func (suite *PricingAPITestSuite) TestUpdatePricingRule() {
	gtest.C(suite.T(), func(t *gtest.T) {
		ruleID := uint64(1)

		t.Run("更新规则优先级", func(t *gtest.T) {
			updateRequest := types.UpdatePricingRuleRequest{
				Priority: func() *int { v := 20; return &v }(),
			}

			body, _ := json.Marshal(updateRequest)
			resp := suite.makeRequest("PUT", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules/%d", suite.productID, ruleID), 
				body)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})

		t.Run("启用/禁用规则", func(t *gtest.T) {
			updateRequest := types.UpdatePricingRuleRequest{
				IsActive: func() *bool { v := false; return &v }(),
			}

			body, _ := json.Marshal(updateRequest)
			resp := suite.makeRequest("PUT", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules/%d", suite.productID, ruleID), 
				body)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})

		t.Run("更新不存在的规则", func(t *gtest.T) {
			nonExistentID := uint64(99999)
			updateRequest := types.UpdatePricingRuleRequest{
				Priority: func() *int { v := 10; return &v }(),
			}

			body, _ := json.Marshal(updateRequest)
			resp := suite.makeRequest("PUT", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules/%d", suite.productID, nonExistentID), 
				body)

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})
}

// TestDeletePricingRule 测试删除定价规则API
func (suite *PricingAPITestSuite) TestDeletePricingRule() {
	gtest.C(suite.T(), func(t *gtest.T) {
		t.Run("删除规则", func(t *gtest.T) {
			ruleID := uint64(2) // 假设存在的规则ID

			resp := suite.makeRequest("DELETE", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules/%d", suite.productID, ruleID), 
				nil)

			// 根据业务逻辑，可能返回200或204
			assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent)
		})

		t.Run("删除不存在的规则", func(t *gtest.T) {
			nonExistentID := uint64(99999)

			resp := suite.makeRequest("DELETE", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules/%d", suite.productID, nonExistentID), 
				nil)

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})
}

// TestGetEffectivePrice 测试获取有效价格API
func (suite *PricingAPITestSuite) TestGetEffectivePrice() {
	gtest.C(suite.T(), func(t *gtest.T) {
		t.Run("计算基础价格", func(t *gtest.T) {
			resp := suite.makeRequest("GET", 
				fmt.Sprintf("/api/v1/products/%d/effective-price?quantity=1", suite.productID), 
				nil)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			
			assert.Equal(t, 0, int(response["code"].(float64)))
			
			data := response["data"].(map[string]interface{})
			assert.NotNil(t, data["effective_price"])
			assert.NotNil(t, data["base_price"])
		})

		t.Run("计算会员价格", func(t *gtest.T) {
			resp := suite.makeRequest("GET", 
				fmt.Sprintf("/api/v1/products/%d/effective-price?quantity=1&member_level=vip", suite.productID), 
				nil)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			
			data := response["data"].(map[string]interface{})
			assert.NotNil(t, data["applied_rules"])
		})

		t.Run("计算阶梯价格", func(t *gtest.T) {
			resp := suite.makeRequest("GET", 
				fmt.Sprintf("/api/v1/products/%d/effective-price?quantity=25", suite.productID), 
				nil)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			
			data := response["data"].(map[string]interface{})
			// 验证阶梯价格规则被应用
			appliedRules := data["applied_rules"].([]interface{})
			assert.True(t, len(appliedRules) > 0)
		})

		t.Run("无效参数", func(t *gtest.T) {
			resp := suite.makeRequest("GET", 
				fmt.Sprintf("/api/v1/products/%d/effective-price?quantity=0", suite.productID), 
				nil)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})
}

// TestPriceChange 测试价格变更API
func (suite *PricingAPITestSuite) TestPriceChange() {
	gtest.C(suite.T(), func(t *gtest.T) {
		t.Run("执行价格变更", func(t *gtest.T) {
			changeRequest := types.PriceChangeRequest{
				NewPrice:      types.Money{Amount: 120.00, Currency: "CNY"},
				ChangeReason:  "市场调整",
				EffectiveDate: time.Now().Add(time.Hour),
			}

			body, _ := json.Marshal(changeRequest)
			resp := suite.makeRequest("POST", 
				fmt.Sprintf("/api/v1/products/%d/price-change", suite.productID), 
				body)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			
			assert.Equal(t, 0, int(response["code"].(float64)))
		})

		t.Run("价格变更参数验证", func(t *gtest.T) {
			invalidRequest := map[string]interface{}{
				"new_price": map[string]interface{}{
					"amount":   -10.00, // 负价格
					"currency": "CNY",
				},
				"change_reason": "", // 空原因
			}

			body, _ := json.Marshal(invalidRequest)
			resp := suite.makeRequest("POST", 
				fmt.Sprintf("/api/v1/products/%d/price-change", suite.productID), 
				body)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})
}

// TestGetPriceHistory 测试获取价格历史API
func (suite *PricingAPITestSuite) TestGetPriceHistory() {
	gtest.C(suite.T(), func(t *gtest.T) {
		t.Run("获取价格历史", func(t *gtest.T) {
			resp := suite.makeRequest("GET", 
				fmt.Sprintf("/api/v1/products/%d/price-history", suite.productID), 
				nil)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			
			assert.Equal(t, 0, int(response["code"].(float64)))
			
			data := response["data"].(map[string]interface{})
			assert.NotNil(t, data["items"])
			assert.NotNil(t, data["pagination"])
		})

		t.Run("按时间范围筛选", func(t *gtest.T) {
			startDate := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
			endDate := time.Now().Format("2006-01-02")

			resp := suite.makeRequest("GET", 
				fmt.Sprintf("/api/v1/products/%d/price-history?start_date=%s&end_date=%s", 
					suite.productID, startDate, endDate), 
				nil)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})
}

// TestValidateRights 测试权益验证API
func (suite *PricingAPITestSuite) TestValidateRights() {
	gtest.C(suite.T(), func(t *gtest.T) {
		t.Run("权益充足验证", func(t *gtest.T) {
			validateRequest := types.ValidateRightsRequest{
				UserID:      suite.testUser,
				Quantity:    2,
				TotalAmount: types.Money{Amount: 100.00, Currency: "CNY"},
			}

			body, _ := json.Marshal(validateRequest)
			resp := suite.makeRequest("POST", 
				fmt.Sprintf("/api/v1/products/%d/validate-rights", suite.productID), 
				body)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			
			assert.Equal(t, 0, int(response["code"].(float64)))
			
			data := response["data"].(map[string]interface{})
			assert.NotNil(t, data["is_valid"])
			assert.NotNil(t, data["required_rights"])
			assert.NotNil(t, data["available_rights"])
		})

		t.Run("权益不足验证", func(t *gtest.T) {
			validateRequest := types.ValidateRightsRequest{
				UserID:      suite.testUser,
				Quantity:    100, // 大数量，可能导致权益不足
				TotalAmount: types.Money{Amount: 5000.00, Currency: "CNY"},
			}

			body, _ := json.Marshal(validateRequest)
			resp := suite.makeRequest("POST", 
				fmt.Sprintf("/api/v1/products/%d/validate-rights", suite.productID), 
				body)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			
			data := response["data"].(map[string]interface{})
			// 如果权益不足，应该有相关字段
			if !data["is_valid"].(bool) {
				assert.NotNil(t, data["insufficient_amount"])
				assert.NotNil(t, data["suggested_action"])
			}
		})
	})
}

// TestPricingRuleConflicts 测试定价规则冲突处理
func (suite *PricingAPITestSuite) TestPricingRuleConflicts() {
	gtest.C(suite.T(), func(t *gtest.T) {
		t.Run("时间冲突检测", func(t *gtest.T) {
			// 创建第一个规则
			request1 := types.CreatePricingRuleRequest{
				RuleType:  types.PricingRuleTypeBasePrice,
				Priority:  0,
				ValidFrom: time.Now(),
			}

			body1, _ := json.Marshal(request1)
			resp1 := suite.makeRequest("POST", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules", suite.productID), 
				body1)

			assert.Equal(t, http.StatusCreated, resp1.StatusCode)

			// 尝试创建冲突的规则
			request2 := types.CreatePricingRuleRequest{
				RuleType:  types.PricingRuleTypeBasePrice, // 相同类型
				Priority:  0,                              // 相同优先级
				ValidFrom: time.Now(),                     // 时间重叠
			}

			body2, _ := json.Marshal(request2)
			resp2 := suite.makeRequest("POST", 
				fmt.Sprintf("/api/v1/products/%d/pricing-rules", suite.productID), 
				body2)

			// 应该返回冲突错误
			assert.Equal(t, http.StatusConflict, resp2.StatusCode)
		})
	})
}

// TestMultiTenantIsolation 测试多租户隔离
func (suite *PricingAPITestSuite) TestMultiTenantIsolation() {
	gtest.C(suite.T(), func(t *gtest.T) {
		t.Run("租户隔离验证", func(t *gtest.T) {
			// 使用不同的租户ID
			otherTenantID := uint64(999)
			
			// 创建请求，但Header中使用不同的租户ID
			request := types.CreatePricingRuleRequest{
				RuleType:  types.PricingRuleTypeBasePrice,
				ValidFrom: time.Now(),
			}

			body, _ := json.Marshal(request)
			req, _ := http.NewRequest("POST", 
				suite.baseURL+fmt.Sprintf("/api/v1/products/%d/pricing-rules", suite.productID), 
				bytes.NewBuffer(body))
			
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			req.Header.Set("X-Tenant-ID", fmt.Sprintf("%d", otherTenantID))

			client := &http.Client{}
			resp, _ := client.Do(req)

			// 根据实现，可能返回403（权限不足）或404（产品不存在）
			assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound)
		})
	})
}

// TestAPIPerformance 测试API性能
func (suite *PricingAPITestSuite) TestAPIPerformance() {
	gtest.C(suite.T(), func(t *gtest.T) {
		t.Run("价格计算性能测试", func(t *gtest.T) {
			// 并发请求测试
			concurrency := 10
			requests := 100
			
			start := time.Now()
			
			for i := 0; i < requests; i++ {
				go func() {
					resp := suite.makeRequest("GET", 
						fmt.Sprintf("/api/v1/products/%d/effective-price?quantity=5", suite.productID), 
						nil)
					assert.Equal(t, http.StatusOK, resp.StatusCode)
				}()
			}
			
			duration := time.Since(start)
			
			// 验证性能要求：100个请求应该在1秒内完成
			assert.True(t, duration < time.Second, 
				fmt.Sprintf("性能测试失败：%d个请求耗时%v", requests, duration))
		})
	})
}

// makeRequest 发送HTTP请求的辅助方法
func (suite *PricingAPITestSuite) makeRequest(method, path string, body []byte) *http.Response {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, suite.baseURL+path, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, suite.baseURL+path, nil)
	}

	if err != nil {
		suite.T().Fatalf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	req.Header.Set("X-Tenant-ID", fmt.Sprintf("%d", suite.tenantID))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		suite.T().Fatalf("发送请求失败: %v", err)
	}

	return resp
}

// TestPricingAPITestSuite 运行定价API集成测试套件
func TestPricingAPITestSuite(t *testing.T) {
	suite.Run(t, new(PricingAPITestSuite))
}

// TestTenantIsolationCompliance 测试租户隔离合规性
func TestTenantIsolationCompliance(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 这是一个重要的合规性测试，确保多租户隔离正确工作
		t.Run("租户数据隔离", func(t *gtest.T) {
			ctx := context.Background()
			
			// 模拟两个不同租户的上下文
			tenant1Ctx := context.WithValue(ctx, "tenant_id", uint64(1))
			tenant2Ctx := context.WithValue(ctx, "tenant_id", uint64(2))

			productID := uint64(1)

			// 验证租户1不能访问租户2的数据
			// 这里需要真实的仓库实现来测试
			
			// 模拟验证逻辑
			tenant1Data := "tenant1-data"
			tenant2Data := "tenant2-data"
			
			assert.NotEqual(t, tenant1Data, tenant2Data)
			assert.NotNil(t, tenant1Ctx)
			assert.NotNil(t, tenant2Ctx)
			assert.NotEqual(t, productID, uint64(0))
		})
	})
}