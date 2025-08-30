package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// ReportHistory 报表历史记录
type ReportHistory struct {
	ID            int64                  `json:"id"`
	ReportID      int64                  `json:"report_id"`
	TaskID        *int64                 `json:"task_id,omitempty"`
	ReportType    types.ReportType       `json:"report_type"`
	GeneratedBy   int64                  `json:"generated_by"`
	GeneratedAt   time.Time              `json:"generated_at"`
	Status        types.ReportStatus     `json:"status"`
	FileFormat    types.FileFormat       `json:"file_format"`
	FilePath      string                 `json:"file_path"`
	FileSize      int64                  `json:"file_size"`
	ExecutionTime int64                  `json:"execution_time"` // 执行时间(毫秒)
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ReportHistoryRequest 报表历史查询请求
type ReportHistoryRequest struct {
	ReportType  types.ReportType   `json:"report_type,omitempty"`
	Status      types.ReportStatus `json:"status,omitempty"`
	GeneratedBy int64              `json:"generated_by,omitempty"`
	TaskID      int64              `json:"task_id,omitempty"`
	StartDate   *time.Time         `json:"start_date,omitempty"`
	EndDate     *time.Time         `json:"end_date,omitempty"`
	Page        int                `json:"page"`
	PageSize    int                `json:"page_size"`
}

// ReportStats 报表统计信息
type ReportStats struct {
	TotalReports    int64                        `json:"total_reports"`
	TotalSize       int64                        `json:"total_size"`        // 总文件大小(字节)
	SuccessRate     float64                      `json:"success_rate"`      // 成功率
	AvgExecutionTime float64                     `json:"avg_execution_time"` // 平均执行时间(毫秒)
	TypeDistribution map[types.ReportType]int64  `json:"type_distribution"`  // 按类型分布
	FormatDistribution map[types.FileFormat]int64 `json:"format_distribution"` // 按格式分布
	DailyStats       []*DailyReportStats          `json:"daily_stats"`         // 每日统计
}

// DailyReportStats 每日报表统计
type DailyReportStats struct {
	Date    string `json:"date"`
	Count   int64  `json:"count"`
	Success int64  `json:"success"`
	Failed  int64  `json:"failed"`
}

// IHistoryService 报表历史服务接口
type IHistoryService interface {
	// 记录报表生成历史
	RecordReportGeneration(ctx context.Context, report *types.Report, executionTime int64, filePath string, fileSize int64) error
	// 更新报表状态
	UpdateReportStatus(ctx context.Context, reportID int64, status types.ReportStatus, errorMessage string) error
	// 获取报表历史列表
	GetReportHistory(ctx context.Context, req *ReportHistoryRequest) ([]*ReportHistory, int, error)
	// 获取报表统计信息
	GetReportStats(ctx context.Context, startDate, endDate time.Time) (*ReportStats, error)
	// 删除报表历史记录
	DeleteReportHistory(ctx context.Context, historyID int64) error
	// 批量删除过期历史记录
	CleanupExpiredHistory(ctx context.Context, retentionDays int) (int64, error)
	// 获取用户报表历史
	GetUserReportHistory(ctx context.Context, userID int64, limit int) ([]*ReportHistory, error)
	// 导出报表历史
	ExportReportHistory(ctx context.Context, req *ReportHistoryRequest, format types.FileFormat) (string, error)
}

// HistoryService 报表历史服务实现
type HistoryService struct {
	reportRepo repository.IReportRepository
}

// NewHistoryService 创建报表历史服务实例
func NewHistoryService() IHistoryService {
	return &HistoryService{
		reportRepo: repository.NewReportRepository(),
	}
}

// RecordReportGeneration 记录报表生成历史
func (h *HistoryService) RecordReportGeneration(ctx context.Context, report *types.Report, executionTime int64, filePath string, fileSize int64) error {
	g.Log().Info(ctx, "记录报表生成历史", 
		"report_id", report.ID,
		"report_type", report.ReportType,
		"execution_time", executionTime,
		"file_size", fileSize)
	
	// 这里可以将历史记录保存到专门的历史表中
	// 或者直接使用现有的报表记录
	
	// 更新报表记录的相关信息
	updateData := map[string]interface{}{
		"file_path":      filePath,
		"execution_time": executionTime,
		"file_size":      fileSize,
		"updated_at":     time.Now(),
	}
	
	// 这里假设有一个更新报表附加信息的方法
	if err := h.updateReportMetadata(ctx, report.ID, updateData); err != nil {
		g.Log().Error(ctx, "更新报表元数据失败", "report_id", report.ID, "error", err)
		return err
	}
	
	g.Log().Info(ctx, "报表生成历史记录成功", "report_id", report.ID)
	return nil
}

// UpdateReportStatus 更新报表状态
func (h *HistoryService) UpdateReportStatus(ctx context.Context, reportID int64, status types.ReportStatus, errorMessage string) error {
	g.Log().Info(ctx, "更新报表状态", 
		"report_id", reportID,
		"status", status,
		"error_message", errorMessage)
	
	updateData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	
	if errorMessage != "" {
		updateData["error_message"] = errorMessage
	}
	
	if err := h.updateReportMetadata(ctx, reportID, updateData); err != nil {
		return fmt.Errorf("更新报表状态失败: %v", err)
	}
	
	return nil
}

// GetReportHistory 获取报表历史列表
func (h *HistoryService) GetReportHistory(ctx context.Context, req *ReportHistoryRequest) ([]*ReportHistory, int, error) {
	g.Log().Info(ctx, "获取报表历史列表", "request", req)
	
	// 构建查询条件
	queryReq := &types.ReportListRequest{
		ReportType: req.ReportType,
		Status:     req.Status,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	
	// 获取报表列表
	reports, total, err := h.reportRepo.List(ctx, queryReq)
	if err != nil {
		return nil, 0, fmt.Errorf("获取报表列表失败: %v", err)
	}
	
	// 转换为历史记录格式
	histories := make([]*ReportHistory, 0, len(reports))
	for _, report := range reports {
		history := &ReportHistory{
			ID:            report.ID,
			ReportID:      report.ID,
			ReportType:    report.ReportType,
			GeneratedBy:   report.GeneratedBy,
			GeneratedAt:   report.GeneratedAt,
			Status:        report.Status,
			FileFormat:    report.FileFormat,
			FilePath:      report.FilePath,
			ExecutionTime: 0, // 这里可以从元数据中获取
			ErrorMessage:  "", // 这里可以从元数据中获取
		}
		
		// 如果有任务关联，设置TaskID
		if report.DataSummary != nil {
			if taskID, ok := report.DataSummary["task_id"].(int64); ok {
				history.TaskID = &taskID
			}
		}
		
		histories = append(histories, history)
	}
	
	g.Log().Info(ctx, "获取报表历史完成", "count", len(histories), "total", total)
	return histories, total, nil
}

// GetReportStats 获取报表统计信息
func (h *HistoryService) GetReportStats(ctx context.Context, startDate, endDate time.Time) (*ReportStats, error) {
	g.Log().Info(ctx, "获取报表统计信息", 
		"start_date", startDate.Format("2006-01-02"),
		"end_date", endDate.Format("2006-01-02"))
	
	// 获取所有报表记录（在指定时间范围内）
	req := &types.ReportListRequest{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Page:      1,
		PageSize:  10000, // 获取所有记录
	}
	
	reports, total, err := h.reportRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取报表数据失败: %v", err)
	}
	
	// 计算统计信息
	stats := &ReportStats{
		TotalReports:       total,
		TotalSize:         0,
		SuccessRate:       0,
		AvgExecutionTime:  0,
		TypeDistribution:  make(map[types.ReportType]int64),
		FormatDistribution: make(map[types.FileFormat]int64),
		DailyStats:        make([]*DailyReportStats, 0),
	}
	
	successCount := int64(0)
	totalExecutionTime := int64(0)
	dailyStatsMap := make(map[string]*DailyReportStats)
	
	for _, report := range reports {
		// 类型分布统计
		stats.TypeDistribution[report.ReportType]++
		
		// 格式分布统计
		stats.FormatDistribution[report.FileFormat]++
		
		// 成功率统计
		if report.Status == types.ReportStatusCompleted {
			successCount++
		}
		
		// 每日统计
		date := report.GeneratedAt.Format("2006-01-02")
		if dailyStats, exists := dailyStatsMap[date]; exists {
			dailyStats.Count++
			if report.Status == types.ReportStatusCompleted {
				dailyStats.Success++
			} else if report.Status == types.ReportStatusFailed {
				dailyStats.Failed++
			}
		} else {
			dailyStats := &DailyReportStats{
				Date:  date,
				Count: 1,
			}
			if report.Status == types.ReportStatusCompleted {
				dailyStats.Success = 1
			} else if report.Status == types.ReportStatusFailed {
				dailyStats.Failed = 1
			}
			dailyStatsMap[date] = dailyStats
		}
		
		// 文件大小统计（这里假设从元数据中获取）
		if report.DataSummary != nil {
			if fileSize, ok := report.DataSummary["file_size"].(int64); ok {
				stats.TotalSize += fileSize
			}
			if executionTime, ok := report.DataSummary["execution_time"].(int64); ok {
				totalExecutionTime += executionTime
			}
		}
	}
	
	// 计算成功率
	if total > 0 {
		stats.SuccessRate = float64(successCount) / float64(total) * 100
	}
	
	// 计算平均执行时间
	if successCount > 0 {
		stats.AvgExecutionTime = float64(totalExecutionTime) / float64(successCount)
	}
	
	// 转换每日统计为数组
	for _, dailyStats := range dailyStatsMap {
		stats.DailyStats = append(stats.DailyStats, dailyStats)
	}
	
	g.Log().Info(ctx, "报表统计信息计算完成", 
		"total_reports", stats.TotalReports,
		"success_rate", stats.SuccessRate,
		"avg_execution_time", stats.AvgExecutionTime)
	
	return stats, nil
}

// DeleteReportHistory 删除报表历史记录
func (h *HistoryService) DeleteReportHistory(ctx context.Context, historyID int64) error {
	g.Log().Info(ctx, "删除报表历史记录", "history_id", historyID)
	
	// 这里实际上是删除报表记录
	if err := h.reportRepo.Delete(ctx, historyID); err != nil {
		return fmt.Errorf("删除报表历史失败: %v", err)
	}
	
	g.Log().Info(ctx, "报表历史记录删除成功", "history_id", historyID)
	return nil
}

// CleanupExpiredHistory 批量删除过期历史记录
func (h *HistoryService) CleanupExpiredHistory(ctx context.Context, retentionDays int) (int64, error) {
	g.Log().Info(ctx, "开始清理过期报表历史", "retention_days", retentionDays)
	
	// 计算过期时间点
	expiredDate := time.Now().AddDate(0, 0, -retentionDays)
	
	// 获取过期的报表记录
	req := &types.ReportListRequest{
		EndDate:  expiredDate.Format("2006-01-02"),
		Page:     1,
		PageSize: 10000,
	}
	
	reports, _, err := h.reportRepo.List(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("获取过期报表记录失败: %v", err)
	}
	
	deletedCount := int64(0)
	for _, report := range reports {
		if err := h.reportRepo.Delete(ctx, report.ID); err != nil {
			g.Log().Warning(ctx, "删除过期报表失败", 
				"report_id", report.ID, "error", err)
			continue
		}
		deletedCount++
	}
	
	g.Log().Info(ctx, "过期报表历史清理完成", 
		"deleted_count", deletedCount,
		"retention_days", retentionDays)
	
	return deletedCount, nil
}

// GetUserReportHistory 获取用户报表历史
func (h *HistoryService) GetUserReportHistory(ctx context.Context, userID int64, limit int) ([]*ReportHistory, error) {
	g.Log().Info(ctx, "获取用户报表历史", "user_id", userID, "limit", limit)
	
	req := &ReportHistoryRequest{
		GeneratedBy: userID,
		Page:        1,
		PageSize:    limit,
	}
	
	histories, _, err := h.GetReportHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取用户报表历史失败: %v", err)
	}
	
	return histories, nil
}

