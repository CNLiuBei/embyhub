// Package handler 支付相关HTTP处理器
package handler

import (
	"net/http"

	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	alipayService *service.AlipayService
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(alipayService *service.AlipayService) *PaymentHandler {
	return &PaymentHandler{
		alipayService: alipayService,
	}
}

// CreateAlipayPaymentRequest 创建支付请求
type CreateAlipayPaymentRequest struct {
	PlanID uint `json:"plan_id" binding:"required"`
}

// CreateAlipayPayment 创建支付宝支付订单
// @Summary 创建支付宝支付订单
// @Tags Payment
// @Accept json
// @Produce json
// @Param request body CreateAlipayPaymentRequest true "支付请求"
// @Success 200 {object} response.Response
// @Router /api/v1/payment/alipay/create [post]
func (h *PaymentHandler) CreateAlipayPayment(c *gin.Context) {
	var req CreateAlipayPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	// 从上下文获取用户ID
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "请先登录")
		return
	}

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "用户信息无效")
		return
	}

	result, err := h.alipayService.CreatePayment(userID, req.PlanID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, result)
}

// AlipayNotify 支付宝异步通知
// @Summary 支付宝异步通知回调
// @Tags Payment
// @Accept x-www-form-urlencoded
// @Produce text/plain
// @Success 200 {string} string "success"
// @Router /api/v1/payment/alipay/notify [post]
func (h *PaymentHandler) AlipayNotify(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	if err := h.alipayService.HandleNotify(c.Request.Form); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

// GetOrderStatus 查询订单状态
// @Summary 查询订单状态
// @Tags Payment
// @Produce json
// @Param order_no path string true "订单号"
// @Success 200 {object} response.Response
// @Router /api/v1/payment/order/{order_no} [get]
func (h *PaymentHandler) GetOrderStatus(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		response.Error(c, http.StatusBadRequest, "订单号不能为空")
		return
	}

	// 从上下文获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "请先登录")
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "用户信息无效")
		return
	}

	result, err := h.alipayService.QueryOrderStatus(orderNo, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, result)
}

// GetOrderList 获取订单列表
// @Summary 获取用户订单列表
// @Tags Payment
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param status query string false "订单状态"
// @Success 200 {object} response.Response
// @Router /api/v1/payment/orders [get]
func (h *PaymentHandler) GetOrderList(c *gin.Context) {
	// 从上下文获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "请先登录")
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "用户信息无效")
		return
	}

	var req service.AlipayOrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	result, err := h.alipayService.GetOrderList(userID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(c, result)
}

// GetVipPlans 获取VIP套餐列表
// @Summary 获取VIP套餐列表
// @Tags Payment
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/payment/plans [get]
func (h *PaymentHandler) GetVipPlans(c *gin.Context) {
	plans, err := h.alipayService.GetVipPlans()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(c, plans)
}

// GetMemberChangeLogs 获取会员变动记录
// @Summary 获取用户会员变动记录
// @Tags Payment
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/v1/payment/member-logs [get]
func (h *PaymentHandler) GetMemberChangeLogs(c *gin.Context) {
	// 从上下文获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "请先登录")
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "用户信息无效")
		return
	}

	result, err := h.alipayService.GetMemberChangeLogs(userID, c)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(c, result)
}
