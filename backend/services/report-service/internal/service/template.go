package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// ITemplateService 报表模板服务接口
type ITemplateService interface {
	CreateTemplate(ctx context.Context, template *types.ReportTemplate) (*types.ReportTemplate, error)
	GetTemplate(ctx context.Context, templateID uint64) (*types.ReportTemplate, error)
	ListTemplates(ctx context.Context, reportType *types.ReportType) ([]*types.ReportTemplate, error)
	UpdateTemplate(ctx context.Context, template *types.ReportTemplate) (*types.ReportTemplate, error)
	DeleteTemplate(ctx context.Context, templateID uint64) error
	ScheduleReport(ctx context.Context, req *types.ReportScheduleRequest) (*types.ReportTemplate, error)
	GetScheduledTemplates(ctx context.Context) ([]*types.ReportTemplate, error)
}

// TemplateService 报表模板服务实现
type TemplateService struct {
	reportRepo repository.IReportRepository
}

// NewTemplateService 创建报表模板服务实例
func NewTemplateService() ITemplateService {
	return &TemplateService{
		reportRepo: repository.NewReportRepository(),
	}
}

// CreateTemplate 创建报表模板
func (s *TemplateService) CreateTemplate(ctx context.Context, template *types.ReportTemplate) (*types.ReportTemplate, error) {
	tenantID := ctx.Value("tenant_id").(uint64)
	userID := ctx.Value("user_id").(uint64)
	
	template.TenantID = tenantID
	template.CreatedBy = userID
	
	// 验证模板配置
	if err := s.validateTemplateConfig(template); err != nil {
		return nil, fmt.Errorf("模板配置验证失败: %v", err)
	}
	
	// 创建模板
	err := s.reportRepo.CreateReportTemplate(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("创建报表模板失败: %v", err)
	}
	
	g.Log().Info(ctx, "报表模板创建成功", 
		"template_id", template.ID,
		"name", template.Name,
		"report_type", template.ReportType)
	
	return template, nil
}

// GetTemplate 获取报表模板
func (s *TemplateService) GetTemplate(ctx context.Context, templateID uint64) (*types.ReportTemplate, error) {
	return s.reportRepo.GetReportTemplate(ctx, templateID)
}

// ListTemplates 获取报表模板列表
func (s *TemplateService) ListTemplates(ctx context.Context, reportType *types.ReportType) ([]*types.ReportTemplate, error) {
	return s.reportRepo.ListReportTemplates(ctx, reportType)
}

// UpdateTemplate 更新报表模板
func (s *TemplateService) UpdateTemplate(ctx context.Context, template *types.ReportTemplate) (*types.ReportTemplate, error) {
	// 验证模板是否存在
	existing, err := s.reportRepo.GetReportTemplate(ctx, template.ID)
	if err != nil {
		return nil, fmt.Errorf("报表模板不存在: %v", err)
	}
	
	// 保持租户ID和创建者不变
	template.TenantID = existing.TenantID
	template.CreatedBy = existing.CreatedBy
	template.CreatedAt = existing.CreatedAt
	
	// 验证模板配置
	if err := s.validateTemplateConfig(template); err != nil {
		return nil, fmt.Errorf("模板配置验证失败: %v", err)
	}
	
	// 更新模板
	err = s.reportRepo.UpdateReportTemplate(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("更新报表模板失败: %v", err)
	}
	
	g.Log().Info(ctx, "报表模板更新成功", 
		"template_id", template.ID,
		"name", template.Name)
	
	return template, nil
}

