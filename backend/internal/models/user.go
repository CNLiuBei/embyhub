// Package models 数据模型定义
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 用户角色常量
const (
	RoleUser       int8 = 0 // 普通用户
	RoleMember     int8 = 1 // 会员用户
	RoleAdmin      int8 = 2 // 管理员
	RoleSuperAdmin int8 = 3 // 超级管理员
)

// 会员等级常量
const (
	MemberNormal int8 = 0 // 普通用户
	MemberMonth  int8 = 1 // 月卡用户
	MemberYear   int8 = 2 // 年卡用户
)

// User 用户模型
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Username     string         `gorm:"size:64;uniqueIndex" json:"username"` // 账号(可选)
	Email        string         `gorm:"size:128;uniqueIndex" json:"email"`   // 邮箱(可选)
	Password     string         `gorm:"size:128;not null" json:"-"`
	Nickname     string         `gorm:"size:64" json:"nickname"`
	Avatar       string         `gorm:"size:256" json:"avatar"`
	Gender       int8           `gorm:"default:0" json:"gender"`                     // 性别 0保密 1男 2女
	Phone        string         `gorm:"size:20" json:"phone"`                        // 手机号
	Bio          string         `gorm:"size:500" json:"bio"`                         // 个人简介
	InviteCode   string         `gorm:"size:16;uniqueIndex" json:"invite_code"`      // 用户专属邀请码
	InvitedBy    *uuid.UUID     `gorm:"type:uuid" json:"invited_by"`                 // 邀请人ID
	InviteCount  int            `gorm:"default:0" json:"invite_count"`               // 邀请人数
	Status       int8           `gorm:"default:1" json:"status"`                     // 1正常 2禁用
	Role         int8           `gorm:"default:0" json:"role"`                       // 0普通用户 1管理员
	Balance      float64        `gorm:"type:decimal(10,2);default:0" json:"balance"` // 账户余额
	Points       int            `gorm:"default:0" json:"points"`                     // 积分
	MemberLevel  int8           `gorm:"default:0" json:"member_level"`               // 0普通 1月卡 2年卡
	MemberExpire *time.Time     `json:"member_expire"`                               // 会员到期时间
	EmbyUserID   string         `gorm:"size:64;index" json:"emby_user_id"`           // Emby用户ID
	LastLoginAt  *time.Time     `json:"last_login_at"`
	LastLoginIP  string         `gorm:"size:64" json:"last_login_ip"`
	RegisterIP   string         `gorm:"size:64" json:"register_ip"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate 创建前生成UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// IsMember 是否是会员
func (u *User) IsMember() bool {
	if u.MemberLevel == MemberNormal {
		return false
	}
	if u.MemberExpire == nil {
		return false
	}
	return u.MemberExpire.After(time.Now())
}

// GetMemberLevelName 获取会员等级名称
func (u *User) GetMemberLevelName() string {
	switch u.MemberLevel {
	case MemberMonth:
		return "月卡会员"
	case MemberYear:
		return "年卡会员"
	default:
		return "普通用户"
	}
}

// LoginLog 登录日志
type LoginLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	IP        string    `gorm:"size:64" json:"ip"`
	Device    string    `gorm:"size:256" json:"device"`
	Location  string    `gorm:"size:128" json:"location"`
	Status    int8      `gorm:"default:1" json:"status"` // 1成功 2失败
	Remark    string    `gorm:"size:256" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}

