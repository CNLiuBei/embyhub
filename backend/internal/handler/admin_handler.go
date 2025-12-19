// Package handler 管理员处理器
package handler

import (
	"fmt"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	adminService  *service.AdminService
	memberService *service.MemberService
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(adminService *service.AdminService, memberService *service.MemberService) *AdminHandler {
	return &AdminHandler{adminService: adminService, memberService: memberService}
}

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Tags 管理员
// @Security Bearer
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Param phone query string false "手机号"
// @Param nickname query string false "昵称"
// @Param member_level query int false "会员等级"
// @Param status query int false "状态"
// @Success 200 {object} response.Response{data=response.PageData}
// @Router /api/v1/admin/user/list [get]
func (h *AdminHandler) GetUserList(c *gin.Context) {
	var req service.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, total, err := h.adminService.GetUserList(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, list, total, req.Page, req.PageSize)
}

// GetUserDetail 获取用户详情
// @Summary 获取用户详情
// @Tags 管理员
// @Security Bearer
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response{data=models.User}
// @Router /api/v1/admin/user/{id} [get]
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	user, err := h.adminService.GetUserDetail(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, user)
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态(启用/禁用)
// @Tags 管理员
// @Security Bearer
// @Accept json
// @Param id path string true "用户ID"
// @Param status body int true "状态 1启用 2禁用"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user/{id}/status [put]
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req struct {
		Status int8 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.adminService.UpdateUserStatus(adminID, userID, req.Status, "管理员"); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "操作成功", nil)
}

// BatchUpdateStatus 批量更新状态
// @Summary 批量更新用户状态
// @Tags 管理员
// @Security Bearer
// @Accept json
// @Param request body object true "批量操作请求"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user/batch-status [put]
func (h *AdminHandler) BatchUpdateStatus(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)

	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
		Status  int8     `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	userIDs := make([]uuid.UUID, 0, len(req.UserIDs))
	for _, idStr := range req.UserIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		userIDs = append(userIDs, id)
	}

	if err := h.adminService.BatchUpdateStatus(adminID, userIDs, req.Status, "管理员"); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "操作成功", nil)
}

// UpdateUserRole 更新用户角色
// @Summary 更新用户角色
// @Tags 管理员
// @Security Bearer
// @Param id path string true "用户ID"
// @Param role body int true "角色"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user/{id}/role [put]
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req struct {
		Role int8 `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.adminService.UpdateUserRole(adminID, userID, req.Role, "管理员"); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "操作成功", nil)
}

// ResetPassword 重置密码
// @Summary 重置用户密码
// @Tags 管理员
// @Security Bearer
// @Param id path string true "用户ID"
// @Param password body string true "新密码"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user/{id}/reset-password [put]
func (h *AdminHandler) ResetPassword(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "密码不能少于8位")
		return
	}

	if err := h.adminService.ResetPassword(adminID, userID, req.Password, "管理员"); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "密码重置成功", nil)
}

// DeleteUser 删除用户
// @Summary 删除用户（同步删除飞牛影视账号）
// @Tags 管理员
// @Security Bearer
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	if err := h.adminService.DeleteUser(adminID, userID, "管理员"); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "用户已删除", nil)
}

// GetLoginLogs 获取登录日志
// @Summary 获取用户登录日志
// @Tags 管理员
// @Security Bearer
// @Param id path string true "用户ID"
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Success 200 {object} response.Response{data=response.PageData}
// @Router /api/v1/admin/user/{id}/login-logs [get]
func (h *AdminHandler) GetLoginLogs(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	page := 1
	pageSize := 20
	c.ShouldBindQuery(&struct {
		Page     *int `form:"page"`
		PageSize *int `form:"page_size"`
	}{&page, &pageSize})

	logs, total, err := h.adminService.GetLoginLogs(userID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, logs, total, page, pageSize)
}

// GetOperationLogs 获取操作日志
// @Summary 获取操作日志
// @Tags 管理员
// @Security Bearer
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Success 200 {object} response.Response{data=response.PageData}
// @Router /api/v1/admin/operation-logs [get]
func (h *AdminHandler) GetOperationLogs(c *gin.Context) {
	page := 1
	pageSize := 20
	c.ShouldBindQuery(&struct {
		Page     *int `form:"page"`
		PageSize *int `form:"page_size"`
	}{&page, &pageSize})

	logs, total, err := h.adminService.GetOperationLogs(page, pageSize, nil)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, logs, total, page, pageSize)
}

