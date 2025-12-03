package handler

import (
	"embyhub/internal/model"
	"embyhub/internal/service"
	"embyhub/internal/util"

	"github.com/gin-gonic/gin"
)

type RegisterHandler struct {
	registerService *service.RegisterService
}

func NewRegisterHandler() *RegisterHandler {
	return &RegisterHandler{
		registerService: service.NewRegisterService(),
	}
}

// Register 用户注册
// @Summary 用户注册
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "注册请求"
// @Success 200 {object} model.Response{data=model.RegisterResponse}
// @Router /api/auth/register [post]
func (h *RegisterHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.registerService.Register(&req)
	if err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessResponse(c, resp)
}
