// Package handler 会员支付处理器
package handler

import (
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// MemberPaymentHandler 会员支付处理器
type MemberPaymentHandler struct {
	service *service.MemberPaymentService
}

// NewMemberPaymentHandler 创建会员支付处理器
func NewMemberPaymentHandler(service *service.MemberPaymentService) *MemberPaymentHandler {
	return &MemberPaymentHandler{
		service: service,
	}
}

// GetMemberPackages 获取会员套餐列表
func (h *MemberPaymentHandler) GetMemberPackages(c *gin.Context) {
	packages := h.service.GetMemberPackages()
	response.Success(c, packages)
}

// PurchaseMemberWithBalance 使用余额购买会员（已迁移至VIP购买功能）
func (h *MemberPaymentHandler) PurchaseMemberWithBalance(c *gin.Context) {
	response.BadRequest(c, "此功能已迁移，请使用VIP购买功能（/api/v1/vip/purchase）")
}
