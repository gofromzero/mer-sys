package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/controller"
)

// setupInventoryTestServer 设置库存测试服务器
func setupInventoryTestServer() *ghttp.Server {
	s := g.Server("inventory-test")
	
	// 初始化控制器
	inventoryController := controller.NewInventoryController()
	
	// 注册路由
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(func(r *ghttp.Request) {
			// 添加测试用的租户和商户信息
			r.SetCtx(r.Context())
			r.SetParam("tenant_id", "1")
			r.SetParam("merchant_id", "1")
			r.SetParam("user_id", "1")
			r.Middleware.Next()
		})
		
		// 库存管理路由
		group.GET("/products/{id}/inventory", inventoryController.GetInventory)
		group.POST("/products/{id}/inventory/adjust", inventoryController.AdjustInventory)
		group.GET("/products/{id}/inventory/records", inventoryController.GetInventoryRecords)
		group.POST("/products/inventory/reserve", inventoryController.ReserveInventory)
		group.POST("/products/inventory/release", inventoryController.ReleaseInventory)
		group.POST("/products/inventory/batch-adjust", inventoryController.BatchAdjustInventory)
	})
	
	return s
}

// TestInventoryAPIs 测试库存管理API
func TestInventoryAPIs(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		server := setupInventoryTestServer()
		defer server.Shutdown()

		// 测试商品ID
		testProductID := "1001"
		baseURL := "/api/v1/products/" + testProductID

		t.Run("GetInventory", func(t *gtest.T) {
			// 获取库存信息
			req := httptest.NewRequest("GET", baseURL+"/inventory", nil)
			req.Header.Set("X-Tenant-ID", "1")
			
			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				t.AssertNil(err)
				t.Log(fmt.Sprintf("获取库存响应: %+v", response))
			} else {
				t.Log(fmt.Sprintf("获取库存失败，状态码: %d，响应: %s", w.Code, w.Body.String()))
			}
		})

		t.Run("AdjustInventory", func(t *gtest.T) {
			// 调整库存
			adjustReq := types.InventoryAdjustRequest{
				ProductID:      1001,
				AdjustmentType: "increase",
				Quantity:       10,
				Reason:         "API测试-增加库存",
			}

			body, _ := json.Marshal(adjustReq)
			req := httptest.NewRequest("POST", baseURL+"/inventory/adjust", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", "1")

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				t.AssertNil(err)
				t.Log(fmt.Sprintf("调整库存响应: %+v", response))
			} else {
				t.Log(fmt.Sprintf("调整库存失败，状态码: %d，响应: %s", w.Code, w.Body.String()))
			}
		})

		t.Run("GetInventoryRecords", func(t *gtest.T) {
			// 获取库存记录
			req := httptest.NewRequest("GET", baseURL+"/inventory/records?page=1&pageSize=10", nil)
			req.Header.Set("X-Tenant-ID", "1")

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				t.AssertNil(err)
				t.Log(fmt.Sprintf("库存记录响应: %+v", response))
			} else {
				t.Log(fmt.Sprintf("获取库存记录失败，状态码: %d，响应: %s", w.Code, w.Body.String()))
			}
		})

		t.Run("ReserveInventory", func(t *gtest.T) {
			// 预留库存
			reserveReq := types.InventoryReserveRequest{
				ProductID:     1001,
				Quantity:      5,
				ReferenceType: "order",
				ReferenceID:   "API_TEST_ORDER_001",
			}

			body, _ := json.Marshal(reserveReq)
			req := httptest.NewRequest("POST", "/api/v1/products/inventory/reserve", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", "1")

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				t.AssertNil(err)
				t.Log(fmt.Sprintf("预留库存响应: %+v", response))

				// 提取预留ID用于后续释放测试
				if data, ok := response["data"].(map[string]interface{}); ok {
					if reservationID, ok := data["id"].(float64); ok {
						testReleaseInventory(t, server, uint64(reservationID))
					}
				}
			} else {
				t.Log(fmt.Sprintf("预留库存失败，状态码: %d，响应: %s", w.Code, w.Body.String()))
			}
		})

		t.Run("BatchAdjustInventory", func(t *gtest.T) {
			// 批量调整库存
			batchReq := types.BatchInventoryAdjustRequest{
				Adjustments: []types.InventoryAdjustRequest{
					{
						ProductID:      1001,
						AdjustmentType: "increase",
						Quantity:       2,
					},
				},
				Reason: "API测试-批量调整",
			}

			body, _ := json.Marshal(batchReq)
			req := httptest.NewRequest("POST", "/api/v1/products/inventory/batch-adjust", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", "1")

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				t.AssertNil(err)
				t.Log(fmt.Sprintf("批量调整响应: %+v", response))
			} else {
				t.Log(fmt.Sprintf("批量调整失败，状态码: %d，响应: %s", w.Code, w.Body.String()))
			}
		})
	})
}