// DeleteTemplate 删除报表模板
func (s *TemplateService) DeleteTemplate(ctx context.Context, templateID uint64) error {
	// 验证模板是否存在
	_, err := s.reportRepo.GetReportTemplate(ctx, templateID)
	if err != nil {
		return fmt.Errorf("报表模板不存在: %v", err)
	}
	
	// 检查是否有关联的报表任务
	pendingJobs, err := s.reportRepo.ListPendingJobs(ctx)
	if err != nil {
		g.Log().Warning(ctx, "检查关联任务失败", "error", err)
	} else {
		for _, job := range pendingJobs {
			if job.TemplateID != nil && *job.TemplateID == templateID {
				return fmt.Errorf("存在未完成的报表任务，无法删除模板")
			}
		}
	}
	
	// 删除模板
	err = s.reportRepo.DeleteReportTemplate(ctx, templateID)
	if err != nil {
		return fmt.Errorf("删除报表模板失败: %v", err)
	}
	
	g.Log().Info(ctx, "报表模板删除成功", "template_id", templateID)
	return nil
}

// ScheduleReport 创建定时报表
func (s *TemplateService) ScheduleReport(ctx context.Context, req *types.ReportScheduleRequest) (*types.ReportTemplate, error) {
	tenantID := ctx.Value("tenant_id").(uint64)
	userID := ctx.Value("user_id").(uint64)
	
	// 构建模板配置JSON
	templateConfigBytes, err := json.Marshal(req.TemplateConfig)
	if err != nil {
		return nil, fmt.Errorf("序列化模板配置失败: %v", err)
	}
	
	// 构建调度配置JSON
	scheduleConfigBytes, err := json.Marshal(req.Schedule)
	if err != nil {
		return nil, fmt.Errorf("序列化调度配置失败: %v", err)
	}
	
	// 构建收件人JSON
	recipientsBytes, err := json.Marshal(req.Recipients)
	if err != nil {
		return nil, fmt.Errorf("序列化收件人失败: %v", err)
	}
	
	// 创建报表模板
	template := &types.ReportTemplate{
		TenantID:       tenantID,
		Name:           req.Name,
		ReportType:     req.ReportType,
		TemplateConfig: templateConfigBytes,
		ScheduleConfig: scheduleConfigBytes,
		Recipients:     recipientsBytes,
		FileFormat:     req.FileFormat,
		Enabled:        req.Enabled,
		CreatedBy:      userID,
	}
	
	// 验证调度配置
	if err := s.validateScheduleConfig(&req.Schedule); err != nil {
		return nil, fmt.Errorf("调度配置验证失败: %v", err)
	}
	
	// 创建模板
	err = s.reportRepo.CreateReportTemplate(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("创建定时报表模板失败: %v", err)
	}
	
	// 创建初始报表任务
	err = s.scheduleNextJob(ctx, template, &req.Schedule)
	if err != nil {
		g.Log().Warning(ctx, "创建初始报表任务失败", "error", err)
		// 不返回错误，模板已经创建成功
	}
	
	g.Log().Info(ctx, "定时报表创建成功", 
		"template_id", template.ID,
		"name", template.Name,
		"frequency", req.Schedule.Frequency)
	
	return template, nil
}

// GetScheduledTemplates 获取已调度的报表模板
func (s *TemplateService) GetScheduledTemplates(ctx context.Context) ([]*types.ReportTemplate, error) {
	templates, err := s.reportRepo.ListReportTemplates(ctx, nil)
	if err != nil {
		return nil, err
	}
	
	// 过滤出有调度配置且启用的模板
	var scheduledTemplates []*types.ReportTemplate
	for _, template := range templates {
		if template.Enabled && len(template.ScheduleConfig) > 0 {
			scheduledTemplates = append(scheduledTemplates, template)
		}
	}
	
	return scheduledTemplates, nil
}

