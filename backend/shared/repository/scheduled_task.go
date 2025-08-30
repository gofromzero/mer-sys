package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"mer-sys/backend/shared/types"
)

// IScheduledTaskRepository 定时任务仓储接口
type IScheduledTaskRepository interface {
	// 创建定时任务
	Create(ctx context.Context, task *types.ScheduledTask) (*types.ScheduledTask, error)
	// 更新定时任务
	Update(ctx context.Context, task *types.ScheduledTask) error
	// 删除定时任务
	Delete(ctx context.Context, taskID int64) error
	// 根据ID获取定时任务
	GetByID(ctx context.Context, taskID int64) (*types.ScheduledTask, error)
	// 获取定时任务列表
	List(ctx context.Context, req *types.ScheduledTaskListRequest) ([]*types.ScheduledTask, int, error)
	// 更新最后执行状态
	UpdateLastExecution(ctx context.Context, taskID int64, status types.TaskStatus, message string) error
}

// scheduledTaskRepository 定时任务仓储实现
type scheduledTaskRepository struct {
	*BaseRepository
}

// NewScheduledTaskRepository 创建定时任务仓储实例
func NewScheduledTaskRepository() IScheduledTaskRepository {
	return &scheduledTaskRepository{
		BaseRepository: NewBaseRepository("scheduled_tasks"),
	}
}

// Create 创建定时任务
func (r *scheduledTaskRepository) Create(ctx context.Context, task *types.ScheduledTask) (*types.ScheduledTask, error) {
	tenantID := r.GetTenantID(ctx)

	data := g.Map{
		"tenant_id":         tenantID,
		"task_name":         task.TaskName,
		"task_description":  task.TaskDescription,
		"report_type":       task.ReportType,
		"cron_expression":   task.CronExpression,
		"report_config":     task.ReportConfig,
		"recipients":        r.encodeJSON(task.Recipients),
		"is_enabled":        task.IsEnabled,
		"last_run_time":     nil,
		"last_run_status":   "",
		"last_run_message":  "",
		"next_run_time":     r.calculateNextRunTime(task.CronExpression),
		"created_at":        time.Now(),
		"updated_at":        time.Now(),
	}

	lastInsertID, err := r.db.Model(r.tableName).Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return nil, fmt.Errorf("创建定时任务失败: %w", err)
	}

	// 返回创建的任务
	return r.GetByID(ctx, lastInsertID)
}

// Update 更新定时任务
func (r *scheduledTaskRepository) Update(ctx context.Context, task *types.ScheduledTask) error {
	tenantID := r.GetTenantID(ctx)

	data := g.Map{
		"updated_at": time.Now(),
	}

	// 只更新非零值字段
	if task.TaskName != "" {
		data["task_name"] = task.TaskName
	}
	if task.TaskDescription != "" {
		data["task_description"] = task.TaskDescription
	}
	if task.ReportType != "" {
		data["report_type"] = task.ReportType
	}
	if task.CronExpression != "" {
		data["cron_expression"] = task.CronExpression
		data["next_run_time"] = r.calculateNextRunTime(task.CronExpression)
	}
	if task.ReportConfig != "" {
		data["report_config"] = task.ReportConfig
	}
	if len(task.Recipients) > 0 {
		data["recipients"] = r.encodeJSON(task.Recipients)
	}
	// IsEnabled 使用指针类型检查是否需要更新
	data["is_enabled"] = task.IsEnabled

	_, err := r.db.Model(r.tableName).Ctx(ctx).
		Where("id = ? AND tenant_id = ?", task.ID, tenantID).
		Data(data).
		Update()

	if err != nil {
		return fmt.Errorf("更新定时任务失败: %w", err)
	}

	return nil
}

