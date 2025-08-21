package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// CategoryRepository 分类数据访问层
type CategoryRepository struct {
	*BaseRepository
}

// NewCategoryRepository 创建分类仓库实例
func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// Create 创建分类
func (r *CategoryRepository) Create(ctx context.Context, category *types.ProductCategory) error {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return fmt.Errorf("missing tenant_id in context")
	}
	
	category.TenantID = tenantID
	
	// 处理层级和路径
	if category.ParentID != nil {
		parent, err := r.GetByID(ctx, *category.ParentID)
		if err != nil {
			return fmt.Errorf("parent category not found: %v", err)
		}
		
		// 检查层级限制（最大5级）
		if parent.Level >= 5 {
			return fmt.Errorf("category level exceeds maximum depth of 5")
		}
		
		category.Level = parent.Level + 1
		category.Path = parent.Path + "/" + category.Name
	} else {
		category.Level = 1
		category.Path = category.Name
	}
	
	result, err := g.DB().Model("product_categories").Ctx(ctx).Insert(category)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	category.ID = uint64(id)
	return nil
}

// GetByID 根据ID获取分类
func (r *CategoryRepository) GetByID(ctx context.Context, id uint64) (*types.ProductCategory, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	var category types.ProductCategory
	err := g.DB().Model("product_categories").
		Ctx(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Scan(&category)
	
	if err != nil {
		return nil, err
	}
	
	if category.ID == 0 {
		return nil, fmt.Errorf("category not found")
	}
	
	return &category, nil
}

// Update 更新分类
func (r *CategoryRepository) Update(ctx context.Context, id uint64, updates map[string]interface{}) error {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return fmt.Errorf("missing tenant_id in context")
	}
	
	// 如果更新了名称或父分类，需要重新计算路径
	if name, hasName := updates["name"]; hasName || updates["parent_id"] != nil {
		category, err := r.GetByID(ctx, id)
		if err != nil {
			return err
		}
		
		newName := category.Name
		if hasName {
			newName = name.(string)
		}
		
		var newPath string
		var newLevel int
		
		if parentID, hasParent := updates["parent_id"]; hasParent {
			if parentID != nil {
				parent, err := r.GetByID(ctx, parentID.(uint64))
				if err != nil {
					return fmt.Errorf("parent category not found: %v", err)
				}
				
				// 检查是否会造成循环引用
				if r.wouldCreateCycle(ctx, id, parentID.(uint64)) {
					return fmt.Errorf("cannot set parent: would create circular reference")
				}
				
				// 检查层级限制
				if parent.Level >= 5 {
					return fmt.Errorf("category level exceeds maximum depth of 5")
				}
				
				newLevel = parent.Level + 1
				newPath = parent.Path + "/" + newName
			} else {
				newLevel = 1
				newPath = newName
			}
		} else {
			// 保持原有父分类，只更新名称
			if category.ParentID != nil {
				parent, err := r.GetByID(ctx, *category.ParentID)
				if err != nil {
					return err
				}
				newPath = parent.Path + "/" + newName
			} else {
				newPath = newName
			}
			newLevel = category.Level
		}
		
		updates["path"] = newPath
		updates["level"] = newLevel
		
		// 同时更新所有子分类的路径
		err = r.updateChildrenPaths(ctx, id, category.Path, newPath)
		if err != nil {
			return err
		}
	}
	
	result, err := g.DB().Model("product_categories").
		Ctx(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update(updates)
	
	if err != nil {
		return err
	}
	
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("category not found")
	}
	
	return nil
}

// Delete 删除分类
func (r *CategoryRepository) Delete(ctx context.Context, id uint64) error {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return fmt.Errorf("missing tenant_id in context")
	}
	
	// 检查是否有子分类
	count, err := g.DB().Model("product_categories").
		Ctx(ctx).
		Where("parent_id = ? AND tenant_id = ?", id, tenantID).
		Count()
	if err != nil {
		return err
	}
	
	if count > 0 {
		return fmt.Errorf("cannot delete category with children")
	}
	
	// 检查是否有商品使用此分类
	productCount, err := g.DB().Model("products").
		Ctx(ctx).
		Where("category_id = ? AND tenant_id = ?", id, tenantID).
		Where("status != ?", types.ProductStatusDeleted).
		Count()
	if err != nil {
		return err
	}
	
	if productCount > 0 {
		return fmt.Errorf("cannot delete category with associated products")
	}
	
	// 删除分类
	result, err := g.DB().Model("product_categories").
		Ctx(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete()
	
	if err != nil {
		return err
	}
	
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("category not found")
	}
	
	return nil
}

