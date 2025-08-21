package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// DashboardService 仪表板服务接口
type DashboardService interface {
	// 获取商户仪表板核心数据
	GetMerchantDashboard(ctx context.Context, tenantID, merchantID uint64) (*types.MerchantDashboardData, error)
	
	// 获取指定时间段业务统计
	GetMerchantStats(ctx context.Context, tenantID, merchantID uint64, period types.TimePeriod) (*types.MerchantDashboardData, error)
	
	// 获取权益使用趋势数据
	GetRightsUsageTrend(ctx context.Context, tenantID, merchantID uint64, days int) ([]types.RightsUsagePoint, error)
	
	// 获取待处理事项汇总
	GetPendingTasks(ctx context.Context, tenantID, merchantID uint64) ([]types.PendingTask, error)
	
	// 获取系统通知和公告
	GetNotifications(ctx context.Context, tenantID, merchantID uint64) (*NotificationsResponse, error)
	
	// 仪表板配置管理
	GetDashboardConfig(ctx context.Context, tenantID, merchantID uint64) (*types.DashboardConfig, error)
	SaveDashboardConfig(ctx context.Context, tenantID, merchantID uint64, config *DashboardConfigRequest) error
	UpdateDashboardConfig(ctx context.Context, tenantID, merchantID uint64, config *DashboardConfigRequest) error
	
	// 标记公告为已读
	MarkAnnouncementAsRead(ctx context.Context, tenantID, merchantID, announcementID uint64) error
}

// dashboardServiceImpl 仪表板服务实现
type dashboardServiceImpl struct {
	dashboardRepo repository.DashboardRepository
	cache         *gcache.Cache
}

// NewDashboardService 创建仪表板服务实例
func NewDashboardService() DashboardService {
	return &dashboardServiceImpl{
		dashboardRepo: repository.NewDashboardRepository(),
		cache:         gcache.New(),
	}
}

// 请求和响应结构

// DashboardConfigRequest 仪表板配置请求
type DashboardConfigRequest struct {
	LayoutConfig      *types.LayoutConfig       `json:"layout_config" binding:"required"`
	WidgetPreferences []types.WidgetPreference  `json:"widget_preferences"`
	RefreshInterval   int                       `json:"refresh_interval" binding:"min=60,max=3600"` // 1分钟到1小时
	MobileLayout      *types.MobileLayoutConfig `json:"mobile_layout"`
}

// NotificationsResponse 通知响应
type NotificationsResponse struct {
	Notifications []types.Notification `json:"notifications"`
	Announcements []types.Announcement `json:"announcements"`
	UnreadCount   int                  `json:"unread_count"`
}

// GetMerchantDashboard 获取商户仪表板核心数据
func (s *dashboardServiceImpl) GetMerchantDashboard(ctx context.Context, tenantID, merchantID uint64) (*types.MerchantDashboardData, error) {
	// 使用缓存键
	cacheKey := s.getDashboardCacheKey(tenantID, merchantID, types.TimePeriodDaily)
	
	// 尝试从缓存获取
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if data, ok := cached.(*types.MerchantDashboardData); ok {
			g.Log().Debugf(ctx, "Dashboard cache hit for merchant %d", merchantID)
			return data, nil
		}
	}
	
	// 从数据库获取
	data, err := s.dashboardRepo.GetMerchantDashboardData(ctx, tenantID, merchantID, types.TimePeriodDaily)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取商户 %d 仪表板数据失败", merchantID)
	}
	
	// 缓存5分钟
	s.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	
	return data, nil
}

// GetMerchantStats 获取指定时间段业务统计
func (s *dashboardServiceImpl) GetMerchantStats(ctx context.Context, tenantID, merchantID uint64, period types.TimePeriod) (*types.MerchantDashboardData, error) {
	cacheKey := s.getDashboardCacheKey(tenantID, merchantID, period)
	
	// 尝试从缓存获取
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if data, ok := cached.(*types.MerchantDashboardData); ok {
			return data, nil
		}
	}
	
	// 从数据库获取
	data, err := s.dashboardRepo.GetMerchantDashboardData(ctx, tenantID, merchantID, period)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取商户 %d 统计数据失败，周期: %v", merchantID, period)
	}
	
	// 根据时间周期设置不同的缓存时间
	var cacheDuration time.Duration
	switch period {
	case types.TimePeriodDaily:
		cacheDuration = 5 * time.Minute
	case types.TimePeriodWeekly:
		cacheDuration = 15 * time.Minute
	case types.TimePeriodMonthly:
		cacheDuration = 30 * time.Minute
	default:
		cacheDuration = 5 * time.Minute
	}
	
	s.cache.Set(ctx, cacheKey, data, cacheDuration)
	
	return data, nil
}

