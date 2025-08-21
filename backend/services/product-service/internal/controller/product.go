package controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ProductController 商品控制器
type ProductController struct {
	productService *service.ProductService
}

// NewProductController 创建商品控制器实例
func NewProductController() *ProductController {
	return &ProductController{
		productService: service.NewProductService(),
	}
}

// CreateProduct 创建商品
func (c *ProductController) CreateProduct(r *ghttp.Request) {
	var req types.CreateProductRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	// 参数验证
	if err := validateCreateProductRequest(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数验证失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	product, err := c.productService.CreateProduct(r.GetCtx(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "创建商品失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "创建成功",
		"data":    product,
	})
}

// GetProduct 获取商品详情
func (c *ProductController) GetProduct(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的商品ID",
			"data":    nil,
		})
		return
	}
	
	product, err := c.productService.GetProduct(r.GetCtx(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    404,
			"message": "商品不存在",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    product,
	})
}

// UpdateProduct 更新商品信息
func (c *ProductController) UpdateProduct(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的商品ID",
			"data":    nil,
		})
		return
	}
	
	var req types.UpdateProductRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	// 参数验证
	if err := validateUpdateProductRequest(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数验证失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	product, err := c.productService.UpdateProduct(r.GetCtx(), id, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新商品失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "更新成功",
		"data":    product,
	})
}

// UpdateProductStatus 更新商品状态
func (c *ProductController) UpdateProductStatus(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的商品ID",
			"data":    nil,
		})
		return
	}
	
	var req types.UpdateProductStatusRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	// 参数验证
	if err := validateUpdateProductStatusRequest(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数验证失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	err = c.productService.UpdateProductStatus(r.GetCtx(), id, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新状态失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "状态更新成功",
		"data":    nil,
	})
}

// DeleteProduct 删除商品
func (c *ProductController) DeleteProduct(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的商品ID",
			"data":    nil,
		})
		return
	}
	
	err = c.productService.DeleteProduct(r.GetCtx(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "删除商品失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}

// ListProducts 获取商品列表
func (c *ProductController) ListProducts(r *ghttp.Request) {
	var req types.ProductListRequest
	
	// 解析查询参数
	req.Page = r.GetQuery("page", 1).Int()
	req.PageSize = r.GetQuery("page_size", 20).Int()
	req.Keyword = r.GetQuery("keyword").String()
	req.SortBy = r.GetQuery("sort_by").String()
	req.SortOrder = r.GetQuery("sort_order").String()
	
	if categoryIDStr := r.GetQuery("category_id").String(); categoryIDStr != "" {
		if categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64); err == nil {
			req.CategoryID = &categoryID
		}
	}
	
	if statusStr := r.GetQuery("status").String(); statusStr != "" {
		req.Status = types.ProductStatus(statusStr)
	}
	
	// 参数验证
	if err := validateProductListRequest(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数验证失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	result, err := c.productService.ListProducts(r.GetCtx(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取商品列表失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    result,
	})
}

// BatchOperation 批量操作商品
func (c *ProductController) BatchOperation(r *ghttp.Request) {
	var req types.ProductBatchOperationRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	// 参数验证
	if err := validateBatchOperationRequest(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数验证失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	err := c.productService.BatchOperation(r.GetCtx(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "批量操作失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "批量操作成功",
		"data":    nil,
	})
}

// UploadImage 上传商品图片
func (c *ProductController) UploadImage(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的商品ID",
			"data":    nil,
		})
		return
	}
	
	// 获取上传的文件
	uploadFile := r.GetUploadFile("image")
	if uploadFile == nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "未找到上传文件",
			"data":    nil,
		})
		return
	}
	
	// 打开文件
	file, err := uploadFile.Open()
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无法打开上传文件",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	defer file.Close()
	
	var req types.UploadImageRequest
	req.AltText = r.GetForm("alt_text").String()
	req.SortOrder = r.GetForm("sort_order").Int()
	req.IsPrimary = r.GetForm("is_primary").Bool()
	
	// 使用产品服务上传图片（包含OSS集成）
	imageInfo, err := c.productService.UploadImageFile(r.GetCtx(), id, file, uploadFile, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "上传图片失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "上传成功",
		"data":    imageInfo,
	})
}

