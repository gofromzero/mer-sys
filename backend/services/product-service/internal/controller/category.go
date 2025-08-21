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

// CategoryController 分类控制器
type CategoryController struct {
	categoryService *service.CategoryService
}

// NewCategoryController 创建分类控制器实例
func NewCategoryController() *CategoryController {
	return &CategoryController{
		categoryService: service.NewCategoryService(),
	}
}

// CreateCategory 创建分类
func (c *CategoryController) CreateCategory(r *ghttp.Request) {
	var req types.CreateCategoryRequest
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
	if err := validateCreateCategoryRequest(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数验证失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	category, err := c.categoryService.CreateCategory(r.GetCtx(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "创建分类失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "创建成功",
		"data":    category,
	})
}

// GetCategory 获取分类详情
func (c *CategoryController) GetCategory(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的分类ID",
			"data":    nil,
		})
		return
	}
	
	category, err := c.categoryService.GetCategory(r.GetCtx(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    404,
			"message": "分类不存在",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    category,
	})
}

// UpdateCategory 更新分类
func (c *CategoryController) UpdateCategory(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的分类ID",
			"data":    nil,
		})
		return
	}
	
	var req types.UpdateCategoryRequest
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
	if err := validateUpdateCategoryRequest(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数验证失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	category, err := c.categoryService.UpdateCategory(r.GetCtx(), id, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新分类失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "更新成功",
		"data":    category,
	})
}

// DeleteCategory 删除分类
func (c *CategoryController) DeleteCategory(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的分类ID",
			"data":    nil,
		})
		return
	}
	
	err = c.categoryService.DeleteCategory(r.GetCtx(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "删除分类失败",
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

// GetCategoryTree 获取分类树
func (c *CategoryController) GetCategoryTree(r *ghttp.Request) {
	tree, err := c.categoryService.GetCategoryTree(r.GetCtx())
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取分类树失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    tree,
	})
}

// GetCategoryList 获取分类扁平列表
func (c *CategoryController) GetCategoryList(r *ghttp.Request) {
	categories, err := c.categoryService.GetCategoryList(r.GetCtx())
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取分类列表失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    categories,
	})
}

// validateCreateCategoryRequest 验证创建分类请求
func validateCreateCategoryRequest(req *types.CreateCategoryRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("分类名称不能为空")
	}
	if len(req.Name) > 100 {
		return fmt.Errorf("分类名称长度不能超过100个字符")
	}
	return nil
}

// validateUpdateCategoryRequest 验证更新分类请求
func validateUpdateCategoryRequest(req *types.UpdateCategoryRequest) error {
	if req.Name != "" {
		if strings.TrimSpace(req.Name) == "" {
			return fmt.Errorf("分类名称不能为空")
		}
		if len(req.Name) > 100 {
			return fmt.Errorf("分类名称长度不能超过100个字符")
		}
	}
	if req.Status != nil {
		if *req.Status != types.CategoryStatusActive && *req.Status != types.CategoryStatusInactive {
			return fmt.Errorf("无效的分类状态")
		}
	}
	return nil
}

// GetCategoryChildren 获取分类的子分类
func (c *CategoryController) GetCategoryChildren(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的分类ID",
			"data":    nil,
		})
		return
	}
	
	children, err := c.categoryService.GetCategoryChildren(r.GetCtx(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取子分类失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    children,
	})
}

// GetCategoryPath 获取分类路径
func (c *CategoryController) GetCategoryPath(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的分类ID",
			"data":    nil,
		})
		return
	}
	
	path, err := c.categoryService.GetCategoryPath(r.GetCtx(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取分类路径失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    path,
	})
}