package model

import "time"

// SystemConfig 系统配置模型
type SystemConfig struct {
	ConfigKey   string    `gorm:"column:config_key;primaryKey;type:varchar(50)" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value;type:text;not null" json:"config_value"`
	Description string    `gorm:"column:description;type:varchar(200)" json:"description"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_configs"
}

// SystemConfigUpdateRequest 更新系统配置请求
type SystemConfigUpdateRequest struct {
	ConfigValue string `json:"config_value"`
}

// SystemConfigListResponse 系统配置列表响应
type SystemConfigListResponse struct {
	Total int             `json:"total"`
	List  []*SystemConfig `json:"list"`
}