// UserDevice 用户设备绑定
type UserDevice struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	DeviceID   string    `gorm:"size:128;index" json:"device_id"`
	DeviceName string    `gorm:"size:128" json:"device_name"`
	DeviceType string    `gorm:"size:32" json:"device_type"` // mobile/pc/tv
	LastIP     string    `gorm:"size:64" json:"last_ip"`
	LastUsedAt time.Time `json:"last_used_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func (UserDevice) TableName() string {
	return "user_devices"
}

// WatchHistory 观影记录
type WatchHistory struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	VideoID   string    `gorm:"size:64;index" json:"video_id"`
	VideoName string    `gorm:"size:256" json:"video_name"`
	Duration  int       `json:"duration"` // 观看时长(秒)
	Progress  int       `json:"progress"` // 观看进度(秒)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WatchHistory) TableName() string {
	return "watch_histories"
}

// Favorite 收藏
type Favorite struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	VideoID   string    `gorm:"size:64;index" json:"video_id"`
	VideoName string    `gorm:"size:256" json:"video_name"`
	Cover     string    `gorm:"size:512" json:"cover"`
	CreatedAt time.Time `json:"created_at"`
}

func (Favorite) TableName() string {
	return "favorites"
}

// Notification 通知消息
type Notification struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Title     string    `gorm:"size:128" json:"title"`
	Content   string    `gorm:"size:1024" json:"content"`
	Type      int8      `json:"type"` // 1系统 2会员 3活动
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

// OperationLog 操作日志(管理员审计)
type OperationLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	AdminID   uuid.UUID `gorm:"type:uuid;index" json:"admin_id"`
	AdminName string    `gorm:"size:64" json:"admin_name"`
	Action    string    `gorm:"size:64" json:"action"`
	Target    string    `gorm:"size:128" json:"target"`
	Detail    string    `gorm:"type:text" json:"detail"`
	IP        string    `gorm:"size:64" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}

// MemberOrder 会员订单(卡密兑换记录)
type MemberOrder struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	OrderNo    string    `gorm:"size:64;uniqueIndex" json:"order_no"`
	UserID     uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	CardID     uint64    `json:"card_id"`                  // 使用的卡密ID
	CardCode   string    `gorm:"size:64" json:"card_code"` // 卡密码
	Level      int8      `json:"level"`                    // 会员等级
	Duration   int       `json:"duration"`                 // 天数
	Status     int8      `json:"status"`                   // 1已兑换 2已过期
	RedeemTime time.Time `json:"redeem_time"`              // 兑换时间
	ExpireTime time.Time `json:"expire_time"`              // 会员到期时间
	CreatedAt  time.Time `json:"created_at"`
}

func (MemberOrder) TableName() string {
	return "member_orders"
}

// 卡密状态常量
const (
	CardStatusUnused  int8 = 0 // 未使用
	CardStatusUsed    int8 = 1 // 已使用
	CardStatusExpired int8 = 2 // 已过期
	CardStatusDisable int8 = 3 // 已禁用
)

// Card 卡密模型
type Card struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	Code      string     `gorm:"size:64;uniqueIndex;not null" json:"code"` // 卡密码
	BatchNo   string     `gorm:"size:32;index" json:"batch_no"`            // 批次号
	CardType  int8       `json:"card_type"`                                // 1月卡 2年卡
	Duration  int        `json:"duration"`                                 // 有效天数
	Status    int8       `gorm:"default:0" json:"status"`                  // 0未使用 1已使用 2已过期 3已禁用
	UsedBy    *uuid.UUID `gorm:"type:uuid" json:"used_by"`                 // 使用者ID
	UsedAt    *time.Time `json:"used_at"`                                  // 使用时间
	ExpireAt  *time.Time `json:"expire_at"`                                // 卡密过期时间(可选)
	CreatedBy uuid.UUID  `gorm:"type:uuid" json:"created_by"`              // 创建者(管理员)
	Remark    string     `gorm:"size:256" json:"remark"`                   // 备注
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (Card) TableName() string {
	return "cards"
}

// GetCardTypeName 获取卡密类型名称
func (c *Card) GetCardTypeName() string {
	switch c.CardType {
	case 1:
		return "月卡"
	case 2:
		return "季卡"
	case 3:
		return "半年卡"
	case 4:
		return "年卡"
	default:
		return "未知"
	}
}

// GetStatusName 获取状态名称
func (c *Card) GetStatusName() string {
	switch c.Status {
	case CardStatusUnused:
		return "未使用"
	case CardStatusUsed:
		return "已使用"
	case CardStatusExpired:
		return "已过期"
	case CardStatusDisable:
		return "已禁用"
	default:
		return "未知"
	}
}

