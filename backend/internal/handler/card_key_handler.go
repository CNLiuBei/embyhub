package handler

import (
	"strconv"

	"embyhub/internal/model"
	"embyhub/internal/service"
	"embyhub/internal/util"

	"github.com/gin-gonic/gin"
)

type CardKeyHandler struct {
	cardKeyService *service.CardKeyService
}

func NewCardKeyHandler() *CardKeyHandler {
	return &CardKeyHandler{
		cardKeyService: service.NewCardKeyService(),
	}
}

// Create 批量生成卡密
// @Summary 批量生成卡密
// @Tags 卡密管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body model.CardKeyCreateRequest true "生成请求"
// @Success 200 {object} model.Response{data=[]model.CardKey}
// @Router /api/card-keys [post]
func (h *CardKeyHandler) Create(c *gin.Context) {
	var req model.CardKeyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取创建者ID
	creatorID, _ := c.Get("user_id")

	cardKeys, err := h.cardKeyService.Create(&req, creatorID.(int))
	if err != nil {
		util.InternalErrorResponse(c, "生成卡密失败: "+err.Error())
		return
	}

	util.SuccessWithMessage(c, "生成成功", cardKeys)
}

// List 获取卡密列表
// @Summary 获取卡密列表
// @Tags 卡密管理
// @Security Bearer
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param status query int false "状态"
// @Param card_type query int false "类型"
// @Param keyword query string false "关键词"
// @Success 200 {object} model.Response{data=model.CardKeyListResponse}
// @Router /api/card-keys [get]
func (h *CardKeyHandler) List(c *gin.Context) {
	var req model.CardKeyListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误")
		return
	}

	result, err := h.cardKeyService.List(&req)
	if err != nil {
		util.InternalErrorResponse(c, "获取卡密列表失败")
		return
	}

	util.SuccessResponse(c, result)
}

// GetByID 获取卡密详情
// @Summary 获取卡密详情
// @Tags 卡密管理
// @Security Bearer
// @Produce json
// @Param id path int true "卡密ID"
// @Success 200 {object} model.Response{data=model.CardKey}
// @Router /api/card-keys/{id} [get]
func (h *CardKeyHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "无效的ID")
		return
	}

	cardKey, err := h.cardKeyService.GetByID(id)
	if err != nil {
		util.NotFoundResponse(c, "卡密不存在")
		return
	}

	util.SuccessResponse(c, cardKey)
}

// Disable 禁用卡密
// @Summary 禁用卡密
// @Tags 卡密管理
// @Security Bearer
// @Param id path int true "卡密ID"
// @Success 200 {object} model.Response
// @Router /api/card-keys/{id}/disable [put]
func (h *CardKeyHandler) Disable(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "无效的ID")
		return
	}

	if err := h.cardKeyService.Disable(id); err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "禁用成功", nil)
}

// Enable 启用卡密
// @Summary 启用卡密
// @Tags 卡密管理
// @Security Bearer
// @Param id path int true "卡密ID"
// @Success 200 {object} model.Response
// @Router /api/card-keys/{id}/enable [put]
func (h *CardKeyHandler) Enable(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "无效的ID")
		return
	}

	if err := h.cardKeyService.Enable(id); err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "启用成功", nil)
}

// Delete 删除卡密
// @Summary 删除卡密
// @Tags 卡密管理
// @Security Bearer
// @Param id path int true "卡密ID"
// @Success 200 {object} model.Response
// @Router /api/card-keys/{id} [delete]
func (h *CardKeyHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequestResponse(c, "无效的ID")
		return
	}

	if err := h.cardKeyService.Delete(id); err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "删除成功", nil)
}

// GetStatistics 获取卡密统计
// @Summary 获取卡密统计
// @Tags 卡密管理
// @Security Bearer
// @Produce json
// @Success 200 {object} model.Response
// @Router /api/card-keys/statistics [get]
func (h *CardKeyHandler) GetStatistics(c *gin.Context) {
	stats, err := h.cardKeyService.GetStatistics()
	if err != nil {
		util.InternalErrorResponse(c, "获取统计失败")
		return
	}

	util.SuccessResponse(c, stats)
}

// Validate 验证卡密（公开接口，用于注册前验证）
// @Summary 验证卡密
// @Tags 卡密管理
// @Accept json
// @Produce json
// @Param request body model.CardKeyUseRequest true "验证请求"
// @Success 200 {object} model.Response
// @Router /api/card-keys/validate [post]
func (h *CardKeyHandler) Validate(c *gin.Context) {
	var req model.CardKeyUseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误")
		return
	}

	cardKey, err := h.cardKeyService.ValidateCardCode(req.CardCode)
	if err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessResponse(c, map[string]interface{}{
		"valid":     true,
		"card_type": cardKey.CardType,
		"duration":  cardKey.Duration,
	})
}

// UseVipCard 使用VIP升级码
// @Summary 使用VIP升级码
// @Tags 卡密管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body model.CardKeyUseRequest true "使用请求"
// @Success 200 {object} model.Response
// @Router /api/card-keys/use-vip [post]
func (h *CardKeyHandler) UseVipCard(c *gin.Context) {
	var req model.CardKeyUseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误")
		return
	}

	// 获取当前用户ID
	userID, _ := c.Get("user_id")

	user, err := h.cardKeyService.UseVipCard(req.CardCode, userID.(int))
	if err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "VIP升级成功", map[string]interface{}{
		"vip_level":     user.VipLevel,
		"vip_expire_at": user.VipExpireAt,
	})
}
