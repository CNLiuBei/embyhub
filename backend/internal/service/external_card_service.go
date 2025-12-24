// Package service 外部卡密API服务
package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExternalCardService 外部卡密API服务
type ExternalCardService struct {
	db *gorm.DB
}

// NewExternalCardService 创建外部卡密API服务
func NewExternalCardService(db *gorm.DB) *ExternalCardService {
	return &ExternalCardService{db: db}
}

// ExternalCardAPISettings 外部卡密API设置
type ExternalCardAPISettings struct {
	Enabled       bool         `json:"enabled"`        // 是否启用外部API
	APIKey        string       `json:"api_key"`        // 主API密钥
	APIKeys       []APIKeyItem `json:"api_keys"`       // 多API密钥列表
	AllowedIPs    string       `json:"allowed_ips"`    // 允许的IP列表（逗号分隔，空表示不限制）
	RateLimit     int          `json:"rate_limit"`     // 每分钟请求限制（0表示不限制）
	DefaultType   int8         `json:"default_type"`   // 默认卡密类型（1月卡 2季卡 3半年卡 4年卡）
	LogEnabled    bool         `json:"log_enabled"`    // 是否记录请求日志
}

// APIKeyItem 单个API密钥项
type APIKeyItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Key        string `json:"key"`
	Enabled    bool   `json:"enabled"`
	CreatedAt  string `json:"created_at"`
	LastUsedAt string `json:"last_used_at,omitempty"`
}

// GetExternalCardAPISettings 获取外部卡密API设置
func (s *ExternalCardService) GetExternalCardAPISettings() (*ExternalCardAPISettings, error) {
	var setting models.Setting
	err := s.db.Where("\"key\" = ?", "external_card_api").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认值
			return &ExternalCardAPISettings{
				Enabled:     false,
				APIKey:      "",
				AllowedIPs:  "",
				RateLimit:   60,
				DefaultType: 1,
				LogEnabled:  true,
			}, nil
		}
		return nil, err
	}

	var settings ExternalCardAPISettings
	if err := json.Unmarshal([]byte(setting.Value), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveExternalCardAPISettings 保存外部卡密API设置
func (s *ExternalCardService) SaveExternalCardAPISettings(settings *ExternalCardAPISettings, updatedBy uuid.UUID) error {
	value, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	var setting models.Setting
	err = s.db.Where("\"key\" = ?", "external_card_api").First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			setting = models.Setting{
				Key:       "external_card_api",
				Value:     string(value),
				Type:      "api",
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	setting.Value = string(value)
	setting.UpdatedBy = updatedBy
	return s.db.Save(&setting).Error
}

// GenerateAPIKey 生成新的API密钥
func (s *ExternalCardService) GenerateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ValidateAPIKey 验证API密钥
func (s *ExternalCardService) ValidateAPIKey(apiKey string) (bool, error) {
	settings, err := s.GetExternalCardAPISettings()
	if err != nil {
		return false, err
	}

	if !settings.Enabled {
		return false, errors.New("外部API未启用")
	}

	// 检查主密钥
	if settings.APIKey != "" && settings.APIKey == apiKey {
		return true, nil
	}

	// 检查多密钥列表
	for _, keyItem := range settings.APIKeys {
		if keyItem.Enabled && keyItem.Key == apiKey {
			// 更新最后使用时间（异步，不阻塞）
			go s.updateAPIKeyLastUsed(keyItem.ID)
			return true, nil
		}
	}

	if settings.APIKey == "" && len(settings.APIKeys) == 0 {
		return false, errors.New("API密钥未配置")
	}

	return false, nil
}

// updateAPIKeyLastUsed 更新API密钥最后使用时间
func (s *ExternalCardService) updateAPIKeyLastUsed(keyID string) {
	settings, err := s.GetExternalCardAPISettings()
	if err != nil {
		return
	}

	for i, keyItem := range settings.APIKeys {
		if keyItem.ID == keyID {
			settings.APIKeys[i].LastUsedAt = time.Now().Format(time.RFC3339)
			break
		}
	}

	value, err := json.Marshal(settings)
	if err != nil {
		return
	}

	s.db.Model(&models.Setting{}).Where("\"key\" = ?", "external_card_api").Update("value", string(value))
}

// FetchCardRequest 获取卡密请求
type FetchCardRequest struct {
	Type  int8 `json:"type" form:"type"`   // 卡密类型（1月卡 2季卡 3半年卡 4年卡），不传则使用默认类型
	Count int  `json:"count" form:"count"` // 获取数量，默认1，最大10
}

// FetchCardResponse 获取卡密响应
type FetchCardResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    *CardData `json:"data,omitempty"`
}

// CardData 卡密数据
type CardData struct {
	Code     string `json:"code"`      // 卡密码
	Type     int8   `json:"type"`      // 卡密类型
	TypeName string `json:"type_name"` // 类型名称
	Duration int    `json:"duration"`  // 有效天数
}

// FetchCardsResponse 批量获取卡密响应
type FetchCardsResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    []CardData  `json:"data,omitempty"`
	Count   int         `json:"count"`
}

