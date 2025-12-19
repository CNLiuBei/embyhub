// Package handler HTTP处理器
package handler

import (
	"strings"
	"time"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Login 用户登录
// @Summary 用户登录
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=service.LoginResponse}
// @Router /api/v1/user/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 获取设备信息
	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()
	ctx := c.Request.Context()

	resp, err := h.userService.Login(&req, ip, userAgent)
	if err != nil {
		// 登录失败，记录失败次数
		middleware.RecordLoginFailure(ctx, req.Account, ip, 15*time.Minute)

		// 检查是否触发IP攻击检测
		if ipBlacklistService, ok := c.Get("ip_blacklist_service"); ok && ipBlacklistService != nil {
			if svc, ok := ipBlacklistService.(*service.IPBlacklistService); ok {
				if blocked, permanent := middleware.CheckIPAttack(ctx, ip, svc); blocked {
					if permanent {
						response.Error(c, 403, "检测到重复攻击行为，您的IP已被永久封禁")
					} else {
						response.Error(c, 403, "检测到异常登录行为，您的IP已被临时封禁24小时")
					}
					return
				}
			}
		}

		response.Error(c, 401, err.Error())
		return
	}

	// 登录成功，清除失败记录
	middleware.ClearLoginFailure(ctx, req.Account, ip)

	response.Success(c, resp)
}

// SendRegisterCode 发送注册验证码
// @Summary 发送注册验证码
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body service.SendRegisterCodeRequest true "发送验证码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/user/send-register-code [post]
func (h *UserHandler) SendRegisterCode(c *gin.Context) {
	var req service.SendRegisterCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的邮箱地址")
		return
	}

	if err := h.userService.SendRegisterCode(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "验证码已发送到您的邮箱", nil)
}

// Register 用户注册
// @Summary 用户注册
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "注册请求"
// @Success 200 {object} response.Response
// @Router /api/v1/user/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	user, err := h.userService.Register(&req, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "注册成功", user)
}

// Logout 退出登录
// @Summary 退出登录
// @Tags 用户
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/user/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

	h.userService.Logout(userID, token)
	response.SuccessWithMessage(c, "退出成功", nil)
}

// GetUserInfo 获取用户信息
// @Summary 获取当前用户信息
// @Tags 用户
// @Security Bearer
// @Success 200 {object} response.Response{data=models.User}
// @Router /api/v1/user/info [get]
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	user, err := h.userService.GetUserInfo(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, user)
}

// UpdateUserInfo 更新用户信息
// @Summary 更新用户信息
// @Tags 用户
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body service.UpdateUserRequest true "更新请求"
// @Success 200 {object} response.Response
// @Router /api/v1/user/update [put]
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.userService.UpdateUserInfo(userID, &req); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags 用户
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body service.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/user/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.userService.ChangePassword(userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "密码修改成功", nil)
}

// RefreshToken 刷新Token
// @Summary 刷新Token
// @Tags 用户
// @Accept json
// @Produce json
// @Param refresh_token body string true "刷新令牌"
// @Success 200 {object} response.Response{data=service.LoginResponse}
// @Router /api/v1/user/refresh-token [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	resp, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// ForgotPassword 忘记密码 - 发送验证码
// @Summary 忘记密码-发送验证码到邮箱
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body service.ForgotPasswordRequest true "忘记密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/user/forgot-password [post]
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req service.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的邮箱地址")
		return
	}

	if err := h.userService.ForgotPassword(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "验证码已发送到您的邮箱", nil)
}

// ResetPassword 重置密码 - 使用验证码重置
// @Summary 重置密码-使用验证码重置密码
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body service.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/user/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.userService.ResetPassword(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "密码重置成功", nil)
}

// UploadAvatar 上传头像
// @Summary 上传用户头像
// @Tags 用户
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "头像文件"
// @Success 200 {object} response.Response
// @Router /api/v1/user/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的图片")
		return
	}

	// 检查文件大小 (最大5MB)
	if file.Size > 5*1024*1024 {
		response.BadRequest(c, "图片大小不能超过5MB")
		return
	}

	// 检查文件类型
	contentType := file.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[contentType] {
		response.BadRequest(c, "只支持 JPG、PNG、GIF、WebP 格式")
		return
	}

	avatarURL, err := h.userService.UploadAvatar(userID, file)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"avatar": avatarURL})
}

// SendChangeEmailCode 发送修改邮箱验证码
// @Summary 发送修改邮箱验证码
// @Tags 用户
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body service.SendChangeEmailCodeRequest true "发送验证码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/user/send-change-email-code [post]
func (h *UserHandler) SendChangeEmailCode(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req service.SendChangeEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的邮箱地址")
		return
	}

	if err := h.userService.SendChangeEmailCode(userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "验证码已发送到新邮箱", nil)
}

// ChangeEmail 修改邮箱
// @Summary 修改邮箱
// @Tags 用户
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body service.ChangeEmailRequest true "修改邮箱请求"
// @Success 200 {object} response.Response
// @Router /api/v1/user/email [put]
func (h *UserHandler) ChangeEmail(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req service.ChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.userService.ChangeEmail(userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "邮箱修改成功", nil)
}
