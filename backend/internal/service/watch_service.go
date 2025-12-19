// Package service 观影记录服务
package service

import (
	"errors"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WatchService 观影记录服务
type WatchService struct {
	db *gorm.DB
}

// NewWatchService 创建观影记录服务
func NewWatchService(db *gorm.DB) *WatchService {
	return &WatchService{db: db}
}

// RecordWatch 记录观影
type RecordWatchRequest struct {
	VideoID   string `json:"video_id" binding:"required"`
	VideoName string `json:"video_name" binding:"required"`
	Duration  int    `json:"duration"`
	Progress  int    `json:"progress"`
}

func (s *WatchService) RecordWatch(userID uuid.UUID, req *RecordWatchRequest) error {
	// 查找是否有现有记录
	var history models.WatchHistory
	err := s.db.Where("user_id = ? AND video_id = ?", userID, req.VideoID).First(&history).Error

	if err == nil {
		// 更新现有记录
		history.Duration = req.Duration
		history.Progress = req.Progress
		history.UpdatedAt = time.Now()
		return s.db.Save(&history).Error
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新记录
		history = models.WatchHistory{
			UserID:    userID,
			VideoID:   req.VideoID,
			VideoName: req.VideoName,
			Duration:  req.Duration,
			Progress:  req.Progress,
		}
		return s.db.Create(&history).Error
	}

	return err
}

// GetWatchHistory 获取观影历史
func (s *WatchService) GetWatchHistory(userID uuid.UUID, page, pageSize int) ([]models.WatchHistory, int64, error) {
	var histories []models.WatchHistory
	var total int64

	query := s.db.Model(&models.WatchHistory{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// DeleteWatchHistory 删除观影记录
func (s *WatchService) DeleteWatchHistory(userID uuid.UUID, historyID uint64) error {
	return s.db.Where("id = ? AND user_id = ?", historyID, userID).Delete(&models.WatchHistory{}).Error
}

// ClearWatchHistory 清空观影记录
func (s *WatchService) ClearWatchHistory(userID uuid.UUID) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.WatchHistory{}).Error
}

// AddFavorite 添加收藏
type AddFavoriteRequest struct {
	VideoID   string `json:"video_id" binding:"required"`
	VideoName string `json:"video_name" binding:"required"`
	Cover     string `json:"cover"`
}

func (s *WatchService) AddFavorite(userID uuid.UUID, req *AddFavoriteRequest) error {
	// 检查是否已收藏
	var count int64
	s.db.Model(&models.Favorite{}).Where("user_id = ? AND video_id = ?", userID, req.VideoID).Count(&count)
	if count > 0 {
		return errors.New("已收藏该影片")
	}

	favorite := &models.Favorite{
		UserID:    userID,
		VideoID:   req.VideoID,
		VideoName: req.VideoName,
		Cover:     req.Cover,
	}
	return s.db.Create(favorite).Error
}

// GetFavorites 获取收藏列表
func (s *WatchService) GetFavorites(userID uuid.UUID, page, pageSize int) ([]models.Favorite, int64, error) {
	var favorites []models.Favorite
	var total int64

	query := s.db.Model(&models.Favorite{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&favorites).Error; err != nil {
		return nil, 0, err
	}

	return favorites, total, nil
}

// RemoveFavorite 取消收藏
func (s *WatchService) RemoveFavorite(userID uuid.UUID, videoID string) error {
	result := s.db.Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&models.Favorite{})
	if result.RowsAffected == 0 {
		return errors.New("未收藏该影片")
	}
	return result.Error
}

// IsFavorite 检查是否已收藏
func (s *WatchService) IsFavorite(userID uuid.UUID, videoID string) bool {
	var count int64
	s.db.Model(&models.Favorite{}).Where("user_id = ? AND video_id = ?", userID, videoID).Count(&count)
	return count > 0
}
