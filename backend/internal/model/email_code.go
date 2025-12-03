package model

import "time"

// EmailCode 邮箱验证码
type EmailCode struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"column:email;type:varchar(100);not null" json:"email"`
	Code      string    `gorm:"column:code;type:varchar(6);not null" json:"code"`
	Type      string    `gorm:"column:type;type:varchar(20);default:register" json:"type"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null" json:"expires_at"`
	Used      bool      `gorm:"column:used;default:false" json:"used"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (EmailCode) TableName() string {
	return "email_codes"
}

// 验证码类型
const (
	CodeTypeRegister      = "register"
	CodeTypeResetPassword = "reset_password"
)

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"omitempty,oneof=register reset_password"`
}

// VerifyCodeRequest 验证码验证请求
type VerifyCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// RegisterWithEmailRequest 邮箱注册请求
type RegisterWithEmailRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}