// CardBatch 卡密批次
type CardBatch struct {
	ID            uint64     `gorm:"primaryKey" json:"id"`
	BatchNo       string     `gorm:"size:32;uniqueIndex;not null" json:"batch_no"` // 批次号
	CardType      int8       `json:"card_type"`                                    // 1月卡 2年卡
	Duration      int        `json:"duration"`                                     // 有效天数
	Quantity      int        `json:"quantity"`                                     // 生成数量
	UsedCount     int        `gorm:"default:0" json:"used_count"`                  // 已使用数量
	ExpireAt      *time.Time `json:"expire_at"`                                    // 批次过期时间
	CreatedBy     uuid.UUID  `gorm:"type:uuid" json:"created_by"`                  // 创建者
	CreatedByName string     `gorm:"size:64" json:"created_by_name"`               // 创建者名称
	Remark        string     `gorm:"size:256" json:"remark"`                       // 备注
	CreatedAt     time.Time  `json:"created_at"`
}

func (CardBatch) TableName() string {
	return "card_batches"
}

// ============= 余额系统相关模型 =============

// 余额变动类型常量
const (
	BalanceTypeRecharge int8 = 1 // 充值
	BalanceTypeConsume  int8 = 2 // 消费
	BalanceTypeRefund   int8 = 3 // 退款
	BalanceTypeReward   int8 = 4 // 奖励
	BalanceTypeAdjust   int8 = 5 // 调整
)

// BalanceRecord 余额变动记录
type BalanceRecord struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;index" json:"user_id"`           // 用户ID
	OrderNo       string    `gorm:"size:64;index" json:"order_no"`            // 关联订单号
	Type          int8      `json:"type"`                                     // 变动类型 1充值 2消费 3退款 4奖励 5调整
	Amount        float64   `gorm:"type:decimal(10,2)" json:"amount"`         // 变动金额（正数为增加，负数为减少）
	BalanceBefore float64   `gorm:"type:decimal(10,2)" json:"balance_before"` // 变动前余额
	BalanceAfter  float64   `gorm:"type:decimal(10,2)" json:"balance_after"`  // 变动后余额
	Remark        string    `gorm:"size:256" json:"remark"`                   // 备注说明
	CreatedAt     time.Time `json:"created_at"`
}

func (BalanceRecord) TableName() string {
	return "balance_records"
}

// GetTypeName 获取类型名称
func (r *BalanceRecord) GetTypeName() string {
	switch r.Type {
	case BalanceTypeRecharge:
		return "充值"
	case BalanceTypeConsume:
		return "消费"
	case BalanceTypeRefund:
		return "退款"
	case BalanceTypeReward:
		return "奖励"
	case BalanceTypeAdjust:
		return "调整"
	default:
		return "未知"
	}
}

// ============= 系统设置相关模型 =============

// Setting 系统设置模型
type Setting struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"size:64;uniqueIndex;not null" json:"key"` // 设置键名
	Value     string    `gorm:"type:text" json:"value"`                  // 设置值(JSON格式)
	Type      string    `gorm:"size:32" json:"type"`                     // 设置类型: email, domain, system
	UpdatedBy uuid.UUID `gorm:"type:uuid" json:"updated_by"`             // 更新者ID
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (Setting) TableName() string {
	return "settings"
}

// ============= 公告系统相关模型 =============

// 公告状态常量
const (
	AnnouncementStatusDraft   int8 = 0 // 草稿
	AnnouncementStatusPublish int8 = 1 // 已发布
	AnnouncementStatusOffline int8 = 2 // 已下线
)

// 公告类型常量
const (
	AnnouncementTypeNotice   int8 = 1 // 通知
	AnnouncementTypeActivity int8 = 2 // 活动
	AnnouncementTypeUpdate   int8 = 3 // 更新
)

// Announcement 公告模型
type Announcement struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:128;not null" json:"title"` // 标题
	Content   string    `gorm:"type:text" json:"content"`       // 内容
	Type      int8      `gorm:"default:1" json:"type"`          // 类型 1通知 2活动 3更新
	IsTop     bool      `gorm:"default:false" json:"is_top"`    // 是否置顶
	Status    int8      `gorm:"default:0" json:"status"`        // 状态 0草稿 1发布 2下线
	CreatedBy uuid.UUID `gorm:"type:uuid" json:"created_by"`    // 创建者
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Announcement) TableName() string {
	return "announcements"
}

// ============= IP黑名单相关模型 =============

