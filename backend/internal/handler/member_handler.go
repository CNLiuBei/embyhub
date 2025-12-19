// Package handler 会员处理器
package handler

import (
	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// MemberHandler 会员处理器
type MemberHandler struct {
	memberService *service.MemberService
}

// NewMemberHandler 创建会员处理器
func NewMemberHandler(memberService *service.MemberService) *MemberHandler {
	return &MemberHandler{memberService: memberService}
}

// GetMemberInfo 获取会员信息
// @Summary 获取当前用户会员信息
// @Tags 会员
// @Security Bearer
// @Success 200 {object} response.Response{data=service.MemberInfo}
// @Router /api/v1/member/info [get]
func (h *MemberHandler) GetMemberInfo(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	info, err := h.memberService.GetMemberInfo(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, info)
}

// GetMemberOrders 获取会员订单(卡密兑换记录)
// @Summary 获取会员订单列表
// @Tags 会员
// @Security Bearer
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Success 200 {object} response.Response{data=response.PageData}
// @Router /api/v1/member/orders [get]
func (h *MemberHandler) GetMemberOrders(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	page := 1
	pageSize := 20
	c.ShouldBindQuery(&struct {
		Page     *int `form:"page"`
		PageSize *int `form:"page_size"`
	}{&page, &pageSize})

	orders, total, err := h.memberService.GetMemberOrders(userID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, orders, total, page, pageSize)
}
