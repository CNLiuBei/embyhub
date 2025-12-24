// Package handler 外部卡密API处理器
package handler

import (
	"encoding/json"
	"strings"
	"time"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// ExternalCardHandler 外部卡密API处理器
type ExternalCardHandler struct {
	externalCardService *service.ExternalCardService
}

// NewExternalCardHandler 创建外部卡密API处理器
func NewExternalCardHandler(externalCardService *service.ExternalCardService) *ExternalCardHandler {
	return &ExternalCardHandler{externalCardService: externalCardService}
}

// FetchCardRequest POST请求体
type FetchCardRequestBody struct {
	Type  int `json:"type"`
	Count int `json:"count"`
}

// FetchCard 获取卡密（供第三方系统调用）
// @Summary 获取卡密
// @Description 第三方系统调用此接口获取卡密，用于自动发货。支持GET和POST两种方式
// @Tags 外部API
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer API密钥"
// @Param type query int false "卡密类型（1月卡 2季卡 3半年卡 4年卡）- GET方式"
// @Param count query int false "获取数量（默认1，最大10）- GET方式"
// @Param body body FetchCardRequestBody false "请求体 - POST方式"
// @Success 200 {object} service.FetchCardResponse
// @Router /api/external/card/fetch [get]
// @Router /api/external/card/fetch [post]
func (h *ExternalCardHandler) FetchCard(c *gin.Context) {
	startTime := time.Now()

	// 获取API密钥
	authHeader := c.GetHeader("Authorization")
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		apiKey = c.Query("api_key") // 也支持query参数
	}

	// 验证API密钥
	valid, err := h.externalCardService.ValidateAPIKey(apiKey)
	if err != nil {
		h.logRequest(c, startTime, 401, err.Error())
		c.JSON(401, service.FetchCardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if !valid {
		h.logRequest(c, startTime, 401, "API密钥无效")
		c.JSON(401, service.FetchCardResponse{
			Success: false,
			Message: "API密钥无效",
		})
		return
	}

	// 检查IP白名单
	if !h.checkIPAllowed(c) {
		h.logRequest(c, startTime, 403, "IP不在白名单中")
		c.JSON(403, service.FetchCardResponse{
			Success: false,
			Message: "IP不在白名单中",
		})
		return
	}

	// 解析请求参数（同时支持GET和POST）
	var req service.FetchCardRequest
	req.Count = 1 // 默认值

	if c.Request.Method == "POST" {
		// POST方式：从请求体读取JSON
		var body FetchCardRequestBody
		if err := c.ShouldBindJSON(&body); err == nil {
			req.Type = int8(body.Type)
			if body.Count > 0 && body.Count <= 10 {
				req.Count = body.Count
			}
		}
	} else {
		// GET方式：从URL参数读取
		if typeStr := c.Query("type"); typeStr != "" {
			switch typeStr {
			case "1":
				req.Type = 1
			case "2":
				req.Type = 2
			case "3":
				req.Type = 3
			case "4":
				req.Type = 4
			}
		}

		if countStr := c.Query("count"); countStr != "" {
			switch countStr {
			case "1":
				req.Count = 1
			case "2":
				req.Count = 2
			case "3":
				req.Count = 3
			case "4":
				req.Count = 4
			case "5":
				req.Count = 5
			case "6":
				req.Count = 6
			case "7":
				req.Count = 7
			case "8":
				req.Count = 8
			case "9":
				req.Count = 9
			case "10":
				req.Count = 10
			}
		}
	}

	// 获取卡密
	if req.Count == 1 {
		card, err := h.externalCardService.FetchCard(req.Type)
		if err != nil {
			h.logRequest(c, startTime, 404, err.Error())
			c.JSON(200, service.FetchCardResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		resp := service.FetchCardResponse{
			Success: true,
			Message: "获取成功",
			Data:    card,
		}
		h.logRequest(c, startTime, 200, "获取成功")
		c.JSON(200, resp)
	} else {
		cards, err := h.externalCardService.FetchCards(req.Type, req.Count)
		if err != nil {
			h.logRequest(c, startTime, 404, err.Error())
			c.JSON(200, service.FetchCardsResponse{
				Success: false,
				Message: err.Error(),
				Count:   0,
			})
			return
		}

		resp := service.FetchCardsResponse{
			Success: true,
			Message: "获取成功",
			Data:    cards,
			Count:   len(cards),
		}
		h.logRequest(c, startTime, 200, "获取成功")
		c.JSON(200, resp)
	}
}

// FetchCardByType 按类型获取卡密（供第三方系统调用）
// @Summary 按类型获取卡密
// @Description 通过URL路径指定卡密类型，适配不支持自定义参数的第三方系统
// @Tags 外部API
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer API密钥"
// @Param type path string true "卡密类型（monthly/quarterly/halfyear/yearly）"
// @Success 200 {object} service.FetchCardResponse
// @Router /api/external/card/fetch/monthly [get]
// @Router /api/external/card/fetch/monthly [post]
// @Router /api/external/card/fetch/quarterly [get]
// @Router /api/external/card/fetch/quarterly [post]
// @Router /api/external/card/fetch/halfyear [get]
// @Router /api/external/card/fetch/halfyear [post]
// @Router /api/external/card/fetch/yearly [get]
// @Router /api/external/card/fetch/yearly [post]
func (h *ExternalCardHandler) FetchCardByType(c *gin.Context) {
	startTime := time.Now()

	// 获取API密钥
	authHeader := c.GetHeader("Authorization")
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	// 验证API密钥
	valid, err := h.externalCardService.ValidateAPIKey(apiKey)
	if err != nil {
		h.logRequest(c, startTime, 401, err.Error())
		c.JSON(401, service.FetchCardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if !valid {
		h.logRequest(c, startTime, 401, "API密钥无效")
		c.JSON(401, service.FetchCardResponse{
			Success: false,
			Message: "API密钥无效",
		})
		return
	}

	// 检查IP白名单
	if !h.checkIPAllowed(c) {
		h.logRequest(c, startTime, 403, "IP不在白名单中")
		c.JSON(403, service.FetchCardResponse{
			Success: false,
			Message: "IP不在白名单中",
		})
		return
	}

	// 从URL路径获取卡密类型
	cardType := c.Param("type")
	var typeCode int8
	switch cardType {
	case "monthly":
		typeCode = 1
	case "quarterly":
		typeCode = 2
	case "halfyear":
		typeCode = 3
	case "yearly":
		typeCode = 4
	default:
		h.logRequest(c, startTime, 400, "无效的卡密类型")
		c.JSON(400, service.FetchCardResponse{
			Success: false,
			Message: "无效的卡密类型，支持: monthly, quarterly, halfyear, yearly",
		})
		return
	}

	// 获取卡密
	card, err := h.externalCardService.FetchCard(typeCode)
	if err != nil {
		h.logRequest(c, startTime, 404, err.Error())
		c.JSON(200, service.FetchCardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	resp := service.FetchCardResponse{
		Success: true,
		Message: "获取成功",
		Data:    card,
	}
	h.logRequest(c, startTime, 200, "获取成功")
	c.JSON(200, resp)
}

// XianyuOrderRequest 咸鱼订单请求参数
type XianyuOrderRequest struct {
	OrderID       string `json:"order_id" form:"order_id"`             // 订单ID
	ItemID        string `json:"item_id" form:"item_id"`               // 商品ID
	ItemDetail    string `json:"item_detail" form:"item_detail"`       // 商品详情/标题
	OrderAmount   string `json:"order_amount" form:"order_amount"`     // 订单金额
	OrderQuantity int    `json:"order_quantity" form:"order_quantity"` // 订单数量
	SpecName      string `json:"spec_name" form:"spec_name"`           // 规格名称
	SpecValue     string `json:"spec_value" form:"spec_value"`         // 规格值
	CookieID      string `json:"cookie_id" form:"cookie_id"`           // Cookie ID
	BuyerID       string `json:"buyer_id" form:"buyer_id"`             // 买家ID
}

// detectCardType 根据请求参数自动识别卡密类型
// 优先级: spec_value > spec_name > item_detail
func (h *ExternalCardHandler) detectCardType(req *XianyuOrderRequest) (int8, string) {
	// 检测关键词的函数
	detectFromText := func(text string) (int8, string, bool) {
		text = strings.ToLower(text)
		// 按优先级检测：年卡 > 半年卡 > 季卡 > 月卡
		if strings.Contains(text, "年卡") || strings.Contains(text, "yearly") || strings.Contains(text, "一年") || strings.Contains(text, "12个月") {
			return 4, "年卡", true
		}
		if strings.Contains(text, "半年卡") || strings.Contains(text, "halfyear") || strings.Contains(text, "半年") || strings.Contains(text, "6个月") {
			return 3, "半年卡", true
		}
		if strings.Contains(text, "季卡") || strings.Contains(text, "quarterly") || strings.Contains(text, "季度") || strings.Contains(text, "3个月") {
			return 2, "季卡", true
		}
		if strings.Contains(text, "月卡") || strings.Contains(text, "monthly") || strings.Contains(text, "一个月") || strings.Contains(text, "1个月") {
			return 1, "月卡", true
		}
		return 0, "", false
	}

	// 优先从 spec_value 检测
	if req.SpecValue != "" {
		if typeCode, typeName, found := detectFromText(req.SpecValue); found {
			return typeCode, typeName
		}
	}

	// 其次从 spec_name 检测
	if req.SpecName != "" {
		if typeCode, typeName, found := detectFromText(req.SpecName); found {
			return typeCode, typeName
		}
	}

	// 最后从 item_detail 检测
	if req.ItemDetail != "" {
		if typeCode, typeName, found := detectFromText(req.ItemDetail); found {
			return typeCode, typeName
		}
	}

	// 默认返回0表示未识别
	return 0, ""
}

// XianyuFetchCard 咸鱼系统专用API - 获取卡密（带类型路径）
// @Summary 咸鱼系统专用 - 获取卡密
// @Description 专为咸鱼自动回复系统设计的API，接受订单参数，返回简化的卡密响应
// @Tags 外部API-咸鱼专用
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer API密钥"
// @Param type path string true "卡密类型（monthly/quarterly/halfyear/yearly）"
// @Success 200 {object} map[string]interface{} "返回格式: {code: '卡密', success: true, message: '获取成功'}"
// @Router /api/external/xianyu/monthly [get]
// @Router /api/external/xianyu/monthly [post]
func (h *ExternalCardHandler) XianyuFetchCard(c *gin.Context) {
	startTime := time.Now()

	// 获取API密钥
	authHeader := c.GetHeader("Authorization")
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	// 验证API密钥
	valid, err := h.externalCardService.ValidateAPIKey(apiKey)
	if err != nil {
		h.logRequest(c, startTime, 401, err.Error())
		c.JSON(401, gin.H{
			"success": false,
			"message": err.Error(),
			"code":    "",
		})
		return
	}
	if !valid {
		h.logRequest(c, startTime, 401, "API密钥无效")
		c.JSON(401, gin.H{
			"success": false,
			"message": "API密钥无效",
			"code":    "",
		})
		return
	}

	// 检查IP白名单
	if !h.checkIPAllowed(c) {
		h.logRequest(c, startTime, 403, "IP不在白名单中")
		c.JSON(403, gin.H{
			"success": false,
			"message": "IP不在白名单中",
			"code":    "",
		})
		return
	}

	// 从URL路径获取卡密类型
	cardType := c.Param("type")
	var typeCode int8
	var typeName string
	switch cardType {
	case "monthly":
		typeCode = 1
		typeName = "月卡"
	case "quarterly":
		typeCode = 2
		typeName = "季卡"
	case "halfyear":
		typeCode = 3
		typeName = "半年卡"
	case "yearly":
		typeCode = 4
		typeName = "年卡"
	default:
		h.logRequest(c, startTime, 400, "无效的卡密类型")
		c.JSON(400, gin.H{
			"success": false,
			"message": "无效的卡密类型，支持: monthly, quarterly, halfyear, yearly",
			"code":    "",
		})
		return
	}

	// 解析咸鱼订单参数（用于日志记录）
	var orderReq XianyuOrderRequest
	if c.Request.Method == "POST" {
		c.ShouldBindJSON(&orderReq)
	} else {
		c.ShouldBindQuery(&orderReq)
	}

	// 获取数量，默认1
	count := orderReq.OrderQuantity
	if count <= 0 {
		count = 1
	}
	if count > 10 {
		count = 10
	}

	// 获取卡密
	if count == 1 {
		card, err := h.externalCardService.FetchCard(typeCode)
		if err != nil {
			h.logRequest(c, startTime, 404, err.Error())
			c.JSON(200, gin.H{
				"success": false,
				"message": err.Error(),
				"code":    "",
			})
			return
		}

		h.logRequest(c, startTime, 200, "获取成功")
		c.JSON(200, gin.H{
			"success":   true,
			"message":   "获取成功",
			"code":      card.Code,
			"type":      typeCode,
			"type_name": typeName,
			"duration":  card.Duration,
			"order_id":  orderReq.OrderID,
			"data":      gin.H{"code": card.Code, "type": typeCode, "duration": card.Duration},
		})
	} else {
		cards, err := h.externalCardService.FetchCards(typeCode, count)
		if err != nil {
			h.logRequest(c, startTime, 404, err.Error())
			c.JSON(200, gin.H{
				"success": false,
				"message": err.Error(),
				"code":    "",
				"codes":   []string{},
			})
			return
		}

		// 提取所有卡密码
		codes := make([]string, len(cards))
		for i, card := range cards {
			codes[i] = card.Code
		}

		h.logRequest(c, startTime, 200, "获取成功")
		c.JSON(200, gin.H{
			"success":   true,
			"message":   "获取成功",
			"code":      strings.Join(codes, "\n"), // 多个卡密用换行分隔
			"codes":     codes,                     // 数组格式
			"count":     len(cards),
			"type":      typeCode,
			"type_name": typeName,
			"order_id":  orderReq.OrderID,
		})
	}
}

// XianyuAutoFetchCard 咸鱼系统专用API - 自动识别类型获取卡密
// @Summary 咸鱼系统专用 - 自动识别类型获取卡密
// @Description 根据订单参数（item_detail/spec_name/spec_value）自动识别卡密类型并返回
// @Tags 外部API-咸鱼专用
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer API密钥"
// @Param order_id body string false "订单ID"
// @Param item_id body string false "商品ID"
// @Param item_detail body string false "商品详情/标题（用于识别卡密类型）"
// @Param order_amount body string false "订单金额"
// @Param order_quantity body int false "订单数量（默认1，最大10）"
// @Param spec_name body string false "规格名称（用于识别卡密类型）"
// @Param spec_value body string false "规格值（用于识别卡密类型，优先级最高）"
// @Param cookie_id body string false "Cookie ID"
// @Param buyer_id body string false "买家ID"
// @Success 200 {object} map[string]interface{} "返回格式: {code: '卡密', success: true, message: '获取成功'}"
// @Router /api/external/xianyu [get]
// @Router /api/external/xianyu [post]
func (h *ExternalCardHandler) XianyuAutoFetchCard(c *gin.Context) {
	startTime := time.Now()

	// 获取API密钥
	authHeader := c.GetHeader("Authorization")
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	// 验证API密钥
	valid, err := h.externalCardService.ValidateAPIKey(apiKey)
	if err != nil {
		h.logRequest(c, startTime, 401, err.Error())
		c.JSON(401, gin.H{
			"success": false,
			"message": err.Error(),
			"code":    "",
		})
		return
	}
	if !valid {
		h.logRequest(c, startTime, 401, "API密钥无效")
		c.JSON(401, gin.H{
			"success": false,
			"message": "API密钥无效",
			"code":    "",
		})
		return
	}

	// 检查IP白名单
	if !h.checkIPAllowed(c) {
		h.logRequest(c, startTime, 403, "IP不在白名单中")
		c.JSON(403, gin.H{
			"success": false,
			"message": "IP不在白名单中",
			"code":    "",
		})
		return
	}

	// 解析咸鱼订单参数
	var orderReq XianyuOrderRequest
	if c.Request.Method == "POST" {
		c.ShouldBindJSON(&orderReq)
	} else {
		c.ShouldBindQuery(&orderReq)
	}

	// 自动识别卡密类型
	typeCode, typeName := h.detectCardType(&orderReq)
	if typeCode == 0 {
		h.logRequest(c, startTime, 400, "无法识别卡密类型")
		c.JSON(200, gin.H{
			"success":     false,
			"message":     "无法识别卡密类型，请在商品标题或规格中包含：月卡/季卡/半年卡/年卡",
			"code":        "",
			"item_detail": orderReq.ItemDetail,
			"spec_name":   orderReq.SpecName,
			"spec_value":  orderReq.SpecValue,
		})
		return
	}

	// 获取数量，默认1
	count := orderReq.OrderQuantity
	if count <= 0 {
		count = 1
	}
	if count > 10 {
		count = 10
	}

	// 获取卡密
	if count == 1 {
		card, err := h.externalCardService.FetchCard(typeCode)
		if err != nil {
			h.logRequest(c, startTime, 404, err.Error())
			c.JSON(200, gin.H{
				"success": false,
				"message": err.Error(),
				"code":    "",
			})
			return
		}

		h.logRequest(c, startTime, 200, "获取成功")
		c.JSON(200, gin.H{
			"success":   true,
			"message":   "获取成功",
			"code":      card.Code,
			"type":      typeCode,
			"type_name": typeName,
			"duration":  card.Duration,
			"order_id":  orderReq.OrderID,
			"data":      gin.H{"code": card.Code, "type": typeCode, "duration": card.Duration},
		})
	} else {
		cards, err := h.externalCardService.FetchCards(typeCode, count)
		if err != nil {
			h.logRequest(c, startTime, 404, err.Error())
			c.JSON(200, gin.H{
				"success": false,
				"message": err.Error(),
				"code":    "",
				"codes":   []string{},
			})
			return
		}

		// 提取所有卡密码
		codes := make([]string, len(cards))
		for i, card := range cards {
			codes[i] = card.Code
		}

		h.logRequest(c, startTime, 200, "获取成功")
		c.JSON(200, gin.H{
			"success":   true,
			"message":   "获取成功",
			"code":      strings.Join(codes, "\n"),
			"codes":     codes,
			"count":     len(cards),
			"type":      typeCode,
			"type_name": typeName,
			"order_id":  orderReq.OrderID,
		})
	}
}

// GetStock 获取卡密库存
// @Summary 获取卡密库存
// @Description 获取各类型卡密的库存数量
// @Tags 外部API
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer API密钥"
// @Success 200 {object} map[string]interface{}
// @Router /api/external/card/stock [get]
func (h *ExternalCardHandler) GetStock(c *gin.Context) {
	startTime := time.Now()

	// 获取API密钥
	authHeader := c.GetHeader("Authorization")
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	// 验证API密钥
	valid, err := h.externalCardService.ValidateAPIKey(apiKey)
	if err != nil {
		h.logRequest(c, startTime, 401, err.Error())
		c.JSON(401, gin.H{"success": false, "message": err.Error()})
		return
	}
	if !valid {
		h.logRequest(c, startTime, 401, "API密钥无效")
		c.JSON(401, gin.H{"success": false, "message": "API密钥无效"})
		return
	}

	// 检查IP白名单
	if !h.checkIPAllowed(c) {
		h.logRequest(c, startTime, 403, "IP不在白名单中")
		c.JSON(403, gin.H{"success": false, "message": "IP不在白名单中"})
		return
	}

	stock, err := h.externalCardService.GetCardStock()
	if err != nil {
		h.logRequest(c, startTime, 500, err.Error())
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logRequest(c, startTime, 200, "获取成功")
	c.JSON(200, gin.H{
		"success": true,
		"message": "获取成功",
		"data":    stock,
	})
}

// checkIPAllowed 检查IP是否在白名单中
func (h *ExternalCardHandler) checkIPAllowed(c *gin.Context) bool {
	settings, err := h.externalCardService.GetExternalCardAPISettings()
	if err != nil {
		return false
	}

	// 如果未配置IP白名单，允许所有IP
	if settings.AllowedIPs == "" {
		return true
	}

	clientIP := c.ClientIP()
	allowedIPs := strings.Split(settings.AllowedIPs, ",")
	for _, ip := range allowedIPs {
		ip = strings.TrimSpace(ip)
		if ip == clientIP {
			return true
		}
	}

	return false
}

// logRequest 记录请求日志
func (h *ExternalCardHandler) logRequest(c *gin.Context, startTime time.Time, status int, message string) {
	duration := time.Since(startTime).Milliseconds()
	params, _ := json.Marshal(c.Request.URL.Query())
	h.externalCardService.LogAPIRequest(
		c.ClientIP(),
		c.Request.Method,
		c.Request.URL.Path,
		string(params),
		message,
		status,
		duration,
	)
}

// ========== 管理员接口 ==========

// GetSettings 获取外部API设置
func (h *ExternalCardHandler) GetSettings(c *gin.Context) {
	settings, err := h.externalCardService.GetExternalCardAPISettings()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, settings)
}

// SaveSettings 保存外部API设置
func (h *ExternalCardHandler) SaveSettings(c *gin.Context) {
	var settings service.ExternalCardAPISettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	adminID, _ := middleware.GetUserID(c)
	if err := h.externalCardService.SaveExternalCardAPISettings(&settings, adminID); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "保存成功", nil)
}

// GenerateAPIKey 生成新的API密钥
func (h *ExternalCardHandler) GenerateAPIKey(c *gin.Context) {
	apiKey := h.externalCardService.GenerateAPIKey()
	response.Success(c, gin.H{"api_key": apiKey})
}

// GetAPILogs 获取API日志
func (h *ExternalCardHandler) GetAPILogs(c *gin.Context) {
	var req struct {
		Page     int `form:"page" binding:"required,min=1"`
		PageSize int `form:"page_size" binding:"required,min=1,max=100"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	logs, total, err := h.externalCardService.GetAPILogs(req.Page, req.PageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, logs, total, req.Page, req.PageSize)
}
