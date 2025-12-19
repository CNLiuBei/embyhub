// Package service 业务逻辑层
package service

import (
	"encoding/json"
	"errors"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SettingService 设置服务
type SettingService struct {
	db *gorm.DB
}

// NewSettingService 创建设置服务实例
func NewSettingService(db *gorm.DB) *SettingService {
	return &SettingService{db: db}
}

// ============= 邮件设置 =============

// EmailSettings 邮件设置结构
type EmailSettings struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	FromName string `json:"from_name"`
}

// GetEmailSettings 获取邮件设置
func (s *SettingService) GetEmailSettings() (*EmailSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "email").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &EmailSettings{
				Enabled: false,
				Port:    587,
			}, nil
		}
		return nil, err
	}

	var settings EmailSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveEmailSettings 保存邮件设置
func (s *SettingService) SaveEmailSettings(settings *EmailSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "email").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "email",
				Value:     string(value),
				Type:      "email",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// ============= 域名白名单设置 =============

// DomainSettings 域名白名单设置结构
type DomainSettings struct {
	Enabled bool     `json:"enabled"` // 是否启用域名白名单
	Domains []string `json:"domains"` // 允许的域名列表
}

// GetDomainSettings 获取域名白名单设置
func (s *SettingService) GetDomainSettings() (*DomainSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "domain_whitelist").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &DomainSettings{
				Enabled: false,
				Domains: []string{"localhost", "127.0.0.1"},
			}, nil
		}
		return nil, err
	}

	var settings DomainSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveDomainSettings 保存域名白名单设置
func (s *SettingService) SaveDomainSettings(settings *DomainSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "domain_whitelist").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "domain_whitelist",
				Value:     string(value),
				Type:      "domain",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// ============= 注册设置 =============

// RegisterSettings 注册设置结构
type RegisterSettings struct {
	Enabled          bool `json:"enabled"`             // 是否允许注册
	GiftMemberDays   int  `json:"gift_member_days"`    // 注册赠送会员天数（0表示不赠送）
	AutoDisableOnExp bool `json:"auto_disable_on_exp"` // 会员到期后自动禁用账户
}

// GetRegisterSettings 获取注册设置
func (s *SettingService) GetRegisterSettings() (*RegisterSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "register").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &RegisterSettings{
				Enabled:          true, // 默认允许注册
				GiftMemberDays:   0,    // 默认不赠送
				AutoDisableOnExp: true, // 默认会员到期后禁用
			}, nil
		}
		return nil, err
	}

	var settings RegisterSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveRegisterSettings 保存注册设置
func (s *SettingService) SaveRegisterSettings(settings *RegisterSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "register").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "register",
				Value:     string(value),
				Type:      "system",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// IsDomainAllowed 检查域名是否允许访问
func (s *SettingService) IsDomainAllowed(domain string) bool {
	settings, err := s.GetDomainSettings()
	if err != nil {
		// 如果获取失败，则允许所有域名（避免配置错误导致无法访问）
		return true
	}

	// 如果未启用白名单，允许所有域名
	if !settings.Enabled {
		// 白名单功能未启用，允许所有域名访问
		return true
	}

	// 白名单已启用，检查域名是否在白名单中
	for _, allowed := range settings.Domains {
		if domain == allowed {
			return true
		}
		// 支持通配符子域名，如 *.example.com
		if len(allowed) > 2 && allowed[0] == '*' && allowed[1] == '.' {
			suffix := allowed[1:] // 去掉 *
			if len(domain) >= len(suffix) && domain[len(domain)-len(suffix):] == suffix {
				return true
			}
		}
	}

	// 域名不在白名单中
	return false
}


// ============= Emby/媒体服务设置 =============

// EmbySettings Emby/媒体服务设置结构
type EmbySettings struct {
	Enabled      bool   `json:"enabled"`       // 是否启用
	Mode         string `json:"mode"`          // 模式: emby 或 feiniu
	BaseURL      string `json:"base_url"`      // 服务器地址
	APIKey       string `json:"api_key"`       // Emby API密钥
	AdminUser    string `json:"admin_user"`    // 管理员用户名（飞牛模式）
	AdminPass    string `json:"admin_pass"`    // 管理员密码（飞牛模式）
	TemplateUser string `json:"template_user"` // 模板用户名（新用户复制其权限）
}

