package handler

import (
	"embyhub/internal/service"
	"embyhub/internal/util"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permissionService *service.PermissionService
}

func NewPermissionHandler() *PermissionHandler {
	return &PermissionHandler{
		permissionService: service.NewPermissionService(),
	}
}

// List 获取权限列表
func (h *PermissionHandler) List(c *gin.Context) {
	resp, err := h.permissionService.List()
	if err != nil {
		util.InternalErrorResponse(c, "获取权限列表失败")
		return
	}

	util.SuccessResponse(c, resp)
}