// IPBlacklist IP黑名单模型
type IPBlacklist struct {
	ID         uint64     `gorm:"primaryKey" json:"id"`
	IP         string     `gorm:"size:64;uniqueIndex;not null" json:"ip"` // IP地址
	Reason     string     `gorm:"size:256" json:"reason"`                 // 封禁原因
	ExpireAt   *time.Time `json:"expire_at"`                              // 过期时间，NULL表示永久
	BlockCount int        `gorm:"default:1" json:"block_count"`           // 被封禁次数，用于判断是否永久封禁
	CreatedBy  uuid.UUID  `gorm:"type:uuid" json:"created_by"`            // 创建者
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (IPBlacklist) TableName() string {
	return "ip_blacklists"
}

// ============= 邀请码相关模型 =============

// 邀请码状态
const (
	InviteCodeStatusUnused   int8 = 0 // 未使用
	InviteCodeStatusUsed     int8 = 1 // 已使用
	InviteCodeStatusDisabled int8 = 2 // 已禁用
)

// InviteCode 邀请码模型
type InviteCode struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	Code      string     `gorm:"size:32;uniqueIndex;not null" json:"code"` // 邀请码
	CreatedBy uuid.UUID  `gorm:"type:uuid" json:"created_by"`              // 创建者
	UsedBy    *uuid.UUID `gorm:"type:uuid" json:"used_by"`                 // 使用者
	UsedAt    *time.Time `json:"used_at"`                                  // 使用时间
	Status    int8       `gorm:"default:0" json:"status"`                  // 状态
	GiftDays  int        `gorm:"default:0" json:"gift_days"`               // 赠送天数
	ExpireAt  *time.Time `json:"expire_at"`                                // 过期时间
	CreatedAt time.Time  `json:"created_at"`
}

func (InviteCode) TableName() string {
	return "invite_codes"
}

// InviteRecord 邀请记录
type InviteRecord struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	InviterID  uuid.UUID `gorm:"type:uuid;index" json:"inviter_id"` // 邀请人ID
	InviteeID  uuid.UUID `gorm:"type:uuid;index" json:"invitee_id"` // 被邀请人ID
	InviteCode string    `gorm:"size:32" json:"invite_code"`        // 使用的邀请码
	RewardDays int       `gorm:"default:0" json:"reward_days"`      // 奖励天数
	CreatedAt  time.Time `json:"created_at"`
}

func (InviteRecord) TableName() string {
	return "invite_records"
}

// ============= 积分系统相关模型 =============

// 积分变动类型常量
const (
	PointsTypeSignIn     int8 = 1  // 签到
	PointsTypeInvite     int8 = 2  // 邀请奖励
	PointsTypeConsume    int8 = 3  // 消费（兑换）
	PointsTypeReward     int8 = 4  // 系统奖励
	PointsTypeAdjust     int8 = 5  // 管理员调整
	PointsTypeRegister   int8 = 6  // 注册奖励
	PointsTypeActivity   int8 = 7  // 活动奖励
	PointsTypeExpire     int8 = 8  // 积分过期
	PointsTypeWatch      int8 = 9  // 观影奖励
	PointsTypeExchange   int8 = 10 // 积分兑换会员
)

// PointsRecord 积分变动记录
type PointsRecord struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;index" json:"user_id"`          // 用户ID
	Type         int8      `json:"type"`                                    // 变动类型
	Points       int       `json:"points"`                                  // 变动积分（正数为增加，负数为减少）
	PointsBefore int       `json:"points_before"`                           // 变动前积分
	PointsAfter  int       `json:"points_after"`                            // 变动后积分
	Remark       string    `gorm:"size:256" json:"remark"`                  // 备注说明
	RelatedID    string    `gorm:"size:64" json:"related_id"`               // 关联ID（如订单号、活动ID等）
	CreatedAt    time.Time `json:"created_at"`
}

func (PointsRecord) TableName() string {
	return "points_records"
}

// GetTypeName 获取积分类型名称
func (r *PointsRecord) GetTypeName() string {
	switch r.Type {
	case PointsTypeSignIn:
		return "签到"
	case PointsTypeInvite:
		return "邀请奖励"
	case PointsTypeConsume:
		return "积分消费"
	case PointsTypeReward:
		return "系统奖励"
	case PointsTypeAdjust:
		return "管理员调整"
	case PointsTypeRegister:
		return "注册奖励"
	case PointsTypeActivity:
		return "活动奖励"
	case PointsTypeExpire:
		return "积分过期"
	case PointsTypeWatch:
		return "观影奖励"
	case PointsTypeExchange:
		return "兑换会员"
	default:
		return "未知"
	}
}

