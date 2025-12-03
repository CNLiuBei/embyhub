package handler

import (
	"embyhub/internal/model"
	"embyhub/internal/service"
	"embyhub/internal/util"

	"github.com/gin-gonic/gin"
)

// EmailHandler 邮件处理器
type EmailHandler struct {
	emailService *service.EmailService
}

// NewEmailHandler 创建邮件处理器
func NewEmailHandler() *EmailHandler {
	return &EmailHandler{
		emailService: service.NewEmailService(),
	}
}

// SendCode 发送验证码
// @Summary 发送邮箱验证码
// @Tags 邮件
// @Accept json
// @Produce json
// @Param request body model.SendCodeRequest true "请求参数"
// @Success 200 {object} model.Response
// @Router /api/email/send-code [post]
func (h *EmailHandler) SendCode(c *gin.Context) {
	var req model.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, 400, "请输入正确的邮箱地址")
		return
	}

	if err := h.emailService.SendVerificationCode(&req); err != nil {
		util.ErrorResponse(c, 400, err.Error())
		return
	}

	util.SuccessResponse(c, gin.H{"message": "验证码已发送"})
}

// SendResetCode 发送密码重置验证码
func (h *EmailHandler) SendResetCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, 400, "请输入正确的邮箱地址")
		return
	}

	if err := h.emailService.SendPasswordResetCode(req.Email); err != nil {
		util.ErrorResponse(c, 400, err.Error())
		return
	}

	util.SuccessResponse(c, gin.H{"message": "重置验证码已发送"})
}

// ResetPassword 重置密码
func (h *EmailHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Code     string `json:"code" binding:"required,len=6"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, 400, "参数错误")
		return
	}

	if err := h.emailService.ResetPassword(req.Email, req.Code, req.Password); err != nil {
		util.ErrorResponse(c, 400, err.Error())
		return
	}

	util.SuccessResponse(c, gin.H{"message": "密码重置成功"})
}

// TestConfig 测试邮件配置
// @Summary 测试邮件配置
// @Tags 邮件
// @Accept json
// @Produce json
// @Param to query string true "收件人邮箱"
// @Success 200 {object} model.Response
// @Router /api/email/test [post]
func (h *EmailHandler) TestConfig(c *gin.Context) {
	to := c.Query("to")
	if to == "" {
		util.ErrorResponse(c, 400, "请输入收件人邮箱")
		return
	}

	if err := h.emailService.TestEmailConfig(to); err != nil {
		util.ErrorResponse(c, 400, err.Error())
		return
	}

	util.SuccessResponse(c, gin.H{"message": "测试邮件发送成功"})
}
