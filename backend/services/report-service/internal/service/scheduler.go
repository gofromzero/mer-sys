package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gctx"
)

// ISchedulerService 调度服务接口
type ISchedulerService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	ProcessPendingJobs(ctx context.Context) error
	ScheduleTemplateJobs(ctx context.Context) error
}

// SchedulerService 调度服务实现
type SchedulerService struct {
	reportRepo      repository.IReportRepository
	generatorService IReportGeneratorService
	templateService  ITemplateService
	cron            *gcron.Cron
	isRunning       bool
}

// NewSchedulerService 创建调度服务实例
func NewSchedulerService() ISchedulerService {
	return &SchedulerService{
		reportRepo:      repository.NewReportRepository(),
		generatorService: NewReportGeneratorService(),
		templateService:  NewTemplateService(),
		cron:            gcron.New(),
		isRunning:       false,
	}
}

// Start 启动调度服务
func (s *SchedulerService) Start(ctx context.Context) error {
	if s.isRunning {
		return fmt.Errorf("调度服务已经在运行")
	}
	
	g.Log().Info(ctx, "启动报表调度服务")
	
	// 每分钟检查一次待执行的任务
	_, err := s.cron.Add(ctx, "*/1 * * * *", func(ctx context.Context) {
		if err := s.ProcessPendingJobs(ctx); err != nil {
			g.Log().Error(ctx, "处理待执行任务失败", "error", err)
		}
	}, "ProcessPendingJobs")
	if err != nil {
		return fmt.Errorf("添加任务处理定时器失败: %v", err)
	}
	
	// 每小时检查一次模板调度
	_, err = s.cron.Add(ctx, "0 * * * *", func(ctx context.Context) {
		if err := s.ScheduleTemplateJobs(ctx); err != nil {
			g.Log().Error(ctx, "模板调度失败", "error", err)
		}
	}, "ScheduleTemplateJobs")
	if err != nil {
		return fmt.Errorf("添加模板调度定时器失败: %v", err)
	}
	
	// 启动定时器
	s.cron.Start()
	s.isRunning = true
	
	g.Log().Info(ctx, "报表调度服务启动成功")
	return nil
}

// Stop 停止调度服务
func (s *SchedulerService) Stop(ctx context.Context) error {
	if !s.isRunning {
		return nil
	}
	
	g.Log().Info(ctx, "停止报表调度服务")
	
	s.cron.Stop()
	s.isRunning = false
	
	g.Log().Info(ctx, "报表调度服务停止成功")
	return nil
}

// ProcessPendingJobs 处理待执行的任务
func (s *SchedulerService) ProcessPendingJobs(ctx context.Context) error {
	g.Log().Debug(ctx, "开始处理待执行的报表任务")
	
	// 获取待执行的任务
	jobs, err := s.reportRepo.ListPendingJobs(ctx)
	if err != nil {
		g.Log().Error(ctx, "获取待执行任务失败", "error", err)
		return err
	}
	
	if len(jobs) == 0 {
		return nil
	}
	
	g.Log().Info(ctx, "找到待执行的报表任务", "count", len(jobs))
	
	// 处理每个任务
	for _, job := range jobs {
		if err := s.processJob(gctx.New(), job); err != nil {
			g.Log().Error(ctx, "处理报表任务失败", 
				"job_id", job.ID, 
				"template_id", job.TemplateID,
				"error", err)
		}
	}
	
	return nil
}

// ScheduleTemplateJobs 为模板创建调度任务
func (s *SchedulerService) ScheduleTemplateJobs(ctx context.Context) error {
	g.Log().Debug(ctx, "检查需要调度的报表模板")
	
	// 获取所有启用的调度模板
	templates, err := s.templateService.GetScheduledTemplates(ctx)
	if err != nil {
		g.Log().Error(ctx, "获取调度模板失败", "error", err)
		return err
	}
	
	for _, template := range templates {
		if err := s.scheduleTemplateIfNeeded(ctx, template); err != nil {
			g.Log().Error(ctx, "调度模板任务失败", 
				"template_id", template.ID,
				"name", template.Name,
				"error", err)
		}
	}
	
	return nil
}