// GetRightsUsageTrend 获取权益使用趋势数据
func (s *dashboardServiceImpl) GetRightsUsageTrend(ctx context.Context, tenantID, merchantID uint64, days int) ([]types.RightsUsagePoint, error) {
	// 验证天数范围
	if days < 1 || days > 365 {
		return nil, gerror.New("天数范围必须在 1-365 之间")
	}
	
	cacheKey := s.getRightsTrendCacheKey(tenantID, merchantID, days)
	
	// 尝试从缓存获取
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if trends, ok := cached.([]types.RightsUsagePoint); ok {
			return trends, nil
		}
	}
	
	// 从数据库获取
	trends, err := s.dashboardRepo.GetRightsUsageTrend(ctx, tenantID, merchantID, days)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取商户 %d 权益趋势失败，天数: %d", merchantID, days)
	}
	
	// 缓存10分钟
	s.cache.Set(ctx, cacheKey, trends, 10*time.Minute)
	
	return trends, nil
}

// GetPendingTasks 获取待处理事项汇总
func (s *dashboardServiceImpl) GetPendingTasks(ctx context.Context, tenantID, merchantID uint64) ([]types.PendingTask, error) {
	cacheKey := s.getPendingTasksCacheKey(tenantID, merchantID)
	
	// 尝试从缓存获取 (较短的缓存时间，因为待处理事项变化频繁)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if tasks, ok := cached.([]types.PendingTask); ok {
			return tasks, nil
		}
	}
	
	// 从数据库获取
	tasks, err := s.dashboardRepo.GetPendingTasks(ctx, tenantID, merchantID)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取商户 %d 待处理事项失败", merchantID)
	}
	
	// 缓存2分钟 (待处理事项需要较新的数据)
	s.cache.Set(ctx, cacheKey, tasks, 2*time.Minute)
	
	return tasks, nil
}

// GetNotifications 获取系统通知和公告
func (s *dashboardServiceImpl) GetNotifications(ctx context.Context, tenantID, merchantID uint64) (*NotificationsResponse, error) {
	cacheKey := s.getNotificationsCacheKey(tenantID, merchantID)
	
	// 尝试从缓存获取
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if resp, ok := cached.(*NotificationsResponse); ok {
			return resp, nil
		}
	}
	
	// 从数据库获取
	notifications, announcements, err := s.dashboardRepo.GetMerchantNotifications(ctx, tenantID, merchantID, 10)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取商户 %d 通知公告失败", merchantID)
	}
	
	// 计算未读数量
	unreadCount := 0
	for _, notification := range notifications {
		if notification.ReadAt == nil {
			unreadCount++
		}
	}
	for _, announcement := range announcements {
		if !announcement.ReadStatus {
			unreadCount++
		}
	}
	
	response := &NotificationsResponse{
		Notifications: notifications,
		Announcements: announcements,
		UnreadCount:   unreadCount,
	}
	
	// 缓存3分钟
	s.cache.Set(ctx, cacheKey, response, 3*time.Minute)
	
	return response, nil
}

// GetDashboardConfig 获取仪表板配置
func (s *dashboardServiceImpl) GetDashboardConfig(ctx context.Context, tenantID, merchantID uint64) (*types.DashboardConfig, error) {
	cacheKey := s.getConfigCacheKey(tenantID, merchantID)
	
	// 尝试从缓存获取
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if config, ok := cached.(*types.DashboardConfig); ok {
			return config, nil
		}
	}
	
	// 从数据库获取
	config, err := s.dashboardRepo.GetDashboardConfig(ctx, tenantID, merchantID)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取商户 %d 仪表板配置失败", merchantID)
	}
	
	// 缓存30分钟 (配置变更不频繁)
	s.cache.Set(ctx, cacheKey, config, 30*time.Minute)
	
	return config, nil
}

