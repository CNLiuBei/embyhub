// Package models 支付宝支付相关数据模型
package models

import (
	"time"

	"github.com/google/uuid"
)

// =====================================================
// 支付宝配置
// =====================================================

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	ID                     uint      `gorm:"primaryKey" json:"id"`
	AppID                  string    `gorm:"size:32;not null" json:"app_id"`                       // 支付宝应用ID
	AppPublicKey           string    `gorm:"type:text" json:"-"`                                   // 应用公钥
	AppPrivateKeyEncrypted string    `gorm:"type:text" json:"-"`                                   // 加密存储的应用私钥
	AlipayPublicKey        string    `gorm:"type:text" json:"-"`                                   // 支付宝公钥
	NotifyURL              string    `gorm:"size:512" json:"notify_url"`                           // 异步通知地址
	Enabled                bool      `gorm:"default:false" json:"enabled"`                         // 是否启用
	IsProduction           bool      `gorm:"default:false" json:"is_production"`                   // 是否生产环境
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AlipayConfig) TableName() string {
	return "alipay_configs"
}

// =====================================================
// 支付日志
// =====================================================

// AlipayLog 支付宝API调用日志
type AlipayLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	OrderNo      string    `gorm:"size:64;index" json:"order_no"`           // 订单号
	Action       string    `gorm:"size:32;not null;index" json:"action"`    // 操作类型: create/notify/query
	RequestBody  string    `gorm:"type:text" json:"request_body"`           // 请求内容
	ResponseBody string    `gorm:"type:text" json:"response_body"`          // 响应内容
	Status       string    `gorm:"size:20" json:"status"`                   // 状态: success/failed
	ErrorMsg     string    `gorm:"size:500" json:"error_msg"`               // 错误信息
	ClientIP     string    `gorm:"size:64" json:"client_ip"`                // 客户端IP
	Duration     int64     `gorm:"default:0" json:"duration"`               // 耗时(ms)
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名
func (AlipayLog) TableName() string {
	return "alipay_logs"
}

// 日志操作类型常量
const (
	AlipayLogActionCreate = "create" // 创建订单
	AlipayLogActionNotify = "notify" // 异步通知
	AlipayLogActionQuery  = "query"  // 查询订单
)

// 日志状态常量
const (
	AlipayLogStatusSuccess = "success" // 成功
	AlipayLogStatusFailed  = "failed"  // 失败
)

// =====================================================
// 扩展 VipOrder 的支付相关字段
// =====================================================

// AlipayVipOrder 支付宝VIP订单扩展信息
type AlipayVipOrder struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	OrderNo       string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"order_no"` // 订单号
	UserID        uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`               // 用户ID
	PlanID        uint       `gorm:"not null" json:"plan_id"`                               // 套餐ID
	PlanName      string     `gorm:"size:50" json:"plan_name"`                              // 套餐名称（冗余）
	Amount        int64      `gorm:"not null" json:"amount"`                                // 支付金额（分）
	PaymentMethod string     `gorm:"size:20;default:'alipay'" json:"payment_method"`        // 支付方式: alipay
	TradeNo       string     `gorm:"size:64;index" json:"trade_no"`                         // 支付宝交易号
	Status        string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending/success/failed/closed
	PaidAt        *time.Time `gorm:"type:timestamp" json:"paid_at"`                         // 支付完成时间
	ExpireAt      *time.Time `gorm:"type:timestamp" json:"expire_at"`                       // 订单过期时间
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (AlipayVipOrder) TableName() string {
	return "alipay_vip_orders"
}

// 订单状态常量
const (
	AlipayOrderStatusPending = "pending" // 待支付
	AlipayOrderStatusSuccess = "success" // 支付成功
	AlipayOrderStatusFailed  = "failed"  // 支付失败
	AlipayOrderStatusClosed  = "closed"  // 已关闭
)

// 支付方式常量
const (
	PaymentMethodAlipay  = "alipay"  // 支付宝
	PaymentMethodBalance = "balance" // 余额
	PaymentMethodCard    = "card"    // 卡密
)
