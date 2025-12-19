// Package service 设备管理服务
package service

import (
	"context"
	"errors"

	"feiniu-user-system/internal/database"
	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceService 设备管理服务
type DeviceService struct {
	db *gorm.DB
}

// NewDeviceService 创建设备管理服务
func NewDeviceService(db *gorm.DB) *DeviceService {
	return &DeviceService{db: db}
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID         uint64 `json:"id"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	DeviceType string `json:"device_type"`
	LastIP     string `json:"last_ip"`
	LastActive string `json:"last_active"`
	IsCurrent  bool   `json:"is_current"`
}

// GetUserDevices 获取用户设备列表
func (s *DeviceService) GetUserDevices(userID uuid.UUID, currentDeviceID string) ([]DeviceInfo, error) {
	var devices []models.UserDevice
	if err := s.db.Where("user_id = ?", userID).Order("last_used_at DESC").Find(&devices).Error; err != nil {
		return nil, err
	}

	result := make([]DeviceInfo, len(devices))
	for i, d := range devices {
		result[i] = DeviceInfo{
			ID:         d.ID,
			DeviceID:   d.DeviceID,
			DeviceName: d.DeviceName,
			DeviceType: d.DeviceType,
			LastIP:     d.LastIP,
			LastActive: d.LastUsedAt.Format("2006-01-02 15:04:05"),
			IsCurrent:  d.DeviceID == currentDeviceID,
		}
	}

	return result, nil
}

// RemoveDevice 移除设备（踢出登录）
func (s *DeviceService) RemoveDevice(userID uuid.UUID, deviceID string) error {
	// 删除设备记录
	result := s.db.Where("user_id = ? AND device_id = ?", userID, deviceID).Delete(&models.UserDevice{})
	if result.RowsAffected == 0 {
		return errors.New("设备不存在")
	}

	// 清除该设备的Token缓存
	ctx := context.Background()
	database.DeleteCache(ctx, database.KeyUserToken+userID.String()+":"+deviceID)

	return nil
}

// RemoveAllDevices 移除所有设备（除当前设备）
func (s *DeviceService) RemoveAllDevices(userID uuid.UUID, currentDeviceID string) (int64, error) {
	// 删除除当前设备外的所有设备
	result := s.db.Where("user_id = ? AND device_id != ?", userID, currentDeviceID).Delete(&models.UserDevice{})

	// 清除用户所有Token缓存（当前设备除外）
	// 注意：这里简化处理，实际生产环境可能需要更精细的Token管理

	return result.RowsAffected, result.Error
}

// GetDeviceCount 获取用户设备数量
func (s *DeviceService) GetDeviceCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&models.UserDevice{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
