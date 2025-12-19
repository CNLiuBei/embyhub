// Package service 邮件服务助手
package service

import (
	"encoding/json"
	"sync"

	"feiniu-user-system/internal/config"
	"feiniu-user-system/internal/models"
	"feiniu-user-system/pkg/email"

	"gorm.io/gorm"
)

// emailServiceCache 邮件服务缓存
var (
	emailServiceInstance *email.Service
	emailServiceMu       sync.Mutex
	emailConfigHash      string
)

// GetEmailServiceFromDB 从数据库获取邮件服务实例
// 会缓存实例，配置变更时自动重建
func GetEmailServiceFromDB(db *gorm.DB) *email.Service {
	emailServiceMu.Lock()
	defer emailServiceMu.Unlock()

	// 从数据库获取邮件设置
	var setting models.Setting
	err := db.Where("\"key\" = ?", "email").First(&setting).Error
	if err != nil {
		return nil
	}

	// 检查配置是否变更
	if emailServiceInstance != nil && emailConfigHash == setting.Value {
		return emailServiceInstance
	}

	// 解析配置
	var settings EmailSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil
	}

	if !settings.Enabled {
		emailServiceInstance = nil
		emailConfigHash = setting.Value
		return nil
	}

	// 创建新的邮件服务实例
	cfg := &config.EmailConfig{
		Enabled:  settings.Enabled,
		Host:     settings.Host,
		Port:     settings.Port,
		Username: settings.Username,
		Password: settings.Password,
		From:     settings.From,
		FromName: settings.FromName,
	}

	// 关闭旧连接
	if emailServiceInstance != nil {
		emailServiceInstance.Close()
	}

	emailServiceInstance = email.NewService(cfg)
	emailConfigHash = setting.Value

	return emailServiceInstance
}
