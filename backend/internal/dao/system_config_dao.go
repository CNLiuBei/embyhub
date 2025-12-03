package dao

import (
	"embyhub/internal/model"
	"embyhub/pkg/database"
)

type SystemConfigDAO struct{}

func NewSystemConfigDAO() *SystemConfigDAO {
	return &SystemConfigDAO{}
}

// Get 根据Key获取配置
func (d *SystemConfigDAO) Get(configKey string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := database.DB.Where("config_key = ?", configKey).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// List 获取所有配置
func (d *SystemConfigDAO) List() ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := database.DB.Order("config_key ASC").Find(&configs).Error
	return configs, err
}

// Update 更新配置
func (d *SystemConfigDAO) Update(config *model.SystemConfig) error {
	return database.DB.Save(config).Error
}

// BatchGet 批量获取配置
func (d *SystemConfigDAO) BatchGet(keys []string) (map[string]string, error) {
	var configs []*model.SystemConfig
	err := database.DB.Where("config_key IN ?", keys).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, config := range configs {
		result[config.ConfigKey] = config.ConfigValue
	}
	return result, nil
}