// GetUserStats 获取用户统计
// @Summary 获取用户统计数据
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response{data=service.UserStats}
// @Router /api/v1/admin/stat/user [get]
func (h *AdminHandler) GetUserStats(c *gin.Context) {
	stats, err := h.adminService.GetUserStats()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// GetDailyStats 获取每日统计
// @Summary 获取每日统计数据
// @Tags 管理员
// @Security Bearer
// @Param days query int false "天数(默认30)"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/stat/daily [get]
func (h *AdminHandler) GetDailyStats(c *gin.Context) {
	days := 30
	c.ShouldBindQuery(&struct {
		Days *int `form:"days"`
	}{&days})

	stats, err := h.adminService.GetDailyStats(days)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// SetMember 设置会员
// @Summary 管理员设置用户会员
// @Tags 管理员
// @Security Bearer
// @Param id path string true "用户ID"
// @Param request body object true "会员设置"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user/{id}/member [put]
func (h *AdminHandler) SetMember(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req struct {
		Days int `json:"days" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.memberService.AdminSetMember(userID, req.Days); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "升级会员成功", nil)
}

// BatchSetMember 批量设置会员
// @Summary 批量设置用户会员（续费）
// @Tags 管理员
// @Security Bearer
// @Accept json
// @Param request body object true "批量续费请求"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/user/batch-member [put]
func (h *AdminHandler) BatchSetMember(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
		Days    int      `json:"days" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要用户ID列表和续费天数")
		return
	}

	userIDs := make([]uuid.UUID, 0, len(req.UserIDs))
	for _, idStr := range req.UserIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		userIDs = append(userIDs, id)
	}

	if len(userIDs) == 0 {
		response.BadRequest(c, "请选择要续费的用户")
		return
	}

	success, failed, errors := h.memberService.BatchSetMember(userIDs, req.Days)

	response.Success(c, gin.H{
		"total":   len(userIDs),
		"success": success,
		"failed":  failed,
		"errors":  errors,
	})
}

// GetEmbyUsers 获取Emby用户列表
// @Summary 获取Emby媒体服务用户列表
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/users [get]
func (h *AdminHandler) GetEmbyUsers(c *gin.Context) {
	users, err := h.adminService.GetEmbyUsers()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, users)
}

// GetEmbyUserByUsername 获取单个Emby用户详情
// @Summary 根据用户名获取Emby用户详情（包含完整权限信息）
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/user/{username} [get]
func (h *AdminHandler) GetEmbyUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "请提供用户名")
		return
	}

	embyUser, err := h.adminService.GetEmbyUserByUsername(username)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, embyUser)
}

// ========== 用户同步接口 ==========

// SyncUserToEmby 同步单个用户到Emby
// @Summary 同步本地用户到Emby（创建Emby账号）
// @Tags 管理员
// @Security Bearer
// @Param id path string true "用户ID"
// @Param password body string false "Emby账号密码（可选，默认使用用户名）"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/sync/user/{id}/to-emby [post]
func (h *AdminHandler) SyncUserToEmby(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req struct {
		Password string `json:"password"` // 可选密码
	}
	c.ShouldBindJSON(&req)

	result, err := h.adminService.SyncUserToEmby(userID, req.Password)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// SyncUserPassword 同步用户密码到Emby
// @Summary 同步用户密码到Emby
// @Tags 管理员
// @Security Bearer
// @Param id path string true "用户ID"
// @Param password body string true "新密码"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/sync/user/{id}/password [post]
func (h *AdminHandler) SyncUserPassword(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "密码不能少于6位")
		return
	}

	result, err := h.adminService.SyncUserPassword(userID, req.Password)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// ImportEmbyUser 导入Emby用户到本地
// @Summary 导入Emby用户到本地数据库
// @Tags 管理员
// @Security Bearer
// @Param username body string true "Emby用户名"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/sync/import-emby-user [post]
func (h *AdminHandler) ImportEmbyUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供Emby用户名")
		return
	}

	result, err := h.adminService.ImportEmbyUser(req.Username)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// SyncUserStatus 同步用户状态
