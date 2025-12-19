// Package models VIP相关数据模型
package models

import (
	"time"

	"github.com/google/uuid"
)

// VipPlan 会员套餐表
type VipPlan struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(50);not null" json:"name"` // 套餐名称
	Price        int64     `gorm:"not null" json:"price"`                 // 价格（分）
	DurationDays int       `gorm:"not null" json:"duration_days"`         // 会员天数
	IsActive     bool      `gorm:"default:true" json:"is_active"`         // 是否启用
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (VipPlan) TableName() string {
	return "vip_plans"
}

// UserVip 用户会员表
type UserVip struct {
	UserID      uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"` // 用户ID
	VipExpireAt *time.Time `gorm:"type:timestamp" json:"vip_expire_at"` // 会员到期时间（NULL表示未开通）
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (UserVip) TableName() string {
	return "user_vips"
}

// IsVipValid 判断会员是否有效
func (uv *UserVip) IsVipValid() bool {
	if uv.VipExpireAt == nil {
		return false
	}
	return uv.VipExpireAt.After(time.Now().UTC())
}

// VipOrder 会员购买订单表
type VipOrder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderNo   string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"order_no"` // 订单号
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`               // 用户ID
	PlanID    uint      `gorm:"not null" json:"plan_id"`                               // 套餐ID
	Amount    int64     `gorm:"not null" json:"amount"`                                // 支付金额（分）
	Status    string    `gorm:"type:varchar(20);not null" json:"status"`               // success / failed / pending
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (VipOrder) TableName() string {
	return "vip_orders"
}

// 订单状态常量
const (
	OrderStatusSuccess = "success"
	OrderStatusFailed  = "failed"
	OrderStatusPending = "pending"
)

// BalanceLog 余额流水表
type BalanceLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"` // 用户ID
	ChangeAmount  int64     `gorm:"not null" json:"change_amount"`           // 变动金额（分，负数表示扣款）
	BeforeBalance int64     `gorm:"not null" json:"before_balance"`          // 变动前余额（分）
	AfterBalance  int64     `gorm:"not null" json:"after_balance"`           // 变动后余额（分）
	Type          string    `gorm:"type:varchar(50);not null" json:"type"`   // 类型：vip_purchase / recharge / refund
	OrderNo       string    `gorm:"type:varchar(64);index" json:"order_no"`  // 关联订单号
	Remark        string    `gorm:"type:varchar(255)" json:"remark"`         // 备注
	CreatedAt     time.Time `json:"created_at"`
}

// TableName 指定表名
func (BalanceLog) TableName() string {
	return "balance_logs"
}

// 余额流水类型常量
const (
	BalanceTypeVipPurchase = "vip_purchase" // VIP购买
)
