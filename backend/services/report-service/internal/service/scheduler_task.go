package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/util/gconv"

	"mer-sys/backend/shared/repository"
	"mer-sys/backend/shared/types"
)

// IScheduledTaskService 定时任务服务接口
type IScheduledTaskService interface {
	// 创建定时任务
	CreateScheduledTask(ctx context.Context, req *types.ScheduledTaskCreateRequest) (*types.ScheduledTask, error)
	// 更新定时任务
	UpdateScheduledTask(ctx context.Context, taskID int64, req *types.ScheduledTaskUpdateRequest) error
	// 删除定时任务
	DeleteScheduledTask(ctx context.Context, taskID int64) error
	// 启用/禁用定时任务
	ToggleScheduledTask(ctx context.Context, taskID int64, enabled bool) error
	// 获取定时任务列表
	GetScheduledTasks(ctx context.Context, req *types.ScheduledTaskListRequest) (*types.ScheduledTaskListResponse, error)
	// 获取定时任务详情
	GetScheduledTask(ctx context.Context, taskID int64) (*types.ScheduledTask, error)
	// 执行定时任务
	ExecuteScheduledTask(ctx context.Context, taskID int64) error
	// 启动所有定时任务
	StartAllScheduledTasks(ctx context.Context) error
	// 停止所有定时任务
	StopAllScheduledTasks(ctx context.Context) error
}

// scheduledTaskService 定时任务服务实现
type scheduledTaskService struct {
	scheduledTaskRepo repository.IScheduledTaskRepository
	reportGenerator   IReportGeneratorService
	cron              *gcron.Cron
	taskJobs          map[int64]string // taskID -> jobName 的映射
}

// NewScheduledTaskService 创建定时任务服务
func NewScheduledTaskService() IScheduledTaskService {
	return &scheduledTaskService{
		scheduledTaskRepo: repository.NewScheduledTaskRepository(),
		reportGenerator:   NewReportGeneratorService(),
		cron:              gcron.New(),
		taskJobs:          make(map[int64]string),
	}
}

// CreateScheduledTask 创建定时任务
func (s *scheduledTaskService) CreateScheduledTask(ctx context.Context, req *types.ScheduledTaskCreateRequest) (*types.ScheduledTask, error) {
	g.Log().Info(ctx, "创建定时任务", "request", req)

	// 验证cron表达式
	if err := s.validateCronExpression(req.CronExpression); err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	// 验证报表配置
	if err := s.validateReportConfig(req.ReportConfig); err != nil {
		return nil, fmt.Errorf("invalid report config: %w", err)
	}

	// 创建定时任务记录
	task, err := s.scheduledTaskRepo.Create(ctx, &types.ScheduledTask{
		TaskName:        req.TaskName,
		TaskDescription: req.TaskDescription,
		ReportType:      req.ReportType,
		CronExpression:  req.CronExpression,
		ReportConfig:    req.ReportConfig,
		Recipients:      req.Recipients,
		IsEnabled:       req.IsEnabled,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("创建定时任务失败: %w", err)
	}

	// 如果任务启用，则添加到定时器
	if task.IsEnabled {
		if err := s.addCronJob(ctx, task); err != nil {
			g.Log().Error(ctx, "添加定时任务失败", "taskID", task.ID, "error", err)
			// 不返回错误，允许手动执行
		}
	}

	g.Log().Info(ctx, "定时任务创建成功", "taskID", task.ID, "taskName", task.TaskName)
	return task, nil
}

// UpdateScheduledTask 更新定时任务
func (s *scheduledTaskService) UpdateScheduledTask(ctx context.Context, taskID int64, req *types.ScheduledTaskUpdateRequest) error {
	g.Log().Info(ctx, "更新定时任务", "taskID", taskID, "request", req)

	// 获取现有任务
	existingTask, err := s.scheduledTaskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取定时任务失败: %w", err)
	}

	// 验证更新内容
	if req.CronExpression != "" {
		if err := s.validateCronExpression(req.CronExpression); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
	}

	if req.ReportConfig != "" {
		if err := s.validateReportConfig(req.ReportConfig); err != nil {
			return fmt.Errorf("invalid report config: %w", err)
		}
	}

	// 停止现有的定时任务
	s.removeCronJob(existingTask.ID)

	// 更新任务
	updateData := &types.ScheduledTask{
		ID:        taskID,
		UpdatedAt: time.Now(),
	}

	if req.TaskName != "" {
		updateData.TaskName = req.TaskName
	}
	if req.TaskDescription != "" {
		updateData.TaskDescription = req.TaskDescription
	}
	if req.CronExpression != "" {
		updateData.CronExpression = req.CronExpression
	}
	if req.ReportConfig != "" {
		updateData.ReportConfig = req.ReportConfig
	}
	if len(req.Recipients) > 0 {
		updateData.Recipients = req.Recipients
	}
	if req.IsEnabled != nil {
		updateData.IsEnabled = *req.IsEnabled
	}

	err = s.scheduledTaskRepo.Update(ctx, updateData)
	if err != nil {
		return fmt.Errorf("更新定时任务失败: %w", err)
	}

	// 重新获取更新后的任务
	updatedTask, err := s.scheduledTaskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取更新后的定时任务失败: %w", err)
	}

	// 如果任务启用，则重新添加到定时器
	if updatedTask.IsEnabled {
		if err := s.addCronJob(ctx, updatedTask); err != nil {
			g.Log().Error(ctx, "重新添加定时任务失败", "taskID", taskID, "error", err)
		}
	}

	g.Log().Info(ctx, "定时任务更新成功", "taskID", taskID)
	return nil
}