// @Summary 同步用户状态到Emby（修复状态不一致）
// @Tags 管理员
// @Security Bearer
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/sync/user/{id}/status [post]
func (h *AdminHandler) SyncUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	result, err := h.adminService.SyncUserStatus(userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// SyncAllUsers 批量同步所有用户
// @Summary 批量同步所有本地用户到Emby
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/sync/all [post]
func (h *AdminHandler) SyncAllUsers(c *gin.Context) {
	result, err := h.adminService.SyncAllUsersToEmby()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// ImportAllEmbyUsers 导入所有Emby用户
// @Summary 导入所有仅存在于Emby的用户到本地
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/sync/import-all [post]
func (h *AdminHandler) ImportAllEmbyUsers(c *gin.Context) {
	result, err := h.adminService.ImportAllEmbyUsers()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// ========== Emby设备和会话管理接口 ==========

// GetEmbySessions 获取所有Emby活动会话
// @Summary 获取所有Emby活动会话
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/sessions [get]
func (h *AdminHandler) GetEmbySessions(c *gin.Context) {
	sessions, err := h.adminService.GetEmbySessions()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, sessions)
}

// GetEmbySessionsByUsername 获取指定用户的Emby活动会话
// @Summary 获取指定用户的Emby活动会话
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/sessions/{username} [get]
func (h *AdminHandler) GetEmbySessionsByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "请提供用户名")
		return
	}

	sessions, err := h.adminService.GetEmbySessionsByUsername(username)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, sessions)
}

// KillEmbySession 终止Emby会话
// @Summary 终止Emby会话（强制下线）
// @Tags 管理员
// @Security Bearer
// @Param session_id path string true "会话ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/sessions/{session_id}/kill [post]
func (h *AdminHandler) KillEmbySession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.BadRequest(c, "请提供会话ID")
		return
	}

	if err := h.adminService.KillEmbySession(sessionID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "会话已终止", nil)
}

// EnforceSessionLimit 强制执行会话限制
// @Summary 强制执行会话限制，终止超限用户的最早会话
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/enforce-session-limit [post]
func (h *AdminHandler) EnforceSessionLimit(c *gin.Context) {
	killedCount, err := h.adminService.CheckAndEnforceSessionLimit()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, fmt.Sprintf("已终止 %d 个超限会话", killedCount), map[string]int{
		"killed_count": killedCount,
	})
}

// GetSessionLimitStatus 获取会话限制状态
// @Summary 获取会话限制状态
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/session-limit-status [get]
func (h *AdminHandler) GetSessionLimitStatus(c *gin.Context) {
	status, err := h.adminService.GetSessionLimitStatus()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, status)
}

// EnforcePlayLimit 强制执行播放数量限制
// @Summary 强制执行播放数量限制，停止超限用户的最早播放
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/enforce-play-limit [post]
func (h *AdminHandler) EnforcePlayLimit(c *gin.Context) {
	stoppedCount, err := h.adminService.CheckAndEnforcePlayLimit()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, fmt.Sprintf("已停止 %d 个超限播放", stoppedCount), map[string]int{
		"stopped_count": stoppedCount,
	})
}

// GetEmbyDevices 获取所有Emby注册设备
// @Summary 获取所有Emby注册设备
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/devices [get]
func (h *AdminHandler) GetEmbyDevices(c *gin.Context) {
	devices, err := h.adminService.GetEmbyDevices()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, devices)
}

// GetEmbyDevicesByUsername 获取指定用户的Emby设备
// @Summary 获取指定用户的Emby设备
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/devices/{username} [get]
func (h *AdminHandler) GetEmbyDevicesByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "请提供用户名")
		return
	}

	devices, err := h.adminService.GetEmbyDevicesByUsername(username)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, devices)
}

// DeleteEmbyDevice 删除Emby设备
// @Summary 删除Emby设备
// @Tags 管理员
// @Security Bearer
// @Param device_id path string true "设备ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/devices/{device_id} [delete]
func (h *AdminHandler) DeleteEmbyDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	if deviceID == "" {
		response.BadRequest(c, "请提供设备ID")
		return
	}

	if err := h.adminService.DeleteEmbyDevice(deviceID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "设备已删除", nil)
}

