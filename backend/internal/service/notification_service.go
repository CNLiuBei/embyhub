// Package service 通知服务
package service

import (
	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationService 通知服务
type NotificationService struct {
	db *gorm.DB
}

// NewNotificationService 创建通知服务
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// GetNotifications 获取通知列表
func (s *NotificationService) GetNotifications(userID uuid.UUID, page, pageSize int, onlyUnread bool) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := s.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	if onlyUnread {
		query = query.Where("is_read = ?", false)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// GetUnreadCount 获取未读数量
func (s *NotificationService) GetUnreadCount(userID uuid.UUID) int64 {
	var count int64
	s.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count)
	return count
}

// MarkAsRead 标记为已读
func (s *NotificationService) MarkAsRead(userID uuid.UUID, notificationID uint64) error {
	return s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

// MarkAllAsRead 全部标记为已读
func (s *NotificationService) MarkAllAsRead(userID uuid.UUID) error {
	return s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

// DeleteNotification 删除通知
func (s *NotificationService) DeleteNotification(userID uuid.UUID, notificationID uint64) error {
	return s.db.Where("id = ? AND user_id = ?", notificationID, userID).Delete(&models.Notification{}).Error
}

// SendNotification 发送通知
func (s *NotificationService) SendNotification(userID uuid.UUID, title, content string, notifyType int8) error {
	notification := &models.Notification{
		UserID:  userID,
		Title:   title,
		Content: content,
		Type:    notifyType,
		IsRead:  false,
	}
	return s.db.Create(notification).Error
}

// SendBatchNotification 批量发送通知
func (s *NotificationService) SendBatchNotification(userIDs []uuid.UUID, title, content string, notifyType int8) error {
	notifications := make([]models.Notification, len(userIDs))
	for i, userID := range userIDs {
		notifications[i] = models.Notification{
			UserID:  userID,
			Title:   title,
			Content: content,
			Type:    notifyType,
			IsRead:  false,
		}
	}
	return s.db.CreateInBatches(notifications, 100).Error
}

// SendSystemNotificationToAll 向所有用户发送系统通知
func (s *NotificationService) SendSystemNotificationToAll(title, content string) error {
	var userIDs []uuid.UUID
	if err := s.db.Model(&models.User{}).Where("status = ?", 1).Pluck("id", &userIDs).Error; err != nil {
		return err
	}
	return s.SendBatchNotification(userIDs, title, content, 1)
}