// DeleteScheduledTask 删除定时任务
func (s *scheduledTaskService) DeleteScheduledTask(ctx context.Context, taskID int64) error {
	g.Log().Info(ctx, "删除定时任务", "taskID", taskID)

	// 停止定时任务
	s.removeCronJob(taskID)

	// 删除任务记录
	err := s.scheduledTaskRepo.Delete(ctx, taskID)
	if err != nil {
		return fmt.Errorf("删除定时任务失败: %w", err)
	}

	g.Log().Info(ctx, "定时任务删除成功", "taskID", taskID)
	return nil
}

// ToggleScheduledTask 启用/禁用定时任务
func (s *scheduledTaskService) ToggleScheduledTask(ctx context.Context, taskID int64, enabled bool) error {
	g.Log().Info(ctx, "切换定时任务状态", "taskID", taskID, "enabled", enabled)

	// 获取任务
	task, err := s.scheduledTaskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取定时任务失败: %w", err)
	}

	// 更新启用状态
	err = s.scheduledTaskRepo.Update(ctx, &types.ScheduledTask{
		ID:        taskID,
		IsEnabled: enabled,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("更新定时任务状态失败: %w", err)
	}

	// 管理定时器
	if enabled {
		task.IsEnabled = true
		if err := s.addCronJob(ctx, task); err != nil {
			g.Log().Error(ctx, "启用定时任务失败", "taskID", taskID, "error", err)
		}
	} else {
		s.removeCronJob(taskID)
	}

	g.Log().Info(ctx, "定时任务状态切换成功", "taskID", taskID, "enabled", enabled)
	return nil
}

// GetScheduledTasks 获取定时任务列表
func (s *scheduledTaskService) GetScheduledTasks(ctx context.Context, req *types.ScheduledTaskListRequest) (*types.ScheduledTaskListResponse, error) {
	tasks, total, err := s.scheduledTaskRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取定时任务列表失败: %w", err)
	}

	return &types.ScheduledTaskListResponse{
		Tasks: tasks,
		Pagination: types.Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetScheduledTask 获取定时任务详情
func (s *scheduledTaskService) GetScheduledTask(ctx context.Context, taskID int64) (*types.ScheduledTask, error) {
	return s.scheduledTaskRepo.GetByID(ctx, taskID)
}

// ExecuteScheduledTask 执行定时任务
func (s *scheduledTaskService) ExecuteScheduledTask(ctx context.Context, taskID int64) error {
	g.Log().Info(ctx, "手动执行定时任务", "taskID", taskID)

	// 获取任务详情
	task, err := s.scheduledTaskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取定时任务失败: %w", err)
	}

	return s.executeTask(ctx, task)
}

// StartAllScheduledTasks 启动所有定时任务
func (s *scheduledTaskService) StartAllScheduledTasks(ctx context.Context) error {
	g.Log().Info(ctx, "启动所有定时任务")

	// 获取所有启用的定时任务
	tasks, _, err := s.scheduledTaskRepo.List(ctx, &types.ScheduledTaskListRequest{
		IsEnabled: gconv.BoolPtr(true),
		Page:      1,
		PageSize:  1000, // 假设不会超过1000个任务
	})
	if err != nil {
		return fmt.Errorf("获取启用的定时任务失败: %w", err)
	}

	// 为每个任务添加定时器
	var errorCount int
	for _, task := range tasks {
		if err := s.addCronJob(ctx, task); err != nil {
			g.Log().Error(ctx, "添加定时任务失败", "taskID", task.ID, "error", err)
			errorCount++
		}
	}

	// 启动定时器
	s.cron.Start()

	if errorCount > 0 {
		return fmt.Errorf("启动定时任务时有 %d 个任务失败", errorCount)
	}

	g.Log().Info(ctx, "所有定时任务启动成功", "count", len(tasks))
	return nil
}