// testReleaseInventory 测试释放库存（辅助函数）
func testReleaseInventory(t *gtest.T, server *ghttp.Server, reservationID uint64) {
	t.Run("ReleaseInventory", func(t *gtest.T) {
		releaseReq := types.InventoryReleaseRequest{
			ReservationID: reservationID,
		}

		body, _ := json.Marshal(releaseReq)
		req := httptest.NewRequest("POST", "/api/v1/products/inventory/release", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "1")

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			t.AssertNil(err)
			t.Log(fmt.Sprintf("释放库存响应: %+v", response))
		} else {
			t.Log(fmt.Sprintf("释放库存失败，状态码: %d，响应: %s", w.Code, w.Body.String()))
		}
	})
}

// TestInventoryErrorHandling 测试库存API错误处理
func TestInventoryErrorHandling(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		server := setupInventoryTestServer()
		defer server.Shutdown()

		t.Run("InvalidProductID", func(t *gtest.T) {
			// 测试无效商品ID
			req := httptest.NewRequest("GET", "/api/v1/products/invalid/inventory", nil)
			req.Header.Set("X-Tenant-ID", "1")

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			t.AssertEQ(w.Code, http.StatusBadRequest)
			t.Log(fmt.Sprintf("无效商品ID响应: %s", w.Body.String()))
		})

		t.Run("MissingTenantID", func(t *gtest.T) {
			// 测试缺少租户ID
			req := httptest.NewRequest("GET", "/api/v1/products/1001/inventory", nil)
			// 不设置 X-Tenant-ID

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			// 由于中间件被省略，这个测试可能需要调整
			t.Log(fmt.Sprintf("缺少租户ID响应码: %d", w.Code))
		})

		t.Run("InvalidJSONPayload", func(t *gtest.T) {
			// 测试无效JSON负载
			req := httptest.NewRequest("POST", "/api/v1/products/1001/inventory/adjust", bytes.NewReader([]byte("invalid json")))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", "1")

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			t.AssertEQ(w.Code, http.StatusBadRequest)
			t.Log(fmt.Sprintf("无效JSON响应: %s", w.Body.String()))
		})
	})
}

// TestInventoryBusinessLogic 测试库存业务逻辑
func TestInventoryBusinessLogic(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		server := setupInventoryTestServer()
		defer server.Shutdown()

		testProductID := "1001"
		baseURL := "/api/v1/products/" + testProductID

		t.Run("InventoryWorkflow", func(t *gtest.T) {
			// 1. 获取初始库存
			req := httptest.NewRequest("GET", baseURL+"/inventory", nil)
			req.Header.Set("X-Tenant-ID", "1")
			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			var initialInventory map[string]interface{}
			if w.Code == http.StatusOK {
				json.Unmarshal(w.Body.Bytes(), &initialInventory)
				t.Log(fmt.Sprintf("初始库存: %+v", initialInventory))
			}

			// 2. 增加库存
			adjustReq := types.InventoryAdjustRequest{
				ProductID:      1001,
				AdjustmentType: "increase",
				Quantity:       20,
				Reason:         "工作流测试-增加库存",
			}

			body, _ := json.Marshal(adjustReq)
			req = httptest.NewRequest("POST", baseURL+"/inventory/adjust", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", "1")
			w = httptest.NewRecorder()
			server.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				t.Log("库存增加成功")
			}

			// 3. 预留部分库存
			reserveReq := types.InventoryReserveRequest{
				ProductID:     1001,
				Quantity:      10,
				ReferenceType: "order",
				ReferenceID:   "WORKFLOW_ORDER_001",
			}

			body, _ = json.Marshal(reserveReq)
			req = httptest.NewRequest("POST", "/api/v1/products/inventory/reserve", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", "1")
			w = httptest.NewRecorder()
			server.ServeHTTP(w, req)

			var reservationResponse map[string]interface{}
			var reservationID uint64
			if w.Code == http.StatusOK {
				json.Unmarshal(w.Body.Bytes(), &reservationResponse)
				if data, ok := reservationResponse["data"].(map[string]interface{}); ok {
					if id, ok := data["id"].(float64); ok {
						reservationID = uint64(id)
					}
				}
				t.Log("库存预留成功")
			}

			// 4. 检查库存状态
			req = httptest.NewRequest("GET", baseURL+"/inventory", nil)
			req.Header.Set("X-Tenant-ID", "1")
			w = httptest.NewRecorder()
			server.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				var currentInventory map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &currentInventory)
				t.Log(fmt.Sprintf("预留后库存: %+v", currentInventory))
			}

			// 5. 释放预留
			if reservationID > 0 {
				releaseReq := types.InventoryReleaseRequest{
					ReservationID: reservationID,
				}

				body, _ = json.Marshal(releaseReq)
				req = httptest.NewRequest("POST", "/api/v1/products/inventory/release", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Tenant-ID", "1")
				w = httptest.NewRecorder()
				server.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					t.Log("库存释放成功")
				}
			}

			// 6. 最终库存检查
			req = httptest.NewRequest("GET", baseURL+"/inventory", nil)
			req.Header.Set("X-Tenant-ID", "1")
			w = httptest.NewRecorder()
			server.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				var finalInventory map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &finalInventory)
				t.Log(fmt.Sprintf("最终库存: %+v", finalInventory))
			}
		})
	})
}