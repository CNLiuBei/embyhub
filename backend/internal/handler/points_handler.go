// Package handler 积分处理器
package handler

import (
	"fmt"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/models"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PointsHandler 积分处理器
type PointsHandler struct {
	pointsService *service.PointsService
}

// NewPointsHandler 创建积分处理器
func NewPointsHandler(pointsService *service.PointsService) *PointsHandler {
	return &PointsHandler{pointsService: pointsService}
}

// GetMyPoints 获取我的积分
func (h *PointsHandler) GetMyPoints(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	points, err := h.pointsService.GetUserPoints(userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"points": points})
}

// SignIn 签到
func (h *PointsHandler) SignIn(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	result, err := h.pointsService.SignIn(userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetSignInStatus 获取签到状态
func (h *PointsHandler) GetSignInStatus(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	status, err := h.pointsService.GetSignInStatus(userID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, status)
}

// GetPointsRecords 获取积分记录
func (h *PointsHandler) GetPointsRecords(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		Page     int   `form:"page" binding:"required,min=1"`
		PageSize int   `form:"page_size" binding:"required,min=1,max=100"`
		Type     *int8 `form:"type"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	records, total, err := h.pointsService.GetPointsRecords(userID, req.Page, req.PageSize, req.Type)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, records, total, req.Page, req.PageSize)
}

// GetExchangeRules 获取兑换规则
func (h *PointsHandler) GetExchangeRules(c *gin.Context) {
	rules, err := h.pointsService.GetExchangeRules()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, rules)
}

// ExchangePoints 积分兑换
func (h *PointsHandler) ExchangePoints(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		RuleID uint64 `json:"rule_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择兑换规则")
		return
	}

	if err := h.pointsService.ExchangePoints(userID, req.RuleID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "兑换成功", nil)
}

// ========== 管理员接口 ==========

// AdminGetPointsStats 获取积分统计
func (h *PointsHandler) AdminGetPointsStats(c *gin.Context) {
	stats, err := h.pointsService.GetPointsStats()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// AdminAdjustPoints 管理员调整积分
func (h *PointsHandler) AdminAdjustPoints(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Points int    `json:"points" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	if req.Remark == "" {
		if req.Points > 0 {
			req.Remark = "管理员增加积分"
		} else {
			req.Remark = "管理员扣除积分"
		}
	}

	if err := h.pointsService.AdminAdjustPoints(userID, req.Points, req.Remark); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "调整成功", nil)
}

// AdminGetExchangeRules 获取所有兑换规则
func (h *PointsHandler) AdminGetExchangeRules(c *gin.Context) {
	rules, err := h.pointsService.GetAllExchangeRules()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, rules)
}

// AdminCreateExchangeRule 创建兑换规则
func (h *PointsHandler) AdminCreateExchangeRule(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Points      int    `json:"points" binding:"required,min=1"`
		MemberDays  int    `json:"member_days" binding:"required,min=1"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	rule := &models.PointsExchangeRule{
		Name:        req.Name,
		Points:      req.Points,
		MemberDays:  req.MemberDays,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Enabled:     true,
	}

	if err := h.pointsService.CreateExchangeRule(rule); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, rule)
}

