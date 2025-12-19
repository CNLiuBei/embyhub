// Package service 会员服务
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"feiniu-user-system/internal/config"
	"feiniu-user-system/internal/database"
	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MemberService 会员服务
type MemberService struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewMemberService 创建会员服务
func NewMemberService(db *gorm.DB, cfg *config.Config) *MemberService {
	return &MemberService{db: db, cfg: cfg}
}

// MemberInfo 会员信息
type MemberInfo struct {
	Level      int8       `json:"level"`
	LevelName  string     `json:"level_name"`
	ExpireAt   *time.Time `json:"expire_at"`
	WatchLimit int        `json:"watch_limit"`
	AdFree     bool       `json:"ad_free"`
	Quality4K  bool       `json:"quality_4k"`
	DaysLeft   int        `json:"days_left"`
}

// GetMemberInfo 获取会员信息
func (s *MemberService) GetMemberInfo(userID uuid.UUID) (*MemberInfo, error) {
	ctx := context.Background()
	cacheKey := database.KeyMemberInfo + userID.String()

	// 先从缓存获取
	cached, err := database.GetCache(ctx, cacheKey)
	if err == nil {
		var info MemberInfo
		if json.Unmarshal([]byte(cached), &info) == nil {
			return &info, nil
		}
	}

	var user models.User
	if err := s.db.Select("member_level", "member_expire").First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 获取等级配置
	levelCfg := s.getLevelConfig(user.MemberLevel)
	info := &MemberInfo{
		Level:      user.MemberLevel,
		LevelName:  levelCfg.Name,
		ExpireAt:   user.MemberExpire,
		WatchLimit: levelCfg.WatchLimit,
		AdFree:     levelCfg.AdFree,
		Quality4K:  levelCfg.Quality4K,
	}

	// 计算剩余天数
	if user.MemberExpire != nil && user.MemberExpire.After(time.Now()) {
		info.DaysLeft = int(time.Until(*user.MemberExpire).Hours() / 24)
	}

	// 缓存
	data, _ := json.Marshal(info)
	database.SetCache(ctx, cacheKey, string(data), 5*time.Minute)

	return info, nil
}

// GetMemberOrders 获取会员订单
func (s *MemberService) GetMemberOrders(userID uuid.UUID, page, pageSize int) ([]models.MemberOrder, int64, error) {
	var orders []models.MemberOrder
	var total int64

	query := s.db.Model(&models.MemberOrder{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// AdminSetMember 管理员设置会员(升级为会员用户并设置到期时间)
// 同时恢复账户状态，适用于禁用用户的会员续费
func (s *MemberService) AdminSetMember(userID uuid.UUID, days int) error {
	// 获取用户当前信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 计算到期时间
	now := time.Now()
	expire := now.AddDate(0, 0, days)

	// 如果用户已是有效会员，则累加时间
	if user.MemberExpire != nil && user.MemberExpire.After(now) {
		expire = user.MemberExpire.AddDate(0, 0, days)
	}

	// 确定会员等级
	memberLevel := models.MemberMonth
	if days >= 365 {
		memberLevel = models.MemberYear
	}

	// 升级为会员用户并设置到期时间，同时恢复账户状态
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"role":          models.RoleMember, // 升级为会员用户
		"member_level":  memberLevel,       // 设置会员等级
		"member_expire": expire,
		"status":        1, // 恢复账户状态为正常
	}).Error; err != nil {
		return err
	}

	// 发送站内通知
	s.sendMemberNotification(userID, memberLevel, expire)

	// 发送邮件通知
	isRenew := user.MemberLevel > 0
	if user.Email != "" {
		go func() {
			emailSvc := GetEmailServiceFromDB(s.db)
			if emailSvc == nil {
				return
			}
			if err := emailSvc.SendMemberActivated(user.Email, user.Nickname, days, expire.Format("2006-01-02"), isRenew); err != nil {
				log.Printf("发送会员开通邮件失败: %v", err)
			}
		}()
	}

	// 清除缓存
	ctx := context.Background()
	database.DeleteCache(ctx, database.KeyMemberInfo+userID.String())
	database.DeleteCache(ctx, database.KeyUserInfo+userID.String())

	return nil
}

// getLevelConfig 获取等级配置
func (s *MemberService) getLevelConfig(level int8) config.MemberLevel {
	for _, l := range s.cfg.Member.Levels {
		if int8(l.ID) == level {
			return l
		}
	}
	return s.cfg.Member.Levels[0]
}

// sendMemberNotification 发送会员通知
func (s *MemberService) sendMemberNotification(userID uuid.UUID, level int8, expire time.Time) {
	levelNames := []string{"普通用户", "月卡会员", "年卡会员"}
	notification := &models.Notification{
		UserID:  userID,
		Title:   "会员开通成功",
		Content: fmt.Sprintf("恭喜您成为%s，有效期至%s", levelNames[level], expire.Format("2006-01-02")),
		Type:    2,
	}
	s.db.Create(notification)
}

// BatchSetMember 批量设置会员
func (s *MemberService) BatchSetMember(userIDs []uuid.UUID, days int) (int, int, []string) {
	success := 0
	failed := 0
	var errors []string

	for _, userID := range userIDs {
		if err := s.AdminSetMember(userID, days); err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("用户 %s: %s", userID.String(), err.Error()))
		} else {
			success++
		}
	}

	return success, failed, errors
}

// CheckMemberExpire 检查会员是否过期(定时任务调用)
// 会员过期后将禁用账户，用户无法登录
func (s *MemberService) CheckMemberExpire() error {
	now := time.Now()

	// 查找已过期的会员（member_level > 0 且 member_expire 已过期）
	var expiredUsers []models.User
	s.db.Where("member_level > 0 AND member_expire < ?", now).Find(&expiredUsers)

	for _, user := range expiredUsers {
		// 会员过期处理：
		// 1. 降级 member_level 为 0
		// 2. 如果是会员角色，降级为普通用户
		// 3. 禁用账户（status = 2）
		updates := map[string]interface{}{
			"member_level": 0,
			"status":       2, // 禁用账户
		}
		// 如果用户角色是会员(role=1)，降级为普通用户(role=0)
		// 管理员(role>=2)不受影响
		if user.Role == models.RoleMember {
			updates["role"] = models.RoleUser
		}

		s.db.Model(&user).Updates(updates)

		// 发送站内到期通知
		notification := &models.Notification{
			UserID:  user.ID,
			Title:   "会员已到期 - 账户已禁用",
			Content: "您的会员已到期，账户已被禁用。请使用卡密续费后重新登录。",
			Type:    2,
		}
		s.db.Create(notification)

		// 发送邮件通知
		if user.Email != "" {
			go func(db *gorm.DB, email, nickname string) {
				emailSvc := GetEmailServiceFromDB(db)
				if emailSvc == nil {
					return
				}
				if err := emailSvc.SendMemberExpired(email, nickname); err != nil {
					log.Printf("发送会员过期邮件失败: %v", err)
				}
			}(s.db, user.Email, user.Nickname)
		}

		// 清除缓存
		ctx := context.Background()
		database.DeleteCache(ctx, database.KeyMemberInfo+user.ID.String())
		database.DeleteCache(ctx, database.KeyUserInfo+user.ID.String())
	}

	return nil
}
