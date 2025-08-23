package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductRepository 商品数据访问层
type ProductRepository struct {
	*BaseRepository
}

// NewProductRepository 创建商品仓库实例
func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// Create 创建商品
func (r *ProductRepository) Create(ctx context.Context, product *types.Product) error {
	db := g.DB()
	
	// 自动注入租户和商户信息
	tenantID := r.GetTenantID(ctx)
	merchantID := r.GetMerchantID(ctx)
	if tenantID == 0 || merchantID == 0 {
		return fmt.Errorf("missing tenant_id or merchant_id in context")
	}
	
	product.TenantID = tenantID
	product.MerchantID = merchantID
	product.Version = 1
	
	result, err := db.Model("products").Ctx(ctx).Insert(product)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	product.ID = uint64(id)
	return nil
}

// GetByID 根据ID获取商品
func (r *ProductRepository) GetByID(ctx context.Context, id uint64) (*types.Product, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	var product types.Product
	err := g.DB().Model("products").
		Ctx(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Scan(&product)
	
	if err != nil {
		return nil, err
	}
	
	if product.ID == 0 {
		return nil, fmt.Errorf("product not found")
	}
	
	return &product, nil
}

// GetByIDWithCategory 根据ID获取商品及其分类信息
func (r *ProductRepository) GetByIDWithCategory(ctx context.Context, id uint64) (*types.ProductResponse, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	// 获取商品信息
	product, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	response := &types.ProductResponse{Product: *product}
	
	// 获取分类信息
	if product.CategoryID != nil {
		var category types.ProductCategory
		err := g.DB().Model("product_categories").
			Ctx(ctx).
			Where("id = ? AND tenant_id = ?", *product.CategoryID, tenantID).
			Scan(&category)
		
		if err == nil && category.ID > 0 {
			response.Category = &category
		}
	}
	
	return response, nil
}

// Update 更新商品
func (r *ProductRepository) Update(ctx context.Context, id uint64, updates map[string]interface{}) error {
	tenantID := r.GetTenantID(ctx)
	merchantID := r.GetMerchantID(ctx)
	if tenantID == 0 || merchantID == 0 {
		return fmt.Errorf("missing tenant_id or merchant_id in context")
	}
	
	// 增加版本号
	if _, exists := updates["version"]; !exists {
		// 获取当前版本号并加1
		var currentVersion int
		err := g.DB().Model("products").
			Ctx(ctx).
			Where("id = ? AND tenant_id = ? AND merchant_id = ?", id, tenantID, merchantID).
			Fields("version").
			Scan(&currentVersion)
		if err != nil {
			return err
		}
		updates["version"] = currentVersion + 1
	}
	
	result, err := g.DB().Model("products").
		Ctx(ctx).
		Where("id = ? AND tenant_id = ? AND merchant_id = ?", id, tenantID, merchantID).
		Update(updates)
	
	if err != nil {
		return err
	}
	
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("product not found or permission denied")
	}
	
	return nil
}

// UpdateStatus 更新商品状态
func (r *ProductRepository) UpdateStatus(ctx context.Context, id uint64, status types.ProductStatus) error {
	return r.Update(ctx, id, map[string]interface{}{
		"status": status,
	})
}

// Delete 软删除商品
func (r *ProductRepository) Delete(ctx context.Context, id uint64) error {
	return r.UpdateStatus(ctx, id, types.ProductStatusDeleted)
}