// validateTemplateConfig 验证模板配置
func (s *TemplateService) validateTemplateConfig(template *types.ReportTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("模板名称不能为空")
	}
	
	if template.ReportType == "" {
		return fmt.Errorf("报表类型不能为空")
	}
	
	// 验证报表类型是否有效
	validTypes := map[types.ReportType]bool{
		types.ReportTypeFinancial:         true,
		types.ReportTypeMerchantOperation: true,
		types.ReportTypeCustomerAnalysis:  true,
	}
	if !validTypes[template.ReportType] {
		return fmt.Errorf("无效的报表类型: %s", template.ReportType)
	}
	
	// 验证文件格式
	validFormats := map[types.FileFormat]bool{
		types.FileFormatExcel: true,
		types.FileFormatPDF:   true,
		types.FileFormatJSON:  true,
	}
	if !validFormats[template.FileFormat] {
		return fmt.Errorf("无效的文件格式: %s", template.FileFormat)
	}
	
	// 验证模板配置JSON格式
	if len(template.TemplateConfig) > 0 {
		var config map[string]interface{}
		if err := json.Unmarshal(template.TemplateConfig, &config); err != nil {
			return fmt.Errorf("模板配置JSON格式无效: %v", err)
		}
	}
	
	return nil
}

// validateScheduleConfig 验证调度配置
func (s *TemplateService) validateScheduleConfig(schedule *types.ScheduleConfig) error {
	if schedule.Frequency == "" {
		return fmt.Errorf("调度频率不能为空")
	}
	
	// 验证频率是否有效
	validFreq := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
	if !validFreq[schedule.Frequency] {
		return fmt.Errorf("无效的调度频率: %s", schedule.Frequency)
	}
	
	// 验证时间格式
	if schedule.Time == "" {
		return fmt.Errorf("调度时间不能为空")
	}
	if _, err := time.Parse("15:04", schedule.Time); err != nil {
		return fmt.Errorf("调度时间格式无效，应为HH:mm格式: %v", err)
	}
	
	// 验证时区
	if schedule.Timezone == "" {
		return fmt.Errorf("时区不能为空")
	}
	if _, err := time.LoadLocation(schedule.Timezone); err != nil {
		return fmt.Errorf("无效的时区: %v", err)
	}
	
	return nil
}

// scheduleNextJob 调度下一个任务
func (s *TemplateService) scheduleNextJob(ctx context.Context, template *types.ReportTemplate, schedule *types.ScheduleConfig) error {
	// 计算下次执行时间
	nextRun := s.calculateNextRun(schedule)
	
	// 创建报表任务
	job := &types.ReportJob{
		TenantID:    template.TenantID,
		TemplateID:  &template.ID,
		Status:      types.JobStatusPending,
		ScheduledAt: nextRun,
	}
	
	return s.reportRepo.CreateReportJob(ctx, job)
}

// calculateNextRun 计算下次执行时间
func (s *TemplateService) calculateNextRun(schedule *types.ScheduleConfig) time.Time {
	now := time.Now()
	
	// 解析调度时间
	scheduleTime, _ := time.Parse("15:04", schedule.Time)
	
	// 获取时区
	loc, err := time.LoadLocation(schedule.Timezone)
	if err != nil {
		loc = time.UTC
	}
	
	// 转换到指定时区
	now = now.In(loc)
	
	var nextRun time.Time
	
	switch schedule.Frequency {
	case "daily":
		// 每天执行
		nextRun = time.Date(now.Year(), now.Month(), now.Day(), 
			scheduleTime.Hour(), scheduleTime.Minute(), 0, 0, loc)
		if nextRun.Before(now) {
			nextRun = nextRun.Add(24 * time.Hour)
		}
		
	case "weekly":
		// 每周执行（周一）
		daysUntilMonday := (7 - int(now.Weekday()) + 1) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		nextRun = time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday,
			scheduleTime.Hour(), scheduleTime.Minute(), 0, 0, loc)
		
	case "monthly":
		// 每月执行（1号）
		nextMonth := now.Month() + 1
		nextYear := now.Year()
		if nextMonth > 12 {
			nextMonth = 1
			nextYear++
		}
		nextRun = time.Date(nextYear, nextMonth, 1,
			scheduleTime.Hour(), scheduleTime.Minute(), 0, 0, loc)
	}
	
	return nextRun
}