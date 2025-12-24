// Package handler 闲管家虚拟货源对接Handler
package handler

import (
	"encoding/json"
	"net/http"

	"feiniu-user-system/internal/service"

	"github.com/gin-gonic/gin"
)

// GoofishHandler 闲管家Handler
type GoofishHandler struct {
	service *service.GoofishService
}

// NewGoofishHandler 创建闲管家Handler
func NewGoofishHandler(svc *service.GoofishService) *GoofishHandler {
	return &GoofishHandler{service: svc}
}

// 统一响应格式
type goofishResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func goofishSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, goofishResponse{
		Code: 0,
		Msg:  "OK",
		Data: data,
	})
}

func goofishError(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, goofishResponse{
		Code: code,
		Msg:  msg,
	})
}

// =====================================================
// 基础接口
// =====================================================

// GetPlatformInfo 查询平台信息
// POST /goofish/open/info
func (h *GoofishHandler) GetPlatformInfo(c *gin.Context) {
	config, _, _, err := h.service.GetDecryptedConfig()
	if err != nil {
		goofishError(c, 500, "获取配置失败")
		return
	}

	// 只返回 app_id（整数类型）
	goofishSuccess(c, gin.H{
		"app_id": config.AppID,
	})
}

// GetMerchantInfo 查询商户信息
// POST /goofish/user/info
func (h *GoofishHandler) GetMerchantInfo(c *gin.Context) {
	// 自研系统固定返回大于0的余额即可
	goofishSuccess(c, gin.H{
		"balance": int64(9999999), // 商户余额（分），固定返回大于0的值
	})
}

// maskSecret 脱敏显示密钥
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// =====================================================
// 商品接口
// =====================================================

// GoodsListRequest 商品列表请求
type GoodsListRequest struct {
	Keyword   string `json:"keyword"`
	GoodsType int    `json:"goods_type"`
	PageNo    int    `json:"page_no"`
	PageSize  int    `json:"page_size"`
}

// GetGoodsList 查询商品列表
// POST /goofish/goods/list
func (h *GoofishHandler) GetGoodsList(c *gin.Context) {
	var req GoodsListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试从context获取body
		bodyStr, _ := c.Get("request_body")
		if bodyStr != nil {
			json.Unmarshal([]byte(bodyStr.(string)), &req)
		}
	}

	// 设置默认值
	if req.PageNo < 1 {
		req.PageNo = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 50
	}

	goods, total, err := h.service.GetGoodsListForAPI(req.GoodsType, req.PageNo, req.PageSize)
	if err != nil {
		goofishError(c, 500, "查询商品列表失败")
		return
	}

	goofishSuccess(c, gin.H{
		"count": total,
		"list":  goods,
	})
}

// GoodsDetailRequest 商品详情请求
type GoodsDetailRequest struct {
	GoodsType int    `json:"goods_type"`
	GoodsNo   string `json:"goods_no" binding:"required"`
}

// GetGoodsDetail 查询商品详情
// POST /goofish/goods/detail
func (h *GoofishHandler) GetGoodsDetail(c *gin.Context) {
	var req GoodsDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试从context获取body
		bodyStr, _ := c.Get("request_body")
		if bodyStr != nil {
			json.Unmarshal([]byte(bodyStr.(string)), &req)
		}
	}

	if req.GoodsNo == "" {
		goofishError(c, 1, "商品编码不能为空")
		return
	}

	goods, err := h.service.GetGoodsDetailWithTime(req.GoodsNo)
	if err != nil {
		goofishError(c, 500, "查询商品详情失败")
		return
	}

	if goods == nil {
		goofishError(c, 1100, "商品不存在")
		return
	}

	goofishSuccess(c, goods)
}

// =====================================================
// 订单接口
// =====================================================

// CreateKamiOrder 创建卡密订单
// POST /goofish/order/purchase/create
func (h *GoofishHandler) CreateKamiOrder(c *gin.Context) {
	var req service.CreateKamiOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试从context获取body
		bodyStr, _ := c.Get("request_body")
		if bodyStr != nil {
			json.Unmarshal([]byte(bodyStr.(string)), &req)
		}
	}

	// 验证必填字段
	if req.OrderNo == "" {
		goofishError(c, 1201, "订单号不能为空")
		return
	}
	if req.GoodsNo == "" {
		goofishError(c, 1201, "商品编码不能为空")
		return
	}
	if req.BuyQuantity < 1 {
		goofishError(c, 1201, "购买数量必须大于0")
		return
	}

	// 创建订单
	resp, code, err := h.service.CreateKamiOrder(&req, c.ClientIP())
	if err != nil {
		goofishError(c, code, err.Error())
		return
	}

	goofishSuccess(c, resp)
}

