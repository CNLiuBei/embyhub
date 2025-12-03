package model

import "time"

// CardKey 卡密模型
type CardKey struct {
	ID        int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CardCode  string     `gorm:"column:card_code;type:varchar(32);not null;uniqueIndex" json:"card_code"`
	CardType  int        `gorm:"column:card_type;type:smallint;not null;default:1" json:"card_type"` // 1=注册码 2=VIP升级码
	Duration  int        `gorm:"column:duration;not null;default:30" json:"duration"`                // 有效期（天）
	Status    int        `gorm:"column:status;type:smallint;not null;default:1" json:"status"`       // 0=已禁用 1=未使用 2=已使用
	UsedBy    *int       `gorm:"column:used_by" json:"used_by,omitempty"`                            // 使用者用户ID
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at,omitempty"`                            // 使用时间
	ExpireAt  *time.Time `gorm:"column:expire_at" json:"expire_at,omitempty"`                        // 过期时间
	Remark    string     `gorm:"column:remark;type:varchar(200)" json:"remark"`                      // 备注
	CreatedBy int        `gorm:"column:created_by;not null" json:"created_by"`                       // 创建者
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// 关联
	UsedByUser    *User `gorm:"foreignKey:UsedBy" json:"used_by_user,omitempty"`
	CreatedByUser *User `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`
}

// TableName 指定表名
func (CardKey) TableName() string {
	return "card_keys"
}

// CardKeyCreateRequest 创建卡密请求
type CardKeyCreateRequest struct {
	Count    int    `json:"count" binding:"required,min=1,max=100"`    // 生成数量
	CardType int    `json:"card_type" binding:"required,oneof=1"`      // 卡密类型（1=VIP会员码）
	Duration int    `json:"duration" binding:"required,min=1,max=365"` // 有效期（天）
	Remark   string `json:"remark" binding:"omitempty,max=200"`        // 备注
}

// CardKeyListRequest 卡密列表请求
type CardKeyListRequest struct {
	Page     int    `form:"page" binding:"omitempty,gt=0"`
	PageSize int    `form:"page_size" binding:"omitempty,gt=0,lte=100"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1 2"`
	CardType *int   `form:"card_type" binding:"omitempty,oneof=1 2"`
	Keyword  string `form:"keyword"`
}

// CardKeyListResponse 卡密列表响应
type CardKeyListResponse struct {
	Total int        `json:"total"`
	List  []*CardKey `json:"list"`
}

// CardKeyUseRequest 使用卡密请求
type CardKeyUseRequest struct {
	CardCode string `json:"card_code" binding:"required"`
}

// CardTypeText 卡密类型文本
func CardTypeText(cardType int) string {
	switch cardType {
	case 1:
		return "VIP会员码"
	default:
		return "VIP会员码"
	}
}

// CardStatusText 卡密状态文本
func CardStatusText(status int) string {
	switch status {
	case 0:
		return "已禁用"
	case 1:
		return "未使用"
	case 2:
		return "已使用"
	default:
		return "未知"
	}
}
