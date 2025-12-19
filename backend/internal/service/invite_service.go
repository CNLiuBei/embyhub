// Package service 用户邀请服务
package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InviteService 邀请服务
type InviteService struct {
	db *gorm.DB
}

// NewInviteService 创建邀请服务
func NewInviteService(db *gorm.DB) *InviteService {
	return &InviteService{db: db}
}

// GenerateUserInviteCode 生成用户专属邀请码
func GenerateUserInviteCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

// EnsureUserInviteCode 确保用户有邀请码
func (s *InviteService) EnsureUserInviteCode(userID uuid.UUID) (string, error) {
	var user models.User
	if err := s.db.Select("id", "invite_code").First(&user, userID).Error; err != nil {
		return "", errors.New("用户不存在")
	}

	// 如果已有邀请码，直接返回
	if user.InviteCode != "" {
		return user.InviteCode, nil
	}

	// 生成新邀请码
	for i := 0; i < 10; i++ { // 最多尝试10次避免冲突
		code := GenerateUserInviteCode()
		var count int64
		s.db.Model(&models.User{}).Where("invite_code = ?", code).Count(&count)
		if count == 0 {
			s.db.Model(&user).Update("invite_code", code)
			return code, nil
		}
	}

	return "", errors.New("生成邀请码失败")
}

// GetUserInviteInfo 获取用户邀请信息
type UserInviteInfo struct {
	InviteCode  string `json:"invite_code"`
	InviteCount int    `json:"invite_count"`
	RewardDays  int    `json:"reward_days"` // 每次邀请奖励天数（从设置读取）
}

func (s *InviteService) GetUserInviteInfo(userID uuid.UUID) (*UserInviteInfo, error) {
	var user models.User
	if err := s.db.Select("invite_code", "invite_count").First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 如果没有邀请码，生成一个
	if user.InviteCode == "" {
		code, err := s.EnsureUserInviteCode(userID)
		if err != nil {
			return nil, err
		}
		user.InviteCode = code
	}

	// 获取邀请奖励设置
	rewardDays := s.getInviteRewardDays()

	return &UserInviteInfo{
		InviteCode:  user.InviteCode,
		InviteCount: user.InviteCount,
		RewardDays:  rewardDays,
	}, nil
}

// getInviteRewardDays 获取邀请奖励天数配置
func (s *InviteService) getInviteRewardDays() int {
	var setting models.Setting
	if err := s.db.Where("key = ?", "invite_reward_days").First(&setting).Error; err != nil {
		return 7 // 默认7天
	}
	// 解析数字
	days := 0
	for _, c := range setting.Value {
		if c >= '0' && c <= '9' {
			days = days*10 + int(c-'0')
		}
	}
	if days <= 0 || days > 365 {
		days = 7
	}
	return days
}

// ProcessInviteOnRegister 注册时处理邀请关系
func (s *InviteService) ProcessInviteOnRegister(newUserID uuid.UUID, inviteCode string) error {
	if inviteCode == "" {
		return nil
	}

	// 查找邀请人
	var inviter models.User
	if err := s.db.Where("invite_code = ?", inviteCode).First(&inviter).Error; err != nil {
		return nil // 邀请码无效，静默忽略
	}

	// 不能邀请自己
	if inviter.ID == newUserID {
		return nil
	}

	// 更新新用户的邀请人
	s.db.Model(&models.User{}).Where("id = ?", newUserID).Update("invited_by", inviter.ID)

	// 增加邀请人的邀请计数
	s.db.Model(&inviter).Update("invite_count", gorm.Expr("invite_count + 1"))

	// 获取奖励天数
	rewardDays := s.getInviteRewardDays()
	if rewardDays <= 0 {
		return nil
	}

	// 给邀请人增加会员天数
	now := time.Now()
	var newExpire time.Time
	if inviter.MemberExpire != nil && inviter.MemberExpire.After(now) {
		newExpire = inviter.MemberExpire.AddDate(0, 0, rewardDays)
	} else {
		newExpire = now.AddDate(0, 0, rewardDays)
	}

	s.db.Model(&inviter).Updates(map[string]interface{}{
		"member_level":  models.MemberMonth,
		"member_expire": newExpire,
	})

	// 创建邀请记录
	record := &models.InviteRecord{
		InviterID:  inviter.ID,
		InviteeID:  newUserID,
		InviteCode: inviteCode,
		RewardDays: rewardDays,
	}
	s.db.Create(record)

	return nil
}