// GetTree 获取分类树
func (r *CategoryRepository) GetTree(ctx context.Context) ([]types.CategoryTreeResponse, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	var categories []types.ProductCategory
	err := g.DB().Model("product_categories").
		Ctx(ctx).
		Where("tenant_id = ?", tenantID).
		Where("status = ?", types.CategoryStatusActive).
		Order("level ASC, sort_order ASC, name ASC").
		Scan(&categories)
	
	if err != nil {
		return nil, err
	}
	
	return r.buildCategoryTree(categories, nil), nil
}

// GetChildren 获取指定分类的直接子分类
func (r *CategoryRepository) GetChildren(ctx context.Context, parentID uint64) ([]types.ProductCategory, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	var children []types.ProductCategory
	err := g.DB().Model("product_categories").
		Ctx(ctx).
		Where("parent_id = ? AND tenant_id = ?", parentID, tenantID).
		Where("status = ?", types.CategoryStatusActive).
		Order("sort_order ASC, name ASC").
		Scan(&children)
	
	if err != nil {
		return nil, err
	}
	
	return children, nil
}

// GetPath 获取分类的完整路径
func (r *CategoryRepository) GetPath(ctx context.Context, categoryID uint64) ([]types.ProductCategory, error) {
	category, err := r.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	
	// 解析路径获取完整的分类链
	pathParts := strings.Split(category.Path, "/")
	var path []types.ProductCategory
	var currentPath string
	
	for _, part := range pathParts {
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath += "/" + part
		}
		
		var cat types.ProductCategory
		err := g.DB().Model("product_categories").
			Ctx(ctx).
			Where("path = ? AND tenant_id = ?", currentPath, category.TenantID).
			Scan(&cat)
		
		if err == nil && cat.ID > 0 {
			path = append(path, cat)
		}
	}
	
	return path, nil
}

// buildCategoryTree 构建分类树结构
func (r *CategoryRepository) buildCategoryTree(categories []types.ProductCategory, parentID *uint64) []types.CategoryTreeResponse {
	var result []types.CategoryTreeResponse
	
	for _, category := range categories {
		// 匹配父分类
		if (parentID == nil && category.ParentID == nil) ||
			(parentID != nil && category.ParentID != nil && *category.ParentID == *parentID) {
			
			node := types.CategoryTreeResponse{
				ProductCategory: category,
			}
			
			// 递归构建子节点
			children := r.buildCategoryTree(categories, &category.ID)
			if len(children) > 0 {
				node.Children = children
			}
			
			result = append(result, node)
		}
	}
	
	return result
}

// wouldCreateCycle 检查设置父分类是否会造成循环引用
func (r *CategoryRepository) wouldCreateCycle(ctx context.Context, categoryID uint64, newParentID uint64) bool {
	if categoryID == newParentID {
		return true
	}
	
	// 检查新父分类的祖先分类中是否包含当前分类
	current := newParentID
	for {
		parent, err := r.GetByID(ctx, current)
		if err != nil || parent.ParentID == nil {
			break
		}
		
		if *parent.ParentID == categoryID {
			return true
		}
		
		current = *parent.ParentID
	}
	
	return false
}

// updateChildrenPaths 更新所有子分类的路径
func (r *CategoryRepository) updateChildrenPaths(ctx context.Context, parentID uint64, oldPath, newPath string) error {
	var children []types.ProductCategory
	err := g.DB().Model("product_categories").
		Ctx(ctx).
		Where("parent_id = ?", parentID).
		Scan(&children)
	
	if err != nil {
		return err
	}
	
	for _, child := range children {
		// 替换路径前缀
		childNewPath := strings.Replace(child.Path, oldPath, newPath, 1)
		
		_, err := g.DB().Model("product_categories").
			Ctx(ctx).
			Where("id = ?", child.ID).
			Update(map[string]interface{}{
				"path": childNewPath,
			})
		
		if err != nil {
			return err
		}
		
		// 递归更新子分类的子分类
		err = r.updateChildrenPaths(ctx, child.ID, child.Path, childNewPath)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// GetFlatList 获取分类扁平列表（用于下拉选择等场景）
func (r *CategoryRepository) GetFlatList(ctx context.Context) ([]types.ProductCategory, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	var categories []types.ProductCategory
	err := g.DB().Model("product_categories").
		Ctx(ctx).
		Where("tenant_id = ?", tenantID).
		Where("status = ?", types.CategoryStatusActive).
		Order("level ASC, sort_order ASC, name ASC").
		Scan(&categories)
	
	if err != nil {
		return nil, err
	}
	
	return categories, nil
}