package service

import (
	"encoding/json"

	"embyhub/internal/model"
	"embyhub/pkg/database"
)

// AuditService 审计日志服务
type AuditService struct{}

// NewAuditService 创建审计日志服务
func NewAuditService() *AuditService {
	return &AuditService{}
}

// Log 记录审计日志
func (s *AuditService) Log(log *model.AuditLog) error {
	return database.DB.Create(log).Error
}

// LogAction 便捷方法：记录操作日志
func (s *AuditService) LogAction(userID *int, username, action, targetType, targetID string, detail interface{}, ip, ua, status string) {
	var detailStr string
	if detail != nil {
		if bytes, err := json.Marshal(detail); err == nil {
			detailStr = string(bytes)
		}
	}

	log := &model.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detailStr,
		IPAddress:  ip,
		UserAgent:  ua,
		Status:     status,
	}

	// 异步写入，不阻塞主业务
	go func() {
		database.DB.Create(log)
	}()
}

// List 查询审计日志列表
func (s *AuditService) List(query *model.AuditLogQuery) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	db := database.DB.Model(&model.AuditLog{})

	// 条件过滤
	if query.UserID != nil {
		db = db.Where("user_id = ?", *query.UserID)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.TargetType != "" {
		db = db.Where("target_type = ?", query.TargetType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	// 统计总数
	db.Count(&total)

	// 分页
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	// 查询
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error

	return logs, total, err
}

// GetUserActions 获取用户操作记录
func (s *AuditService) GetUserActions(userID int, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetRecentActions 获取最近操作记录
func (s *AuditService) GetRecentActions(limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := database.DB.Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetActionStats 获取操作统计
func (s *AuditService) GetActionStats() (map[string]int64, error) {
	var results []struct {
		Action string
		Count  int64
	}

	err := database.DB.Model(&model.AuditLog{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Find(&results).Error

	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Action] = r.Count
	}

	return stats, err
}

// 全局审计服务实例
var auditService = NewAuditService()

// Audit 全局审计记录函数
func Audit(userID *int, username, action, targetType, targetID string, detail interface{}, ip, ua, status string) {
	auditService.LogAction(userID, username, action, targetType, targetID, detail, ip, ua, status)
}