// GetEmbySettings 获取Emby设置
func (s *SettingService) GetEmbySettings() (*EmbySettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "emby").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &EmbySettings{
				Enabled: false,
				Mode:    "emby",
				BaseURL: "http://localhost:8096",
			}, nil
		}
		return nil, err
	}

	var settings EmbySettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveEmbySettings 保存Emby设置
func (s *SettingService) SaveEmbySettings(settings *EmbySettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "emby").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "emby",
				Value:     string(value),
				Type:      "emby",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// IsEmbyMode 是否为Emby官方模式
func (s *EmbySettings) IsEmbyMode() bool {
	return s.Mode == "emby" || s.Mode == ""
}

// IsFeiniuMode 是否为飞牛影视模式
func (s *EmbySettings) IsFeiniuMode() bool {
	return s.Mode == "feiniu"
}

// ============= 客户端白名单设置 =============

// ClientWhitelistItem 客户端白名单项
type ClientWhitelistItem struct {
	Name        string `json:"name"`        // 客户端名称（用于匹配）
	DisplayName string `json:"display_name"` // 显示名称
	Enabled     bool   `json:"enabled"`     // 是否启用
}

// ClientWhitelistSettings 客户端白名单设置结构
type ClientWhitelistSettings struct {
	Enabled bool                  `json:"enabled"` // 是否启用客户端白名单
	Clients []ClientWhitelistItem `json:"clients"` // 允许的客户端列表
}

// GetClientWhitelistSettings 获取客户端白名单设置
func (s *SettingService) GetClientWhitelistSettings() (*ClientWhitelistSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "client_whitelist").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值（只保留常用客户端）
			return &ClientWhitelistSettings{
				Enabled: false,
				Clients: []ClientWhitelistItem{
					{Name: "VidHub", DisplayName: "VidHub", Enabled: true},
					{Name: "SenPlayer", DisplayName: "SenPlayer", Enabled: true},
					{Name: "Infuse", DisplayName: "Infuse", Enabled: true},
					{Name: "Infuse-Direct", DisplayName: "Infuse (直连)", Enabled: true},
				},
			}, nil
		}
		return nil, err
	}

	var settings ClientWhitelistSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveClientWhitelistSettings 保存客户端白名单设置
func (s *SettingService) SaveClientWhitelistSettings(settings *ClientWhitelistSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "client_whitelist").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "client_whitelist",
				Value:     string(value),
				Type:      "emby",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// IsClientAllowed 检查客户端是否允许
func (s *SettingService) IsClientAllowed(clientName string) bool {
	settings, err := s.GetClientWhitelistSettings()
	if err != nil {
		// 如果获取失败，则允许所有客户端
		return true
	}

	// 如果未启用白名单，允许所有客户端
	if !settings.Enabled {
		return true
	}

	// 检查客户端是否在白名单中且已启用
	for _, client := range settings.Clients {
		if client.Enabled && clientName == client.Name {
			return true
		}
		// 支持部分匹配（客户端名称可能包含版本号等）
		if client.Enabled && len(clientName) >= len(client.Name) {
			if clientName[:len(client.Name)] == client.Name {
				return true
			}
		}
	}

	return false
}

// AddClientToWhitelist 添加客户端到白名单
func (s *SettingService) AddClientToWhitelist(name, displayName string, updatedBy uuid.UUID) error {
	settings, err := s.GetClientWhitelistSettings()
	if err != nil {
		return err
	}

	// 检查是否已存在
	for _, client := range settings.Clients {
		if client.Name == name {
			return errors.New("客户端已存在")
		}
	}

	// 添加新客户端
	settings.Clients = append(settings.Clients, ClientWhitelistItem{
		Name:        name,
		DisplayName: displayName,
		Enabled:     true,
	})

	return s.SaveClientWhitelistSettings(settings, updatedBy)
}