// List 获取商品列表
func (r *ProductRepository) List(ctx context.Context, req *types.ProductListRequest) (*types.ProductListResponse, error) {
	tenantID := r.GetTenantID(ctx)
	merchantID := r.GetMerchantID(ctx)
	if tenantID == 0 || merchantID == 0 {
		return nil, fmt.Errorf("missing tenant_id or merchant_id in context")
	}
	
	db := g.DB().Model("products p").
		LeftJoin("product_categories c", "p.category_id = c.id").
		Where("p.tenant_id = ? AND p.merchant_id = ?", tenantID, merchantID).
		Where("p.status != ?", types.ProductStatusDeleted)
	
	// 添加筛选条件
	if req.CategoryID != nil {
		db = db.Where("p.category_id = ?", *req.CategoryID)
	}
	
	if req.Status != "" {
		db = db.Where("p.status = ?", req.Status)
	}
	
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("(p.name LIKE ? OR p.description LIKE ?)", keyword, keyword)
	}
	
	// 计算总数
	total, err := db.Count()
	if err != nil {
		return nil, err
	}
	
	// 排序
	orderBy := "p.created_at DESC"
	if req.SortBy != "" {
		order := "ASC"
		if req.SortOrder == "desc" {
			order = "DESC"
		}
		
		validSortFields := map[string]string{
			"created_at": "p.created_at",
			"updated_at": "p.updated_at",
			"name":       "p.name",
			"price":      "p.price",
		}
		
		if dbField, valid := validSortFields[req.SortBy]; valid {
			orderBy = fmt.Sprintf("%s %s", dbField, order)
		}
	}
	
	// 分页
	offset := (req.Page - 1) * req.PageSize
	
	// 查询数据
	var results []struct {
		types.Product
		CategoryName *string `json:"category_name"`
	}
	
	err = db.Fields("p.*, c.name as category_name").
		Order(orderBy).
		Limit(req.PageSize).
		Offset(offset).
		Scan(&results)
	
	if err != nil {
		return nil, err
	}
	
	// 转换响应格式
	products := make([]types.ProductResponse, len(results))
	for i, result := range results {
		products[i] = types.ProductResponse{Product: result.Product}
		if result.CategoryName != nil && result.CategoryID != nil {
			products[i].Category = &types.ProductCategory{
				ID:   *result.CategoryID,
				Name: *result.CategoryName,
			}
		}
	}
	
	return &types.ProductListResponse{
		Products: products,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// BatchUpdateStatus 批量更新商品状态
func (r *ProductRepository) BatchUpdateStatus(ctx context.Context, productIDs []uint64, status types.ProductStatus) error {
	tenantID := r.GetTenantID(ctx)
	merchantID := r.GetMerchantID(ctx)
	if tenantID == 0 || merchantID == 0 {
		return fmt.Errorf("missing tenant_id or merchant_id in context")
	}
	
	if len(productIDs) == 0 {
		return fmt.Errorf("product IDs cannot be empty")
	}
	
	// 构建 IN 条件的占位符
	placeholders := strings.Repeat("?,", len(productIDs))
	placeholders = strings.TrimSuffix(placeholders, ",")
	
	// 构建参数列表
	args := make([]interface{}, 0, len(productIDs)+3)
	for _, id := range productIDs {
		args = append(args, id)
	}
	args = append(args, tenantID, merchantID, status)
	
	result, err := g.DB().Exec(ctx, 
		fmt.Sprintf("UPDATE products SET status = ?, version = version + 1 WHERE id IN (%s) AND tenant_id = ? AND merchant_id = ?", placeholders),
		append([]interface{}{status}, args[:len(productIDs)+2]...)...,
	)
	
	if err != nil {
		return err
	}
	
	affected, _ := result.RowsAffected()
	if affected != int64(len(productIDs)) {
		return fmt.Errorf("some products were not updated, possibly due to permissions")
	}
	
	return nil
}

// GetMerchantProducts 获取商户的所有商品（用于商户用户管理）
func (r *ProductRepository) GetMerchantProducts(ctx context.Context, merchantID uint64, status types.ProductStatus) ([]types.Product, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	db := g.DB().Model("products").
		Where("tenant_id = ? AND merchant_id = ?", tenantID, merchantID)
	
	if status != "" {
		db = db.Where("status = ?", status)
	} else {
		db = db.Where("status != ?", types.ProductStatusDeleted)
	}
	
	var products []types.Product
	err := db.Ctx(ctx).Scan(&products)
	if err != nil {
		return nil, err
	}
	
	return products, nil
}

// AddImage 添加商品图片
func (r *ProductRepository) AddImage(ctx context.Context, productID uint64, image types.ProductImage) error {
	product, err := r.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	
	// 如果设置为主图，先将其他图片的主图标识设为false
	if image.IsPrimary {
		for i := range product.Images {
			product.Images[i].IsPrimary = false
		}
	}
	
	// 添加新图片
	product.Images = append(product.Images, image)
	
	return r.Update(ctx, productID, map[string]interface{}{
		"images": product.Images,
	})
}

// UpdateInventoryInfo 更新商品库存信息
func (r *ProductRepository) UpdateInventoryInfo(ctx context.Context, productID uint64, inventoryInfo types.ExtendedInventoryInfo) error {
	tenantID := r.GetTenantID(ctx)
	merchantID := r.GetMerchantID(ctx)
	if tenantID == 0 || merchantID == 0 {
		return fmt.Errorf("missing tenant_id or merchant_id in context")
	}
	
	return r.Update(ctx, productID, map[string]interface{}{
		"inventory_info": inventoryInfo,
	})
}

// GetInventoryInfo 获取商品库存信息  
func (r *ProductRepository) GetInventoryInfo(ctx context.Context, productID uint64) (*types.ExtendedInventoryInfo, error) {
	product, err := r.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	
	if product.InventoryInfo == nil {
		return nil, fmt.Errorf("商品库存信息为空")
	}
	
	// 转换为扩展库存信息
	extendedInfo := &types.ExtendedInventoryInfo{
		StockQuantity:    product.InventoryInfo.StockQuantity,
		ReservedQuantity: product.InventoryInfo.ReservedQuantity,
		TrackInventory:   product.InventoryInfo.TrackInventory,
	}
	
	return extendedInfo, nil
}

// AdjustInventory 调整商品库存（原子操作）
func (r *ProductRepository) AdjustInventory(ctx context.Context, productID uint64, adjustment int, reason string) (*types.ExtendedInventoryInfo, error) {
	tenantID := r.GetTenantID(ctx)
	merchantID := r.GetMerchantID(ctx)
	if tenantID == 0 || merchantID == 0 {
		return nil, fmt.Errorf("missing tenant_id or merchant_id in context")
	}
	
	// 在事务中执行库存调整
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	
	// 获取当前库存信息（行级锁）
	var product types.Product
	err = tx.Model("products").
		Where("id = ? AND tenant_id = ? AND merchant_id = ?", productID, tenantID, merchantID).
		LockUpdate().
		Scan(&product)
	if err != nil {
		return nil, err
	}
	
	if product.ID == 0 {
		return nil, fmt.Errorf("product not found")
	}
	
	// 计算新的库存数量
	oldQuantity := product.InventoryInfo.StockQuantity
	newQuantity := oldQuantity + adjustment
	
	// 检查库存不能为负数
	if newQuantity < 0 {
		return nil, fmt.Errorf("insufficient inventory: current=%d, adjustment=%d", oldQuantity, adjustment)
	}
	
	// 更新库存
	product.InventoryInfo.StockQuantity = newQuantity
	_, err = tx.Model("products").
		Where("id = ? AND tenant_id = ? AND merchant_id = ?", productID, tenantID, merchantID).
		Update(g.Map{
			"inventory_info": product.InventoryInfo,
			"version": gdb.Raw("version + 1"),
		})
	if err != nil {
		return nil, err
	}
	
	// 转换为扩展库存信息
	extendedInfo := &types.ExtendedInventoryInfo{
		StockQuantity:    product.InventoryInfo.StockQuantity,
		ReservedQuantity: product.InventoryInfo.ReservedQuantity,
		TrackInventory:   product.InventoryInfo.TrackInventory,
	}
	
	return extendedInfo, nil
}

// ReserveInventory 预留库存
func (r *ProductRepository) ReserveInventory(ctx context.Context, productID uint64, quantity int) error {
	tenantID := r.GetTenantID(ctx)
	merchantID := r.GetMerchantID(ctx)
	if tenantID == 0 || merchantID == 0 {
		return fmt.Errorf("missing tenant_id or merchant_id in context")
	}
	
	// 在事务中执行库存预留
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 获取当前库存信息（行级锁）
		var product types.Product
		err := tx.Model("products").
			Where("id = ? AND tenant_id = ? AND merchant_id = ?", productID, tenantID, merchantID).
			LockUpdate().
			Scan(&product)
		if err != nil {
			return err
		}
		
		if product.ID == 0 {
			return fmt.Errorf("product not found")
		}
		
		// 检查可用库存
		availableStock := product.InventoryInfo.StockQuantity - product.InventoryInfo.ReservedQuantity
		if availableStock < quantity {
			return fmt.Errorf("insufficient available inventory: available=%d, requested=%d", availableStock, quantity)
		}
		
		// 增加预留数量
		product.InventoryInfo.ReservedQuantity += quantity
		
		// 更新库存信息
		_, err = tx.Model("products").
			Where("id = ? AND tenant_id = ? AND merchant_id = ?", productID, tenantID, merchantID).
			Update(g.Map{
				"inventory_info": product.InventoryInfo,
				"version": gdb.Raw("version + 1"),
			})
		
		return err
	})
}

