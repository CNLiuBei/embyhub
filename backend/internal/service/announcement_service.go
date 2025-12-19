// Package service 公告服务
package service

import (
	"errors"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AnnouncementService 公告服务
type AnnouncementService struct {
	db *gorm.DB
}

// NewAnnouncementService 创建公告服务
func NewAnnouncementService(db *gorm.DB) *AnnouncementService {
	return &AnnouncementService{db: db}
}

// CreateAnnouncementRequest 创建公告请求
type CreateAnnouncementRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Type    int8   `json:"type"`
	IsTop   bool   `json:"is_top"`
	Status  int8   `json:"status"`
}

// Create 创建公告
func (s *AnnouncementService) Create(adminID uuid.UUID, req *CreateAnnouncementRequest) (*models.Announcement, error) {
	if req.Type == 0 {
		req.Type = models.AnnouncementTypeNotice
	}

	announcement := &models.Announcement{
		Title:     req.Title,
		Content:   req.Content,
		Type:      req.Type,
		IsTop:     req.IsTop,
		Status:    req.Status,
		CreatedBy: adminID,
	}

	if err := s.db.Create(announcement).Error; err != nil {
		return nil, errors.New("创建公告失败")
	}

	return announcement, nil
}

// Update 更新公告
func (s *AnnouncementService) Update(id uint64, req *CreateAnnouncementRequest) error {
	updates := map[string]interface{}{
		"title":      req.Title,
		"content":    req.Content,
		"type":       req.Type,
		"is_top":     req.IsTop,
		"status":     req.Status,
		"updated_at": time.Now(),
	}

	result := s.db.Model(&models.Announcement{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return errors.New("更新公告失败")
	}
	if result.RowsAffected == 0 {
		return errors.New("公告不存在")
	}

	return nil
}

// Delete 删除公告
func (s *AnnouncementService) Delete(id uint64) error {
	result := s.db.Delete(&models.Announcement{}, id)
	if result.Error != nil {
		return errors.New("删除公告失败")
	}
	if result.RowsAffected == 0 {
		return errors.New("公告不存在")
	}

	return nil
}

// GetList 获取公告列表（管理员）
func (s *AnnouncementService) GetList(page, pageSize int, status *int8) ([]models.Announcement, int64, error) {
	var announcements []models.Announcement
	var total int64

	query := s.db.Model(&models.Announcement{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("is_top DESC, created_at DESC").Offset(offset).Limit(pageSize).Find(&announcements).Error; err != nil {
		return nil, 0, err
	}

	return announcements, total, nil
}

// GetPublished 获取已发布的公告（用户端）
func (s *AnnouncementService) GetPublished(limit int) ([]models.Announcement, error) {
	var announcements []models.Announcement

	if limit <= 0 {
		limit = 10
	}

	if err := s.db.Where("status = ?", models.AnnouncementStatusPublish).
		Order("is_top DESC, created_at DESC").
		Limit(limit).
		Find(&announcements).Error; err != nil {
		return nil, err
	}

	return announcements, nil
}

// GetByID 获取公告详情
func (s *AnnouncementService) GetByID(id uint64) (*models.Announcement, error) {
	var announcement models.Announcement
	if err := s.db.First(&announcement, id).Error; err != nil {
		return nil, errors.New("公告不存在")
	}
	return &announcement, nil
}

// Publish 发布公告
func (s *AnnouncementService) Publish(id uint64) error {
	result := s.db.Model(&models.Announcement{}).Where("id = ?", id).Update("status", models.AnnouncementStatusPublish)
	if result.RowsAffected == 0 {
		return errors.New("公告不存在")
	}
	return result.Error
}

// Offline 下线公告
func (s *AnnouncementService) Offline(id uint64) error {
	result := s.db.Model(&models.Announcement{}).Where("id = ?", id).Update("status", models.AnnouncementStatusOffline)
	if result.RowsAffected == 0 {
		return errors.New("公告不存在")
	}
	return result.Error
}
