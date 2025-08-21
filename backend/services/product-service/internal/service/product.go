package service

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/gofromzero/mer-sys/backend/shared/oss"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ProductService 商品服务
type ProductService struct {
	productRepo  *repository.ProductRepository
	categoryRepo *repository.CategoryRepository
	historyRepo  *repository.ProductHistoryRepository
	ossService   *oss.OSSService
}

// NewProductService 创建商品服务实例
func NewProductService() *ProductService {
	return &ProductService{
		productRepo:  repository.NewProductRepository(),
		categoryRepo: repository.NewCategoryRepository(),
		historyRepo:  repository.NewProductHistoryRepository(),
		ossService:   oss.NewOSSService(),
	}
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, req *types.CreateProductRequest) (*types.Product, error) {
	// 验证分类是否存在
	if req.CategoryID != nil {
		_, err := s.categoryRepo.GetByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id: %v", err)
		}
	}
	
	// 构建商品对象，直接使用Money结构
	product := &types.Product{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Tags:        types.StringArray(req.Tags),
		PriceAmount: float64(req.Price.Amount) / 100, // 转换为元
		PriceCurrency: req.Price.Currency,
		RightsCost:  float64(req.RightsCost) / 100, // 转换为元
		InventoryInfo: &req.Inventory,
		Status:      types.ProductStatusDraft, // 默认为草稿状态
		Images:      types.ProductImages{}, // 初始化空图片数组
	}
	
	// 设置分类路径（这里暂时不使用，但保留逻辑供将来扩展）
	
	// 创建商品
	err := s.productRepo.Create(ctx, product)
	if err != nil {
		return nil, err
	}
	
	// 记录创建历史
	err = s.historyRepo.RecordChange(ctx, product.ID, types.ChangeOperationCreate, map[string]interface{}{
		"name":        product.Name,
		"description": product.Description,
		"price_amount": product.PriceAmount,
		"price_currency": product.PriceCurrency,
		"status":      product.Status,
	})
	if err != nil {
		// 记录历史失败不应该影响商品创建，只记录日志
		// TODO: 添加日志记录
	}
	
	return product, nil
}

// GetProduct 获取商品详情
func (s *ProductService) GetProduct(ctx context.Context, id uint64) (*types.ProductResponse, error) {
	return s.productRepo.GetByIDWithCategory(ctx, id)
}

// UpdateProduct 更新商品信息
func (s *ProductService) UpdateProduct(ctx context.Context, id uint64, req *types.UpdateProductRequest) (*types.Product, error) {
	// 获取原商品信息用于记录变更历史
	oldProduct, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// 构建更新字段
	updates := make(map[string]interface{})
	changes := make(map[string]interface{})
	
	if req.Name != "" && req.Name != oldProduct.Name {
		updates["name"] = req.Name
		changes["name"] = map[string]interface{}{
			"old": oldProduct.Name,
			"new": req.Name,
		}
	}
	
	if req.Description != oldProduct.Description {
		updates["description"] = req.Description
		changes["description"] = map[string]interface{}{
			"old": oldProduct.Description,
			"new": req.Description,
		}
	}
	
	if req.CategoryID != nil && (oldProduct.CategoryID == nil || *req.CategoryID != *oldProduct.CategoryID) {
		// 验证分类是否存在
		category, err := s.categoryRepo.GetByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id: %v", err)
		}
		
		updates["category_id"] = *req.CategoryID
		updates["category_path"] = category.Path
		changes["category"] = map[string]interface{}{
			"old": oldProduct.CategoryID,
			"new": *req.CategoryID,
		}
	}
	
	if len(req.Tags) > 0 {
		updates["tags"] = types.StringArray(req.Tags)
		changes["tags"] = map[string]interface{}{
			"old": oldProduct.Tags,
			"new": req.Tags,
		}
	}
	
	if req.Price != nil {
		updates["price_amount"] = float64(req.Price.Amount) / 100
		updates["price_currency"] = req.Price.Currency
		changes["price"] = map[string]interface{}{
			"old": map[string]interface{}{
				"amount": oldProduct.PriceAmount * 100,
				"currency": oldProduct.PriceCurrency,
			},
			"new": map[string]interface{}{
				"amount": req.Price.Amount,
				"currency": req.Price.Currency,
			},
		}
	}
	
	if req.RightsCost != nil && float64(*req.RightsCost)/100 != oldProduct.RightsCost {
		updates["rights_cost"] = float64(*req.RightsCost) / 100
		changes["rights_cost"] = map[string]interface{}{
			"old": oldProduct.RightsCost * 100,
			"new": *req.RightsCost,
		}
	}
	
	if req.Inventory != nil {
		updates["inventory_info"] = *req.Inventory
		changes["inventory"] = map[string]interface{}{
			"old": oldProduct.InventoryInfo,
			"new": *req.Inventory,
		}
	}
	
	if len(updates) == 0 {
		// 没有任何更新
		return oldProduct, nil
	}
	
	// 执行更新
	err = s.productRepo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}
	
	// 记录变更历史
	for field, change := range changes {
		err = s.historyRepo.RecordFieldChange(ctx, id, field, change)
		if err != nil {
			// 记录历史失败不应该影响更新，只记录日志
		}
	}
	
	// 返回更新后的商品信息
	return s.productRepo.GetByID(ctx, id)
}