// RemoveClientFromWhitelist 从白名单移除客户端
func (s *SettingService) RemoveClientFromWhitelist(name string, updatedBy uuid.UUID) error {
	settings, err := s.GetClientWhitelistSettings()
	if err != nil {
		return err
	}

	// 查找并移除
	newClients := make([]ClientWhitelistItem, 0)
	found := false
	for _, client := range settings.Clients {
		if client.Name != name {
			newClients = append(newClients, client)
		} else {
			found = true
		}
	}

	if !found {
		return errors.New("客户端不存在")
	}

	settings.Clients = newClients
	return s.SaveClientWhitelistSettings(settings, updatedBy)
}

// UpdateClientStatus 更新客户端状态
func (s *SettingService) UpdateClientStatus(name string, enabled bool, updatedBy uuid.UUID) error {
	settings, err := s.GetClientWhitelistSettings()
	if err != nil {
		return err
	}

	// 查找并更新
	found := false
	for i, client := range settings.Clients {
		if client.Name == name {
			settings.Clients[i].Enabled = enabled
			found = true
			break
		}
	}

	if !found {
		return errors.New("客户端不存在")
	}

	return s.SaveClientWhitelistSettings(settings, updatedBy)
}


// ============= 会话限制设置 =============

// SessionLimitSettings 会话限制设置结构
type SessionLimitSettings struct {
	Enabled        bool `json:"enabled"`          // 是否启用会话限制
	MaxSessions    int  `json:"max_sessions"`     // 每用户最大会话数
	AutoKillOldest bool `json:"auto_kill_oldest"` // 超出限制时自动终止最早的会话
	// 播放限制
	PlayLimitEnabled   bool `json:"play_limit_enabled"`    // 是否启用播放数量限制
	MaxPlayingSessions int  `json:"max_playing_sessions"`  // 每用户最大同时播放数
	AutoStopOldestPlay bool `json:"auto_stop_oldest_play"` // 超出播放限制时自动停止最早的播放
}

// GetSessionLimitSettings 获取会话限制设置
func (s *SettingService) GetSessionLimitSettings() (*SessionLimitSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "session_limit").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &SessionLimitSettings{
				Enabled:            false,
				MaxSessions:        3,
				AutoKillOldest:     true,
				PlayLimitEnabled:   false,
				MaxPlayingSessions: 1,
				AutoStopOldestPlay: true,
			}, nil
		}
		return nil, err
	}

	var settings SessionLimitSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	// 设置默认值（兼容旧数据）
	if settings.MaxPlayingSessions == 0 {
		settings.MaxPlayingSessions = 1
	}

	return &settings, nil
}

// SaveSessionLimitSettings 保存会话限制设置
func (s *SettingService) SaveSessionLimitSettings(settings *SessionLimitSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "session_limit").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "session_limit",
				Value:     string(value),
				Type:      "emby",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// ============= 全局设备策略设置 =============

// GlobalDevicePolicySettings 全局设备策略设置
type GlobalDevicePolicySettings struct {
	EnableAllDevices bool     `json:"enable_all_devices"` // 是否允许所有设备
	EnabledClients   []string `json:"enabled_clients"`    // 允许的客户端名称列表
}

// GetGlobalDevicePolicySettings 获取全局设备策略设置
func (s *SettingService) GetGlobalDevicePolicySettings() (*GlobalDevicePolicySettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "global_device_policy").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值：允许所有设备
			return &GlobalDevicePolicySettings{
				EnableAllDevices: true,
				EnabledClients:   []string{},
			}, nil
		}
		return nil, err
	}

	var settings GlobalDevicePolicySettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveGlobalDevicePolicySettings 保存全局设备策略设置
func (s *SettingService) SaveGlobalDevicePolicySettings(settings *GlobalDevicePolicySettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "global_device_policy").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "global_device_policy",
				Value:     string(value),
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// ============= 网站设置 =============

