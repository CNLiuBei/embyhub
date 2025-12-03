package handler

import (
	"embyhub/internal/model"
	"embyhub/internal/service"
	"embyhub/internal/util"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

// Login 管理员登录
// @Summary 管理员登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "登录请求"
// @Success 200 {object} model.Response{data=model.LoginResponse}
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessResponse(c, resp)
}

// Logout 管理员登出
// @Summary 管理员登出
// @Tags 认证
// @Security Bearer
// @Success 200 {object} model.Response
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")

	if err := h.authService.Logout(userID.(int)); err != nil {
		util.InternalErrorResponse(c, "登出失败")
		return
	}

	util.SuccessWithMessage(c, "登出成功", nil)
}

// GetCurrentUser 获取当前登录用户信息
// @Summary 获取当前用户信息
// @Tags 认证
// @Security Bearer
// @Success 200 {object} model.Response{data=model.User}
// @Router /api/auth/current [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		util.UnauthorizedResponse(c, "用户未登录")
		return
	}

	// 从数据库获取完整的用户信息（包含角色和权限）
	user, err := h.authService.GetUserByID(userID.(int))
	if err != nil {
		util.InternalErrorResponse(c, "获取用户信息失败")
		return
	}

	util.SuccessResponse(c, user)
}

// RefreshToken 刷新Token
// @Summary 刷新Token
// @Tags 认证
// @Security Bearer
// @Produce json
// @Success 200 {object} model.Response{data=object{token=string,expires_in=int64}}
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从Header获取当前Token
	tokenString := c.GetHeader("Authorization")
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// 刷新Token
	newToken, err := util.RefreshToken(tokenString)
	if err != nil {
		util.UnauthorizedResponse(c, "Token无效或已过期")
		return
	}

	// 获取新Token的剩余时间
	expiresIn := util.GetTokenRemainingTime(newToken)

	util.SuccessResponse(c, gin.H{
		"token":      newToken,
		"expires_in": expiresIn,
	})
}

// ChangePassword 用户修改自己的密码
// @Summary 修改密码
// @Tags 认证
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body model.UserPasswordRequest true "密码请求"
// @Success 200 {object} model.Response
// @Router /api/auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		util.UnauthorizedResponse(c, "未登录")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "密码格式错误，至少6位")
		return
	}

	if err := h.authService.ChangePassword(userID.(int), req.Password); err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "密码修改成功", nil)
}
