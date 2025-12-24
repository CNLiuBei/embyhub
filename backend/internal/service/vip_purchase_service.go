// Package service VIP购买业务逻辑层
package service

import (
	"errors"
	"fmt"
	"time"

	"feiniu-user-system/internal/models"
	"feiniu-user-system/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// VipPurchaseService VIP购买服务
type VipPurchaseService struct {
	db      *gorm.DB
	vipRepo *repository.VipRepository
}

// NewVipPurchaseService 创建VIP购买服务实例
func NewVipPurchaseService(db *gorm.DB) *VipPurchaseService {
	return &VipPurchaseService{
		db:      db,
		vipRepo: repository.NewVipRepository(db),
	}
}

// PurchaseVipRequest 购买VIP请求
type PurchaseVipRequest struct {
	UserID uuid.UUID `json:"-"`                                // 从上下文获取
	PlanID uint      `json:"plan_id" binding:"required,min=1"` // 套餐ID
}

// PurchaseVipResponse 购买VIP响应
type PurchaseVipResponse struct {
	VipExpireAt time.Time `json:"vip_expire_at"` // 会员到期时间
	Balance     int64     `json:"balance"`       // 剩余余额（分）
	OrderNo     string    `json:"order_no"`      // 订单号
}

// PurchaseVip 购买VIP会员（核心业务逻辑）
func (s *VipPurchaseService) PurchaseVip(req *PurchaseVipRequest) (*PurchaseVipResponse, error) {
	var response *PurchaseVipResponse

	// 使用事务保证数据一致性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查询套餐信息
		plan, err := s.vipRepo.GetPlanByID(tx, req.PlanID)
		if err != nil {
			return fmt.Errorf("查询套餐失败: %w", err)
		}

		// 2. 获取用户余额（带行锁，防止并发超卖）
		beforeBalance, err := s.vipRepo.GetUserBalanceWithLock(tx, req.UserID)
		if err != nil {
			return fmt.Errorf("查询用户余额失败: %w", err)
		}

		// 3. 校验余额是否充足
		if beforeBalance < plan.Price {
			return fmt.Errorf("余额不足：当前余额 %d 分，需要 %d 分", beforeBalance, plan.Price)
		}

		// 4. 计算扣款后余额
		afterBalance := beforeBalance - plan.Price

		// 5. 生成订单号
		orderNo := s.generateOrderNo(req.UserID)

		// 6. 创建订单记录
		order := &models.VipOrder{
			OrderNo:   orderNo,
			UserID:    req.UserID,
			PlanID:    plan.ID,
			Amount:    plan.Price,
			Status:    models.OrderStatusPending, // 初始为pending
			CreatedAt: time.Now().UTC(),
		}
		if err := s.vipRepo.CreateOrder(tx, order); err != nil {
			return fmt.Errorf("创建订单失败: %w", err)
		}

		// 7. 扣减用户余额
		if err := s.vipRepo.UpdateUserBalance(tx, req.UserID, afterBalance); err != nil {
			return fmt.Errorf("扣减余额失败: %w", err)
		}

		// 8. 记录余额流水
		balanceLog := &models.BalanceLog{
			UserID:        req.UserID,
			ChangeAmount:  -plan.Price, // 负数表示扣款
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			Type:          models.BalanceTypeVipPurchase,
			OrderNo:       orderNo,
			Remark:        fmt.Sprintf("购买VIP套餐：%s", plan.Name),
			CreatedAt:     time.Now().UTC(),
		}
		if err := s.vipRepo.CreateBalanceLog(tx, balanceLog); err != nil {
			return fmt.Errorf("记录余额流水失败: %w", err)
		}

		// 9. 获取用户当前会员状态（从 users 表，带行锁）
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id, member_expire, member_level").
			Where("id = ?", req.UserID).
			First(&user).Error; err != nil {
			return fmt.Errorf("查询用户信息失败: %w", err)
		}

		beforeExpire := user.MemberExpire // 记录变动前的到期时间

		// 10. 计算新的会员到期时间
		newExpireAt := repository.CalculateNewExpireTime(user.MemberExpire, plan.DurationDays)

		// 11. 更新用户会员信息（统一使用 users 表）
		if err := tx.Model(&models.User{}).Where("id = ?", req.UserID).Updates(map[string]interface{}{
			"member_expire": newExpireAt,
			"member_level":  models.MemberMonth,
			"role":          models.RoleMember,
			"status":        1,
		}).Error; err != nil {
			return fmt.Errorf("更新会员信息失败: %w", err)
		}

		// 12. 记录会员变动日志
		changeLog := models.MemberChangeLog{
			UserID:         req.UserID,
			Source:         models.MemberSourceBalance,
			OrderNo:        orderNo,
			ChangeDays:     plan.DurationDays,
			Amount:         plan.Price,
			BeforeExpireAt: beforeExpire,
			AfterExpireAt:  &newExpireAt,
			Remark:         fmt.Sprintf("余额购买VIP: %s", plan.Name),
		}
		if err := tx.Create(&changeLog).Error; err != nil {
			return fmt.Errorf("记录会员变动失败: %w", err)
		}

		// 13. 更新订单状态为成功
		order.Status = models.OrderStatusSuccess
		if err := tx.Model(order).Where("order_no = ?", orderNo).Update("status", models.OrderStatusSuccess).Error; err != nil {
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		// 14. 构造响应
		response = &PurchaseVipResponse{
			VipExpireAt: newExpireAt,
			Balance:     afterBalance,
			OrderNo:     orderNo,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// generateOrderNo 生成订单号
func (s *VipPurchaseService) generateOrderNo(userID uuid.UUID) string {
	// 格式：VIP + 时间戳 + 用户ID前6位
	timestamp := time.Now().UTC().Format("20060102150405")
	userIDStr := userID.String()[:6]
	return fmt.Sprintf("VIP%s%s", timestamp, userIDStr)
}

// GetVipPlans 获取VIP套餐列表
func (s *VipPurchaseService) GetVipPlans() ([]models.VipPlan, error) {
	var plans []models.VipPlan
	err := s.db.Where("is_active = ?", true).Find(&plans).Error
	if err != nil {
		return nil, errors.New("查询套餐列表失败")
	}
	return plans, nil
}

// GetUserVipInfo 获取用户VIP信息（从 users 表读取）
func (s *VipPurchaseService) GetUserVipInfo(userID uuid.UUID) (*models.UserVip, error) {
	var user models.User
	err := s.db.Select("id, member_expire").Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.UserVip{
				UserID:      userID,
				VipExpireAt: nil,
			}, nil
		}
		return nil, errors.New("查询VIP信息失败")
	}
	return &models.UserVip{
		UserID:      userID,
		VipExpireAt: user.MemberExpire,
	}, nil
}
