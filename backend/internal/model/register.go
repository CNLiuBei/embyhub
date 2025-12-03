package model

// RegisterRequest 注册请求（邮箱验证方式）
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`           // 邮箱（必填）
	Code     string `json:"code" binding:"required,len=6"`            // 验证码（必填）
	Username string `json:"username" binding:"required,min=3,max=50"` // 用户名
	Password string `json:"password" binding:"required,min=6,max=50"` // 密码
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	UserID     int    `json:"user_id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	EmbyUserID string `json:"emby_user_id,omitempty"`
	Message    string `json:"message"`
}