// Delete 删除定时任务
func (r *scheduledTaskRepository) Delete(ctx context.Context, taskID int64) error {
	tenantID := r.GetTenantID(ctx)

	_, err := r.db.Model(r.tableName).Ctx(ctx).
		Where("id = ? AND tenant_id = ?", taskID, tenantID).
		Delete()

	if err != nil {
		return fmt.Errorf("删除定时任务失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取定时任务
func (r *scheduledTaskRepository) GetByID(ctx context.Context, taskID int64) (*types.ScheduledTask, error) {
	tenantID := r.GetTenantID(ctx)

	var task *types.ScheduledTask
	err := r.db.Model(r.tableName).Ctx(ctx).
		Where("id = ? AND tenant_id = ?", taskID, tenantID).
		Scan(&task)

	if err != nil {
		return nil, fmt.Errorf("获取定时任务失败: %w", err)
	}

	if task == nil {
		return nil, fmt.Errorf("定时任务不存在")
	}

	// 解析JSON字段
	if task.RecipientsJSON != "" {
		task.Recipients = r.decodeJSONToStringSlice(task.RecipientsJSON)
	}

	return task, nil
}

// List 获取定时任务列表
func (r *scheduledTaskRepository) List(ctx context.Context, req *types.ScheduledTaskListRequest) ([]*types.ScheduledTask, int, error) {
	tenantID := r.GetTenantID(ctx)

	query := r.db.Model(r.tableName).Ctx(ctx).Where("tenant_id = ?", tenantID)

	// 构建查询条件
	if req.TaskName != "" {
		query = query.WhereLike("task_name", "%"+req.TaskName+"%")
	}

	if req.ReportType != "" {
		query = query.Where("report_type = ?", req.ReportType)
	}

	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}

	if req.LastRunStatus != "" {
		query = query.Where("last_run_status = ?", req.LastRunStatus)
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("统计定时任务数量失败: %w", err)
	}

	// 分页查询
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize
	query = query.OrderDesc("created_at").Limit(offset, req.PageSize)

	var tasks []*types.ScheduledTask
	err = query.Scan(&tasks)
	if err != nil {
		return nil, 0, fmt.Errorf("查询定时任务列表失败: %w", err)
	}

	// 解析JSON字段
	for _, task := range tasks {
		if task.RecipientsJSON != "" {
			task.Recipients = r.decodeJSONToStringSlice(task.RecipientsJSON)
		}
	}

	return tasks, total, nil
}

// UpdateLastExecution 更新最后执行状态
func (r *scheduledTaskRepository) UpdateLastExecution(ctx context.Context, taskID int64, status types.TaskStatus, message string) error {
	tenantID := r.GetTenantID(ctx)

	data := g.Map{
		"last_run_time":    time.Now(),
		"last_run_status":  string(status),
		"last_run_message": message,
		"updated_at":       time.Now(),
	}

	// 如果任务成功完成，计算下次运行时间
	if status == types.TaskStatusCompleted {
		// 获取任务的cron表达式
		var cronExpression string
		err := r.db.Model(r.tableName).Ctx(ctx).
			Fields("cron_expression").
			Where("id = ? AND tenant_id = ?", taskID, tenantID).
			Scan(&cronExpression)

		if err == nil && cronExpression != "" {
			data["next_run_time"] = r.calculateNextRunTime(cronExpression)
		}
	}

	_, err := r.db.Model(r.tableName).Ctx(ctx).
		Where("id = ? AND tenant_id = ?", taskID, tenantID).
		Data(data).
		Update()

	if err != nil {
		return fmt.Errorf("更新任务执行状态失败: %w", err)
	}

	return nil
}

// calculateNextRunTime 计算下次运行时间
func (r *scheduledTaskRepository) calculateNextRunTime(cronExpression string) *time.Time {
	// 这里应该使用cron表达式解析库来计算下次运行时间
	// 为了简化，这里只是返回一小时后的时间
	// 在实际实现中，应该使用专门的cron解析库
	nextTime := time.Now().Add(time.Hour)
	return &nextTime
}

// encodeJSON 编码JSON字符串
func (r *scheduledTaskRepository) encodeJSON(data interface{}) string {
	jsonBytes, err := gconv.Bytes(data)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

// decodeJSONToStringSlice 解码JSON为字符串切片
func (r *scheduledTaskRepository) decodeJSONToStringSlice(jsonStr string) []string {
	if jsonStr == "" {
		return nil
	}

	var result []string
	err := gconv.Struct(jsonStr, &result)
	if err != nil {
		// 如果解析失败，尝试按逗号分割
		return strings.Split(strings.Trim(jsonStr, "[]\""), "\",\"")
	}

	return result
}