// processJob 处理单个任务
func (s *SchedulerService) processJob(ctx context.Context, job *types.ReportJob) error {
	// 设置租户上下文
	ctx = context.WithValue(ctx, "tenant_id", job.TenantID)
	
	g.Log().Info(ctx, "开始处理报表任务", 
		"job_id", job.ID,
		"template_id", job.TemplateID,
		"scheduled_at", job.ScheduledAt)
	
	// 更新任务状态为运行中
	job.Status = types.JobStatusRunning
	startTime := time.Now()
	job.StartedAt = &startTime
	
	err := s.reportRepo.UpdateReportJob(ctx, job)
	if err != nil {
		g.Log().Error(ctx, "更新任务状态失败", "job_id", job.ID, "error", err)
		return err
	}
	
	var reportErr error
	
	defer func() {
		// 更新任务完成状态
		endTime := time.Now()
		job.CompletedAt = &endTime
		
		if reportErr != nil {
			job.Status = types.JobStatusFailed
			job.ErrorMessage = reportErr.Error()
			job.RetryCount++
		} else {
			job.Status = types.JobStatusCompleted
		}
		
		if err := s.reportRepo.UpdateReportJob(ctx, job); err != nil {
			g.Log().Error(ctx, "更新任务完成状态失败", "job_id", job.ID, "error", err)
		}
		
		// 为模板任务调度下一次执行
		if job.TemplateID != nil && reportErr == nil {
			if err := s.scheduleNextTemplateJob(ctx, *job.TemplateID); err != nil {
				g.Log().Error(ctx, "调度下一次模板任务失败", "template_id", *job.TemplateID, "error", err)
			}
		}
	}()
	
	// 如果是模板任务，需要获取模板配置
	if job.TemplateID != nil {
		reportErr = s.processTemplateJob(ctx, job)
	} else if job.ReportID != nil {
		reportErr = s.processReportJob(ctx, job)
	} else {
		reportErr = fmt.Errorf("任务没有关联的模板或报表")
	}
	
	return reportErr
}

// processTemplateJob 处理模板任务
func (s *SchedulerService) processTemplateJob(ctx context.Context, job *types.ReportJob) error {
	// 获取模板配置
	template, err := s.templateService.GetTemplate(ctx, *job.TemplateID)
	if err != nil {
		return fmt.Errorf("获取报表模板失败: %v", err)
	}
	
	// 解析模板配置
	var templateConfig map[string]interface{}
	if err := json.Unmarshal(template.TemplateConfig, &templateConfig); err != nil {
		return fmt.Errorf("解析模板配置失败: %v", err)
	}
	
	// 构建报表生成请求
	req := &types.ReportCreateRequest{
		ReportType: template.ReportType,
		FileFormat: template.FileFormat,
		Config:     templateConfig,
	}
	
	// 设置时间范围（根据调度频率）
	s.setTimeRangeForScheduledReport(req, template)
	
	// 生成报表
	report, err := s.generatorService.GenerateReport(ctx, req)
	if err != nil {
		return fmt.Errorf("生成调度报表失败: %v", err)
	}
	
	// 更新任务关联的报表ID
	job.ReportID = &report.ID
	
	g.Log().Info(ctx, "调度报表生成成功", 
		"job_id", job.ID,
		"report_id", report.ID,
		"template_name", template.Name)
	
	// TODO: 发送报表邮件通知
	// s.sendReportNotification(ctx, template, report)
	
	return nil
}

// processReportJob 处理报表任务
func (s *SchedulerService) processReportJob(ctx context.Context, job *types.ReportJob) error {
	// 获取报表信息
	report, err := s.generatorService.GetReport(ctx, *job.ReportID)
	if err != nil {
		return fmt.Errorf("获取报表信息失败: %v", err)
	}
	
	// 检查报表状态
	if report.Status == types.ReportStatusCompleted {
		g.Log().Info(ctx, "报表已完成", "report_id", report.ID)
		return nil
	}
	
	if report.Status == types.ReportStatusFailed {
		return fmt.Errorf("关联的报表生成失败")
	}
	
	// 报表还在生成中，继续等待
	g.Log().Info(ctx, "报表仍在生成中", "report_id", report.ID, "status", report.Status)
	return fmt.Errorf("报表仍在生成中")
}

