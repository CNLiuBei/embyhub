package handler

import (
	"embyhub/internal/model"
	"embyhub/internal/service"
	"embyhub/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(),
	}
}

// Create 创建用户
// @Summary 创建用户
// @Tags 用户管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body model.UserCreateRequest true "创建用户请求"
// @Success 200 {object} model.Response{data=model.User}
// @Router /api/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req model.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	user, err := h.userService.Create(&req)
	if err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "创建用户成功", user)
}

// GetByID 获取用户详情
// @Summary 获取用户详情
// @Tags 用户管理
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} model.Response{data=model.User}
// @Router /api/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "用户ID格式错误")
		return
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		util.NotFoundResponse(c, "用户不存在")
		return
	}

	util.SuccessResponse(c, user)
}

// Update 更新用户
// @Summary 更新用户
// @Tags 用户管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body model.UserUpdateRequest true "更新用户请求"
// @Success 200 {object} model.Response{data=model.User}
// @Router /api/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "用户ID格式错误")
		return
	}

	var req model.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	user, err := h.userService.Update(id, &req)
	if err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "更新用户成功", user)
}

// Delete 删除用户
// @Summary 删除用户
// @Tags 用户管理
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} model.Response
// @Router /api/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "用户ID格式错误")
		return
	}

	if err := h.userService.Delete(id); err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "删除用户成功", nil)
}

// List 获取用户列表
// @Summary 获取用户列表
// @Tags 用户管理
// @Security Bearer
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param keyword query string false "关键词"
// @Param status query int false "状态"
// @Param role_id query int false "角色ID"
// @Success 200 {object} model.Response{data=model.UserListResponse}
// @Router /api/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var req model.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.userService.List(&req)
	if err != nil {
		util.InternalErrorResponse(c, "获取用户列表失败")
		return
	}

	util.SuccessResponse(c, resp)
}

// ResetPassword 重置用户密码
// @Summary 重置用户密码
// @Tags 用户管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body model.UserPasswordRequest true "密码请求"
// @Success 200 {object} model.Response
// @Router /api/users/{id}/password [put]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "用户ID格式错误")
		return
	}

	var req model.UserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.userService.ResetPassword(id, req.Password); err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "重置密码成功", nil)
}

// BatchUpdateStatus 批量更新用户状态
// @Summary 批量更新用户状态
// @Tags 用户管理
// @Security Bearer
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Router /api/users/batch/status [put]
func (h *UserHandler) BatchUpdateStatus(c *gin.Context) {
	var req struct {
		UserIDs []int `json:"user_ids" binding:"required"`
		Status  int   `json:"status" binding:"required,oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.userService.BatchUpdateStatus(req.UserIDs, req.Status); err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "批量更新成功", nil)
}

// SetVip 设置用户VIP
// @Summary 设置用户VIP
// @Tags 用户管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body object true "VIP设置"
// @Success 200 {object} model.Response
// @Router /api/users/{id}/vip [put]
func (h *UserHandler) SetVip(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "用户ID格式错误")
		return
	}

	var req struct {
		Days int `json:"days" binding:"required,min=1,max=3650"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请输入有效的天数（1-3650）")
		return
	}

	user, err := h.userService.SetVip(id, req.Days)
	if err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "VIP设置成功", map[string]interface{}{
		"vip_level":     user.VipLevel,
		"vip_expire_at": user.VipExpireAt,
	})
}