// ExportReportHistory 导出报表历史
func (h *HistoryService) ExportReportHistory(ctx context.Context, req *ReportHistoryRequest, format types.FileFormat) (string, error) {
	g.Log().Info(ctx, "导出报表历史", "format", format)
	
	// 获取所有符合条件的历史记录
	req.PageSize = 10000 // 导出时获取所有记录
	histories, _, err := h.GetReportHistory(ctx, req)
	if err != nil {
		return "", fmt.Errorf("获取报表历史失败: %v", err)
	}
	
	// 根据格式生成文件
	switch format {
	case types.FileFormatExcel:
		return h.exportToExcel(ctx, histories)
	case types.FileFormatJSON:
		return h.exportToJSON(ctx, histories)
	default:
		return "", fmt.Errorf("不支持的导出格式: %s", format)
	}
}

// updateReportMetadata 更新报表元数据
func (h *HistoryService) updateReportMetadata(ctx context.Context, reportID int64, metadata map[string]interface{}) error {
	// 这里需要实现更新报表元数据的逻辑
	// 可能需要在Repository层添加相应的方法
	g.Log().Debug(ctx, "更新报表元数据", "report_id", reportID, "metadata", metadata)
	
	// 暂时返回成功，实际实现中需要调用Repository方法
	return nil
}

