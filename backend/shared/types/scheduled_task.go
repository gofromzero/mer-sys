package types

import (
	"time"
)

// TaskStatus 任务状态枚举
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 待执行
	TaskStatusRunning   TaskStatus = "running"   // 运行中
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusFailed    TaskStatus = "failed"    // 失败
)

// ScheduledTask 定时任务模型
type ScheduledTask struct {
	ID              int64      `json:"id" db:"id"`
	TenantID        int64      `json:"tenant_id" db:"tenant_id"`
	TaskName        string     `json:"task_name" db:"task_name"`
	TaskDescription string     `json:"task_description" db:"task_description"`
	ReportType      ReportType `json:"report_type" db:"report_type"`
	CronExpression  string     `json:"cron_expression" db:"cron_expression"`
	ReportConfig    string     `json:"report_config" db:"report_config"` // JSON格式的报表配置
	Recipients      []string   `json:"recipients" db:"-"`                // 接收人列表
	RecipientsJSON  string     `json:"-" db:"recipients"`                // 数据库存储的JSON格式
	IsEnabled       bool       `json:"is_enabled" db:"is_enabled"`
	LastRunTime     *time.Time `json:"last_run_time" db:"last_run_time"`
	LastRunStatus   string     `json:"last_run_status" db:"last_run_status"`
	LastRunMessage  string     `json:"last_run_message" db:"last_run_message"`
	NextRunTime     *time.Time `json:"next_run_time" db:"next_run_time"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// ScheduledTaskCreateRequest 创建定时任务请求
type ScheduledTaskCreateRequest struct {
	TaskName        string     `json:"task_name" binding:"required,min=1,max=100" label:"任务名称"`
	TaskDescription string     `json:"task_description" binding:"max=500" label:"任务描述"`
	ReportType      ReportType `json:"report_type" binding:"required" label:"报表类型"`
	CronExpression  string     `json:"cron_expression" binding:"required" label:"Cron表达式"`
	ReportConfig    string     `json:"report_config" binding:"required" label:"报表配置"`
	Recipients      []string   `json:"recipients" binding:"required,min=1" label:"接收人列表"`
	IsEnabled       bool       `json:"is_enabled" label:"是否启用"`
}

// ScheduledTaskUpdateRequest 更新定时任务请求
type ScheduledTaskUpdateRequest struct {
	TaskName        string     `json:"task_name" binding:"max=100" label:"任务名称"`
	TaskDescription string     `json:"task_description" binding:"max=500" label:"任务描述"`
	CronExpression  string     `json:"cron_expression" label:"Cron表达式"`
	ReportConfig    string     `json:"report_config" label:"报表配置"`
	Recipients      []string   `json:"recipients" label:"接收人列表"`
	IsEnabled       *bool      `json:"is_enabled" label:"是否启用"`
}

// ScheduledTaskListRequest 定时任务列表请求
type ScheduledTaskListRequest struct {
	TaskName      string  `json:"task_name" form:"task_name" label:"任务名称"`
	ReportType    string  `json:"report_type" form:"report_type" label:"报表类型"`
	IsEnabled     *bool   `json:"is_enabled" form:"is_enabled" label:"是否启用"`
	LastRunStatus string  `json:"last_run_status" form:"last_run_status" label:"最后运行状态"`
	Page          int     `json:"page" form:"page" binding:"min=1" label:"页码"`
	PageSize      int     `json:"page_size" form:"page_size" binding:"min=1,max=100" label:"每页数量"`
}

// ScheduledTaskListResponse 定时任务列表响应
type ScheduledTaskListResponse struct {
	Tasks      []*ScheduledTask `json:"tasks"`
	Pagination Pagination       `json:"pagination"`
}

// ScheduledTaskExecuteRequest 执行定时任务请求
type ScheduledTaskExecuteRequest struct {
	TaskID int64 `json:"task_id" uri:"id" binding:"required,min=1" label:"任务ID"`
}

// ReportScheduleConfig 报表调度配置
type ReportScheduleConfig struct {
	StartDate  string                 `json:"start_date" binding:"required" label:"开始日期"`
	EndDate    string                 `json:"end_date" binding:"required" label:"结束日期"`
	Format     string                 `json:"format" binding:"required,oneof=excel pdf json" label:"报表格式"`
	Filters    map[string]interface{} `json:"filters,omitempty" label:"过滤条件"`
	MerchantID int64                  `json:"merchant_id,omitempty" label:"商户ID"`
}

// 常用的Cron表达式预设
var CommonCronExpressions = map[string]string{
	"daily_9am":      "0 9 * * *",        // 每天上午9点
	"daily_6pm":      "0 18 * * *",       // 每天下午6点
	"weekly_monday":  "0 9 * * 1",        // 每周一上午9点
	"monthly_first":  "0 9 1 * *",        // 每月1号上午9点
	"quarterly":      "0 9 1 */3 *",      // 每季度第一天上午9点
	"yearly":         "0 9 1 1 *",        // 每年1月1号上午9点
	"hourly":         "0 * * * *",        // 每小时
	"every_30min":    "*/30 * * * *",     // 每30分钟
	"business_daily": "0 9 * * 1-5",      // 工作日每天上午9点
}