// SetEmbyUserStreamLimit 设置Emby用户同时在线流数限制
// @Summary 设置Emby用户同时在线流数限制
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Param limit body int true "流数限制（0表示无限制）"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/user/{username}/stream-limit [put]
func (h *AdminHandler) SetEmbyUserStreamLimit(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "请提供用户名")
		return
	}

	var req struct {
		Limit int `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.Limit < 0 {
		response.BadRequest(c, "流数限制不能为负数")
		return
	}

	if err := h.adminService.SetEmbyUserStreamLimit(username, req.Limit); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	msg := "已设置同时在线流数限制"
	if req.Limit == 0 {
		msg = "已取消同时在线流数限制"
	}
	response.SuccessWithMessage(c, msg, nil)
}

// GetEmbyUserStreamLimit 获取Emby用户同时在线流数限制
// @Summary 获取Emby用户同时在线流数限制
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/user/{username}/stream-limit [get]
func (h *AdminHandler) GetEmbyUserStreamLimit(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "请提供用户名")
		return
	}

	limit, err := h.adminService.GetEmbyUserStreamLimit(username)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"username": username,
		"limit":    limit,
	})
}

// EnforceClientWhitelist 强制执行客户端白名单
// @Summary 强制执行客户端白名单（踢出所有未授权客户端）
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/enforce-client-whitelist [post]
func (h *AdminHandler) EnforceClientWhitelist(c *gin.Context) {
	result, err := h.adminService.EnforceClientWhitelist()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// ========== 用户设备策略管理接口（使用Emby的EnableAllDevices/EnabledDevices） ==========

// GetEmbyUserDevicePolicy 获取用户设备策略
// @Summary 获取用户设备策略（EnableAllDevices和EnabledDevices）
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/user/{username}/device-policy [get]
func (h *AdminHandler) GetEmbyUserDevicePolicy(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "请提供用户名")
		return
	}

	policy, err := h.adminService.GetEmbyUserDevicePolicy(username)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, policy)
}

// SetEmbyUserDevicePolicy 设置用户设备策略
// @Summary 设置用户设备策略（EnableAllDevices和EnabledDevices）
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Param request body object true "设备策略"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/user/{username}/device-policy [put]
func (h *AdminHandler) SetEmbyUserDevicePolicy(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "请提供用户名")
		return
	}

	var req struct {
		EnableAllDevices bool     `json:"enable_all_devices"`
		EnabledDevices   []string `json:"enabled_devices"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.adminService.SetEmbyUserDevicePolicy(username, req.EnableAllDevices, req.EnabledDevices); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	msg := "设备策略已更新"
	if req.EnableAllDevices {
		msg = "已允许所有设备"
	} else {
		msg = "已限制为指定设备"
	}
	response.SuccessWithMessage(c, msg, nil)
}

// AddDeviceToEmbyUserWhitelist 添加设备到用户白名单
// @Summary 添加设备到用户白名单
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Param device_id body string true "设备ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/user/{username}/device-whitelist [post]
func (h *AdminHandler) AddDeviceToEmbyUserWhitelist(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "请提供用户名")
		return
	}

	var req struct {
		DeviceID string `json:"device_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供设备ID")
		return
	}

	if err := h.adminService.AddDeviceToEmbyUserWhitelist(username, req.DeviceID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "设备已添加到白名单", nil)
}

// RemoveDeviceFromEmbyUserWhitelist 从用户白名单移除设备
// @Summary 从用户白名单移除设备
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Param device_id path string true "设备ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/user/{username}/device-whitelist/{device_id} [delete]
func (h *AdminHandler) RemoveDeviceFromEmbyUserWhitelist(c *gin.Context) {
	username := c.Param("username")
	deviceID := c.Param("device_id")
	if username == "" || deviceID == "" {
		response.BadRequest(c, "请提供用户名和设备ID")
		return
	}

	if err := h.adminService.RemoveDeviceFromEmbyUserWhitelist(username, deviceID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "设备已从白名单移除", nil)
}

// ApplyGlobalDeviceWhitelistToUser 将全局客户端白名单应用到用户
// @Summary 将全局客户端白名单应用到指定用户
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/user/{username}/apply-global-whitelist [post]
func (h *AdminHandler) ApplyGlobalDeviceWhitelistToUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "请提供用户名")
		return
	}

	if err := h.adminService.ApplyGlobalDeviceWhitelistToUser(username); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "全局设备白名单已应用到用户", nil)
}

// ApplyGlobalDeviceWhitelistToAllUsers 将全局客户端白名单应用到所有用户
// @Summary 将全局客户端白名单应用到所有用户
// @Tags 管理员
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/apply-global-whitelist-all [post]
func (h *AdminHandler) ApplyGlobalDeviceWhitelistToAllUsers(c *gin.Context) {
	result, err := h.adminService.ApplyGlobalDeviceWhitelistToAllUsers()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// SetEmbyUserClientPolicy 按客户端名称设置用户设备策略
// @Summary 按客户端名称设置用户设备策略
// @Tags 管理员
// @Security Bearer
// @Param username path string true "用户名"
// @Param request body object true "客户端策略"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/emby/user/{username}/client-policy [put]
func (h *AdminHandler) SetEmbyUserClientPolicy(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "请提供用户名")
		return
	}

	var req struct {
		EnableAllDevices bool     `json:"enable_all_devices"`
		EnabledClients   []string `json:"enabled_clients"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.adminService.SetEmbyUserClientPolicy(username, req.EnableAllDevices, req.EnabledClients); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	msg := "客户端策略已更新"
	if req.EnableAllDevices {
		msg = "已允许所有设备"
	} else if len(req.EnabledClients) > 0 {
		msg = "已限制为指定客户端"
	}
	response.SuccessWithMessage(c, msg, nil)
}