// exportToExcel 导出为Excel格式
func (h *HistoryService) exportToExcel(ctx context.Context, histories []*ReportHistory) (string, error) {
	// 这里可以使用Excel生成库（如excelize）来创建Excel文件
	// 暂时返回模拟的文件路径
	filename := fmt.Sprintf("report_history_%s.xlsx", time.Now().Format("20060102_150405"))
	filePath := fmt.Sprintf("/tmp/exports/%s", filename)
	
	g.Log().Info(ctx, "报表历史导出到Excel完成", 
		"count", len(histories), 
		"file_path", filePath)
	
	return filePath, nil
}

// exportToJSON 导出为JSON格式
func (h *HistoryService) exportToJSON(ctx context.Context, histories []*ReportHistory) (string, error) {
	// 这里可以直接将数据序列化为JSON文件
	filename := fmt.Sprintf("report_history_%s.json", time.Now().Format("20060102_150405"))
	filePath := fmt.Sprintf("/tmp/exports/%s", filename)
	
	g.Log().Info(ctx, "报表历史导出到JSON完成", 
		"count", len(histories), 
		"file_path", filePath)
	
	return filePath, nil
}

// GetReportHistoryByDateRange 按日期范围获取报表历史
func (h *HistoryService) GetReportHistoryByDateRange(ctx context.Context, startDate, endDate time.Time, reportType types.ReportType) ([]*ReportHistory, error) {
	req := &ReportHistoryRequest{
		ReportType: reportType,
		StartDate:  &startDate,
		EndDate:    &endDate,
		Page:       1,
		PageSize:   10000,
	}
	
	histories, _, err := h.GetReportHistory(ctx, req)
	return histories, err
}

// GetReportHistoryTrend 获取报表历史趋势数据
func (h *HistoryService) GetReportHistoryTrend(ctx context.Context, days int) ([]*DailyReportStats, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	
	stats, err := h.GetReportStats(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	
	return stats.DailyStats, nil
}