// StopAllScheduledTasks 停止所有定时任务
func (s *scheduledTaskService) StopAllScheduledTasks(ctx context.Context) error {
	g.Log().Info(ctx, "停止所有定时任务")

	s.cron.Stop()
	s.taskJobs = make(map[int64]string) // 清空任务映射

	g.Log().Info(ctx, "所有定时任务已停止")
	return nil
}

// validateCronExpression 验证cron表达式
func (s *scheduledTaskService) validateCronExpression(cronExpr string) error {
	_, err := gcron.New().Add(context.Background(), cronExpr, func(ctx context.Context) {}, "validation")
	if err != nil {
		return err
	}
	return nil
}

// validateReportConfig 验证报表配置
func (s *scheduledTaskService) validateReportConfig(configStr string) error {
	var config map[string]interface{}
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// 检查必需的字段
	if _, ok := config["start_date"]; !ok {
		return fmt.Errorf("missing required field: start_date")
	}
	if _, ok := config["end_date"]; !ok {
		return fmt.Errorf("missing required field: end_date")
	}

	return nil
}

// addCronJob 添加定时任务到定时器
func (s *scheduledTaskService) addCronJob(ctx context.Context, task *types.ScheduledTask) error {
	jobName := fmt.Sprintf("scheduled_task_%d", task.ID)

	_, err := s.cron.Add(ctx, task.CronExpression, func(ctx context.Context) {
		if err := s.executeTask(ctx, task); err != nil {
			g.Log().Error(ctx, "执行定时任务失败", "taskID", task.ID, "error", err)
		}
	}, jobName)

	if err != nil {
		return fmt.Errorf("添加定时任务失败: %w", err)
	}

	// 记录任务映射
	s.taskJobs[task.ID] = jobName

	g.Log().Info(ctx, "定时任务已添加到定时器", "taskID", task.ID, "jobName", jobName, "cron", task.CronExpression)
	return nil
}

// removeCronJob 从定时器中移除定时任务
func (s *scheduledTaskService) removeCronJob(taskID int64) {
	if jobName, exists := s.taskJobs[taskID]; exists {
		s.cron.Remove(jobName)
		delete(s.taskJobs, taskID)
		g.Log().Info(context.Background(), "定时任务已从定时器移除", "taskID", taskID, "jobName", jobName)
	}
}

// executeTask 执行任务
func (s *scheduledTaskService) executeTask(ctx context.Context, task *types.ScheduledTask) error {
	g.Log().Info(ctx, "开始执行定时任务", "taskID", task.ID, "taskName", task.TaskName)

	// 更新任务执行状态
	err := s.scheduledTaskRepo.UpdateLastExecution(ctx, task.ID, types.TaskStatusRunning, "")
	if err != nil {
		g.Log().Error(ctx, "更新任务执行状态失败", "taskID", task.ID, "error", err)
	}

	// 解析报表配置
	var reportConfig map[string]interface{}
	if err := json.Unmarshal([]byte(task.ReportConfig), &reportConfig); err != nil {
		s.handleTaskFailure(ctx, task.ID, fmt.Sprintf("解析报表配置失败: %v", err))
		return fmt.Errorf("解析报表配置失败: %w", err)
	}

	// 构建报表生成请求
	reportReq := &types.ReportCreateRequest{
		ReportType: task.ReportType,
		StartDate:  gconv.String(reportConfig["start_date"]),
		EndDate:    gconv.String(reportConfig["end_date"]),
		Filters:    gconv.Map(reportConfig["filters"]),
		Format:     gconv.String(reportConfig["format"]),
		MerchantID: gconv.Int64(reportConfig["merchant_id"]),
	}

	// 生成报表
	report, err := s.reportGenerator.GenerateReport(ctx, reportReq)
	if err != nil {
		s.handleTaskFailure(ctx, task.ID, fmt.Sprintf("生成报表失败: %v", err))
		return fmt.Errorf("生成报表失败: %w", err)
	}

	// 这里可以添加报表发送逻辑（邮件、钉钉、企业微信等）
	// TODO: 实现报表发送功能

	// 更新任务执行成功状态
	err = s.scheduledTaskRepo.UpdateLastExecution(ctx, task.ID, types.TaskStatusCompleted, 
		fmt.Sprintf("报表生成成功，报表UUID: %s", report.UUID))
	if err != nil {
		g.Log().Error(ctx, "更新任务执行成功状态失败", "taskID", task.ID, "error", err)
	}

	g.Log().Info(ctx, "定时任务执行成功", "taskID", task.ID, "reportUUID", report.UUID)
	return nil
}

// handleTaskFailure 处理任务执行失败
func (s *scheduledTaskService) handleTaskFailure(ctx context.Context, taskID int64, errorMsg string) {
	err := s.scheduledTaskRepo.UpdateLastExecution(ctx, taskID, types.TaskStatusFailed, errorMsg)
	if err != nil {
		g.Log().Error(ctx, "更新任务失败状态失败", "taskID", taskID, "error", err)
	}
}