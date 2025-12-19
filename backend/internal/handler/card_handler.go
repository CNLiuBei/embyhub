// Package handler 卡密处理器
package handler

import (
	"strconv"
	"strings"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// CardHandler 卡密处理器
type CardHandler struct {
	cardService       *service.CardService
	cardExportService *service.CardExportService
}

// NewCardHandler 创建卡密处理器
func NewCardHandler(cardService *service.CardService, cardExportService *service.CardExportService) *CardHandler {
	return &CardHandler{
		cardService:       cardService,
		cardExportService: cardExportService,
	}
}

// Redeem 用户兑换卡密
// @Summary 兑换卡密
// @Tags 卡密
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body service.RedeemRequest true "兑换请求"
// @Success 200 {object} response.Response
// @Router /api/v1/card/redeem [post]
func (h *CardHandler) Redeem(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req service.RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入卡密")
		return
	}

	order, err := h.cardService.Redeem(userID, req.Code)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "兑换成功", order)
}

// GetRedeemHistory 获取兑换记录
// @Summary 获取兑换记录
// @Tags 卡密
// @Security Bearer
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Success 200 {object} response.Response{data=response.PageData}
// @Router /api/v1/card/history [get]
func (h *CardHandler) GetRedeemHistory(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	page := 1
	pageSize := 20
	c.ShouldBindQuery(&struct {
		Page     *int `form:"page"`
		PageSize *int `form:"page_size"`
	}{&page, &pageSize})

	orders, total, err := h.cardService.GetUserRedeemHistory(userID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, orders, total, page, pageSize)
}

// ========== 管理员接口 ==========

// CreateBatch 批量生成卡密
// @Summary 批量生成卡密
// @Tags 管理员-卡密
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body service.CreateBatchRequest true "创建请求"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/card/batch [post]
func (h *CardHandler) CreateBatch(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)
	adminName := c.GetString("nickname")

	var req service.CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	batch, cards, err := h.cardService.CreateBatch(adminID, adminName, &req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	// 返回批次信息和卡密码列表
	codes := make([]string, len(cards))
	for i, card := range cards {
		codes[i] = card.Code
	}

	response.Success(c, gin.H{
		"batch": batch,
		"codes": codes,
	})
}

// GetBatchList 获取批次列表
// @Summary 获取批次列表
// @Tags 管理员-卡密
// @Security Bearer
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Success 200 {object} response.Response{data=response.PageData}
// @Router /api/v1/admin/card/batch/list [get]
func (h *CardHandler) GetBatchList(c *gin.Context) {
	page := 1
	pageSize := 20
	c.ShouldBindQuery(&struct {
		Page     *int `form:"page"`
		PageSize *int `form:"page_size"`
	}{&page, &pageSize})

	batches, total, err := h.cardService.GetBatchList(page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, batches, total, page, pageSize)
}

// GetCardList 获取卡密列表
// @Summary 获取卡密列表
// @Tags 管理员-卡密
// @Security Bearer
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Param batch_no query string false "批次号"
// @Param card_type query int false "卡密类型"
// @Param status query int false "状态"
// @Param code query string false "卡密码"
// @Success 200 {object} response.Response{data=response.PageData}
// @Router /api/v1/admin/card/list [get]
func (h *CardHandler) GetCardList(c *gin.Context) {
	var req service.CardListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}

	cards, total, err := h.cardService.GetCardList(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, cards, total, req.Page, req.PageSize)
}

// DisableCard 禁用卡密
// @Summary 禁用卡密
// @Tags 管理员-卡密
// @Security Bearer
// @Param id path int true "卡密ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/card/{id}/disable [post]
func (h *CardHandler) DisableCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的卡密ID")
		return
	}

	if err := h.cardService.DisableCard(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "禁用成功", nil)
}

// EnableCard 启用卡密
// @Summary 启用卡密
// @Tags 管理员-卡密
// @Security Bearer
// @Param id path int true "卡密ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/card/{id}/enable [post]
func (h *CardHandler) EnableCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的卡密ID")
		return
	}

	if err := h.cardService.EnableCard(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "启用成功", nil)
}

