package model

import "time"

// User 用户模型
type User struct {
	UserID       int        `gorm:"column:user_id;primaryKey;autoIncrement" json:"user_id"`
	Username     string     `gorm:"column:username;type:varchar(50);not null;uniqueIndex" json:"username"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(100);not null" json:"-"`
	Email        string     `gorm:"column:email;type:varchar(100);uniqueIndex" json:"email"`
	EmbyUserID   string     `gorm:"column:emby_user_id;type:varchar(50);index" json:"emby_user_id"`
	RoleID       int        `gorm:"column:role_id;not null" json:"role_id"`
	Status       int        `gorm:"column:status;type:smallint;not null;default:1" json:"status"`
	VipLevel     int        `gorm:"column:vip_level;default:0" json:"vip_level"`         // VIP等级：0=普通用户 1=VIP会员
	VipExpireAt  *time.Time `gorm:"column:vip_expire_at" json:"vip_expire_at,omitempty"` // VIP到期时间
	CreatedAt    time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// 关联
	Role *Role `gorm:"foreignKey:RoleID;references:RoleID" json:"role,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserCreateRequest 创建用户请求
type UserCreateRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=50"`
	Password   string `json:"password" binding:"required,min=6,max=50"`
	Email      string `json:"email" binding:"omitempty,email"`
	EmbyUserID string `json:"emby_user_id" binding:"omitempty"`
	RoleID     int    `json:"role_id" binding:"required,gt=0"`
}

// UserUpdateRequest 更新用户请求
type UserUpdateRequest struct {
	Email      string `json:"email" binding:"omitempty,email"`
	EmbyUserID string `json:"emby_user_id" binding:"omitempty"`
	RoleID     int    `json:"role_id" binding:"omitempty,gt=0"`
	Status     *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

// UserListRequest 用户列表查询请求
type UserListRequest struct {
	Page     int    `form:"page" binding:"omitempty,gt=0"`
	PageSize int    `form:"page_size" binding:"omitempty,gt=0,lte=100"`
	Keyword  string `form:"keyword"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1"`
	RoleID   int    `form:"role_id" binding:"omitempty,gt=0"`
}

// UserPasswordRequest 修改密码请求
type UserPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6,max=50"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Total int     `json:"total"`
	List  []*User `json:"list"`
}
