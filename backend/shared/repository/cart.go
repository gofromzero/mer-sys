package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ICartRepository 购物车仓储接口
type ICartRepository interface {
	GetOrCreate(ctx context.Context, customerID uint64) (*types.Cart, error)
	AddItem(ctx context.Context, cartID uint64, productID uint64, quantity int) error
	UpdateItemQuantity(ctx context.Context, itemID uint64, quantity int) error
	RemoveItem(ctx context.Context, itemID uint64) error
	ClearCart(ctx context.Context, cartID uint64) error
	GetCartItems(ctx context.Context, cartID uint64) ([]types.CartItem, error)
	GetItemByProductID(ctx context.Context, cartID uint64, productID uint64) (*types.CartItem, error)
	CleanExpiredCarts(ctx context.Context) error
}

// CartRepository 购物车仓储实现
type CartRepository struct {
	*BaseRepository
}

// NewCartRepository 创建购物车仓储实例
func NewCartRepository() ICartRepository {
	return &CartRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// GetOrCreate 获取或创建购物车
func (r *CartRepository) GetOrCreate(ctx context.Context, customerID uint64) (*types.Cart, error) {
	tenantID := r.GetTenantID(ctx)
	
	// 先尝试获取现有购物车
	var cart types.Cart
	err := g.DB().Model("carts").Ctx(ctx).
		Where("tenant_id = ? AND customer_id = ?", tenantID, customerID).
		Scan(&cart)
	
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("查询购物车失败: %v", err)
	}
	
	// 如果购物车不存在，创建新的
	if err == sql.ErrNoRows {
		expiresAt := gtime.Now().Add(7 * 24 * 3600) // 7天后过期
		result, err := g.DB().Model("carts").Ctx(ctx).Insert(gdb.Map{
			"tenant_id":   tenantID,
			"customer_id": customerID,
			"created_at":  gtime.Now(),
			"updated_at":  gtime.Now(),
			"expires_at":  expiresAt,
		})
		
		if err != nil {
			return nil, fmt.Errorf("创建购物车失败: %v", err)
		}
		
		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("获取购物车ID失败: %v", err)
		}
		
		cart = types.Cart{
			ID:         uint64(id),
			TenantID:   tenantID,
			CustomerID: customerID,
			CreatedAt:  gtime.Now().Time,
			UpdatedAt:  gtime.Now().Time,
			ExpiresAt:  expiresAt.Time,
		}
	}
	
	// 获取购物车项
	items, err := r.GetCartItems(ctx, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("获取购物车项失败: %v", err)
	}
	cart.Items = items
	
	return &cart, nil
}

// AddItem 添加商品到购物车
func (r *CartRepository) AddItem(ctx context.Context, cartID uint64, productID uint64, quantity int) error {
	tenantID := r.GetTenantID(ctx)
	
	// 检查商品是否已在购物车中
	existingItem, err := r.GetItemByProductID(ctx, cartID, productID)
	if err != nil && err.Error() != "购物车项不存在" {
		return fmt.Errorf("检查购物车项失败: %v", err)
	}
	
	if existingItem != nil {
		// 如果已存在，更新数量
		return r.UpdateItemQuantity(ctx, existingItem.ID, existingItem.Quantity+quantity)
	}
	
	// 添加新项
	_, err = g.DB().Model("cart_items").Ctx(ctx).Insert(gdb.Map{
		"tenant_id":  tenantID,
		"cart_id":    cartID,
		"product_id": productID,
		"quantity":   quantity,
		"added_at":   gtime.Now(),
	})
	
	if err != nil {
		return fmt.Errorf("添加购物车项失败: %v", err)
	}
	
	// 更新购物车的更新时间
	_, err = g.DB().Model("carts").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", cartID, tenantID).
		Update(gdb.Map{"updated_at": gtime.Now()})
	
	return err
}

// UpdateItemQuantity 更新购物车项数量
func (r *CartRepository) UpdateItemQuantity(ctx context.Context, itemID uint64, quantity int) error {
	tenantID := r.GetTenantID(ctx)
	
	_, err := g.DB().Model("cart_items").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", itemID, tenantID).
		Update(gdb.Map{"quantity": quantity})
	
	if err != nil {
		return fmt.Errorf("更新购物车项数量失败: %v", err)
	}
	
	return nil
}

// RemoveItem 从购物车中移除商品
func (r *CartRepository) RemoveItem(ctx context.Context, itemID uint64) error {
	tenantID := r.GetTenantID(ctx)
	
	_, err := g.DB().Model("cart_items").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", itemID, tenantID).
		Delete()
	
	if err != nil {
		return fmt.Errorf("删除购物车项失败: %v", err)
	}
	
	return nil
}

// ClearCart 清空购物车
func (r *CartRepository) ClearCart(ctx context.Context, cartID uint64) error {
	tenantID := r.GetTenantID(ctx)
	
	_, err := g.DB().Model("cart_items").Ctx(ctx).
		Where("cart_id = ? AND tenant_id = ?", cartID, tenantID).
		Delete()
	
	if err != nil {
		return fmt.Errorf("清空购物车失败: %v", err)
	}
	
	// 更新购物车的更新时间
	_, err = g.DB().Model("carts").Ctx(ctx).
		Where("id = ? AND tenant_id = ?", cartID, tenantID).
		Update(gdb.Map{"updated_at": gtime.Now()})
	
	return err
}

// GetCartItems 获取购物车项列表
func (r *CartRepository) GetCartItems(ctx context.Context, cartID uint64) ([]types.CartItem, error) {
	tenantID := r.GetTenantID(ctx)
	
	var items []types.CartItem
	err := g.DB().Model("cart_items").Ctx(ctx).
		Where("cart_id = ? AND tenant_id = ?", cartID, tenantID).
		Order("added_at ASC").
		Scan(&items)
	
	if err != nil {
		return nil, fmt.Errorf("查询购物车项失败: %v", err)
	}
	
	return items, nil
}

// GetItemByProductID 根据商品ID获取购物车项
func (r *CartRepository) GetItemByProductID(ctx context.Context, cartID uint64, productID uint64) (*types.CartItem, error) {
	tenantID := r.GetTenantID(ctx)
	
	var item types.CartItem
	err := g.DB().Model("cart_items").Ctx(ctx).
		Where("cart_id = ? AND product_id = ? AND tenant_id = ?", cartID, productID, tenantID).
		Scan(&item)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("购物车项不存在")
		}
		return nil, fmt.Errorf("查询购物车项失败: %v", err)
	}
	
	return &item, nil
}

// CleanExpiredCarts 清理过期的购物车
func (r *CartRepository) CleanExpiredCarts(ctx context.Context) error {
	// 删除过期的购物车项
	_, err := g.DB().Model("cart_items").Ctx(ctx).
		Where("cart_id IN (SELECT id FROM carts WHERE expires_at < ?)", gtime.Now()).
		Delete()
	
	if err != nil {
		return fmt.Errorf("删除过期购物车项失败: %v", err)
	}
	
	// 删除过期的购物车
	_, err = g.DB().Model("carts").Ctx(ctx).
		Where("expires_at < ?", gtime.Now()).
		Delete()
	
	if err != nil {
		return fmt.Errorf("删除过期购物车失败: %v", err)
	}
	
	return nil
}