// GetMyInvites 获取我的邀请记录
func (s *InviteService) GetMyInvites(userID uuid.UUID) ([]map[string]interface{}, int64, error) {
	var records []models.InviteRecord
	var total int64

	s.db.Model(&models.InviteRecord{}).Where("inviter_id = ?", userID).Count(&total)
	s.db.Where("inviter_id = ?", userID).Order("created_at DESC").Limit(50).Find(&records)

	var result []map[string]interface{}
	for _, r := range records {
		var invitee models.User
		s.db.Select("username", "nickname", "created_at").First(&invitee, r.InviteeID)

		result = append(result, map[string]interface{}{
			"id":          r.ID,
			"invitee":     invitee.Nickname,
			"reward_days": r.RewardDays,
			"created_at":  r.CreatedAt,
		})
	}

	return result, total, nil
}

// GetInviteRanking 获取邀请排行榜
func (s *InviteService) GetInviteRanking(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}

	var users []models.User
	s.db.Select("id", "username", "nickname", "avatar", "invite_count").
		Where("invite_count > 0").
		Order("invite_count DESC").
		Limit(limit).
		Find(&users)

	var result []map[string]interface{}
	for i, u := range users {
		result = append(result, map[string]interface{}{
			"rank":         i + 1,
			"nickname":     u.Nickname,
			"avatar":       u.Avatar,
			"invite_count": u.InviteCount,
		})
	}

	return result, nil
}

// Admin: GetInviteStats 获取邀请统计
func (s *InviteService) GetInviteStats() map[string]interface{} {
	var totalUsers, invitedUsers, totalInvites int64

	s.db.Model(&models.User{}).Count(&totalUsers)
	s.db.Model(&models.User{}).Where("invited_by IS NOT NULL").Count(&invitedUsers)
	s.db.Model(&models.InviteRecord{}).Count(&totalInvites)

	return map[string]interface{}{
		"total_users":   totalUsers,
		"invited_users": invitedUsers,
		"total_invites": totalInvites,
		"reward_days":   s.getInviteRewardDays(),
	}
}

// Admin: SetInviteRewardDays 设置邀请奖励天数
func (s *InviteService) SetInviteRewardDays(days int) error {
	var setting models.Setting
	err := s.db.Where("key = ?", "invite_reward_days").First(&setting).Error

	valueStr := strconv.Itoa(days)

	if err == gorm.ErrRecordNotFound {
		setting = models.Setting{
			Key:   "invite_reward_days",
			Value: valueStr,
			Type:  "invite",
		}
		return s.db.Create(&setting).Error
	}

	return s.db.Model(&setting).Update("value", valueStr).Error
}

// Admin: GetInviteRecords 获取邀请记录列表
func (s *InviteService) GetInviteRecords(page, pageSize int) ([]map[string]interface{}, int64, error) {
	var records []models.InviteRecord
	var total int64

	s.db.Model(&models.InviteRecord{}).Count(&total)

	offset := (page - 1) * pageSize
	s.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records)

	var result []map[string]interface{}
	for _, r := range records {
		var inviter, invitee models.User
		s.db.Select("nickname").First(&inviter, r.InviterID)
		s.db.Select("nickname").First(&invitee, r.InviteeID)

		result = append(result, map[string]interface{}{
			"id":          r.ID,
			"inviter":     inviter.Nickname,
			"invitee":     invitee.Nickname,
			"reward_days": r.RewardDays,
			"created_at":  r.CreatedAt,
		})
	}

	return result, total, nil
}