// SiteSettings 网站设置结构
type SiteSettings struct {
	Title       string `json:"title"`       // 网站标题
	Description string `json:"description"` // 网站描述
	Keywords    string `json:"keywords"`    // SEO关键词
	Logo        string `json:"logo"`        // Logo URL
	Favicon     string `json:"favicon"`     // Favicon URL
	Footer      string `json:"footer"`      // 页脚文字
	ICP         string `json:"icp"`         // ICP备案号
	GithubURL   string `json:"github_url"`  // GitHub链接
	TelegramURL string `json:"telegram_url"` // Telegram链接
	QQURL       string `json:"qq_url"`      // QQ群链接
}

// GetSiteSettings 获取网站设置
func (s *SettingService) GetSiteSettings() (*SiteSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "site").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &SiteSettings{
				Title:       "EmbyHub - 用户管理系统",
				Description: "EmbyHub用户管理系统",
				Keywords:    "EmbyHub,Emby,媒体服务",
				Logo:        "/uploads/logo/lightsail-logo.svg",
				Favicon:     "/uploads/logo/lightsail-logo.svg",
				Footer:      "© 2025 EmbyHub",
			}, nil
		}
		return nil, err
	}

	var settings SiteSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveSiteSettings 保存网站设置
func (s *SettingService) SaveSiteSettings(settings *SiteSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "site").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "site",
				Value:     string(value),
				Type:      "site",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// ============= 播放限制设置 =============

// PlayLimitSettings 播放限制设置结构
type PlayLimitSettings struct {
	Enabled      bool `json:"enabled"`       // 是否启用播放限制
	MaxPlaying   int  `json:"max_playing"`   // 每用户最大同时播放数
	SpeedEnabled bool `json:"speed_enabled"` // 是否启用速率限制
	// 按角色设置速率限制 (Mbps)
	SpeedUser       int `json:"speed_user"`        // 普通用户速率限制
	SpeedMember     int `json:"speed_member"`      // 会员用户速率限制
	SpeedAdmin      int `json:"speed_admin"`       // 管理员速率限制 (0表示不限制)
}

// GetPlayLimitSettings 获取播放限制设置
func (s *SettingService) GetPlayLimitSettings() (*PlayLimitSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "play_limit").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &PlayLimitSettings{
				Enabled:      false,
				MaxPlaying:   1,
				SpeedEnabled: false,
				SpeedUser:    10,  // 普通用户默认 10 Mbps
				SpeedMember:  30,  // 会员用户默认 30 Mbps
				SpeedAdmin:   0,   // 管理员不限制
			}, nil
		}
		return nil, err
	}

	var settings PlayLimitSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	// 设置默认值
	if settings.MaxPlaying == 0 {
		settings.MaxPlaying = 1
	}
	if settings.SpeedUser == 0 {
		settings.SpeedUser = 10 // 普通用户默认 10 Mbps
	}
	if settings.SpeedMember == 0 {
		settings.SpeedMember = 30 // 会员用户默认 30 Mbps
	}
	// SpeedAdmin 为 0 表示不限制，所以不设置默认值

	return &settings, nil
}

// SavePlayLimitSettings 保存播放限制设置
func (s *SettingService) SavePlayLimitSettings(settings *PlayLimitSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "play_limit").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "play_limit",
				Value:     string(value),
				Type:      "emby",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// UserCleanupSettings 用户清理设置结构
type UserCleanupSettings struct {
	Enabled           bool `json:"enabled"`             // 是否启用自动清理
	InactiveDays      int  `json:"inactive_days"`       // 未登录天数阈值（超过此天数且会员过期的用户将被清理）
	ExpiredDays       int  `json:"expired_days"`        // 会员过期天数阈值（会员过期超过此天数的用户将被清理）
	DeleteEmbyAccount bool `json:"delete_emby_account"` // 是否同时删除Emby账号
}

// GetUserCleanupSettings 获取用户清理设置
func (s *SettingService) GetUserCleanupSettings() (*UserCleanupSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "user_cleanup").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &UserCleanupSettings{
				Enabled:           false,
				InactiveDays:      90,  // 默认90天未登录
				ExpiredDays:       30,  // 默认会员过期30天
				DeleteEmbyAccount: true,
			}, nil
		}
		return nil, err
	}

	var settings UserCleanupSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	// 设置默认值
	if settings.InactiveDays == 0 {
		settings.InactiveDays = 90
	}
	if settings.ExpiredDays == 0 {
		settings.ExpiredDays = 30
	}

	return &settings, nil
}

