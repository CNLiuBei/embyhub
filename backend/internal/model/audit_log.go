package model

import "time"

// AuditLog 操作审计日志
type AuditLog struct {
	LogID      int       `gorm:"column:log_id;primaryKey;autoIncrement" json:"log_id"`
	UserID     *int      `gorm:"column:user_id" json:"user_id"`
	Username   string    `gorm:"column:username;type:varchar(50)" json:"username"`
	Action     string    `gorm:"column:action;type:varchar(50);not null" json:"action"`
	TargetType string    `gorm:"column:target_type;type:varchar(50)" json:"target_type"`
	TargetID   string    `gorm:"column:target_id;type:varchar(50)" json:"target_id"`
	Detail     string    `gorm:"column:detail;type:text" json:"detail"`
	IPAddress  string    `gorm:"column:ip_address;type:varchar(45)" json:"ip_address"`
	UserAgent  string    `gorm:"column:user_agent;type:text" json:"user_agent"`
	Status     string    `gorm:"column:status;type:varchar(20);default:success" json:"status"`
	CreatedAt  time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// 操作类型常量
const (
	ActionLogin       = "login"
	ActionLogout      = "logout"
	ActionLoginFailed = "login_failed"
	ActionCreateUser  = "create_user"
	ActionUpdateUser  = "update_user"
	ActionDeleteUser  = "delete_user"
	ActionResetPwd    = "reset_password"
	ActionSetVip      = "set_vip"
	ActionCreateCard  = "create_card_key"
	ActionDeleteCard  = "delete_card_key"
	ActionDisableCard = "disable_card_key"
	ActionEnableCard  = "enable_card_key"
	ActionUseVipCard  = "use_vip_card"
	ActionCreateRole  = "create_role"
	ActionUpdateRole  = "update_role"
	ActionDeleteRole  = "delete_role"
	ActionAssignPerms = "assign_permissions"
)

// 目标类型常量
const (
	TargetUser    = "user"
	TargetRole    = "role"
	TargetCardKey = "card_key"
	TargetSystem  = "system"
)

// AuditLogQuery 审计日志查询请求
type AuditLogQuery struct {
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
	UserID     *int   `form:"user_id" json:"user_id"`
	Action     string `form:"action" json:"action"`
	TargetType string `form:"target_type" json:"target_type"`
	StartTime  string `form:"start_time" json:"start_time"`
	EndTime    string `form:"end_time" json:"end_time"`
	Status     string `form:"status" json:"status"`
}