// OrderDetailRequest 订单详情请求
type OrderDetailRequest struct {
	OrderType  int    `json:"order_type"`
	OrderNo    string `json:"order_no"`
	OutOrderNo string `json:"out_order_no"`
}

// GetOrderDetail 查询订单详情
// POST /goofish/order/detail
func (h *GoofishHandler) GetOrderDetail(c *gin.Context) {
	var req OrderDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试从context获取body
		bodyStr, _ := c.Get("request_body")
		if bodyStr != nil {
			json.Unmarshal([]byte(bodyStr.(string)), &req)
		}
	}

	// order_no 和 out_order_no 必传其一
	if req.OrderNo == "" && req.OutOrderNo == "" {
		goofishError(c, 1, "订单号不能为空")
		return
	}

	// 优先使用 order_no
	orderNo := req.OrderNo
	if orderNo == "" {
		orderNo = req.OutOrderNo
	}

	order, err := h.service.GetOrderDetailForAPI(orderNo, req.OutOrderNo)
	if err != nil {
		goofishError(c, 500, "查询订单详情失败")
		return
	}

	if order == nil {
		goofishError(c, 1200, "订单不存在")
		return
	}

	goofishSuccess(c, order)
}

// =====================================================
// 商品订阅接口
// =====================================================

// SubscribeGoods 订阅商品变更通知
// POST /goofish/goods/change/subscribe
func (h *GoofishHandler) SubscribeGoods(c *gin.Context) {
	var req service.SubscribeGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试从context获取body
		bodyStr, _ := c.Get("request_body")
		if bodyStr != nil {
			json.Unmarshal([]byte(bodyStr.(string)), &req)
		}
	}

	// 验证必填字段
	if req.GoodsType == 0 {
		goofishError(c, 1, "商品类型不能为空")
		return
	}
	if req.GoodsNo == "" {
		goofishError(c, 1, "商品编码不能为空")
		return
	}
	if req.Token == "" && req.NotifyURL == "" {
		goofishError(c, 1, "token和notify_url至少需要一个")
		return
	}

	if err := h.service.SubscribeGoods(&req); err != nil {
		if err.Error() == "商品不存在" {
			goofishError(c, 1100, err.Error())
			return
		}
		goofishError(c, 500, err.Error())
		return
	}

	goofishSuccess(c, nil)
}

// UnsubscribeGoods 取消商品变更通知订阅
// POST /goofish/goods/change/unsubscribe
func (h *GoofishHandler) UnsubscribeGoods(c *gin.Context) {
	var req service.UnsubscribeGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试从context获取body
		bodyStr, _ := c.Get("request_body")
		if bodyStr != nil {
			json.Unmarshal([]byte(bodyStr.(string)), &req)
		}
	}

	// 验证必填字段
	if req.GoodsType == 0 {
		goofishError(c, 1, "商品类型不能为空")
		return
	}
	if req.GoodsNo == "" {
		goofishError(c, 1, "商品编码不能为空")
		return
	}
	if req.Token == "" {
		goofishError(c, 1, "token不能为空")
		return
	}

	if err := h.service.UnsubscribeGoods(&req); err != nil {
		if err.Error() == "订阅不存在" {
			goofishError(c, 1, err.Error())
			return
		}
		goofishError(c, 500, err.Error())
		return
	}

	goofishSuccess(c, nil)
}

// GetSubscriptionList 查询商品订阅列表
// POST /goofish/goods/change/subscribe/list
func (h *GoofishHandler) GetSubscriptionList(c *gin.Context) {
	var req service.SubscriptionListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试从context获取body
		bodyStr, _ := c.Get("request_body")
		if bodyStr != nil {
			json.Unmarshal([]byte(bodyStr.(string)), &req)
		}
	}

	// 设置默认值
	if req.PageNo < 1 {
		req.PageNo = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 50
	}

	list, total, err := h.service.GetSubscriptionList(&req)
	if err != nil {
		goofishError(c, 500, "查询订阅列表失败")
		return
	}

	goofishSuccess(c, gin.H{
		"list":  list,
		"count": total,
	})
}
