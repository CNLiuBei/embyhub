package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"embyhub/internal/dao"
	"embyhub/internal/model"
)

type CardKeyService struct {
	cardKeyDAO *dao.CardKeyDAO
	userDAO    *dao.UserDAO
}

func NewCardKeyService() *CardKeyService {
	return &CardKeyService{
		cardKeyDAO: dao.NewCardKeyDAO(),
		userDAO:    dao.NewUserDAO(),
	}
}

// GenerateCardCode 生成卡密码（TL|24位字符）
func generateCardCode() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	code := strings.ToUpper(hex.EncodeToString(bytes))
	// 格式: TL|24位字符
	return "TL|" + code
}

// Create 批量创建卡密
func (s *CardKeyService) Create(req *model.CardKeyCreateRequest, creatorID int) ([]*model.CardKey, error) {
	cardKeys := make([]*model.CardKey, req.Count)

	for i := 0; i < req.Count; i++ {
		cardKeys[i] = &model.CardKey{
			CardCode:  generateCardCode(),
			CardType:  req.CardType,
			Duration:  req.Duration,
			Status:    1, // 未使用
			Remark:    req.Remark,
			CreatedBy: creatorID,
			CreatedAt: time.Now(),
		}
	}

	if err := s.cardKeyDAO.BatchCreate(cardKeys); err != nil {
		return nil, err
	}

	return cardKeys, nil
}

// GetByID 获取卡密详情
func (s *CardKeyService) GetByID(id int) (*model.CardKey, error) {
	return s.cardKeyDAO.GetByID(id)
}

// GetByCode 根据卡密码获取
func (s *CardKeyService) GetByCode(cardCode string) (*model.CardKey, error) {
	return s.cardKeyDAO.GetByCode(cardCode)
}

// List 获取卡密列表
func (s *CardKeyService) List(req *model.CardKeyListRequest) (*model.CardKeyListResponse, error) {
	cardKeys, total, err := s.cardKeyDAO.List(req)
	if err != nil {
		return nil, err
	}

	return &model.CardKeyListResponse{
		Total: int(total),
		List:  cardKeys,
	}, nil
}

// Use 使用卡密
func (s *CardKeyService) Use(cardCode string, userID int) (*model.CardKey, error) {
	cardKey, err := s.cardKeyDAO.GetByCode(cardCode)
	if err != nil {
		return nil, errors.New("卡密不存在")
	}

	// 检查状态
	if cardKey.Status == 0 {
		return nil, errors.New("卡密已被禁用")
	}
	if cardKey.Status == 2 {
		return nil, errors.New("卡密已被使用")
	}

	// 检查过期时间
	if cardKey.ExpireAt != nil && time.Now().After(*cardKey.ExpireAt) {
		return nil, errors.New("卡密已过期")
	}

	// 标记为已使用
	now := time.Now()
	cardKey.Status = 2
	cardKey.UsedBy = &userID
	cardKey.UsedAt = &now

	if err := s.cardKeyDAO.Update(cardKey); err != nil {
		return nil, err
	}

	return cardKey, nil
}

// Disable 禁用卡密
func (s *CardKeyService) Disable(id int) error {
	cardKey, err := s.cardKeyDAO.GetByID(id)
	if err != nil {
		return errors.New("卡密不存在")
	}

	if cardKey.Status == 2 {
		return errors.New("卡密已被使用，无法禁用")
	}

	cardKey.Status = 0
	return s.cardKeyDAO.Update(cardKey)
}

// Enable 启用卡密
func (s *CardKeyService) Enable(id int) error {
	cardKey, err := s.cardKeyDAO.GetByID(id)
	if err != nil {
		return errors.New("卡密不存在")
	}

	if cardKey.Status == 2 {
		return errors.New("卡密已被使用，无法启用")
	}

	cardKey.Status = 1
	return s.cardKeyDAO.Update(cardKey)
}

// Delete 删除卡密
func (s *CardKeyService) Delete(id int) error {
	cardKey, err := s.cardKeyDAO.GetByID(id)
	if err != nil {
		return errors.New("卡密不存在")
	}

	if cardKey.Status == 2 {
		return errors.New("卡密已被使用，无法删除")
	}

	return s.cardKeyDAO.Delete(id)
}

// BatchDelete 批量删除未使用的卡密
func (s *CardKeyService) BatchDelete(ids []int) (int, error) {
	deleted := 0
	for _, id := range ids {
		cardKey, err := s.cardKeyDAO.GetByID(id)
		if err != nil {
			continue
		}
		if cardKey.Status != 2 {
			if err := s.cardKeyDAO.Delete(id); err == nil {
				deleted++
			}
		}
	}
	return deleted, nil
}

// GetStatistics 获取卡密统计
func (s *CardKeyService) GetStatistics() (map[string]interface{}, error) {
	counts, err := s.cardKeyDAO.CountByStatus()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":    counts[0] + counts[1] + counts[2],
		"unused":   counts[1],
		"used":     counts[2],
		"disabled": counts[0],
	}, nil
}

// ValidateCardCode 验证卡密是否可用（用于注册验证）
func (s *CardKeyService) ValidateCardCode(cardCode string) (*model.CardKey, error) {
	cardKey, err := s.cardKeyDAO.GetByCode(cardCode)
	if err != nil {
		return nil, errors.New("卡密不存在")
	}

	if cardKey.Status == 0 {
		return nil, errors.New("卡密已被禁用")
	}
	if cardKey.Status == 2 {
		return nil, errors.New("卡密已被使用")
	}
	if cardKey.ExpireAt != nil && time.Now().After(*cardKey.ExpireAt) {
		return nil, errors.New("卡密已过期")
	}

	return cardKey, nil
}

// UseVipCard 使用VIP升级码
func (s *CardKeyService) UseVipCard(cardCode string, userID int) (*model.User, error) {
	// 验证卡密
	cardKey, err := s.ValidateCardCode(cardCode)
	if err != nil {
		return nil, err
	}

	// 检查是否为VIP升级码
	if cardKey.CardType != 2 {
		return nil, errors.New("该卡密不是VIP升级码")
	}

	// 获取用户
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 计算VIP到期时间
	now := time.Now()
	var vipExpireAt time.Time
	if user.VipExpireAt != nil && user.VipExpireAt.After(now) {
		// 如果当前VIP未过期，在原基础上增加
		vipExpireAt = user.VipExpireAt.AddDate(0, 0, cardKey.Duration)
	} else {
		// 从现在开始计算
		vipExpireAt = now.AddDate(0, 0, cardKey.Duration)
	}

	// 更新用户VIP状态
	user.VipLevel = 1
	user.VipExpireAt = &vipExpireAt
	user.UpdatedAt = now
	if err := s.userDAO.Update(user); err != nil {
		return nil, errors.New("升级VIP失败")
	}

	// 标记卡密为已使用
	cardKey.Status = 2
	cardKey.UsedBy = &userID
	cardKey.UsedAt = &now
	s.cardKeyDAO.Update(cardKey)

	return user, nil
}
