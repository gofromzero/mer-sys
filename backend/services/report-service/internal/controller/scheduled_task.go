package controller

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"mer-sys/backend/services/report-service/internal/service"
	"mer-sys/backend/shared/middleware"
	"mer-sys/backend/shared/types"
)

// ScheduledTaskController 定时任务控制器
type ScheduledTaskController struct {
	scheduledTaskService service.IScheduledTaskService
}

// NewScheduledTaskController 创建定时任务控制器
func NewScheduledTaskController() *ScheduledTaskController {
	return &ScheduledTaskController{
		scheduledTaskService: service.NewScheduledTaskService(),
	}
}

// RegisterRoutes 注册路由
func (c *ScheduledTaskController) RegisterRoutes(group *ghttp.RouterGroup) {
	taskGroup := group.Group("/scheduled-tasks")

	// 应用认证和租户中间件
	taskGroup.Middleware(
		middleware.JWTAuth(),
		middleware.TenantContext(),
	)

	// 定时任务路由
	taskGroup.GET("/", c.GetScheduledTasks)           // 获取定时任务列表
	taskGroup.POST("/", c.CreateScheduledTask)        // 创建定时任务
	taskGroup.GET("/:id", c.GetScheduledTask)         // 获取定时任务详情
	taskGroup.PUT("/:id", c.UpdateScheduledTask)      // 更新定时任务
	taskGroup.DELETE("/:id", c.DeleteScheduledTask)   // 删除定时任务
	taskGroup.POST("/:id/toggle", c.ToggleScheduledTask) // 启用/禁用定时任务
	taskGroup.POST("/:id/execute", c.ExecuteScheduledTask) // 手动执行定时任务

	// 系统管理路由（仅管理员）
	adminGroup := group.Group("/admin/scheduled-tasks")
	adminGroup.Middleware(
		middleware.JWTAuth(),
		middleware.TenantContext(),
		// middleware.RequireRole("admin"), // 需要管理员角色
	)

	adminGroup.POST("/start-all", c.StartAllScheduledTasks) // 启动所有定时任务
	adminGroup.POST("/stop-all", c.StopAllScheduledTasks)   // 停止所有定时任务
}

// GetScheduledTasks 获取定时任务列表
func (c *ScheduledTaskController) GetScheduledTasks(r *ghttp.Request) {
	var req types.ScheduledTaskListRequest
	if err := r.Parse(&req); err != nil {
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	response, err := c.scheduledTaskService.GetScheduledTasks(r.Context(), &req)
	if err != nil {
		g.Log().Error(r.Context(), "获取定时任务列表失败", "error", err)
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取定时任务列表失败",
		})
		return
	}

	g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "success",
		"data":    response,
	})
}

// CreateScheduledTask 创建定时任务
func (c *ScheduledTaskController) CreateScheduledTask(r *ghttp.Request) {
	var req types.ScheduledTaskCreateRequest
	if err := r.Parse(&req); err != nil {
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	task, err := c.scheduledTaskService.CreateScheduledTask(r.Context(), &req)
	if err != nil {
		g.Log().Error(r.Context(), "创建定时任务失败", "error", err)
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "创建定时任务失败: " + err.Error(),
		})
		return
	}

	g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "定时任务创建成功",
		"data":    task,
	})
}

// GetScheduledTask 获取定时任务详情
func (c *ScheduledTaskController) GetScheduledTask(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "任务ID参数无效",
		})
		return
	}

	task, err := c.scheduledTaskService.GetScheduledTask(r.Context(), taskID)
	if err != nil {
		g.Log().Error(r.Context(), "获取定时任务详情失败", "taskID", taskID, "error", err)
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取定时任务详情失败",
		})
		return
	}

	g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "success",
		"data":    task,
	})
}

// UpdateScheduledTask 更新定时任务
func (c *ScheduledTaskController) UpdateScheduledTask(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "任务ID参数无效",
		})
		return
	}

	var req types.ScheduledTaskUpdateRequest
	if err := r.Parse(&req); err != nil {
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	err = c.scheduledTaskService.UpdateScheduledTask(r.Context(), taskID, &req)
	if err != nil {
		g.Log().Error(r.Context(), "更新定时任务失败", "taskID", taskID, "error", err)
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新定时任务失败: " + err.Error(),
		})
		return
	}

	g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "定时任务更新成功",
	})
}

// DeleteScheduledTask 删除定时任务
func (c *ScheduledTaskController) DeleteScheduledTask(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "任务ID参数无效",
		})
		return
	}

	err = c.scheduledTaskService.DeleteScheduledTask(r.Context(), taskID)
	if err != nil {
		g.Log().Error(r.Context(), "删除定时任务失败", "taskID", taskID, "error", err)
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "删除定时任务失败",
		})
		return
	}

	g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "定时任务删除成功",
	})
}

// ToggleScheduledTask 启用/禁用定时任务
func (c *ScheduledTaskController) ToggleScheduledTask(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "任务ID参数无效",
		})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := r.Parse(&req); err != nil {
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	err = c.scheduledTaskService.ToggleScheduledTask(r.Context(), taskID, req.Enabled)
	if err != nil {
		g.Log().Error(r.Context(), "切换定时任务状态失败", "taskID", taskID, "enabled", req.Enabled, "error", err)
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "切换定时任务状态失败",
		})
		return
	}

	action := "启用"
	if !req.Enabled {
		action = "禁用"
	}

	g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "定时任务" + action + "成功",
	})
}

// ExecuteScheduledTask 手动执行定时任务
func (c *ScheduledTaskController) ExecuteScheduledTask(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "任务ID参数无效",
		})
		return
	}

	err = c.scheduledTaskService.ExecuteScheduledTask(r.Context(), taskID)
	if err != nil {
		g.Log().Error(r.Context(), "手动执行定时任务失败", "taskID", taskID, "error", err)
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "执行定时任务失败: " + err.Error(),
		})
		return
	}

	g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "定时任务执行成功",
	})
}

// StartAllScheduledTasks 启动所有定时任务
func (c *ScheduledTaskController) StartAllScheduledTasks(r *ghttp.Request) {
	err := c.scheduledTaskService.StartAllScheduledTasks(r.Context())
	if err != nil {
		g.Log().Error(r.Context(), "启动所有定时任务失败", "error", err)
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "启动所有定时任务失败: " + err.Error(),
		})
		return
	}

	g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "所有定时任务启动成功",
	})
}

// StopAllScheduledTasks 停止所有定时任务
func (c *ScheduledTaskController) StopAllScheduledTasks(r *ghttp.Request) {
	err := c.scheduledTaskService.StopAllScheduledTasks(r.Context())
	if err != nil {
		g.Log().Error(r.Context(), "停止所有定时任务失败", "error", err)
		g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "停止所有定时任务失败",
		})
		return
	}

	g.RequestFromCtx(r.Context()).Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "所有定时任务已停止",
	})
}