// UpdateProductStatus 更新商品状态
func (s *ProductService) UpdateProductStatus(ctx context.Context, id uint64, req *types.UpdateProductStatusRequest) error {
	// 获取原状态
	oldProduct, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	if oldProduct.Status == req.Status {
		return nil // 状态未变更
	}
	
	// 状态流转验证
	err = s.validateStatusTransition(oldProduct.Status, req.Status)
	if err != nil {
		return err
	}
	
	// 更新状态
	err = s.productRepo.UpdateStatus(ctx, id, req.Status)
	if err != nil {
		return err
	}
	
	// 记录状态变更历史
	err = s.historyRepo.RecordStatusChange(ctx, id, oldProduct.Status, req.Status)
	if err != nil {
		// 记录历史失败不应该影响状态更新，只记录日志
	}
	
	return nil
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(ctx context.Context, id uint64) error {
	// 获取商品信息
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	// 执行软删除
	err = s.productRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	
	// 记录删除历史
	err = s.historyRepo.RecordChange(ctx, id, types.ChangeOperationDelete, map[string]interface{}{
		"name":   product.Name,
		"status": "deleted",
	})
	if err != nil {
		// 记录历史失败不应该影响删除，只记录日志
	}
	
	return nil
}

// ListProducts 获取商品列表
func (s *ProductService) ListProducts(ctx context.Context, req *types.ProductListRequest) (*types.ProductListResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	
	return s.productRepo.List(ctx, req)
}

// BatchOperation 批量操作商品
func (s *ProductService) BatchOperation(ctx context.Context, req *types.ProductBatchOperationRequest) error {
	if len(req.ProductIDs) == 0 {
		return fmt.Errorf("product IDs cannot be empty")
	}
	
	var status types.ProductStatus
	var operation types.ChangeOperation
	
	switch req.Operation {
	case "activate":
		status = types.ProductStatusActive
		operation = types.ChangeOperationStatusChange
	case "deactivate":
		status = types.ProductStatusInactive
		operation = types.ChangeOperationStatusChange
	case "delete":
		status = types.ProductStatusDeleted
		operation = types.ChangeOperationDelete
	default:
		return fmt.Errorf("invalid operation: %s", req.Operation)
	}
	
	// 批量更新状态
	err := s.productRepo.BatchUpdateStatus(ctx, req.ProductIDs, status)
	if err != nil {
		return err
	}
	
	// 记录批量操作历史
	for _, productID := range req.ProductIDs {
		err = s.historyRepo.RecordChange(ctx, productID, operation, map[string]interface{}{
			"batch_operation": req.Operation,
			"new_status":     status,
		})
		if err != nil {
			// 记录历史失败不应该影响操作，只记录日志
		}
	}
	
	return nil
}

// UploadImageFile 上传商品图片文件（完整的文件上传到OSS）  
func (s *ProductService) UploadImageFile(ctx context.Context, productID uint64, file multipart.File, uploadFile *ghttp.UploadFile, req *types.UploadImageRequest) (*oss.UploadFileInfo, error) {
	// 验证商品是否存在
	_, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %v", err)
	}
	
	// 设置上传选项
	options := &oss.UploadOptions{
		Directory: fmt.Sprintf("products/%d/images", productID),
		Metadata: map[string]string{
			"product_id": fmt.Sprintf("%d", productID),
			"alt_text":   req.AltText,
		},
	}
	
	// 创建multipart.FileHeader
	header := &multipart.FileHeader{
		Filename: uploadFile.Filename,
		Size:     uploadFile.Size,
		Header:   uploadFile.Header,
	}
	
	// 上传到OSS
	uploadInfo, err := s.ossService.UploadFile(ctx, file, header, options)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to OSS: %v", err)
	}
	
	// 创建商品图片记录
	err = s.UploadImage(ctx, productID, req, uploadInfo.URL)
	if err != nil {
		// 如果数据库操作失败，尝试删除已上传的文件（最佳实践）
		s.ossService.DeleteFile(ctx, fmt.Sprintf("%s/%s", options.Directory, uploadInfo.FileName))
		return nil, fmt.Errorf("failed to save image info: %v", err)
	}
	
	return uploadInfo, nil
}

