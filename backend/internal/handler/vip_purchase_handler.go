// Package handler VIP购买控制器
package handler

import (
	"net/http"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// VipPurchaseHandler VIP购买处理器
type VipPurchaseHandler struct {
	vipService *service.VipPurchaseService
}

// NewVipPurchaseHandler 创建VIP购买处理器
func NewVipPurchaseHandler(vipService *service.VipPurchaseService) *VipPurchaseHandler {
	return &VipPurchaseHandler{
		vipService: vipService,
	}
}

// PurchaseVip 购买VIP会员
// @Summary 购买VIP会员
// @Description 使用账户余额购买VIP会员套餐
// @Tags VIP
// @Accept json
// @Produce json
// @Param request body service.PurchaseVipRequest true "购买请求"
// @Success 200 {object} response.Response{data=service.PurchaseVipResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/vip/purchase [post]
func (h *VipPurchaseHandler) PurchaseVip(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未登录或登录已过期")
		return
	}

	// 2. 解析请求参数
	var req service.PurchaseVipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	req.UserID = userID

	// 3. 调用业务逻辑
	result, err := h.vipService.PurchaseVip(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 4. 返回成功响应
	response.SuccessWithMessage(c, "购买成功", result)
}

// GetVipPlans 获取VIP套餐列表
// @Summary 获取VIP套餐列表
// @Description 获取所有可用的VIP套餐
// @Tags VIP
// @Produce json
// @Success 200 {object} response.Response{data=[]models.VipPlan}
// @Router /api/v1/vip/plans [get]
func (h *VipPurchaseHandler) GetVipPlans(c *gin.Context) {
	plans, err := h.vipService.GetVipPlans()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询套餐列表失败")
		return
	}

	response.Success(c, plans)
}

// GetUserVipInfo 获取用户VIP信息
// @Summary 获取用户VIP信息
// @Description 获取当前用户的VIP会员状态
// @Tags VIP
// @Produce json
// @Success 200 {object} response.Response{data=models.UserVip}
// @Router /api/v1/vip/info [get]
func (h *VipPurchaseHandler) GetUserVipInfo(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "未登录或登录已过期")
		return
	}

	userVip, err := h.vipService.GetUserVipInfo(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询VIP信息失败")
		return
	}

	// 返回响应，包含会员状态
	data := gin.H{
		"user_id":       userVip.UserID,
		"vip_expire_at": userVip.VipExpireAt,
		"is_vip":        userVip.IsVipValid(),
	}

	response.Success(c, data)
}