// GetProductHistory 获取商品变更历史
func (c *ProductController) GetProductHistory(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的商品ID",
			"data":    nil,
		})
		return
	}
	
	history, err := c.productService.GetProductHistory(r.GetCtx(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取变更历史失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    history,
	})
}

// validateCreateProductRequest 验证创建商品请求
func validateCreateProductRequest(req *types.CreateProductRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("商品名称不能为空")
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("商品名称长度不能超过255个字符")
	}
	if len(req.Description) > 2000 {
		return fmt.Errorf("商品描述长度不能超过2000个字符")
	}
	if req.Price.Amount <= 0 {
		return fmt.Errorf("商品价格必须大于0")
	}
	if req.RightsCost < 0 {
		return fmt.Errorf("权益成本不能为负数")
	}
	if req.Inventory.StockQuantity < 0 {
		return fmt.Errorf("库存数量不能为负数")
	}
	if len(req.Tags) > 20 {
		return fmt.Errorf("商品标签不能超过20个")
	}
	for _, tag := range req.Tags {
		if len(tag) > 50 {
			return fmt.Errorf("标签长度不能超过50个字符")
		}
	}
	return nil
}

// validateUpdateProductRequest 验证更新商品请求
func validateUpdateProductRequest(req *types.UpdateProductRequest) error {
	if req.Name != "" {
		if strings.TrimSpace(req.Name) == "" {
			return fmt.Errorf("商品名称不能为空")
		}
		if len(req.Name) > 255 {
			return fmt.Errorf("商品名称长度不能超过255个字符")
		}
	}
	if len(req.Description) > 2000 {
		return fmt.Errorf("商品描述长度不能超过2000个字符")
	}
	if req.Price != nil && req.Price.Amount <= 0 {
		return fmt.Errorf("商品价格必须大于0")
	}
	if req.RightsCost != nil && *req.RightsCost < 0 {
		return fmt.Errorf("权益成本不能为负数")
	}
	if req.Inventory != nil && req.Inventory.StockQuantity < 0 {
		return fmt.Errorf("库存数量不能为负数")
	}
	if len(req.Tags) > 20 {
		return fmt.Errorf("商品标签不能超过20个")
	}
	for _, tag := range req.Tags {
		if len(tag) > 50 {
			return fmt.Errorf("标签长度不能超过50个字符")
		}
	}
	return nil
}

// validateUpdateProductStatusRequest 验证状态更新请求
func validateUpdateProductStatusRequest(req *types.UpdateProductStatusRequest) error {
	validStatuses := []types.ProductStatus{
		types.ProductStatusDraft,
		types.ProductStatusActive,
		types.ProductStatusInactive,
		types.ProductStatusDeleted,
	}
	
	for _, status := range validStatuses {
		if req.Status == status {
			return nil
		}
	}
	return fmt.Errorf("无效的商品状态: %s", req.Status)
}

// validateProductListRequest 验证商品列表请求
func validateProductListRequest(req *types.ProductListRequest) error {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		return fmt.Errorf("每页数量不能超过100")
	}
	
	if req.SortBy != "" {
		validSortFields := []string{"created_at", "updated_at", "name", "price"}
		valid := false
		for _, field := range validSortFields {
			if req.SortBy == field {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("无效的排序字段: %s", req.SortBy)
		}
	}
	
	if req.SortOrder != "" && req.SortOrder != "asc" && req.SortOrder != "desc" {
		return fmt.Errorf("排序方向只能是asc或desc")
	}
	
	return nil
}

// validateBatchOperationRequest 验证批量操作请求
func validateBatchOperationRequest(req *types.ProductBatchOperationRequest) error {
	if len(req.ProductIDs) == 0 {
		return fmt.Errorf("商品ID列表不能为空")
	}
	if len(req.ProductIDs) > 100 {
		return fmt.Errorf("批量操作商品数量不能超过100个")
	}
	
	validOperations := []string{"activate", "deactivate", "delete"}
	valid := false
	for _, op := range validOperations {
		if req.Operation == op {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("无效的操作类型: %s", req.Operation)
	}
	
	return nil
}