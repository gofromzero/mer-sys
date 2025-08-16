package health

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/spume/mer-sys/shared/config"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  string                 `json:"duration"`
}

// SystemHealth 系统整体健康状态
type SystemHealth struct {
	Status     HealthStatus               `json:"status"`
	Version    string                     `json:"version"`
	Service    string                     `json:"service"`
	Timestamp  time.Time                  `json:"timestamp"`
	Uptime     string                     `json:"uptime"`
	Components map[string]ComponentHealth `json:"components"`
	System     SystemInfo                 `json:"system"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	MemoryMB     string `json:"memory_mb"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	startTime   time.Time
	serviceName string
	version     string
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(serviceName, version string) *HealthChecker {
	return &HealthChecker{
		startTime:   time.Now(),
		serviceName: serviceName,
		version:     version,
	}
}

// CheckHealth 执行健康检查
func (h *HealthChecker) CheckHealth(ctx context.Context) *SystemHealth {
	components := make(map[string]ComponentHealth)
	overallStatus := HealthStatusHealthy

	// 检查数据库连接
	dbHealth := h.checkDatabase(ctx)
	components["database"] = dbHealth
	if dbHealth.Status != HealthStatusHealthy {
		overallStatus = HealthStatusDegraded
	}

	// 检查Redis连接
	redisHealth := h.checkRedis(ctx)
	components["redis"] = redisHealth
	if redisHealth.Status != HealthStatusHealthy {
		if overallStatus == HealthStatusDegraded {
			overallStatus = HealthStatusUnhealthy
		} else {
			overallStatus = HealthStatusDegraded
		}
	}

	// 检查系统资源
	systemHealth := h.checkSystem(ctx)
	components["system"] = systemHealth

	return &SystemHealth{
		Status:     overallStatus,
		Version:    h.version,
		Service:    h.serviceName,
		Timestamp:  time.Now(),
		Uptime:     time.Since(h.startTime).String(),
		Components: components,
		System:     h.getSystemInfo(),
	}
}

// checkDatabase 检查数据库连接
func (h *HealthChecker) checkDatabase(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:      "database",
		Timestamp: time.Now(),
	}

	defer func() {
		health.Duration = time.Since(start).String()
	}()

	// 检查数据库连接
	db := config.GetDB()
	if db == nil {
		health.Status = HealthStatusUnhealthy
		health.Message = "数据库连接未初始化"
		return health
	}

	// 执行简单查询测试连接
	_, err := db.GetOne(ctx, "SELECT 1")
	if err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = fmt.Sprintf("数据库连接失败: %v", err)
		return health
	}

	// 获取数据库统计信息
	stats := db.GetStats()
	health.Status = HealthStatusHealthy
	health.Message = "数据库连接正常"
	health.Details = map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
	}

	return health
}

// checkRedis 检查Redis连接
func (h *HealthChecker) checkRedis(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:      "redis",
		Timestamp: time.Now(),
	}

	defer func() {
		health.Duration = time.Since(start).String()
	}()

	// 检查Redis连接
	redis := config.GetRedis()
	if redis == nil {
		health.Status = HealthStatusUnhealthy
		health.Message = "Redis连接未初始化"
		return health
	}

	// 执行PING命令测试连接
	result, err := redis.Do(ctx, "PING")
	if err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = fmt.Sprintf("Redis连接失败: %v", err)
		return health
	}

	if result != "PONG" {
		health.Status = HealthStatusUnhealthy
		health.Message = "Redis PING 响应异常"
		return health
	}

	// 获取Redis信息
	info, err := redis.Do(ctx, "INFO", "server")
	if err == nil {
		health.Details = map[string]interface{}{
			"ping_response": result,
			"server_info":   info,
		}
	}

	health.Status = HealthStatusHealthy
	health.Message = "Redis连接正常"

	return health
}

// checkSystem 检查系统资源
func (h *HealthChecker) checkSystem(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:      "system",
		Timestamp: time.Now(),
		Status:    HealthStatusHealthy,
		Message:   "系统资源正常",
	}

	defer func() {
		health.Duration = time.Since(start).String()
	}()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	health.Details = map[string]interface{}{
		"goroutines":   runtime.NumGoroutine(),
		"memory_alloc": fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
		"memory_total": fmt.Sprintf("%.2f MB", float64(m.TotalAlloc)/1024/1024),
		"memory_sys":   fmt.Sprintf("%.2f MB", float64(m.Sys)/1024/1024),
		"gc_runs":      m.NumGC,
		"last_gc":      time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
	}

	// 检查Goroutine数量是否过多
	if runtime.NumGoroutine() > 10000 {
		health.Status = HealthStatusDegraded
		health.Message = "Goroutine数量过多"
	}

	// 检查内存使用是否过高（超过1GB）
	if m.Alloc > 1024*1024*1024 {
		health.Status = HealthStatusDegraded
		health.Message = "内存使用过高"
	}

	return health
}

// getSystemInfo 获取系统信息
func (h *HealthChecker) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		MemoryMB:     fmt.Sprintf("%.2f", float64(m.Alloc)/1024/1024),
	}
}

// CheckDependency 检查单个依赖项
func (h *HealthChecker) CheckDependency(ctx context.Context, name string) ComponentHealth {
	switch name {
	case "database", "db", "mysql":
		return h.checkDatabase(ctx)
	case "redis", "cache":
		return h.checkRedis(ctx)
	case "system":
		return h.checkSystem(ctx)
	default:
		return ComponentHealth{
			Name:      name,
			Status:    HealthStatusUnhealthy,
			Message:   "未知的依赖项",
			Timestamp: time.Now(),
		}
	}
}

// IsHealthy 检查系统是否健康
func (h *HealthChecker) IsHealthy(ctx context.Context) bool {
	health := h.CheckHealth(ctx)
	return health.Status == HealthStatusHealthy
}

// GetReadiness 获取就绪状态（更严格的检查）
func (h *HealthChecker) GetReadiness(ctx context.Context) *SystemHealth {
	health := h.CheckHealth(ctx)

	// 就绪检查要求所有组件都必须健康
	for _, component := range health.Components {
		if component.Status != HealthStatusHealthy {
			health.Status = HealthStatusUnhealthy
			break
		}
	}

	return health
}

// GetLiveness 获取存活状态（基础检查）
func (h *HealthChecker) GetLiveness(ctx context.Context) *SystemHealth {
	// 存活检查只检查基本的系统状态
	health := &SystemHealth{
		Status:     HealthStatusHealthy,
		Version:    h.version,
		Service:    h.serviceName,
		Timestamp:  time.Now(),
		Uptime:     time.Since(h.startTime).String(),
		Components: make(map[string]ComponentHealth),
		System:     h.getSystemInfo(),
	}

	// 只检查系统资源
	systemCheck := h.checkSystem(ctx)
	health.Components["system"] = systemCheck
	health.Status = systemCheck.Status

	return health
}
