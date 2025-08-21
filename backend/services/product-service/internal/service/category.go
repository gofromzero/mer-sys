package service

import (
	"context"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// CategoryService 分类服务
type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

// NewCategoryService 创建分类服务实例
func NewCategoryService() *CategoryService {
	return &CategoryService{
		categoryRepo: repository.NewCategoryRepository(),
	}
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(ctx context.Context, req *types.CreateCategoryRequest) (*types.ProductCategory, error) {
	category := &types.ProductCategory{
		Name:      req.Name,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
		Status:    types.CategoryStatusActive, // 默认为启用状态
	}
	
	err := s.categoryRepo.Create(ctx, category)
	if err != nil {
		return nil, err
	}
	
	return category, nil
}

// GetCategory 获取分类详情
func (s *CategoryService) GetCategory(ctx context.Context, id uint64) (*types.ProductCategory, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(ctx context.Context, id uint64, req *types.UpdateCategoryRequest) (*types.ProductCategory, error) {
	// 构建更新字段
	updates := make(map[string]interface{})
	
	if req.Name != "" {
		updates["name"] = req.Name
	}
	
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	
	if len(updates) == 0 {
		// 没有任何更新，直接返回原数据
		return s.categoryRepo.GetByID(ctx, id)
	}
	
	// 执行更新
	err := s.categoryRepo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}
	
	// 返回更新后的分类信息
	return s.categoryRepo.GetByID(ctx, id)
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(ctx context.Context, id uint64) error {
	return s.categoryRepo.Delete(ctx, id)
}

// GetCategoryTree 获取分类树
func (s *CategoryService) GetCategoryTree(ctx context.Context) ([]types.CategoryTreeResponse, error) {
	return s.categoryRepo.GetTree(ctx)
}

// GetCategoryList 获取分类扁平列表
func (s *CategoryService) GetCategoryList(ctx context.Context) ([]types.ProductCategory, error) {
	return s.categoryRepo.GetFlatList(ctx)
}

// GetCategoryChildren 获取分类的子分类
func (s *CategoryService) GetCategoryChildren(ctx context.Context, parentID uint64) ([]types.ProductCategory, error) {
	return s.categoryRepo.GetChildren(ctx, parentID)
}

// GetCategoryPath 获取分类路径
func (s *CategoryService) GetCategoryPath(ctx context.Context, categoryID uint64) ([]types.ProductCategory, error) {
	return s.categoryRepo.GetPath(ctx, categoryID)
}