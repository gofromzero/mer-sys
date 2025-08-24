package service

import (
	"context"
	"fmt"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// ICartService 购物车服务接口
type ICartService interface {
	GetCart(ctx context.Context, customerID uint64) (*types.Cart, error)
	AddItem(ctx context.Context, customerID uint64, productID uint64, quantity int) error
	UpdateItemQuantity(ctx context.Context, itemID uint64, quantity int) error
	RemoveItem(ctx context.Context, itemID uint64) error
	ClearCart(ctx context.Context, customerID uint64) error
	GetCartWithProductDetails(ctx context.Context, customerID uint64) (*types.Cart, error)
}

// CartService 购物车服务实现
type CartService struct {
	cartRepo repository.ICartRepository
}

// NewCartService 创建购物车服务实例
func NewCartService() ICartService {
	return &CartService{
		cartRepo: repository.NewCartRepository(),
	}
}

// GetCart 获取购物车
func (s *CartService) GetCart(ctx context.Context, customerID uint64) (*types.Cart, error) {
	return s.cartRepo.GetOrCreate(ctx, customerID)
}

// AddItem 添加商品到购物车
func (s *CartService) AddItem(ctx context.Context, customerID uint64, productID uint64, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("商品数量必须大于0")
	}

	// TODO: 调用Product Service验证商品是否存在和可用
	// 暂时跳过商品验证

	// 获取或创建购物车
	cart, err := s.cartRepo.GetOrCreate(ctx, customerID)
	if err != nil {
		return fmt.Errorf("获取购物车失败: %v", err)
	}

	// 添加商品到购物车
	return s.cartRepo.AddItem(ctx, cart.ID, productID, quantity)
}

// UpdateItemQuantity 更新购物车商品数量
func (s *CartService) UpdateItemQuantity(ctx context.Context, itemID uint64, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("商品数量必须大于0")
	}

	return s.cartRepo.UpdateItemQuantity(ctx, itemID, quantity)
}

// RemoveItem 从购物车中移除商品
func (s *CartService) RemoveItem(ctx context.Context, itemID uint64) error {
	return s.cartRepo.RemoveItem(ctx, itemID)
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(ctx context.Context, customerID uint64) error {
	// 获取购物车
	cart, err := s.cartRepo.GetOrCreate(ctx, customerID)
	if err != nil {
		return fmt.Errorf("获取购物车失败: %v", err)
	}

	return s.cartRepo.ClearCart(ctx, cart.ID)
}

// GetCartWithProductDetails 获取带商品详情的购物车
func (s *CartService) GetCartWithProductDetails(ctx context.Context, customerID uint64) (*types.Cart, error) {
	cart, err := s.cartRepo.GetOrCreate(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("获取购物车失败: %v", err)
	}

	// TODO: 这里应该调用Product Service获取商品详情
	// 暂时返回基础购物车信息

	return cart, nil
}