// SignInRecord 签到记录
type SignInRecord struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;index" json:"user_id"`       // 用户ID
	SignDate      string    `gorm:"size:10;index" json:"sign_date"`       // 签到日期 YYYY-MM-DD
	Points        int       `json:"points"`                               // 获得积分
	ContinueDays  int       `json:"continue_days"`                        // 连续签到天数
	CreatedAt     time.Time `json:"created_at"`
}

func (SignInRecord) TableName() string {
	return "sign_in_records"
}

// PointsExchangeRule 积分兑换规则
type PointsExchangeRule struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:64;not null" json:"name"`           // 规则名称
	Points      int       `json:"points"`                                 // 所需积分
	MemberDays  int       `json:"member_days"`                            // 兑换会员天数
	Description string    `gorm:"size:256" json:"description"`            // 描述
	Enabled     bool      `gorm:"default:true" json:"enabled"`            // 是否启用
	SortOrder   int       `gorm:"default:0" json:"sort_order"`            // 排序
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (PointsExchangeRule) TableName() string {
	return "points_exchange_rules"
}

// ============= 积分卡密相关模型 =============

// PointsCard 积分卡密
type PointsCard struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	Code      string     `gorm:"size:32;uniqueIndex;not null" json:"code"` // 卡密码
	BatchNo   string     `gorm:"size:32;index" json:"batch_no"`            // 批次号
	Points    int        `json:"points"`                                   // 积分数量
	Status    int8       `gorm:"default:0" json:"status"`                  // 0未使用 1已使用 3已禁用
	UsedBy    *uuid.UUID `gorm:"type:uuid" json:"used_by"`                 // 使用者ID
	UsedAt    *time.Time `json:"used_at"`                                  // 使用时间
	ExpireAt  *time.Time `json:"expire_at"`                                // 卡密过期时间(可选)
	CreatedBy uuid.UUID  `gorm:"type:uuid" json:"created_by"`              // 创建者(管理员)
	Remark    string     `gorm:"size:256" json:"remark"`                   // 备注
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (PointsCard) TableName() string {
	return "points_cards"
}

// GetStatusName 获取状态名称
func (c *PointsCard) GetStatusName() string {
	switch c.Status {
	case 0:
		return "未使用"
	case 1:
		return "已使用"
	case 3:
		return "已禁用"
	default:
		return "未知"
	}
}

// PointsCardBatch 积分卡密批次
type PointsCardBatch struct {
	ID            uint64     `gorm:"primaryKey" json:"id"`
	BatchNo       string     `gorm:"size:32;uniqueIndex;not null" json:"batch_no"` // 批次号
	Points        int        `json:"points"`                                       // 积分数量
	Quantity      int        `json:"quantity"`                                     // 生成数量
	UsedCount     int        `gorm:"default:0" json:"used_count"`                  // 已使用数量
	ExpireAt      *time.Time `json:"expire_at"`                                    // 批次过期时间
	CreatedBy     uuid.UUID  `gorm:"type:uuid" json:"created_by"`                  // 创建者
	CreatedByName string     `gorm:"size:64" json:"created_by_name"`               // 创建者名称
	Remark        string     `gorm:"size:256" json:"remark"`                       // 备注
	CreatedAt     time.Time  `json:"created_at"`
}

func (PointsCardBatch) TableName() string {
	return "points_card_batches"
}


// ============= 积分自动赠送规则相关模型 =============

// 自动赠送规则类型
const (
	GiftRuleTypeDaily    int8 = 1 // 每日赠送
	GiftRuleTypeWeekly   int8 = 2 // 每周赠送
	GiftRuleTypeMonthly  int8 = 3 // 每月赠送
	GiftRuleTypeYearly   int8 = 4 // 每年赠送
	GiftRuleTypeBirthday int8 = 5 // 生日赠送（预留）
)