// AdminUpdateExchangeRule 更新兑换规则
func (h *PointsHandler) AdminUpdateExchangeRule(c *gin.Context) {
	idStr := c.Param("id")
	var id uint64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Points      *int    `json:"points"`
		MemberDays  *int    `json:"member_days"`
		Description *string `json:"description"`
		Enabled     *bool   `json:"enabled"`
		SortOrder   *int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Points != nil {
		updates["points"] = *req.Points
	}
	if req.MemberDays != nil {
		updates["member_days"] = *req.MemberDays
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if err := h.pointsService.UpdateExchangeRule(id, updates); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// AdminDeleteExchangeRule 删除兑换规则
func (h *PointsHandler) AdminDeleteExchangeRule(c *gin.Context) {
	idStr := c.Param("id")
	var id uint64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	if err := h.pointsService.DeleteExchangeRule(id); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetPointsRanking 获取积分排行榜（分页）
func (h *PointsHandler) GetPointsRanking(c *gin.Context) {
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	ranking, total, err := h.pointsService.GetPointsRanking(page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, ranking, total, page, pageSize)
}

// GetMyPointsRank 获取我的积分排名
func (h *PointsHandler) GetMyPointsRank(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	rank, err := h.pointsService.GetUserPointsRank(userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"rank": rank})
}

// AdminGiftPointsToAll 给用户赠送积分
func (h *PointsHandler) AdminGiftPointsToAll(c *gin.Context) {
	var req struct {
		Points            int    `json:"points" binding:"required,min=1"`
		Remark            string `json:"remark"`
		TargetType        string `json:"target_type"`         // all, member, non_member, role
		MemberLevel       *int8  `json:"member_level"`        // 会员等级
		Role              *int8  `json:"role"`                // 角色
		SendNotification  bool   `json:"send_notification"`   // 发送站内通知
		NotificationTitle string `json:"notification_title"`  // 通知标题
		NotificationBody  string `json:"notification_body"`   // 通知内容
		SendEmail         bool   `json:"send_email"`          // 发送邮件
		EmailTitle        string `json:"email_title"`
		EmailBody         string `json:"email_body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.Remark == "" {
		req.Remark = "活动赠送积分"
	}
	if req.TargetType == "" {
		req.TargetType = "all"
	}

	// 构建请求
	giftReq := &service.GiftPointsRequest{
		Points:            req.Points,
		Remark:            req.Remark,
		TargetType:        req.TargetType,
		MemberLevel:       req.MemberLevel,
		Role:              req.Role,
		SendNotification:  req.SendNotification,
		NotificationTitle: req.NotificationTitle,
		NotificationBody:  req.NotificationBody,
		SendEmail:         req.SendEmail,
		EmailTitle:        req.EmailTitle,
		EmailBody:         req.EmailBody,
	}

	result, userIDs, err := h.pointsService.GiftPointsToUsers(giftReq)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	// 发送站内通知
	notificationSent := 0
	if req.SendNotification && len(userIDs) > 0 {
		title := req.NotificationTitle
		if title == "" {
			title = "🎁 积分到账通知"
		}
		body := req.NotificationBody
		if body == "" {
			body = fmt.Sprintf("恭喜您获得 %d 积分！\n原因：%s\n\n积分已到账，请前往积分中心查看。", req.Points, req.Remark)
		}
		// 批量创建通知
		notifications := make([]models.Notification, len(userIDs))
		for i, userID := range userIDs {
			notifications[i] = models.Notification{
				UserID:  userID,
				Title:   title,
				Content: body,
				Type:    3, // 活动类型
				IsRead:  false,
			}
		}
		if err := h.pointsService.CreateBatchNotifications(notifications); err == nil {
			notificationSent = len(userIDs)
		}
	}

	// 如果需要发送邮件通知
	emailSent := 0
	if req.SendEmail {
		emails, _ := h.pointsService.GetAllUserEmails()
		if len(emails) > 0 {
			emailSent = len(emails)
			// 异步发送邮件（实际发送逻辑需要邮件服务支持）
		}
	}

	response.Success(c, gin.H{
		"total_users":       result.TotalUsers,
		"success_count":     result.SuccessCount,
		"failed_count":      result.FailedCount,
		"notification_sent": notificationSent,
		"email_sent":        emailSent,
	})
}

// ========== 自动赠送规则管理 ==========

// AdminGetGiftRules 获取自动赠送规则列表
func (h *PointsHandler) AdminGetGiftRules(c *gin.Context) {
	rules, err := h.pointsService.GetGiftRules()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, rules)
}

// AdminCreateGiftRule 创建自动赠送规则
func (h *PointsHandler) AdminCreateGiftRule(c *gin.Context) {
	var req struct {
		Name              string `json:"name" binding:"required"`
		RuleType          int8   `json:"rule_type" binding:"required,min=1,max=5"`
		Points            int    `json:"points" binding:"required,min=1"`
		TargetType        string `json:"target_type"`
		MemberLevel       *int8  `json:"member_level"`
		ExecuteTime       string `json:"execute_time"`
		ExecuteDay        int    `json:"execute_day"`
		ExecuteMonth      int    `json:"execute_month"`
		SendNotification  bool   `json:"send_notification"`
		NotificationTitle string `json:"notification_title"`
		NotificationBody  string `json:"notification_body"`
		Enabled           bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	adminID, _ := middleware.GetUserID(c)

	if req.TargetType == "" {
		req.TargetType = "all"
	}
	if req.ExecuteTime == "" {
		req.ExecuteTime = "08:00"
	}

	rule := &models.PointsGiftRule{
		Name:              req.Name,
		RuleType:          req.RuleType,
		Points:            req.Points,
		TargetType:        req.TargetType,
		MemberLevel:       req.MemberLevel,
		ExecuteTime:       req.ExecuteTime,
		ExecuteDay:        req.ExecuteDay,
		ExecuteMonth:      req.ExecuteMonth,
		SendNotification:  req.SendNotification,
		NotificationTitle: req.NotificationTitle,
		NotificationBody:  req.NotificationBody,
		Enabled:           req.Enabled,
		CreatedBy:         adminID,
	}

	if err := h.pointsService.CreateGiftRule(rule); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, rule)
}

// AdminUpdateGiftRule 更新自动赠送规则
func (h *PointsHandler) AdminUpdateGiftRule(c *gin.Context) {
	idStr := c.Param("id")
	var id uint64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	var req struct {
		Name              *string `json:"name"`
		RuleType          *int8   `json:"rule_type"`
		Points            *int    `json:"points"`
		TargetType        *string `json:"target_type"`
		MemberLevel       *int8   `json:"member_level"`
		ExecuteTime       *string `json:"execute_time"`
		ExecuteDay        *int    `json:"execute_day"`
		ExecuteMonth      *int    `json:"execute_month"`
		SendNotification  *bool   `json:"send_notification"`
		NotificationTitle *string `json:"notification_title"`
		NotificationBody  *string `json:"notification_body"`
		Enabled           *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.RuleType != nil {
		updates["rule_type"] = *req.RuleType
	}
	if req.Points != nil {
		updates["points"] = *req.Points
	}
	if req.TargetType != nil {
		updates["target_type"] = *req.TargetType
	}
	if req.MemberLevel != nil {
		updates["member_level"] = *req.MemberLevel
	}
	if req.ExecuteTime != nil {
		updates["execute_time"] = *req.ExecuteTime
	}
	if req.ExecuteDay != nil {
		updates["execute_day"] = *req.ExecuteDay
	}
	if req.ExecuteMonth != nil {
		updates["execute_month"] = *req.ExecuteMonth
	}
	if req.SendNotification != nil {
		updates["send_notification"] = *req.SendNotification
	}
	if req.NotificationTitle != nil {
		updates["notification_title"] = *req.NotificationTitle
	}
	if req.NotificationBody != nil {
		updates["notification_body"] = *req.NotificationBody
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := h.pointsService.UpdateGiftRule(id, updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// AdminDeleteGiftRule 删除自动赠送规则
func (h *PointsHandler) AdminDeleteGiftRule(c *gin.Context) {
	idStr := c.Param("id")
	var id uint64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	if err := h.pointsService.DeleteGiftRule(id); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// AdminToggleGiftRule 切换规则启用状态
func (h *PointsHandler) AdminToggleGiftRule(c *gin.Context) {
	idStr := c.Param("id")
	var id uint64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	if err := h.pointsService.ToggleGiftRule(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "操作成功", nil)
}

// AdminExecuteGiftRule 手动执行赠送规则
func (h *PointsHandler) AdminExecuteGiftRule(c *gin.Context) {
	idStr := c.Param("id")
	var id uint64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	rule, err := h.pointsService.GetGiftRule(id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.pointsService.ExecuteGiftRule(rule)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"total_users":       result.TotalUsers,
		"success_count":     result.SuccessCount,
		"failed_count":      result.FailedCount,
		"notification_sent": result.NotificationSent,
	})
}

// AdminGetGiftLogs 获取赠送执行日志
func (h *PointsHandler) AdminGetGiftLogs(c *gin.Context) {
	var ruleID uint64
	if idStr := c.Query("rule_id"); idStr != "" {
		fmt.Sscanf(idStr, "%d", &ruleID)
	}

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	logs, total, err := h.pointsService.GetGiftLogs(ruleID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, logs, total, page, pageSize)
}
