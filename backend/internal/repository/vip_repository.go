// Package repository VIP数据访问层
package repository

import (
	"errors"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// VipRepository VIP数据仓库
type VipRepository struct {
	db *gorm.DB
}

// NewVipRepository 创建VIP仓库实例
func NewVipRepository(db *gorm.DB) *VipRepository {
	return &VipRepository{db: db}
}

// GetPlanByID 根据ID获取套餐（带缓存友好设计）
func (r *VipRepository) GetPlanByID(tx *gorm.DB, planID uint) (*models.VipPlan, error) {
	var plan models.VipPlan
	err := tx.Where("id = ? AND is_active = ?", planID, true).First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("套餐不存在或已下架")
		}
		return nil, err
	}
	return &plan, nil
}

// GetUserVipWithLock 获取用户会员信息（带行锁，防止并发）
func (r *VipRepository) GetUserVipWithLock(tx *gorm.DB, userID uuid.UUID) (*models.UserVip, error) {
	var userVip models.UserVip
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&userVip).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户未开通过VIP，返回新记录
			return &models.UserVip{
				UserID:      userID,
				VipExpireAt: nil,
			}, nil
		}
		return nil, err
	}
	return &userVip, nil
}

// CreateOrUpdateUserVip 创建或更新用户VIP
func (r *VipRepository) CreateOrUpdateUserVip(tx *gorm.DB, userVip *models.UserVip) error {
	return tx.Save(userVip).Error
}

// CreateOrder 创建订单
func (r *VipRepository) CreateOrder(tx *gorm.DB, order *models.VipOrder) error {
	return tx.Create(order).Error
}

// CreateBalanceLog 创建余额流水
func (r *VipRepository) CreateBalanceLog(tx *gorm.DB, log *models.BalanceLog) error {
	return tx.Create(log).Error
}

// GetUserBalanceWithLock 获取用户余额（带行锁）
func (r *VipRepository) GetUserBalanceWithLock(tx *gorm.DB, userID uuid.UUID) (int64, error) {
	var user models.User
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("balance").
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("用户不存在")
		}
		return 0, err
	}

	// 将float64余额转换为分（整数）
	// 假设数据库存储的是元，需要 * 100 转为分
	balanceInCents := int64(user.Balance * 100)
	return balanceInCents, nil
}

// UpdateUserBalance 更新用户余额
func (r *VipRepository) UpdateUserBalance(tx *gorm.DB, userID uuid.UUID, newBalanceInCents int64) error {
	// 将分转换为元
	newBalanceInYuan := float64(newBalanceInCents) / 100.0

	result := tx.Model(&models.User{}).
		Where("id = ?", userID).
		Update("balance", newBalanceInYuan)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("更新余额失败，用户不存在")
	}
	return nil
}

// CalculateNewExpireTime 计算新的会员到期时间
func CalculateNewExpireTime(currentExpireAt *time.Time, durationDays int) time.Time {
	now := time.Now().UTC()

	// 如果没有会员或已过期，从现在开始计算
	if currentExpireAt == nil || currentExpireAt.Before(now) {
		return now.AddDate(0, 0, durationDays)
	}

	// 如果还在有效期内，从到期时间顺延
	return currentExpireAt.AddDate(0, 0, durationDays)
}