// DeleteCard 删除卡密
// @Summary 删除卡密
// @Tags 管理员-卡密
// @Security Bearer
// @Param id path int true "卡密ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/card/{id} [delete]
func (h *CardHandler) DeleteCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的卡密ID")
		return
	}

	if err := h.cardService.DeleteCard(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// ExportCards 导出卡密
// @Summary 导出卡密
// @Tags 管理员-卡密
// @Security Bearer
// @Param batch_no query string true "批次号"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/card/export [get]
func (h *CardHandler) ExportCards(c *gin.Context) {
	batchNo := c.Query("batch_no")
	if batchNo == "" {
		response.BadRequest(c, "请指定批次号")
		return
	}

	codes, err := h.cardService.ExportCards(batchNo)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	// 返回纯文本格式，每行一个卡密
	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", "attachment; filename=cards_"+batchNo+".txt")
	c.String(200, strings.Join(codes, "\n"))
}

// GetCardStats 获取卡密统计
// @Summary 获取卡密统计
// @Tags 管理员-卡密
// @Security Bearer
// @Success 200 {object} response.Response{data=service.CardStats}
// @Router /api/v1/admin/card/stats [get]
func (h *CardHandler) GetCardStats(c *gin.Context) {
	stats, err := h.cardService.GetCardStats()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// ExportToCSV 导出为CSV
// @Summary 导出卡密为CSV格式
// @Tags 管理员-卡密
// @Security Bearer
// @Param batch_no query string false "批次号"
// @Param card_type query int false "卡密类型"
// @Param status query int false "状态"
// @Router /api/v1/admin/card/export/csv [get]
func (h *CardHandler) ExportToCSV(c *gin.Context) {
	filter := &service.ExportFilter{
		BatchNo: c.Query("batch_no"),
	}

	if cardType := c.Query("card_type"); cardType != "" {
		ct, _ := strconv.Atoi(cardType)
		filter.CardType = int8(ct)
	}

	if status := c.Query("status"); status != "" {
		st, _ := strconv.Atoi(status)
		s := int8(st)
		filter.Status = &s
	}

	data, filename, err := h.cardExportService.ExportToCSV(filter)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(200, "text/csv", data)
}

// ExportToExcel 导出为Excel
// @Summary 导出卡密为Excel格式
// @Tags 管理员-卡密
// @Security Bearer
// @Param batch_no query string false "批次号"
// @Param card_type query int false "卡密类型"
// @Param status query int false "状态"
// @Router /api/v1/admin/card/export/excel [get]
func (h *CardHandler) ExportToExcel(c *gin.Context) {
	filter := &service.ExportFilter{
		BatchNo: c.Query("batch_no"),
	}

	if cardType := c.Query("card_type"); cardType != "" {
		ct, _ := strconv.Atoi(cardType)
		filter.CardType = int8(ct)
	}

	if status := c.Query("status"); status != "" {
		st, _ := strconv.Atoi(status)
		s := int8(st)
		filter.Status = &s
	}

	data, filename, err := h.cardExportService.ExportToExcel(filter)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ExportCodesOnly 仅导出卡密码
// @Summary 导出卡密码列表（纯文本）
// @Tags 管理员-卡密
// @Security Bearer
// @Param batch_no query string true "批次号"
// @Router /api/v1/admin/card/export/codes [get]
func (h *CardHandler) ExportCodesOnly(c *gin.Context) {
	batchNo := c.Query("batch_no")
	if batchNo == "" {
		response.BadRequest(c, "请指定批次号")
		return
	}

	content, filename, err := h.cardExportService.ExportCodesOnly(batchNo)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.String(200, content)
}

// GenerateUsageReport 生成使用报告
// @Summary 生成批次使用情况报告
// @Tags 管理员-卡密
// @Security Bearer
// @Param batch_no query string true "批次号"
// @Router /api/v1/admin/card/export/report [get]
func (h *CardHandler) GenerateUsageReport(c *gin.Context) {
	batchNo := c.Query("batch_no")
	if batchNo == "" {
		response.BadRequest(c, "请指定批次号")
		return
	}

	data, filename, err := h.cardExportService.GenerateUsageReport(batchNo)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// RenewByCard 使用卡密续费（公开接口，无需登录）
// @Summary 禁用用户使用卡密续费
// @Description 允许被禁用的用户通过账号+卡密来续费并恢复账户
// @Tags 卡密
// @Accept json
// @Produce json
// @Param request body service.RenewByCardRequest true "续费请求"
// @Success 200 {object} response.Response
// @Router /api/v1/card/renew [post]
func (h *CardHandler) RenewByCard(c *gin.Context) {
	var req service.RenewByCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入账号和卡密")
		return
	}

	order, err := h.cardService.RenewByCard(req.Account, req.Code)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "续费成功，账户已恢复，请重新登录", gin.H{
		"order_no":    order.OrderNo,
		"expire_time": order.ExpireTime.Format("2006-01-02 15:04:05"),
		"duration":    order.Duration,
	})
}
