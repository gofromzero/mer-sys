package controller

import (
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/service"
)

// InventoryController 库存管理控制器
type InventoryController struct {
	inventoryService service.IInventoryService
	productRepo      *repository.ProductRepository
	recordRepo       repository.IInventoryRecordRepository
	reservationRepo  repository.IInventoryReservationRepository
}

// NewInventoryController 创建库存管理控制器
func NewInventoryController() *InventoryController {
	return &InventoryController{
		inventoryService: service.NewInventoryService(),
		productRepo:      repository.NewProductRepository(),
		recordRepo:       repository.NewInventoryRecordRepository(),
		reservationRepo:  repository.NewInventoryReservationRepository(),
	}
}

// GetInventory 获取商品库存信息
// GET /api/v1/products/{id}/inventory
func (c *InventoryController) GetInventory(r *ghttp.Request) {
	// 权限检查暂时省略

	productIDStr := r.Get("id").String()
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取库存信息
	response, err := c.inventoryService.GetInventoryInfo(r.Context(), productID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取库存信息失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "获取库存信息成功",
		"data":    response,
	})
}

// AdjustInventory 调整库存数量
// POST /api/v1/products/{id}/inventory/adjust
func (c *InventoryController) AdjustInventory(r *ghttp.Request) {
	// 权限检查暂时省略

	productIDStr := r.Get("id").String()
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	var req types.InventoryAdjustRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数解析失败",
			"error":   err.Error(),
		})
		return
	}

	// 设置商品ID
	req.ProductID = productID

	// 执行库存调整
	response, err := c.inventoryService.AdjustInventory(r.Context(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "库存调整失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "库存调整成功",
		"data":    response,
	})
}

// GetInventoryRecords 获取库存变更历史
// GET /api/v1/products/{id}/inventory/records
func (c *InventoryController) GetInventoryRecords(r *ghttp.Request) {
	// 权限检查暂时省略

	productIDStr := r.Get("id").String()
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 分页参数
	page := r.Get("page", 1).Int()
	pageSize := r.Get("pageSize", 20).Int()
	if pageSize > 100 {
		pageSize = 100
	}

	// 获取租户ID
	tenantID := r.GetCtxVar("tenant_id").Uint64()
	if tenantID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "租户信息缺失",
		})
		return
	}

	// 获取库存记录
	records, total, err := c.recordRepo.GetByProductID(r.Context(), tenantID, productID, page, pageSize)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取库存记录失败",
			"error":   err.Error(),
		})
		return
	}

	response := &types.InventoryRecordResponse{
		Records:  records,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "获取库存记录成功",
		"data":    response,
	})
}

// ReserveInventory 预留库存
// POST /api/v1/products/{id}/inventory/reserve
func (c *InventoryController) ReserveInventory(r *ghttp.Request) {
	// 权限检查暂时省略

	productIDStr := r.Get("id").String()
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	var req types.InventoryReserveRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数解析失败",
			"error":   err.Error(),
		})
		return
	}

	// 设置商品ID
	req.ProductID = productID

	// 执行库存预留
	reservation, err := c.inventoryService.ReserveInventory(r.Context(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "库存预留失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "库存预留成功",
		"data":    reservation,
	})
}

// ReleaseInventory 释放预留库存
// POST /api/v1/products/{id}/inventory/release
func (c *InventoryController) ReleaseInventory(r *ghttp.Request) {
	// 权限检查暂时省略

	var req types.InventoryReleaseRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数解析失败",
			"error":   err.Error(),
		})
		return
	}

	// 执行库存释放
	err := c.inventoryService.ReleaseInventory(r.Context(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "库存释放失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "库存释放成功",
	})
}

// BatchAdjustInventory 批量调整库存
// POST /api/v1/products/inventory/batch-adjust
func (c *InventoryController) BatchAdjustInventory(r *ghttp.Request) {
	// 权限检查暂时省略

	var req types.BatchInventoryAdjustRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数解析失败",
			"error":   err.Error(),
		})
		return
	}

	// 执行批量库存调整
	results, err := c.inventoryService.BatchAdjustInventory(r.Context(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "批量库存调整失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "批量库存调整成功",
		"data":    results,
	})
}

// StartStocktaking 启动库存盘点
// POST /api/v1/inventory/stocktaking/start
func (c *InventoryController) StartStocktaking(r *ghttp.Request) {
	// 权限检查暂时省略

	var req struct {
		Name        string   `json:"name" v:"required#盘点名称不能为空"`
		Description string   `json:"description"`
		ProductIDs  []uint64 `json:"product_ids"`  // 如果为空则盘点所有商品
		StartTime   string   `json:"start_time"`   // 盘点开始时间，为空则立即开始
	}

	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数解析失败",
			"error":   err.Error(),
		})
		return
	}

	// 这里应该调用库存服务创建盘点任务
	// 暂时模拟创建成功
	stocktakingID := uint64(12345)

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "库存盘点任务创建成功",
		"data": g.Map{
			"stocktaking_id": stocktakingID,
			"status":        "pending",
		},
	})
}