// PointsGiftRule 积分自动赠送规则
type PointsGiftRule struct {
	ID                uint64     `gorm:"primaryKey" json:"id"`
	Name              string     `gorm:"size:64;not null" json:"name"`              // 规则名称
	RuleType          int8       `json:"rule_type"`                                 // 规则类型 1每日 2每周 3每月 4每年
	Points            int        `json:"points"`                                    // 赠送积分
	TargetType        string     `gorm:"size:32" json:"target_type"`                // 目标类型: all, member, non_member
	MemberLevel       *int8      `json:"member_level"`                              // 会员等级筛选
	ExecuteTime       string     `gorm:"size:8" json:"execute_time"`                // 执行时间 HH:MM
	ExecuteDay        int        `json:"execute_day"`                               // 执行日（周几1-7 或 每月几号1-31）
	ExecuteMonth      int        `json:"execute_month"`                             // 执行月份（每年赠送时使用，1-12）
	SendNotification  bool       `gorm:"default:true" json:"send_notification"`     // 是否发送站内通知
	NotificationTitle string     `gorm:"size:128" json:"notification_title"`        // 通知标题
	NotificationBody  string     `gorm:"size:512" json:"notification_body"`         // 通知内容
	Enabled           bool       `gorm:"default:true" json:"enabled"`               // 是否启用
	LastExecuteAt     *time.Time `json:"last_execute_at"`                           // 上次执行时间
	NextExecuteAt     *time.Time `json:"next_execute_at"`                           // 下次执行时间
	CreatedBy         uuid.UUID  `gorm:"type:uuid" json:"created_by"`               // 创建者
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (PointsGiftRule) TableName() string {
	return "points_gift_rules"
}

// GetRuleTypeName 获取规则类型名称
func (r *PointsGiftRule) GetRuleTypeName() string {
	switch r.RuleType {
	case GiftRuleTypeDaily:
		return "每日赠送"
	case GiftRuleTypeWeekly:
		return "每周赠送"
	case GiftRuleTypeMonthly:
		return "每月赠送"
	case GiftRuleTypeYearly:
		return "每年赠送"
	case GiftRuleTypeBirthday:
		return "生日赠送"
	default:
		return "未知"
	}
}

// PointsGiftLog 积分赠送执行日志
type PointsGiftLog struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	RuleID       uint64    `gorm:"index" json:"rule_id"`                   // 规则ID
	RuleName     string    `gorm:"size:64" json:"rule_name"`               // 规则名称
	Points       int       `json:"points"`                                 // 赠送积分
	TotalUsers   int       `json:"total_users"`                            // 目标用户数
	SuccessCount int       `json:"success_count"`                          // 成功数
	FailedCount  int       `json:"failed_count"`                           // 失败数
	ExecuteAt    time.Time `json:"execute_at"`                             // 执行时间
	CreatedAt    time.Time `json:"created_at"`
}

func (PointsGiftLog) TableName() string {
	return "points_gift_logs"
}

// ============= 论坛系统相关模型 =============

// 话题节点（板块）
type ForumNode struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:32;not null" json:"name"`        // 节点名称
	Description string    `gorm:"size:256" json:"description"`         // 节点描述
	Icon        string    `gorm:"size:64" json:"icon"`                 // 图标
	SortOrder   int       `gorm:"default:0" json:"sort_order"`         // 排序
	TopicCount  int       `gorm:"default:0" json:"topic_count"`        // 话题数量
	Status      int8      `gorm:"default:1" json:"status"`             // 状态 1正常 0禁用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ForumNode) TableName() string {
	return "forum_nodes"
}

// 话题类型
const (
	TopicTypeNormal int8 = 0 // 普通话题
	TopicTypeImage  int8 = 1 // 图片话题
)

// 话题状态
const (
	TopicStatusNormal  int8 = 0 // 正常
	TopicStatusDeleted int8 = 1 // 已删除
	TopicStatusPending int8 = 2 // 待审核
)

