package task

import (
	"log"
	"time"

	"embyhub/pkg/database"
)

// CleanupTask 数据清理任务
type CleanupTask struct {
	interval time.Duration
	stopChan chan struct{}
}

// NewCleanupTask 创建清理任务
func NewCleanupTask(interval time.Duration) *CleanupTask {
	return &CleanupTask{
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start 启动清理任务
func (t *CleanupTask) Start() {
	log.Printf("[CleanupTask] 数据清理任务已启动，间隔: %v", t.interval)

	// 启动后延迟1分钟执行，避免启动时负载过高
	time.AfterFunc(1*time.Minute, func() {
		t.run()
	})

	// 定时执行
	ticker := time.NewTicker(t.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				t.run()
			case <-t.stopChan:
				ticker.Stop()
				log.Println("[CleanupTask] 数据清理任务已停止")
				return
			}
		}
	}()
}

// Stop 停止任务
func (t *CleanupTask) Stop() {
	close(t.stopChan)
}

// run 执行清理
func (t *CleanupTask) run() {
	startTime := time.Now()
	log.Println("[CleanupTask] 开始数据清理...")

	var totalCleaned int64

	// 1. 清理90天前的访问记录
	accessCleaned := t.cleanAccessRecords(90)
	totalCleaned += accessCleaned

	// 2. 清理30天前的审计日志（可选，根据需求调整）
	// auditCleaned := t.cleanAuditLogs(30)
	// totalCleaned += auditCleaned

	duration := time.Since(startTime)
	log.Printf("[CleanupTask] 数据清理完成，耗时: %v，清理记录: %d", duration, totalCleaned)
}

// cleanAccessRecords 清理访问记录
func (t *CleanupTask) cleanAccessRecords(days int) int64 {
	cutoff := time.Now().AddDate(0, 0, -days)

	result := database.DB.Exec(`
		DELETE FROM access_records 
		WHERE access_time < ?
	`, cutoff)

	if result.Error != nil {
		log.Printf("[CleanupTask] 清理访问记录失败: %v", result.Error)
		return 0
	}

	if result.RowsAffected > 0 {
		log.Printf("[CleanupTask] 清理访问记录: %d 条（%d天前）", result.RowsAffected, days)
	}

	return result.RowsAffected
}

// cleanAuditLogs 清理审计日志
func (t *CleanupTask) cleanAuditLogs(days int) int64 {
	cutoff := time.Now().AddDate(0, 0, -days)

	result := database.DB.Exec(`
		DELETE FROM audit_logs 
		WHERE created_at < ?
	`, cutoff)

	if result.Error != nil {
		log.Printf("[CleanupTask] 清理审计日志失败: %v", result.Error)
		return 0
	}

	if result.RowsAffected > 0 {
		log.Printf("[CleanupTask] 清理审计日志: %d 条（%d天前）", result.RowsAffected, days)
	}

	return result.RowsAffected
}

// CleanupStats 清理统计信息
type CleanupStats struct {
	AccessRecords int64  `json:"access_records"`
	AuditLogs     int64  `json:"audit_logs"`
	TotalCleaned  int64  `json:"total_cleaned"`
	Duration      string `json:"duration"`
}

// RunManualCleanup 手动执行清理（可通过API调用）
func (t *CleanupTask) RunManualCleanup(accessDays, auditDays int) *CleanupStats {
	startTime := time.Now()

	stats := &CleanupStats{}
	stats.AccessRecords = t.cleanAccessRecords(accessDays)
	stats.AuditLogs = t.cleanAuditLogs(auditDays)
	stats.TotalCleaned = stats.AccessRecords + stats.AuditLogs
	stats.Duration = time.Since(startTime).String()

	return stats
}