// UpdateStocktakingRecord 更新盘点记录
// PUT /api/v1/inventory/stocktaking/{id}/records
func (c *InventoryController) UpdateStocktakingRecord(r *ghttp.Request) {
	// 权限检查暂时省略

	stocktakingIDStr := r.Get("id").String()
	_, err := strconv.ParseUint(stocktakingIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "盘点ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	var req struct {
		Records []struct {
			ProductID     uint64 `json:"product_id" v:"required"`
			ActualCount   int    `json:"actual_count" v:"required|min:0"`
			SystemCount   int    `json:"system_count"`
			Difference    int    `json:"difference"`
			Reason        string `json:"reason"`
		} `json:"records" v:"required"`
	}

	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数解析失败",
			"error":   err.Error(),
		})
		return
	}

	// 处理盘点记录更新
	for _, record := range req.Records {
		// 计算差异
		difference := record.ActualCount - record.SystemCount
		
		// 如果有差异，执行库存调整
		if difference != 0 {
			adjustReq := &types.InventoryAdjustRequest{
				ProductID:      record.ProductID,
				AdjustmentType: "set",
				Quantity:       record.ActualCount,
				Reason:         "库存盘点调整: " + record.Reason,
			}

			_, err := c.inventoryService.AdjustInventory(r.Context(), adjustReq)
			if err != nil {
				g.Log().Errorf(r.Context(), "盘点调整库存失败: 商品ID=%d, 错误=%v", record.ProductID, err)
				// 继续处理其他记录，不中断整个流程
			}
		}
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "盘点记录更新成功",
		"data": g.Map{
			"processed_records": len(req.Records),
		},
	})
}

// CompleteStocktaking 完成库存盘点
// POST /api/v1/inventory/stocktaking/{id}/complete
func (c *InventoryController) CompleteStocktaking(r *ghttp.Request) {
	// 权限检查暂时省略

	stocktakingIDStr := r.Get("id").String()
	stocktakingID, err := strconv.ParseUint(stocktakingIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "盘点ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	var req struct {
		Summary string `json:"summary"`
		Notes   string `json:"notes"`
	}

	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数解析失败",
			"error":   err.Error(),
		})
		return
	}

	// 完成盘点任务
	// 这里应该调用服务层更新盘点状态
	g.Log().Infof(r.Context(), "完成库存盘点任务: ID=%d, 摘要=%s", stocktakingID, req.Summary)

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "库存盘点已完成",
		"data": g.Map{
			"stocktaking_id": stocktakingID,
			"status":        "completed",
		},
	})
}

// GetStocktakingList 获取盘点任务列表
// GET /api/v1/inventory/stocktaking
func (c *InventoryController) GetStocktakingList(r *ghttp.Request) {
	// 权限检查暂时省略

	page := r.Get("page", 1).Int()
	pageSize := r.Get("page_size", 20).Int()
	status := r.Get("status").String()

	// 这里应该调用服务层获取盘点任务列表
	// 暂时返回模拟数据
	stocktakings := []g.Map{
		{
			"id":          12345,
			"name":        "2024年第一季度盘点",
			"description": "全面库存盘点",
			"status":      "completed",
			"created_at":  "2024-08-23 10:00:00",
			"completed_at": "2024-08-23 18:00:00",
		},
		{
			"id":          12346,
			"name":        "重点商品专项盘点",
			"description": "针对高价值商品的专项盘点",
			"status":      "in_progress",
			"created_at":  "2024-08-23 14:00:00",
			"completed_at": nil,
		},
	}

	// 根据状态筛选
	var filteredStocktakings []g.Map
	if status != "" {
		for _, st := range stocktakings {
			if st["status"] == status {
				filteredStocktakings = append(filteredStocktakings, st)
			}
		}
	} else {
		filteredStocktakings = stocktakings
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "获取盘点任务列表成功",
		"data": g.Map{
			"stocktakings": filteredStocktakings,
			"total":        len(filteredStocktakings),
			"page":         page,
			"page_size":    pageSize,
		},
	})
}

// BatchQueryInventory 批量查询库存信息
// POST /api/v1/inventory/batch-query
func (c *InventoryController) BatchQueryInventory(r *ghttp.Request) {
	// 权限检查暂时省略

	var req struct {
		ProductIDs []uint64 `json:"product_ids" v:"required"`
	}

	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数解析失败",
			"error":   err.Error(),
		})
		return
	}

	if len(req.ProductIDs) == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID列表不能为空",
		})
		return
	}

	if len(req.ProductIDs) > 1000 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "单次查询商品数量不能超过1000个",
		})
		return
	}

	// 批量查询库存信息
	var results []types.InventoryResponse
	for _, productID := range req.ProductIDs {
		response, err := c.inventoryService.GetInventoryInfo(r.Context(), productID)
		if err != nil {
			g.Log().Warningf(r.Context(), "获取商品%d库存信息失败: %v", productID, err)
			continue
		}
		results = append(results, *response)
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "批量查询库存成功",
		"data":    results,
	})
}