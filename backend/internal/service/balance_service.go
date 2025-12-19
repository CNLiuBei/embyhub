// Package service 余额服务
package service

import (
	"errors"
	"fmt"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BalanceService 余额服务
type BalanceService struct {
	db *gorm.DB
}

// NewBalanceService 创建余额服务
func NewBalanceService(db *gorm.DB) *BalanceService {
	return &BalanceService{db: db}
}

// GetBalance 获取用户余额
func (s *BalanceService) GetBalance(userID uuid.UUID) (float64, error) {
	var user models.User
	if err := s.db.Select("balance").First(&user, userID).Error; err != nil {
		return 0, errors.New("用户不存在")
	}
	return user.Balance, nil
}

// AddBalance 增加余额
func (s *BalanceService) AddBalance(userID uuid.UUID, amount float64, orderNo string, balanceType int8, remark string) error {
	if amount <= 0 {
		return errors.New("金额必须大于0")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询用户当前余额
	var user models.User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&user, userID).Error; err != nil {
		tx.Rollback()
		return errors.New("用户不存在")
	}

	balanceBefore := user.Balance
	balanceAfter := balanceBefore + amount

	// 更新用户余额
	if err := tx.Model(&user).Update("balance", balanceAfter).Error; err != nil {
		tx.Rollback()
		return errors.New("更新余额失败")
	}

	// 记录余额变动
	record := &models.BalanceRecord{
		UserID:        userID,
		OrderNo:       orderNo,
		Type:          balanceType,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Remark:        remark,
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		return errors.New("记录余额变动失败")
	}

	return tx.Commit().Error
}

// ReduceBalance 减少余额
func (s *BalanceService) ReduceBalance(userID uuid.UUID, amount float64, orderNo string, balanceType int8, remark string) error {
	if amount <= 0 {
		return errors.New("金额必须大于0")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询用户当前余额
	var user models.User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&user, userID).Error; err != nil {
		tx.Rollback()
		return errors.New("用户不存在")
	}

	if user.Balance < amount {
		tx.Rollback()
		return errors.New("余额不足")
	}

	balanceBefore := user.Balance
	balanceAfter := balanceBefore - amount

	// 更新用户余额
	if err := tx.Model(&user).Update("balance", balanceAfter).Error; err != nil {
		tx.Rollback()
		return errors.New("更新余额失败")
	}

	// 记录余额变动（金额为负数）
	record := &models.BalanceRecord{
		UserID:        userID,
		OrderNo:       orderNo,
		Type:          balanceType,
		Amount:        -amount, // 负数表示减少
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Remark:        remark,
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		return errors.New("记录余额变动失败")
	}

	return tx.Commit().Error
}

// GetBalanceRecords 获取余额变动记录
func (s *BalanceService) GetBalanceRecords(userID uuid.UUID, page, pageSize int) ([]models.BalanceRecord, int64, error) {
	var records []models.BalanceRecord
	var total int64

	query := s.db.Model(&models.BalanceRecord{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// AdjustBalance 调整余额（管理员）
func (s *BalanceService) AdjustBalance(userID uuid.UUID, amount float64, remark string, adminID uuid.UUID) error {
	orderNo := fmt.Sprintf("ADJ%s%06d", time.Now().Format("20060102150405"), userID.ID()%1000000)

	if amount > 0 {
		return s.AddBalance(userID, amount, orderNo, models.BalanceTypeAdjust, remark)
	} else if amount < 0 {
		return s.ReduceBalance(userID, -amount, orderNo, models.BalanceTypeAdjust, remark)
	}

	return errors.New("调整金额不能为0")
}

// ReduceBalanceWithTx 减少余额（使用外部事务）
func (s *BalanceService) ReduceBalanceWithTx(tx *gorm.DB, userID uuid.UUID, amount float64, orderNo string, balanceType int8, remark string) error {
	if amount <= 0 {
		return errors.New("金额必须大于0")
	}

	// 查询用户当前余额（外部事务已经加锁，不需要再加FOR UPDATE）
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.Balance < amount {
		return errors.New("余额不足")
	}

	balanceBefore := user.Balance
	balanceAfter := balanceBefore - amount

	// 更新用户余额（使用原始SQL避免锁问题）
	result := tx.Exec("UPDATE users SET balance = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		balanceAfter, time.Now(), userID)
	if result.Error != nil {
		return errors.New("更新余额失败")
	}
	if result.RowsAffected == 0 {
		return errors.New("用户不存在或已删除")
	}

	// 记录余额变动（金额为负数）
	record := &models.BalanceRecord{
		UserID:        userID,
		OrderNo:       orderNo,
		Type:          balanceType,
		Amount:        -amount, // 负数表示减少
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Remark:        remark,
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(record).Error; err != nil {
		return errors.New("记录余额变动失败")
	}

	return nil
}
