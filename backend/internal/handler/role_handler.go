package handler

import (
	"embyhub/internal/model"
	"embyhub/internal/service"
	"embyhub/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleService *service.RoleService
}

func NewRoleHandler() *RoleHandler {
	return &RoleHandler{
		roleService: service.NewRoleService(),
	}
}

// Create 创建角色
func (h *RoleHandler) Create(c *gin.Context) {
	var req model.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	role, err := h.roleService.Create(&req)
	if err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "创建角色成功", role)
}

// GetByID 获取角色详情
func (h *RoleHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "角色ID格式错误")
		return
	}

	role, err := h.roleService.GetByID(id)
	if err != nil {
		util.NotFoundResponse(c, "角色不存在")
		return
	}

	util.SuccessResponse(c, role)
}

// Update 更新角色
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "角色ID格式错误")
		return
	}

	var req model.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	role, err := h.roleService.Update(id, &req)
	if err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "更新角色成功", role)
}

// Delete 删除角色
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "角色ID格式错误")
		return
	}

	if err := h.roleService.Delete(id); err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "删除角色成功", nil)
}

// List 获取角色列表
func (h *RoleHandler) List(c *gin.Context) {
	resp, err := h.roleService.List()
	if err != nil {
		util.InternalErrorResponse(c, "获取角色列表失败")
		return
	}

	util.SuccessResponse(c, resp)
}

// AssignPermissions 为角色分配权限
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "角色ID格式错误")
		return
	}

	var req model.RolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.roleService.AssignPermissions(id, req.PermissionIDs); err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "分配权限成功", nil)
}
