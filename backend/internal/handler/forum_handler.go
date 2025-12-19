// Package handler 论坛处理器
package handler

import (
	"net/http"
	"strconv"

	"feiniu-user-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ForumHandler 论坛处理器
type ForumHandler struct {
	forumService *service.ForumService
}

// NewForumHandler 创建论坛处理器
func NewForumHandler(forumService *service.ForumService) *ForumHandler {
	return &ForumHandler{forumService: forumService}
}

// GetNodes 获取节点列表
func (h *ForumHandler) GetNodes(c *gin.Context) {
	nodes, err := h.forumService.GetNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取节点失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nodes})
}

// CreateTopic 创建话题
func (h *ForumHandler) CreateTopic(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		NodeID      uint64   `json:"node_id" binding:"required"`
		Title       string   `json:"title" binding:"required,max=128"`
		Content     string   `json:"content" binding:"required"`
		ContentType string   `json:"content_type"`
		Images      []string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if req.ContentType == "" {
		req.ContentType = "html"
	}

	topic, err := h.forumService.CreateTopic(userID, req.NodeID, req.Title, req.Content, req.ContentType, req.Images, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "发布成功", "data": topic})
}

// GetTopicList 获取话题列表
func (h *ForumHandler) GetTopicList(c *gin.Context) {
	nodeID, _ := strconv.ParseUint(c.Query("node_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	orderBy := c.DefaultQuery("order_by", "latest")

	if pageSize > 50 {
		pageSize = 50
	}

	var currentUserID *uuid.UUID
	if uid, exists := c.Get("userID"); exists {
		id := uid.(uuid.UUID)
		currentUserID = &id
	}

	topics, total, err := h.forumService.GetTopicList(nodeID, page, pageSize, orderBy, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取话题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  topics,
			"total": total,
		},
	})
}

// GetTopicDetail 获取话题详情
func (h *ForumHandler) GetTopicDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var currentUserID *uuid.UUID
	if uid, exists := c.Get("userID"); exists {
		id := uid.(uuid.UUID)
		currentUserID = &id
	}

	topic, err := h.forumService.GetTopicDetail(id, currentUserID, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": topic})
}

// UpdateTopic 更新话题
func (h *ForumHandler) UpdateTopic(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req struct {
		Title   string   `json:"title" binding:"required,max=128"`
		Content string   `json:"content" binding:"required"`
		Images  []string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.forumService.UpdateTopic(id, userID, req.Title, req.Content, req.Images); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

// DeleteTopic 删除话题
func (h *ForumHandler) DeleteTopic(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.forumService.DeleteTopic(id, userID, false); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// CreateComment 创建评论
func (h *ForumHandler) CreateComment(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		TopicID     uint64    `json:"topic_id" binding:"required"`
		Content     string    `json:"content" binding:"required"`
		Images      []string  `json:"images"`
		ParentID    uint64    `json:"parent_id"`
		ReplyToUser *string   `json:"reply_to_user"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var replyToUser *uuid.UUID
	if req.ReplyToUser != nil && *req.ReplyToUser != "" {
		id, err := uuid.Parse(*req.ReplyToUser)
		if err == nil {
			replyToUser = &id
		}
	}

	comment, err := h.forumService.CreateComment(userID, req.TopicID, req.Content, req.Images, req.ParentID, replyToUser, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "评论成功", "data": comment})
}

// GetCommentList 获取评论列表
func (h *ForumHandler) GetCommentList(c *gin.Context) {
	topicID, err := strconv.ParseUint(c.Query("topic_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if pageSize > 50 {
		pageSize = 50
	}

	var currentUserID *uuid.UUID
	if uid, exists := c.Get("userID"); exists {
		id := uid.(uuid.UUID)
		currentUserID = &id
	}

	comments, total, err := h.forumService.GetCommentList(topicID, page, pageSize, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取评论失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  comments,
			"total": total,
		},
	})
}

// GetCommentReplies 获取评论回复
func (h *ForumHandler) GetCommentReplies(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	replies, total, err := h.forumService.GetCommentReplies(commentID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取回复失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  replies,
			"total": total,
		},
	})
}

// DeleteComment 删除评论
func (h *ForumHandler) DeleteComment(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.forumService.DeleteComment(id, userID, false); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// LikeTopic 点赞话题
func (h *ForumHandler) LikeTopic(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	liked, err := h.forumService.LikeTopic(userID, topicID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	msg := "点赞成功"
	if !liked {
		msg = "取消点赞"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg, "data": gin.H{"liked": liked}})
}

// LikeComment 点赞评论
func (h *ForumHandler) LikeComment(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	liked, err := h.forumService.LikeComment(userID, commentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	msg := "点赞成功"
	if !liked {
		msg = "取消点赞"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg, "data": gin.H{"liked": liked}})
}

// FavoriteTopic 收藏话题
func (h *ForumHandler) FavoriteTopic(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	faved, err := h.forumService.FavoriteTopic(userID, topicID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	msg := "收藏成功"
	if !faved {
		msg = "取消收藏"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg, "data": gin.H{"favorited": faved}})
}

// GetMyFavorites 获取我的收藏
func (h *ForumHandler) GetMyFavorites(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	topics, total, err := h.forumService.GetMyFavorites(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取收藏失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  topics,
			"total": total,
		},
	})
}

// GetMyTopics 获取我的话题
func (h *ForumHandler) GetMyTopics(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	topics, total, err := h.forumService.GetMyTopics(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取话题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  topics,
			"total": total,
		},
	})
}

// ============= 管理员接口 =============

// AdminGetNodes 管理员获取所有节点
func (h *ForumHandler) AdminGetNodes(c *gin.Context) {
	nodes, err := h.forumService.GetAllNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取节点失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nodes})
}

// AdminCreateNode 管理员创建节点
func (h *ForumHandler) AdminCreateNode(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,max=32"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		SortOrder   int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	node, err := h.forumService.CreateNode(req.Name, req.Description, req.Icon, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "创建成功", "data": node})
}

// AdminUpdateNode 管理员更新节点
func (h *ForumHandler) AdminUpdateNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		SortOrder   *int   `json:"sort_order"`
		Status      *int8  `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.forumService.UpdateNode(id, updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

// AdminDeleteNode 管理员删除节点
func (h *ForumHandler) AdminDeleteNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.forumService.DeleteNode(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// AdminGetTopicList 管理员获取话题列表
func (h *ForumHandler) AdminGetTopicList(c *gin.Context) {
	nodeID, _ := strconv.ParseUint(c.Query("node_id"), 10, 64)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	topics, total, err := h.forumService.AdminGetTopicList(nodeID, int8(status), keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取话题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  topics,
			"total": total,
		},
	})
}

// AdminDeleteTopic 管理员删除话题
func (h *ForumHandler) AdminDeleteTopic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.forumService.DeleteTopic(id, uuid.Nil, true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// AdminSetTopicTop 管理员设置话题置顶
func (h *ForumHandler) AdminSetTopicTop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req struct {
		IsTop bool `json:"is_top"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.forumService.AdminSetTopicTop(id, req.IsTop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "设置成功"})
}

// AdminSetTopicRecommend 管理员设置话题推荐
func (h *ForumHandler) AdminSetTopicRecommend(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req struct {
		IsRecommend bool `json:"is_recommend"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.forumService.AdminSetTopicRecommend(id, req.IsRecommend); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "设置成功"})
}

// AdminDeleteComment 管理员删除评论
func (h *ForumHandler) AdminDeleteComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.forumService.DeleteComment(id, uuid.Nil, true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}