// ForumTopic 论坛话题/帖子
type ForumTopic struct {
	ID              uint64         `gorm:"primaryKey" json:"id"`
	NodeID          uint64         `gorm:"index:idx_topic_node_status" json:"node_id"`                       // 节点ID
	UserID          uuid.UUID      `gorm:"type:uuid;index:idx_topic_user_status" json:"user_id"`             // 发布者ID
	Title           string         `gorm:"size:128;index:idx_topic_title" json:"title"`                      // 标题
	Content         string         `gorm:"type:text" json:"content"`                                         // 内容
	ContentType     string         `gorm:"size:16;default:html" json:"content_type"`                         // 内容类型 html/markdown
	Images          string         `gorm:"type:text" json:"images"`                                          // 图片列表JSON
	TopicType       int8           `gorm:"default:0" json:"topic_type"`                                      // 话题类型
	ViewCount       int            `gorm:"default:0" json:"view_count"`                                      // 浏览数
	CommentCount    int            `gorm:"default:0;index:idx_topic_hot" json:"comment_count"`               // 评论数
	LikeCount       int            `gorm:"default:0" json:"like_count"`                                      // 点赞数
	FavoriteCount   int            `gorm:"default:0" json:"favorite_count"`                                  // 收藏数
	IsTop           bool           `gorm:"default:false;index:idx_topic_top" json:"is_top"`                  // 是否置顶
	IsRecommend     bool           `gorm:"default:false;index:idx_topic_recommend" json:"is_recommend"`      // 是否推荐
	Status          int8           `gorm:"default:0;index:idx_topic_node_status;index:idx_topic_user_status" json:"status"` // 状态
	LastCommentTime *time.Time     `json:"last_comment_time"`                                                // 最后评论时间
	LastCommentUser *uuid.UUID     `gorm:"type:uuid" json:"last_comment_user"`                               // 最后评论用户
	IP              string         `gorm:"size:64" json:"-"`                                                 // 发布IP（不返回给前端）
	Location        string         `gorm:"size:64" json:"location"`                                          // 发布地区
	CreatedAt       time.Time      `gorm:"index:idx_topic_created" json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联数据（非数据库字段）
	User     *User      `gorm:"-" json:"user,omitempty"`
	Node     *ForumNode `gorm:"-" json:"node,omitempty"`
	IsLiked  bool       `gorm:"-" json:"is_liked"`
	IsFaved  bool       `gorm:"-" json:"is_faved"`
}

func (ForumTopic) TableName() string {
	return "forum_topics"
}

// ForumComment 论坛评论
type ForumComment struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	TopicID      uint64         `gorm:"index" json:"topic_id"`                    // 话题ID
	UserID       uuid.UUID      `gorm:"type:uuid;index" json:"user_id"`           // 评论者ID
	Content      string         `gorm:"type:text;not null" json:"content"`        // 评论内容
	Images       string         `gorm:"type:text" json:"images"`                  // 图片列表JSON
	ParentID     uint64         `gorm:"default:0;index" json:"parent_id"`         // 父评论ID（回复）
	ReplyToUser  *uuid.UUID     `gorm:"type:uuid" json:"reply_to_user"`           // 回复的用户ID
	LikeCount    int            `gorm:"default:0" json:"like_count"`              // 点赞数
	ReplyCount   int            `gorm:"default:0" json:"reply_count"`             // 回复数
	Status       int8           `gorm:"default:0" json:"status"`                  // 状态 0正常 1删除
	IP           string         `gorm:"size:64" json:"-"`                         // 发布IP（不返回给前端）
	Location     string         `gorm:"size:64" json:"location"`                  // 发布地区
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联数据
	User        *User          `gorm:"-" json:"user,omitempty"`
	ReplyToName string         `gorm:"-" json:"reply_to_name,omitempty"`
	Replies     []ForumComment `gorm:"-" json:"replies,omitempty"`
	IsLiked     bool           `gorm:"-" json:"is_liked"`
}

func (ForumComment) TableName() string {
	return "forum_comments"
}

// ForumLike 点赞记录
type ForumLike struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_forum_like" json:"user_id"`
	EntityType string    `gorm:"size:16;uniqueIndex:idx_forum_like" json:"entity_type"` // topic/comment
	EntityID   uint64    `gorm:"uniqueIndex:idx_forum_like" json:"entity_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ForumLike) TableName() string {
	return "forum_likes"
}

// ForumFavorite 收藏记录
type ForumFavorite struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_forum_fav" json:"user_id"`
	TopicID   uint64    `gorm:"uniqueIndex:idx_forum_fav" json:"topic_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (ForumFavorite) TableName() string {
	return "forum_favorites"
}

// TopicView 话题浏览记录（用于防刷浏览量）
type TopicView struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TopicID   uint64    `gorm:"uniqueIndex:idx_topic_view" json:"topic_id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_topic_view" json:"user_id"`
	IP        string    `gorm:"size:64;uniqueIndex:idx_topic_view" json:"ip"` // 未登录用户用IP
	ViewedAt  time.Time `gorm:"index" json:"viewed_at"`                       // 最后浏览时间
	CreatedAt time.Time `json:"created_at"`
}

func (TopicView) TableName() string {
	return "topic_views"
}

// ============= 私信系统相关模型 =============

// 私信状态
const (
	MessageStatusUnread    int8 = 0 // 未读
	MessageStatusRead      int8 = 1 // 已读
	MessageStatusDeleted   int8 = 2 // 已删除
	MessageStatusRecalled  int8 = 3 // 已撤回
)

// 消息类型
const (
	MessageTypeText   int8 = 1 // 文本
	MessageTypeImage  int8 = 2 // 图片
	MessageTypeFile   int8 = 3 // 文件
	MessageTypeSystem int8 = 4 // 系统消息
)

// 撤回时间限制（分钟）
const MessageRecallTimeLimit = 5

// PrivateMessage 私信消息
type PrivateMessage struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	FromUserID  uuid.UUID      `gorm:"type:uuid;index" json:"from_user_id"`       // 发送者ID
	ToUserID    uuid.UUID      `gorm:"type:uuid;index" json:"to_user_id"`         // 接收者ID
	Content     string         `gorm:"type:text;not null" json:"content"`         // 消息内容
	ContentType int8           `gorm:"default:1" json:"content_type"`             // 消息类型
	Images      string         `gorm:"type:text" json:"images"`                   // 图片列表JSON
	Status      int8           `gorm:"default:0" json:"status"`                   // 状态
	ReadAt      *time.Time     `json:"read_at"`                                   // 阅读时间
	RecalledAt  *time.Time     `json:"recalled_at"`                               // 撤回时间
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联数据
	FromUser *User `gorm:"-" json:"from_user,omitempty"`
	ToUser   *User `gorm:"-" json:"to_user,omitempty"`
}

func (PrivateMessage) TableName() string {
	return "private_messages"
}

// Conversation 会话（用于私信列表）
type Conversation struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	User1ID       uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_conversation" json:"user1_id"` // 用户1（ID较小的）
	User2ID       uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_conversation" json:"user2_id"` // 用户2（ID较大的）
	LastMessageID uint64    `json:"last_message_id"`                                        // 最后一条消息ID
	LastMessageAt time.Time `json:"last_message_at"`                                        // 最后消息时间
	User1Unread   int       `gorm:"default:0" json:"user1_unread"`                          // 用户1未读数
	User2Unread   int       `gorm:"default:0" json:"user2_unread"`                          // 用户2未读数
	User1Muted    bool      `gorm:"default:false" json:"user1_muted"`                       // 用户1是否静音
	User2Muted    bool      `gorm:"default:false" json:"user2_muted"`                       // 用户2是否静音
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联数据
	OtherUser   *User           `gorm:"-" json:"other_user,omitempty"`
	LastMessage *PrivateMessage `gorm:"-" json:"last_message,omitempty"`
	UnreadCount int             `gorm:"-" json:"unread_count"`
	IsMuted     bool            `gorm:"-" json:"is_muted"`
}

func (Conversation) TableName() string {
	return "conversations"
}

// UserBlacklist 用户黑名单
type UserBlacklist struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_blacklist" json:"user_id"`    // 用户ID
	BlockedID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_blacklist" json:"blocked_id"` // 被拉黑的用户ID
	Reason    string    `gorm:"size:256" json:"reason"`                                // 拉黑原因
	CreatedAt time.Time `json:"created_at"`

	// 关联数据
	BlockedUser *User `gorm:"-" json:"blocked_user,omitempty"`
}

func (UserBlacklist) TableName() string {
	return "user_blacklists"
}

// UserFollow 用户关注
type UserFollow struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_follow" json:"user_id"`   // 关注者
	FollowID   uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_follow" json:"follow_id"` // 被关注者
	CreatedAt  time.Time `json:"created_at"`
}

func (UserFollow) TableName() string {
	return "user_follows"
}