// SaveUserCleanupSettings 保存用户清理设置
func (s *SettingService) SaveUserCleanupSettings(settings *UserCleanupSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "user_cleanup").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "user_cleanup",
				Value:     string(value),
				Type:      "system",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// ============= 充值链接设置 =============

// RechargeLink 单个充值链接
type RechargeLink struct {
	CardType int    `json:"card_type"` // 卡密类型: 1=月卡, 2=季卡, 3=半年卡, 4=年卡
	Name     string `json:"name"`      // 显示名称
	URL      string `json:"url"`       // 跳转链接
	Enabled  bool   `json:"enabled"`   // 是否启用
}

// RechargeLinksSettings 充值链接设置
type RechargeLinksSettings struct {
	Links []RechargeLink `json:"links"` // 充值链接列表
}

// GetRechargeLinksSettings 获取充值链接设置
func (s *SettingService) GetRechargeLinksSettings() (*RechargeLinksSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "recharge_links").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &RechargeLinksSettings{
				Links: []RechargeLink{
					{CardType: 1, Name: "月卡", URL: "", Enabled: false},
					{CardType: 2, Name: "季卡", URL: "", Enabled: false},
					{CardType: 3, Name: "半年卡", URL: "", Enabled: false},
					{CardType: 4, Name: "年卡", URL: "", Enabled: false},
				},
			}, nil
		}
		return nil, err
	}

	var settings RechargeLinksSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveRechargeLinksSettings 保存充值链接设置
func (s *SettingService) SaveRechargeLinksSettings(settings *RechargeLinksSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "recharge_links").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "recharge_links",
				Value:     string(value),
				Type:      "system",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}


// ============= 积分卡购买链接设置 =============

// PointsRechargeLink 积分卡购买链接
type PointsRechargeLink struct {
	Points  int    `json:"points"`  // 积分数量
	Name    string `json:"name"`    // 显示名称
	URL     string `json:"url"`     // 跳转链接
	Enabled bool   `json:"enabled"` // 是否启用
}

// PointsRechargeLinksSettings 积分卡购买链接设置
type PointsRechargeLinksSettings struct {
	Links []PointsRechargeLink `json:"links"` // 购买链接列表
}

// GetPointsRechargeLinksSettings 获取积分卡购买链接设置
func (s *SettingService) GetPointsRechargeLinksSettings() (*PointsRechargeLinksSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "points_recharge_links").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &PointsRechargeLinksSettings{
				Links: []PointsRechargeLink{
					{Points: 100, Name: "100积分", URL: "", Enabled: false},
					{Points: 500, Name: "500积分", URL: "", Enabled: false},
					{Points: 1000, Name: "1000积分", URL: "", Enabled: false},
					{Points: 5000, Name: "5000积分", URL: "", Enabled: false},
				},
			}, nil
		}
		return nil, err
	}

	var settings PointsRechargeLinksSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SavePointsRechargeLinksSettings 保存积分卡购买链接设置
func (s *SettingService) SavePointsRechargeLinksSettings(settings *PointsRechargeLinksSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "points_recharge_links").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "points_recharge_links",
				Value:     string(value),
				Type:      "system",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// ============= 图床设置 =============

// ImageHostSettings 图床设置结构
type ImageHostSettings struct {
	Enabled bool   `json:"enabled"`  // 是否启用外部图床
	BaseURL string `json:"base_url"` // 图床地址，如 https://img.liubei.org
}

// GetImageHostSettings 获取图床设置
func (s *SettingService) GetImageHostSettings() (*ImageHostSettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "image_host").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &ImageHostSettings{
				Enabled: true,
				BaseURL: "https://img.liubei.org",
			}, nil
		}
		return nil, err
	}

	var settings ImageHostSettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveImageHostSettings 保存图床设置
func (s *SettingService) SaveImageHostSettings(settings *ImageHostSettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "image_host").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			setting = models.Setting{
				Key:       "image_host",
				Value:     string(value),
				Type:      "forum",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	// 更新现有记录
	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}
