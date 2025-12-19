// Package handler 积分卡密处理器
package handler

import (
	"fmt"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// PointsCardHandler 积分卡密处理器
type PointsCardHandler struct {
	pointsCardService *service.PointsCardService
}

// NewPointsCardHandler 创建积分卡密处理器
func NewPointsCardHandler(pointsCardService *service.PointsCardService) *PointsCardHandler {
	return &PointsCardHandler{pointsCardService: pointsCardService}
}

// Redeem 兑换积分卡密
func (h *PointsCardHandler) Redeem(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入卡密")
		return
	}

	points, err := h.pointsCardService.Redeem(userID, req.Code)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "兑换成功", gin.H{"points": points})
}

// ========== 管理员接口 ==========

// AdminCreateBatch 批量生成积分卡密
func (h *PointsCardHandler) AdminCreateBatch(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)
	adminName := c.GetString("username")

	var req struct {
		Points   int    `json:"points" binding:"required,min=1"`
		Quantity int    `json:"quantity" binding:"required,min=1,max=1000"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	batch, err := h.pointsCardService.CreateBatch(adminID, adminName, req.Points, req.Quantity, req.Remark)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, batch)
}

// AdminGetBatchList 获取批次列表
func (h *PointsCardHandler) AdminGetBatchList(c *gin.Context) {
	var req struct {
		Page     int `form:"page" binding:"required,min=1"`
		PageSize int `form:"page_size" binding:"required,min=1,max=100"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	batches, total, err := h.pointsCardService.GetBatchList(req.Page, req.PageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, batches, total, req.Page, req.PageSize)
}

// AdminGetCardList 获取卡密列表
func (h *PointsCardHandler) AdminGetCardList(c *gin.Context) {
	var req struct {
		Page     int    `form:"page" binding:"required,min=1"`
		PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
		BatchNo  string `form:"batch_no"`
		Status   *int8  `form:"status"`
		Keyword  string `form:"keyword"` // 卡密模糊查询
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	cards, total, err := h.pointsCardService.GetCardList(req.Page, req.PageSize, req.BatchNo, req.Status, req.Keyword)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, cards, total, req.Page, req.PageSize)
}

// AdminDisableCard 禁用卡密
func (h *PointsCardHandler) AdminDisableCard(c *gin.Context) {
	idStr := c.Param("id")
	var id uint64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.pointsCardService.DisableCard(id); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "禁用成功", nil)
}

// AdminEnableCard 启用卡密
func (h *PointsCardHandler) AdminEnableCard(c *gin.Context) {
	idStr := c.Param("id")
	var id uint64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.pointsCardService.EnableCard(id); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "启用成功", nil)
}

// AdminGetStats 获取统计
func (h *PointsCardHandler) AdminGetStats(c *gin.Context) {
	stats, err := h.pointsCardService.GetStats()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// AdminExportCards 导出卡密
func (h *PointsCardHandler) AdminExportCards(c *gin.Context) {
	batchNo := c.Query("batch_no")
	if batchNo == "" {
		response.BadRequest(c, "请指定批次号")
		return
	}

	cards, err := h.pointsCardService.ExportCards(batchNo)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	// 返回卡密码列表
	codes := make([]string, len(cards))
	for i, card := range cards {
		codes[i] = card.Code
	}

	response.Success(c, gin.H{
		"batch_no": batchNo,
		"codes":    codes,
		"count":    len(codes),
	})
}

// AdminDeleteBatch 删除批次
func (h *PointsCardHandler) AdminDeleteBatch(c *gin.Context) {
	batchNo := c.Param("batch_no")
	if batchNo == "" {
		response.BadRequest(c, "请指定批次号")
		return
	}

	if err := h.pointsCardService.DeleteBatch(batchNo); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// AdminDeleteCard 删除卡密
func (h *PointsCardHandler) AdminDeleteCard(c *gin.Context) {
	idStr := c.Param("id")
	var id uint64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.pointsCardService.DeleteCard(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
