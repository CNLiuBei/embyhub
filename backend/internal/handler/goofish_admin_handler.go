// Package handler 闲管家管理后台Handler
package handler

import (
	"net/http"
	"strconv"

	"feiniu-user-system/internal/service"

	"github.com/gin-gonic/gin"
)

// GoofishAdminHandler 闲管家管理后台Handler
type GoofishAdminHandler struct {
	service *service.GoofishService
}

// NewGoofishAdminHandler 创建闲管家管理后台Handler
func NewGoofishAdminHandler(svc *service.GoofishService) *GoofishAdminHandler {
	return &GoofishAdminHandler{service: svc}
}

// =====================================================
// 配置管理
// =====================================================

// GetConfig 获取配置
// GET /api/v1/admin/goofish/config
func (h *GoofishAdminHandler) GetConfig(c *gin.Context) {
	config, err := h.service.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配置失败"})
		return
	}

	// 添加网关地址
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	config.GatewayURL = scheme + "://" + c.Request.Host + "/goofish"

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": config})
}

// SaveConfig 保存配置
// POST /api/v1/admin/goofish/config
func (h *GoofishAdminHandler) SaveConfig(c *gin.Context) {
	var req service.SaveConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := h.service.SaveConfig(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "保存成功"})
}

// =====================================================
// 商品管理
// =====================================================

// GetGoodsList 获取商品列表
// GET /api/v1/admin/goofish/goods
func (h *GoofishAdminHandler) GetGoodsList(c *gin.Context) {
	var req service.GoodsListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	goods, total, err := h.service.GetGoodsList(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 为每个商品计算库存
	type GoodsWithStock struct {
		ID              uint   `json:"id"`
		GoodsNo         string `json:"goods_no"`
		GoodsName       string `json:"goods_name"`
		CardType        int8   `json:"card_type"`
		Price           int64  `json:"price"`
		Status          int8   `json:"status"`
		AutoGenerate    bool   `json:"auto_generate"`
		CardPrefix      string `json:"card_prefix"`
		Duration        int    `json:"duration"`
		MaxAutoGenerate int    `json:"max_auto_generate"`
		Stock           int64  `json:"stock"`
		CreatedAt       string `json:"created_at"`
		UpdatedAt       string `json:"updated_at"`
	}

	result := make([]GoodsWithStock, len(goods))
	for i, g := range goods {
		stock, _ := h.service.CalculateStock(g.CardType)
		result[i] = GoodsWithStock{
			ID:              g.ID,
			GoodsNo:         g.GoodsNo,
			GoodsName:       g.GoodsName,
			CardType:        g.CardType,
			Price:           g.Price,
			Status:          g.Status,
			AutoGenerate:    g.AutoGenerate,
			CardPrefix:      g.CardPrefix,
			Duration:        g.Duration,
			MaxAutoGenerate: g.MaxAutoGenerate,
			Stock:           stock,
			CreatedAt:       g.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       g.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total": total,
			"list":  result,
		},
	})
}

// CreateGoods 创建商品
// POST /api/v1/admin/goofish/goods
func (h *GoofishAdminHandler) CreateGoods(c *gin.Context) {
	var req service.CreateGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := h.service.CreateGoods(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功"})
}

// UpdateGoods 更新商品
// PUT /api/v1/admin/goofish/goods/:id
func (h *GoofishAdminHandler) UpdateGoods(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	var req service.UpdateGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := h.service.UpdateGoods(uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteGoods 删除商品
// DELETE /api/v1/admin/goofish/goods/:id
func (h *GoofishAdminHandler) DeleteGoods(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	if err := h.service.DeleteGoods(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// AutoGenerateGoods 自动生成商品映射
// POST /api/v1/admin/goofish/goods/auto-generate
func (h *GoofishAdminHandler) AutoGenerateGoods(c *gin.Context) {
	resp, err := h.service.AutoGenerateGoods()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "自动生成完成",
		"data":    resp,
	})
}

// =====================================================
// 订单管理
// =====================================================

// GetOrderList 获取订单列表
// GET /api/v1/admin/goofish/orders
func (h *GoofishAdminHandler) GetOrderList(c *gin.Context) {
	var req service.OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	orders, total, err := h.service.GetOrderList(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total": total,
			"list":  orders,
		},
	})
}

// GetOrderDetailAdmin 获取订单详情（管理后台）
// GET /api/v1/admin/goofish/orders/:order_no
func (h *GoofishAdminHandler) GetOrderDetailAdmin(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "订单号不能为空"})
		return
	}

	order, orderCards, err := h.service.GetOrderDetail(orderNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"order": order,
			"cards": orderCards,
		},
	})
}

// =====================================================
// 日志管理
// =====================================================

// GetAPILogs 获取API日志
// GET /api/v1/admin/goofish/logs
func (h *GoofishAdminHandler) GetAPILogs(c *gin.Context) {
	var req service.LogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	logs, total, err := h.service.GetAPILogs(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total": total,
			"list":  logs,
		},
	})
}

// CleanOldLogs 清理旧日志
// POST /api/v1/admin/goofish/logs/clean
func (h *GoofishAdminHandler) CleanOldLogs(c *gin.Context) {
	count, err := h.service.CleanOldLogs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "清理失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "清理成功",
		"data": gin.H{
			"count": count,
		},
	})
}

// =====================================================
// 商品变更通知
// =====================================================

// NotifyGoodsChange 通知单个商品变更
// POST /api/v1/admin/goofish/goods/:goods_no/notify
func (h *GoofishAdminHandler) NotifyGoodsChange(c *gin.Context) {
	goodsNo := c.Param("goods_no")
	if goodsNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "商品编码不能为空"})
		return
	}

	if err := h.service.NotifyGoodsChange(goodsNo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "通知已发送"})
}

// NotifyAllGoodsChange 通知所有商品变更
// POST /api/v1/admin/goofish/goods/notify-all
func (h *GoofishAdminHandler) NotifyAllGoodsChange(c *gin.Context) {
	if err := h.service.NotifyAllGoodsChange(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "通知已发送"})
}
