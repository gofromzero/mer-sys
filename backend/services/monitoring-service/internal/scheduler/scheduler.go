package scheduler

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/os/gtimer"

	"mer-demo/services/monitoring-service/internal/service"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	monitoringService service.MonitoringService
	ctx               context.Context
}

// NewScheduler 创建调度器实例
func NewScheduler(ctx context.Context) *Scheduler {
	return &Scheduler{
		monitoringService: service.NewMonitoringService(),
		ctx:               ctx,
	}
}

// Start 启动定时任务
func (s *Scheduler) Start() {
	g.Log().Info(s.ctx, "Starting monitoring scheduler")

	// 每5分钟执行一次预警检查
	gtimer.Add(s.ctx, time.Minute*5, s.runPeriodicChecks)

	// 每小时的第1分钟执行使用数据收集
	gtimer.Add(s.ctx, time.Hour, s.runUsageDataCollection)

	// 每天凌晨2点执行数据清理
	s.scheduleDataCleanup()

	g.Log().Info(s.ctx, "Monitoring scheduler started")
}

// Stop 停止定时任务
func (s *Scheduler) Stop() {
	g.Log().Info(s.ctx, "Stopping monitoring scheduler")
	// GoFrame的定时器会在context取消时自动停止
}

// runPeriodicChecks 执行周期性检查
func (s *Scheduler) runPeriodicChecks(ctx context.Context) {
	g.Log().Debug(ctx, "Running periodic monitoring checks")
	
	if err := s.monitoringService.RunPeriodicChecks(ctx); err != nil {
		g.Log().Error(ctx, "Periodic checks failed", err)
	}
}

// runUsageDataCollection 执行使用数据收集
func (s *Scheduler) runUsageDataCollection(ctx context.Context) {
	// 只在每小时的第1分钟执行
	if gtime.Now().Minute() != 1 {
		return
	}

	g.Log().Debug(ctx, "Running usage data collection")
	
	if err := s.monitoringService.CollectUsageData(ctx); err != nil {
		g.Log().Error(ctx, "Usage data collection failed", err)
	}
}

// scheduleDataCleanup 安排数据清理任务
func (s *Scheduler) scheduleDataCleanup() {
	// 计算到下一个凌晨2点的时间
	now := gtime.Now()
	next2AM := now.StartOfDay().Add(time.Hour * 2)
	if now.After(next2AM.Time) {
		next2AM = next2AM.AddDate(0, 0, 1)
	}
	
	duration := next2AM.Sub(now.Time)
	
	// 首次执行
	gtimer.SetTimeout(s.ctx, duration, func(ctx context.Context) {
		s.runDataCleanup(ctx)
		
		// 之后每24小时执行一次
		gtimer.Add(ctx, time.Hour*24, s.runDataCleanup)
	})

	g.Log().Info(s.ctx, "Data cleanup scheduled", g.Map{
		"next_run": next2AM.Format("2006-01-02 15:04:05"),
	})
}

// runDataCleanup 执行数据清理
func (s *Scheduler) runDataCleanup(ctx context.Context) {
	g.Log().Info(ctx, "Running data cleanup")
	
	// 清理30天前的已解决预警
	cutoffDate := gtime.Now().AddDate(0, 0, -30)
	
	// 这里应该调用Repository方法清理旧数据
	// 为简化实现，只记录日志
	g.Log().Info(ctx, "Would cleanup alerts before", g.Map{
		"cutoff_date": cutoffDate.Format("2006-01-02"),
	})
	
	// 清理90天前的使用统计数据
	statsCutoffDate := gtime.Now().AddDate(0, 0, -90)
	g.Log().Info(ctx, "Would cleanup usage stats before", g.Map{
		"cutoff_date": statsCutoffDate.Format("2006-01-02"),
	})
}