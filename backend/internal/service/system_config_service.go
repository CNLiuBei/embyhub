package service

import (
	"embyhub/internal/dao"
	"embyhub/internal/model"
	"time"
)

type SystemConfigService struct {
	configDAO *dao.SystemConfigDAO
}

func NewSystemConfigService() *SystemConfigService {
	return &SystemConfigService{
		configDAO: dao.NewSystemConfigDAO(),
	}
}

// Get 获取配置
func (s *SystemConfigService) Get(configKey string) (*model.SystemConfig, error) {
	return s.configDAO.Get(configKey)
}

// List 获取所有配置
func (s *SystemConfigService) List() (*model.SystemConfigListResponse, error) {
	configs, err := s.configDAO.List()
	if err != nil {
		return nil, err
	}

	return &model.SystemConfigListResponse{
		Total: len(configs),
		List:  configs,
	}, nil
}

// Update 更新配置（不存在则创建）
func (s *SystemConfigService) Update(configKey string, configValue string) error {
	config, err := s.configDAO.Get(configKey)
	if err != nil {
		// 配置项不存在，创建新的
		config = &model.SystemConfig{
			ConfigKey:   configKey,
			ConfigValue: configValue,
			Description: "",
			UpdatedAt:   time.Now(),
		}
	} else {
		config.ConfigValue = configValue
		config.UpdatedAt = time.Now()
	}

	return s.configDAO.Update(config)
}