// scheduleTemplateIfNeeded 检查模板是否需要调度新任务
func (s *SchedulerService) scheduleTemplateIfNeeded(ctx context.Context, template *types.ReportTemplate) error {
	// 解析调度配置
	var scheduleConfig types.ScheduleConfig
	if err := json.Unmarshal(template.ScheduleConfig, &scheduleConfig); err != nil {
		return fmt.Errorf("解析调度配置失败: %v", err)
	}
	
	// 检查是否已有未来的任务
	pendingJobs, err := s.reportRepo.ListPendingJobs(ctx)
	if err != nil {
		return fmt.Errorf("检查待执行任务失败: %v", err)
	}
	
	// 检查是否已有该模板的未来任务
	now := time.Now()
	hasFutureJob := false
	
	for _, job := range pendingJobs {
		if job.TemplateID != nil && *job.TemplateID == template.ID && job.ScheduledAt.After(now) {
			hasFutureJob = true
			break
		}
	}
	
	// 如果没有未来任务，创建一个
	if !hasFutureJob {
		nextRun := s.calculateNextRun(&scheduleConfig)
		
		job := &types.ReportJob{
			TenantID:    template.TenantID,
			TemplateID:  &template.ID,
			Status:      types.JobStatusPending,
			ScheduledAt: nextRun,
		}
		
		err := s.reportRepo.CreateReportJob(ctx, job)
		if err != nil {
			return fmt.Errorf("创建调度任务失败: %v", err)
		}
		
		g.Log().Info(ctx, "为模板创建调度任务", 
			"template_id", template.ID,
			"template_name", template.Name,
			"next_run", nextRun)
	}
	
	return nil
}

// scheduleNextTemplateJob 为模板调度下一次任务
func (s *SchedulerService) scheduleNextTemplateJob(ctx context.Context, templateID uint64) error {
	template, err := s.templateService.GetTemplate(ctx, templateID)
	if err != nil {
		return err
	}
	
	return s.scheduleTemplateIfNeeded(ctx, template)
}

// setTimeRangeForScheduledReport 为调度报表设置时间范围
func (s *SchedulerService) setTimeRangeForScheduledReport(req *types.ReportCreateRequest, template *types.ReportTemplate) {
	now := time.Now()
	
	// 解析调度配置
	var scheduleConfig types.ScheduleConfig
	if err := json.Unmarshal(template.ScheduleConfig, &scheduleConfig); err != nil {
		// 使用默认时间范围
		req.StartDate = now.AddDate(0, 0, -7)  // 最近7天
		req.EndDate = now
		req.PeriodType = types.PeriodTypeDaily
		return
	}
	
	// 根据调度频率设置时间范围
	switch scheduleConfig.Frequency {
	case "daily":
		// 昨天的数据
		req.StartDate = now.AddDate(0, 0, -1)
		req.EndDate = now.AddDate(0, 0, -1)
		req.PeriodType = types.PeriodTypeDaily
		
	case "weekly":
		// 上周的数据
		req.StartDate = now.AddDate(0, 0, -7)
		req.EndDate = now.AddDate(0, 0, -1)
		req.PeriodType = types.PeriodTypeWeekly
		
	case "monthly":
		// 上月的数据
		lastMonth := now.AddDate(0, -1, 0)
		req.StartDate = time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location())
		req.EndDate = req.StartDate.AddDate(0, 1, -1)
		req.PeriodType = types.PeriodTypeMonthly
		
	default:
		// 默认最近7天
		req.StartDate = now.AddDate(0, 0, -7)
		req.EndDate = now
		req.PeriodType = types.PeriodTypeDaily
	}
}

// calculateNextRun 计算下次执行时间（与template.go中的实现相同）
func (s *SchedulerService) calculateNextRun(schedule *types.ScheduleConfig) time.Time {
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