// FetchCard 获取一张卡密（标记为已发放但未使用）
func (s *ExternalCardService) FetchCard(cardType int8) (*CardData, error) {
	settings, err := s.GetExternalCardAPISettings()
	if err != nil {
		return nil, err
	}

	// 如果未指定类型，使用默认类型
	if cardType == 0 {
		cardType = settings.DefaultType
	}

	// 查找一张未使用的卡密
	var card models.Card
	err = s.db.Where("status = ? AND card_type = ?", models.CardStatusUnused, cardType).
		Order("id ASC").
		First(&card).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("没有可用的卡密")
		}
		return nil, err
	}

	// 检查卡密是否过期
	if card.ExpireAt != nil && card.ExpireAt.Before(time.Now()) {
		s.db.Model(&card).Update("status", models.CardStatusExpired)
		return nil, errors.New("没有可用的卡密")
	}

	// 更新卡密状态为"已发放"（新增状态4）
	// 注意：这里我们不改变状态，只是返回卡密，让用户自己兑换
	// 如果需要标记为已发放，可以添加一个新字段

	return &CardData{
		Code:     card.Code,
		Type:     card.CardType,
		TypeName: card.GetCardTypeName(),
		Duration: card.Duration,
	}, nil
}

// FetchCards 批量获取卡密
func (s *ExternalCardService) FetchCards(cardType int8, count int) ([]CardData, error) {
	settings, err := s.GetExternalCardAPISettings()
	if err != nil {
		return nil, err
	}

	// 限制数量
	if count <= 0 {
		count = 1
	}
	if count > 10 {
		count = 10
	}

	// 如果未指定类型，使用默认类型
	if cardType == 0 {
		cardType = settings.DefaultType
	}

	// 查找多张未使用的卡密
	var cards []models.Card
	err = s.db.Where("status = ? AND card_type = ?", models.CardStatusUnused, cardType).
		Order("id ASC").
		Limit(count).
		Find(&cards).Error
	if err != nil {
		return nil, err
	}

	if len(cards) == 0 {
		return nil, errors.New("没有可用的卡密")
	}

	result := make([]CardData, len(cards))
	for i, card := range cards {
		result[i] = CardData{
			Code:     card.Code,
			Type:     card.CardType,
			TypeName: card.GetCardTypeName(),
			Duration: card.Duration,
		}
	}

	return result, nil
}

// GetCardStock 获取卡密库存
func (s *ExternalCardService) GetCardStock() (map[string]int64, error) {
	stock := make(map[string]int64)

	// 统计各类型未使用的卡密数量
	var results []struct {
		CardType int8
		Count    int64
	}

	err := s.db.Model(&models.Card{}).
		Select("card_type, COUNT(*) as count").
		Where("status = ?", models.CardStatusUnused).
		Group("card_type").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	typeNames := map[int8]string{
		1: "月卡",
		2: "季卡",
		3: "半年卡",
		4: "年卡",
	}

	for _, r := range results {
		name := typeNames[r.CardType]
		if name == "" {
			name = "未知"
		}
		stock[name] = r.Count
	}

	return stock, nil
}

// LogAPIRequest 记录API请求
func (s *ExternalCardService) LogAPIRequest(ip, method, path, params, response string, status int, duration int64) error {
	settings, err := s.GetExternalCardAPISettings()
	if err != nil || !settings.LogEnabled {
		return nil
	}

	log := &models.ExternalAPILog{
		IP:       ip,
		Method:   method,
		Path:     path,
		Params:   params,
		Response: response,
		Status:   status,
		Duration: duration,
	}

	return s.db.Create(log).Error
}

// GetAPILogs 获取API日志
func (s *ExternalCardService) GetAPILogs(page, pageSize int) ([]models.ExternalAPILog, int64, error) {
	var logs []models.ExternalAPILog
	var total int64

	s.db.Model(&models.ExternalAPILog{}).Count(&total)

	offset := (page - 1) * pageSize
	err := s.db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