// UploadImage 上传商品图片（用于已有URL的场景）
func (s *ProductService) UploadImage(ctx context.Context, productID uint64, req *types.UploadImageRequest, imageURL string) error {
	// 生成图片ID
	imageID := fmt.Sprintf("img_%d_%d", productID, len(imageURL)) // 简单的ID生成，实际应该使用UUID
	
	image := types.ProductImage{
		ID:        imageID,
		URL:       imageURL,
		AltText:   req.AltText,
		SortOrder: req.SortOrder,
		IsPrimary: req.IsPrimary,
	}
	
	err := s.productRepo.AddImage(ctx, productID, image)
	if err != nil {
		return err
	}
	
	// 记录图片上传历史
	err = s.historyRepo.RecordChange(ctx, productID, types.ChangeOperationUpdate, map[string]interface{}{
		"action":   "add_image",
		"image_id": imageID,
		"image_url": imageURL,
	})
	if err != nil {
		// 记录历史失败不应该影响上传，只记录日志
	}
	
	return nil
}

// GetProductHistory 获取商品变更历史
func (s *ProductService) GetProductHistory(ctx context.Context, productID uint64) ([]types.ProductHistory, error) {
	return s.historyRepo.GetProductHistory(ctx, productID)
}

// validateStatusTransition 验证状态流转是否合法
func (s *ProductService) validateStatusTransition(from, to types.ProductStatus) error {
	// 定义允许的状态流转
	allowedTransitions := map[types.ProductStatus][]types.ProductStatus{
		types.ProductStatusDraft: {
			types.ProductStatusActive,
			types.ProductStatusDeleted,
		},
		types.ProductStatusActive: {
			types.ProductStatusInactive,
			types.ProductStatusDeleted,
		},
		types.ProductStatusInactive: {
			types.ProductStatusActive,
			types.ProductStatusDeleted,
		},
		types.ProductStatusDeleted: {}, // 删除状态不能流转到其他状态
	}
	
	allowed, exists := allowedTransitions[from]
	if !exists {
		return fmt.Errorf("invalid current status: %s", from)
	}
	
	for _, allowedStatus := range allowed {
		if allowedStatus == to {
			return nil
		}
	}
	
	return fmt.Errorf("invalid status transition from %s to %s", from, to)
}