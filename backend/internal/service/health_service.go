// Package service 系统健康监控服务
package service

import (
	"context"
	"runtime"
	"time"

	"feiniu-user-system/internal/database"

	"gorm.io/gorm"
)

// HealthService 健康监控服务
type HealthService struct {
	db        *gorm.DB
	startTime time.Time
}

// NewHealthService 创建健康监控服务
func NewHealthService(db *gorm.DB) *HealthService {
	return &HealthService{
		db:        db,
		startTime: time.Now(),
	}
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string            `json:"status"`
	Uptime    string            `json:"uptime"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
	System    SystemInfo        `json:"system"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	MemAlloc     string `json:"mem_alloc"`
	MemSys       string `json:"mem_sys"`
}

// GetHealth 获取健康状态
func (s *HealthService) GetHealth() *HealthStatus {
	status := &HealthStatus{
		Status:    "healthy",
		Uptime:    s.formatDuration(time.Since(s.startTime)),
		Timestamp: time.Now().Format(time.RFC3339),
		Services:  make(map[string]string),
	}

	// 检查数据库
	if err := s.checkDatabase(); err != nil {
		status.Services["database"] = "unhealthy: " + err.Error()
		status.Status = "unhealthy"
	} else {
		status.Services["database"] = "healthy"
	}

	// 检查Redis
	if err := s.checkRedis(); err != nil {
		status.Services["redis"] = "unhealthy: " + err.Error()
		status.Status = "unhealthy"
	} else {
		status.Services["redis"] = "healthy"
	}

	// 获取系统信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status.System = SystemInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		MemAlloc:     s.formatBytes(m.Alloc),
		MemSys:       s.formatBytes(m.Sys),
	}

	return status
}

// checkDatabase 检查数据库连接
func (s *HealthService) checkDatabase() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// checkRedis 检查Redis连接
func (s *HealthService) checkRedis() error {
	ctx := context.Background()
	return database.GetRedis().Ping(ctx).Err()
}

// formatDuration 格式化时间间隔
func (s *HealthService) formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return formatTimePart(days, "天") + formatTimePart(hours, "小时") + formatTimePart(minutes, "分钟")
	}
	if hours > 0 {
		return formatTimePart(hours, "小时") + formatTimePart(minutes, "分钟")
	}
	return formatTimePart(minutes, "分钟")
}

func formatTimePart(value int, unit string) string {
	if value > 0 {
		return string(rune('0'+value/10)) + string(rune('0'+value%10)) + unit
	}
	return ""
}

// formatBytes 格式化字节数
func (s *HealthService) formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return formatUint(b) + " B"
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return formatFloat(float64(b)/float64(div)) + " " + string("KMGTPE"[exp]) + "B"
}

func formatUint(n uint64) string {
	return string(rune('0' + n))
}

func formatFloat(f float64) string {
	whole := int(f)
	frac := int((f - float64(whole)) * 10)
	return string(rune('0'+whole/10)) + string(rune('0'+whole%10)) + "." + string(rune('0'+frac))
}

// GetStats 获取统计信息
type SystemStats struct {
	TotalUsers     int64 `json:"total_users"`
	TotalMembers   int64 `json:"total_members"`
	TotalCards     int64 `json:"total_cards"`
	UsedCards      int64 `json:"used_cards"`
	TodayLogins    int64 `json:"today_logins"`
	TodayRegisters int64 `json:"today_registers"`
	ActiveSessions int64 `json:"active_sessions"`
	Announcements  int64 `json:"announcements"`
	BlockedIPs     int64 `json:"blocked_ips"`
}

func (s *HealthService) GetStats() *SystemStats {
	stats := &SystemStats{}
	today := time.Now().Format("2006-01-02")

	s.db.Table("users").Where("deleted_at IS NULL").Count(&stats.TotalUsers)
	s.db.Table("users").Where("deleted_at IS NULL AND member_level > 0").Count(&stats.TotalMembers)
	s.db.Table("cards").Count(&stats.TotalCards)
	s.db.Table("cards").Where("status = 1").Count(&stats.UsedCards)
	s.db.Table("login_logs").Where("DATE(created_at) = ?", today).Count(&stats.TodayLogins)
	s.db.Table("users").Where("DATE(created_at) = ?", today).Count(&stats.TodayRegisters)
	s.db.Table("announcements").Where("status = 1").Count(&stats.Announcements)
	s.db.Table("ip_blacklists").Count(&stats.BlockedIPs)

	return stats
}
