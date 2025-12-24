// Package handler 支付宝管理后台处理器
package handler

import (
	"net/http"

	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// AlipayAdminHandler 支付宝管理处理器
type AlipayAdminHandler struct {
	alipayService *service.AlipayService
}

// NewAlipayAdminHandler 创建支付宝管理处理器
func NewAlipayAdminHandler(alipayService *service.AlipayService) *AlipayAdminHandler {
	return &AlipayAdminHandler{
		alipayService: alipayService,
	}
}

// GetConfig 获取支付宝配置
// @Summary 获取支付宝配置
// @Tags Admin-Alipay
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/admin/alipay/config [get]
func (h *AlipayAdminHandler) GetConfig(c *gin.Context) {
	config, err := h.alipayService.GetConfig()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取配置失败")
		return
	}

	response.Success(c, config)
}

// SaveConfig 保存支付宝配置
// @Summary 保存支付宝配置
// @Tags Admin-Alipay
// @Accept json
// @Produce json
// @Param request body service.AlipaySaveConfigRequest true "配置信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/alipay/config [put]
func (h *AlipayAdminHandler) SaveConfig(c *gin.Context) {
	var req service.AlipaySaveConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.alipayService.SaveConfig(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// TestConnection 测试支付宝连接
// @Summary 测试支付宝连接
// @Tags Admin-Alipay
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/admin/alipay/test [post]
func (h *AlipayAdminHandler) TestConnection(c *gin.Context) {
	if err := h.alipayService.TestConnection(); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "连接成功"})
}

// GetLogs 获取支付日志
// @Summary 获取支付日志
// @Tags Admin-Alipay
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param order_no query string false "订单号"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/alipay/logs [get]
func (h *AlipayAdminHandler) GetLogs(c *gin.Context) {
	page := 1
	pageSize := 20
	orderNo := c.Query("order_no")

	if p := c.Query("page"); p != "" {
		if _, err := c.GetQuery("page"); err {
			page = 1
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if _, err := c.GetQuery("page_size"); err {
			pageSize = 20
		}
	}

	logs, total, err := h.alipayService.GetAPILogs(page, pageSize, orderNo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(c, gin.H{
		"list":  logs,
		"total": total,
	})
}

// =====================================================
// VIP套餐管理
// =====================================================

// GetVipPlans 获取所有VIP套餐
// @Summary 获取所有VIP套餐
// @Tags Admin-Alipay
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/admin/alipay/plans [get]
func (h *AlipayAdminHandler) GetVipPlans(c *gin.Context) {
	plans, err := h.alipayService.GetAllVipPlans()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, plans)
}

// CreateVipPlanRequest 创建套餐请求
type CreateVipPlanRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	Price        int64  `json:"price" binding:"required,min=1"`
	DurationDays int    `json:"duration_days" binding:"required,min=1"`
}

// CreateVipPlan 创建VIP套餐
// @Summary 创建VIP套餐
// @Tags Admin-Alipay
// @Accept json
// @Produce json
// @Param request body CreateVipPlanRequest true "套餐信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/alipay/plans [post]
func (h *AlipayAdminHandler) CreateVipPlan(c *gin.Context) {
	var req CreateVipPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	plan, err := h.alipayService.CreateVipPlan(req.Name, req.Price, req.DurationDays, req.Description)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}

	response.Success(c, plan)
}

// UpdateVipPlanRequest 更新套餐请求
type UpdateVipPlanRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	Price        int64  `json:"price" binding:"required,min=1"`
	DurationDays int    `json:"duration_days" binding:"required,min=1"`
	IsActive     bool   `json:"is_active"`
}

// UpdateVipPlan 更新VIP套餐
// @Summary 更新VIP套餐
// @Tags Admin-Alipay
// @Accept json
// @Produce json
// @Param id path int true "套餐ID"
// @Param request body UpdateVipPlanRequest true "套餐信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/alipay/plans/{id} [put]
func (h *AlipayAdminHandler) UpdateVipPlan(c *gin.Context) {
	idStr := c.Param("id")
	var id uint
	if _, err := c.GetQuery("id"); !err {
		// 从路径参数获取
	}
	if n, err := parseUint(idStr); err == nil {
		id = n
	} else {
		response.Error(c, http.StatusBadRequest, "无效的套餐ID")
		return
	}

	var req UpdateVipPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	plan, err := h.alipayService.UpdateVipPlan(id, req.Name, req.Price, req.DurationDays, req.Description, req.IsActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}

	response.Success(c, plan)
}

// DeleteVipPlan 删除VIP套餐
// @Summary 删除VIP套餐
// @Tags Admin-Alipay
// @Produce json
// @Param id path int true "套餐ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/alipay/plans/{id} [delete]
func (h *AlipayAdminHandler) DeleteVipPlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的套餐ID")
		return
	}

	if err := h.alipayService.DeleteVipPlan(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}

	response.Success(c, nil)
}

// ToggleVipPlanStatus 切换VIP套餐状态
// @Summary 切换VIP套餐状态
// @Tags Admin-Alipay
// @Produce json
// @Param id path int true "套餐ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/alipay/plans/{id}/toggle [post]
func (h *AlipayAdminHandler) ToggleVipPlanStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的套餐ID")
		return
	}

	plan, err := h.alipayService.ToggleVipPlanStatus(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "操作失败")
		return
	}

	response.Success(c, plan)
}

// parseUint 解析无符号整数
func parseUint(s string) (uint, error) {
	var n uint
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + uint(c-'0')
	}
	return n, nil
}