// SaveDashboardConfig 保存仪表板配置
func (s *dashboardServiceImpl) SaveDashboardConfig(ctx context.Context, tenantID, merchantID uint64, request *DashboardConfigRequest) error {
	// 验证配置
	if err := s.validateDashboardConfig(request); err != nil {
		return gerror.Wrap(err, "仪表板配置验证失败")
	}
	
	// 转换为实体
	config := &types.DashboardConfig{
		MerchantID:        merchantID,
		LayoutConfig:      request.LayoutConfig,
		WidgetPreferences: request.WidgetPreferences,
		RefreshInterval:   request.RefreshInterval,
		MobileLayout:      request.MobileLayout,
	}
	
	// 保存到数据库
	if err := s.dashboardRepo.SaveDashboardConfig(ctx, tenantID, merchantID, config); err != nil {
		return gerror.Wrapf(err, "保存商户 %d 仪表板配置失败", merchantID)
	}
	
	// 清除缓存
	s.cache.Remove(ctx, s.getConfigCacheKey(tenantID, merchantID))
	s.cache.Remove(ctx, s.getDashboardCacheKey(tenantID, merchantID, types.TimePeriodDaily))
	
	g.Log().Infof(ctx, "商户 %d 仪表板配置保存成功", merchantID)
	
	return nil
}

// UpdateDashboardConfig 更新仪表板配置
func (s *dashboardServiceImpl) UpdateDashboardConfig(ctx context.Context, tenantID, merchantID uint64, request *DashboardConfigRequest) error {
	return s.SaveDashboardConfig(ctx, tenantID, merchantID, request)
}

// MarkAnnouncementAsRead 标记公告为已读
func (s *dashboardServiceImpl) MarkAnnouncementAsRead(ctx context.Context, tenantID, merchantID, announcementID uint64) error {
	if err := s.dashboardRepo.MarkAnnouncementAsRead(ctx, tenantID, merchantID, announcementID); err != nil {
		return gerror.Wrapf(err, "标记商户 %d 公告 %d 已读失败", merchantID, announcementID)
	}
	
	// 清除通知缓存
	s.cache.Remove(ctx, s.getNotificationsCacheKey(tenantID, merchantID))
	
	g.Log().Infof(ctx, "商户 %d 公告 %d 标记已读成功", merchantID, announcementID)
	
	return nil
}

// 辅助方法

// validateDashboardConfig 验证仪表板配置
func (s *dashboardServiceImpl) validateDashboardConfig(config *DashboardConfigRequest) error {
	if config.LayoutConfig == nil {
		return gerror.New("布局配置不能为空")
	}
	
	if config.LayoutConfig.Columns < 1 || config.LayoutConfig.Columns > 12 {
		return gerror.New("布局列数必须在 1-12 之间")
	}
	
	// 验证组件数量
	if len(config.LayoutConfig.Widgets) > 20 {
		return gerror.New("仪表板组件数量不能超过20个")
	}
	
	// 验证组件类型
	validTypes := map[types.WidgetType]bool{
		types.WidgetTypeSalesOverview:  true,
		types.WidgetTypeRightsBalance:  true,
		types.WidgetTypeRightsTrend:    true,
		types.WidgetTypePendingTasks:   true,
		types.WidgetTypeRecentOrders:   true,
		types.WidgetTypeAnnouncements:  true,
		types.WidgetTypeQuickActions:   true,
	}
	
	for _, widget := range config.LayoutConfig.Widgets {
		if !validTypes[widget.Type] {
			return gerror.Newf("无效的组件类型: %v", widget.Type)
		}
		
		if widget.Size.Width < 1 || widget.Size.Width > config.LayoutConfig.Columns {
			return gerror.Newf("组件 %s 宽度无效", widget.ID)
		}
		
		if widget.Size.Height < 1 || widget.Size.Height > 10 {
			return gerror.Newf("组件 %s 高度无效", widget.ID)
		}
	}
	
	return nil
}

// 缓存键生成方法

func (s *dashboardServiceImpl) getDashboardCacheKey(tenantID, merchantID uint64, period types.TimePeriod) string {
	return fmt.Sprintf("dashboard:merchant:%d:%d:%v", tenantID, merchantID, period)
}

func (s *dashboardServiceImpl) getRightsTrendCacheKey(tenantID, merchantID uint64, days int) string {
	return fmt.Sprintf("rights_trend:merchant:%d:%d:%d", tenantID, merchantID, days)
}

func (s *dashboardServiceImpl) getPendingTasksCacheKey(tenantID, merchantID uint64) string {
	return fmt.Sprintf("pending_tasks:merchant:%d:%d", tenantID, merchantID)
}

func (s *dashboardServiceImpl) getNotificationsCacheKey(tenantID, merchantID uint64) string {
	return fmt.Sprintf("notifications:merchant:%d:%d", tenantID, merchantID)
}

func (s *dashboardServiceImpl) getConfigCacheKey(tenantID, merchantID uint64) string {
	return fmt.Sprintf("dashboard_config:merchant:%d:%d", tenantID, merchantID)
}