// ReleaseInventory 释放预留库存
func (r *ProductRepository) ReleaseInventory(ctx context.Context, productID uint64, quantity int) error {
	tenantID := r.GetTenantID(ctx)
	merchantID := r.GetMerchantID(ctx)
	if tenantID == 0 || merchantID == 0 {
		return fmt.Errorf("missing tenant_id or merchant_id in context")
	}
	
	// 在事务中执行库存释放
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 获取当前库存信息（行级锁）
		var product types.Product
		err := tx.Model("products").
			Where("id = ? AND tenant_id = ? AND merchant_id = ?", productID, tenantID, merchantID).
			LockUpdate().
			Scan(&product)
		if err != nil {
			return err
		}
		
		if product.ID == 0 {
			return fmt.Errorf("product not found")
		}
		
		// 检查预留数量是否足够
		if product.InventoryInfo.ReservedQuantity < quantity {
			return fmt.Errorf("insufficient reserved inventory: reserved=%d, release=%d", product.InventoryInfo.ReservedQuantity, quantity)
		}
		
		// 减少预留数量
		product.InventoryInfo.ReservedQuantity -= quantity
		
		// 更新库存信息
		_, err = tx.Model("products").
			Where("id = ? AND tenant_id = ? AND merchant_id = ?", productID, tenantID, merchantID).
			Update(g.Map{
				"inventory_info": product.InventoryInfo,
				"version": gdb.Raw("version + 1"),
			})
		
		return err
	})
}

// GetLowStockProducts 获取低库存商品
func (r *ProductRepository) GetLowStockProducts(ctx context.Context) ([]types.Product, error) {
	tenantID := r.GetTenantID(ctx)
	merchantID := r.GetMerchantID(ctx)
	if tenantID == 0 || merchantID == 0 {
		return nil, fmt.Errorf("missing tenant_id or merchant_id in context")
	}
	
	var products []types.Product
	err := g.DB().Model("products").
		Ctx(ctx).
		Where("tenant_id = ? AND merchant_id = ?", tenantID, merchantID).
		Where("status != ?", types.ProductStatusDeleted).
		Where("JSON_EXTRACT(inventory_info, '$.track_inventory') = true").
		Where("JSON_EXTRACT(inventory_info, '$.stock_quantity') - JSON_EXTRACT(inventory_info, '$.reserved_quantity') <= COALESCE(JSON_EXTRACT(inventory_info, '$.low_stock_threshold'), 0)").
		Scan(&products)
	if err != nil {
		return nil, err
	}
	
	return products, nil
}