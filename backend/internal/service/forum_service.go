// Package service 论坛服务
package service

import (
	"encoding/json"
	"errors"
	"time"

	"feiniu-user-system/internal/models"
	"feiniu-user-system/pkg/ipgeo"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ForumService 论坛服务
type ForumService struct {
	db *gorm.DB
}

// NewForumService 创建论坛服务
func NewForumService(db *gorm.DB) *ForumService {
	return &ForumService{db: db}
}

// ============= 节点管理 =============

// GetNodes 获取所有节点
func (s *ForumService) GetNodes() ([]models.ForumNode, error) {
	var nodes []models.ForumNode
	err := s.db.Where("status = ?", 1).Order("sort_order, id").Find(&nodes).Error
	return nodes, err
}

// GetAllNodes 获取所有节点（管理员）
func (s *ForumService) GetAllNodes() ([]models.ForumNode, error) {
	var nodes []models.ForumNode
	err := s.db.Order("sort_order, id").Find(&nodes).Error
	return nodes, err
}

// CreateNode 创建节点
func (s *ForumService) CreateNode(name, description, icon string, sortOrder int) (*models.ForumNode, error) {
	node := &models.ForumNode{
		Name:        name,
		Description: description,
		Icon:        icon,
		SortOrder:   sortOrder,
		Status:      1,
	}
	if err := s.db.Create(node).Error; err != nil {
		return nil, err
	}
	return node, nil
}

// UpdateNode 更新节点
func (s *ForumService) UpdateNode(id uint64, updates map[string]interface{}) error {
	return s.db.Model(&models.ForumNode{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteNode 删除节点
func (s *ForumService) DeleteNode(id uint64) error {
	// 检查是否有话题
	var count int64
	s.db.Model(&models.ForumTopic{}).Where("node_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该节点下有话题，无法删除")
	}
	return s.db.Delete(&models.ForumNode{}, id).Error
}

// ============= 发帖限制常量 =============
const (
	MinTitleLength      = 2     // 标题最小长度
	MaxTitleLength      = 128   // 标题最大长度
	MinContentLength    = 1     // 内容最小长度
	MaxContentLength    = 50000 // 内容最大长度
	MinCommentLength    = 1     // 评论最小长度
	MaxCommentLength    = 5000  // 评论最大长度
	MaxTopicsPerDay     = 20    // 每用户每天最多发帖数
	MaxCommentsPerHour  = 60    // 每用户每小时最多评论数
	NewUserPostWaitDays = 0     // 新用户注册后等待天数才能发帖（0表示不限制）
)

// ============= 话题管理 =============

// CheckPostPermission 检查发帖权限
func (s *ForumService) CheckPostPermission(userID uuid.UUID) error {
	// 检查用户状态
	var user models.User
	if err := s.db.Select("id, status, role, created_at").First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.Status != 1 {
		return errors.New("账号状态异常，无法发帖")
	}

	// 新用户限制（可选）
	if NewUserPostWaitDays > 0 {
		if time.Since(user.CreatedAt) < time.Duration(NewUserPostWaitDays)*24*time.Hour {
			return errors.New("新注册用户需等待一段时间才能发帖")
		}
	}

	// 检查今日发帖数（管理员不限制）
	if user.Role < 2 {
		var todayCount int64
		today := time.Now().Truncate(24 * time.Hour)
		s.db.Model(&models.ForumTopic{}).
			Where("user_id = ? AND created_at >= ? AND status = ?", userID, today, models.TopicStatusNormal).
			Count(&todayCount)
		if todayCount >= MaxTopicsPerDay {
			return errors.New("今日发帖数已达上限")
		}
	}

	return nil
}

// CheckCommentPermission 检查评论权限
func (s *ForumService) CheckCommentPermission(userID uuid.UUID) error {
	// 检查用户状态
	var user models.User
	if err := s.db.Select("id, status, role").First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.Status != 1 {
		return errors.New("账号状态异常，无法评论")
	}

	// 检查本小时评论数（管理员不限制）
	if user.Role < 2 {
		var hourCount int64
		hourAgo := time.Now().Add(-time.Hour)
		s.db.Model(&models.ForumComment{}).
			Where("user_id = ? AND created_at >= ? AND status = 0", userID, hourAgo).
			Count(&hourCount)
		if hourCount >= MaxCommentsPerHour {
			return errors.New("评论过于频繁，请稍后再试")
		}
	}

	return nil
}

// ValidateTopicContent 验证话题内容
func (s *ForumService) ValidateTopicContent(title, content string) error {
	titleLen := len([]rune(title))
	if titleLen < MinTitleLength {
		return errors.New("标题太短，至少需要2个字符")
	}
	if titleLen > MaxTitleLength {
		return errors.New("标题太长，最多128个字符")
	}

	contentLen := len([]rune(content))
	if contentLen < MinContentLength {
		return errors.New("内容不能为空")
	}
	if contentLen > MaxContentLength {
		return errors.New("内容太长，最多50000个字符")
	}

	return nil
}

// ValidateCommentContent 验证评论内容
func (s *ForumService) ValidateCommentContent(content string) error {
	contentLen := len([]rune(content))
	if contentLen < MinCommentLength {
		return errors.New("评论内容太短")
	}
	if contentLen > MaxCommentLength {
		return errors.New("评论内容太长，最多5000个字符")
	}

	return nil
}

// CreateTopic 创建话题
func (s *ForumService) CreateTopic(userID uuid.UUID, nodeID uint64, title, content, contentType string, images []string, ip string) (*models.ForumTopic, error) {
	// 检查发帖权限
	if err := s.CheckPostPermission(userID); err != nil {
		return nil, err
	}

	// 验证内容
	if err := s.ValidateTopicContent(title, content); err != nil {
		return nil, err
	}

	// 检查节点是否存在
	var node models.ForumNode
	if err := s.db.First(&node, nodeID).Error; err != nil {
		return nil, errors.New("节点不存在")
	}

	imagesJSON := "[]"
	if len(images) > 0 {
		b, _ := json.Marshal(images)
		imagesJSON = string(b)
	}

	topicType := models.TopicTypeNormal
	if len(images) > 0 {
		topicType = models.TopicTypeImage
	}

	// 解析IP地理位置
	location := ipgeo.GetInstance().GetLocation(ip)

	topic := &models.ForumTopic{
		NodeID:      nodeID,
		UserID:      userID,
		Title:       title,
		Content:     content,
		ContentType: contentType,
		Images:      imagesJSON,
		TopicType:   topicType,
		IP:          ip,
		Location:    location,
		Status:      models.TopicStatusNormal,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(topic).Error; err != nil {
			return err
		}
		// 更新节点话题数
		return tx.Model(&node).Update("topic_count", gorm.Expr("topic_count + 1")).Error
	})

	if err != nil {
		return nil, err
	}
	return topic, nil
}

// GetTopicList 获取话题列表
func (s *ForumService) GetTopicList(nodeID uint64, page, pageSize int, orderBy string, currentUserID *uuid.UUID) ([]models.ForumTopic, int64, error) {
	var topics []models.ForumTopic
	var total int64

	query := s.db.Model(&models.ForumTopic{}).Where("status = ?", models.TopicStatusNormal)
	if nodeID > 0 {
		query = query.Where("node_id = ?", nodeID)
	}

	query.Count(&total)

	// 排序
	switch orderBy {
	case "hot":
		query = query.Order("is_top DESC, comment_count DESC, created_at DESC")
	case "recommend":
		query = query.Where("is_recommend = ?", true).Order("is_top DESC, created_at DESC")
	default:
		query = query.Order("is_top DESC, created_at DESC")
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&topics).Error; err != nil {
		return nil, 0, err
	}

	// 填充用户信息和节点信息
	s.fillTopicUsers(topics)
	s.fillTopicNodes(topics)

	// 填充当前用户的点赞和收藏状态
	if currentUserID != nil {
		s.fillTopicLikeStatus(topics, *currentUserID)
	}

	return topics, total, nil
}

// GetTopicDetail 获取话题详情
func (s *ForumService) GetTopicDetail(id uint64, currentUserID *uuid.UUID, ip string) (*models.ForumTopic, error) {
	var topic models.ForumTopic
	if err := s.db.First(&topic, id).Error; err != nil {
		return nil, errors.New("话题不存在")
	}

	if topic.Status != models.TopicStatusNormal {
		return nil, errors.New("话题已删除")
	}

	// 增加浏览量（同一用户/IP 30分钟内只计一次）
	s.incrementViewCount(id, currentUserID, ip)

	// 填充用户和节点信息
	s.fillTopicUser(&topic)
	s.fillTopicNode(&topic)

	// 填充点赞和收藏状态
	if currentUserID != nil {
		var likeCount int64
		s.db.Model(&models.ForumLike{}).Where("user_id = ? AND entity_type = ? AND entity_id = ?", *currentUserID, "topic", id).Count(&likeCount)
		topic.IsLiked = likeCount > 0

		var favCount int64
		s.db.Model(&models.ForumFavorite{}).Where("user_id = ? AND topic_id = ?", *currentUserID, id).Count(&favCount)
		topic.IsFaved = favCount > 0
	}

	return &topic, nil
}

// incrementViewCount 增加浏览量（防刷：同一用户/IP 30分钟内只计一次）
func (s *ForumService) incrementViewCount(topicID uint64, userID *uuid.UUID, ip string) {
	now := time.Now()
	threshold := now.Add(-30 * time.Minute) // 30分钟内不重复计数

	var view models.TopicView
	var err error

	if userID != nil {
		// 登录用户：按用户ID查询
		err = s.db.Where("topic_id = ? AND user_id = ?", topicID, *userID).First(&view).Error
	} else if ip != "" {
		// 未登录用户：按IP查询
		emptyUUID := uuid.UUID{}
		err = s.db.Where("topic_id = ? AND user_id = ? AND ip = ?", topicID, emptyUUID, ip).First(&view).Error
	} else {
		// 无法识别用户，直接增加浏览量
		s.db.Model(&models.ForumTopic{}).Where("id = ?", topicID).Update("view_count", gorm.Expr("view_count + 1"))
		return
	}

	if err != nil {
		// 没有记录，创建新记录并增加浏览量
		newView := models.TopicView{
			TopicID:  topicID,
			IP:       ip,
			ViewedAt: now,
		}
		if userID != nil {
			newView.UserID = *userID
		}
		s.db.Create(&newView)
		s.db.Model(&models.ForumTopic{}).Where("id = ?", topicID).Update("view_count", gorm.Expr("view_count + 1"))
		return
	}

	// 有记录，检查是否超过阈值时间
	if view.ViewedAt.Before(threshold) {
		// 超过30分钟，更新时间并增加浏览量
		s.db.Model(&view).Update("viewed_at", now)
		s.db.Model(&models.ForumTopic{}).Where("id = ?", topicID).Update("view_count", gorm.Expr("view_count + 1"))
	}
	// 30分钟内，不增加浏览量
}

// UpdateTopic 更新话题
func (s *ForumService) UpdateTopic(id uint64, userID uuid.UUID, title, content string, images []string) error {
	var topic models.ForumTopic
	if err := s.db.First(&topic, id).Error; err != nil {
		return errors.New("话题不存在")
	}

	if topic.UserID != userID {
		return errors.New("无权修改")
	}

	updates := map[string]interface{}{
		"title":   title,
		"content": content,
	}

	if images != nil {
		b, _ := json.Marshal(images)
		updates["images"] = string(b)
	}

	return s.db.Model(&topic).Updates(updates).Error
}

// DeleteTopic 删除话题
func (s *ForumService) DeleteTopic(id uint64, userID uuid.UUID, isAdmin bool) error {
	var topic models.ForumTopic
	if err := s.db.First(&topic, id).Error; err != nil {
		return errors.New("话题不存在")
	}

	if !isAdmin && topic.UserID != userID {
		return errors.New("无权删除")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 软删除话题
		if err := tx.Model(&topic).Update("status", models.TopicStatusDeleted).Error; err != nil {
			return err
		}
		// 更新节点话题数
		return tx.Model(&models.ForumNode{}).Where("id = ?", topic.NodeID).Update("topic_count", gorm.Expr("topic_count - 1")).Error
	})
}

// ============= 评论管理 =============

// CreateComment 创建评论
func (s *ForumService) CreateComment(userID uuid.UUID, topicID uint64, content string, images []string, parentID uint64, replyToUser *uuid.UUID, ip string) (*models.ForumComment, error) {
	// 检查评论权限
	if err := s.CheckCommentPermission(userID); err != nil {
		return nil, err
	}

	// 验证评论内容
	if err := s.ValidateCommentContent(content); err != nil {
		return nil, err
	}

	// 检查话题是否存在
	var topic models.ForumTopic
	if err := s.db.First(&topic, topicID).Error; err != nil {
		return nil, errors.New("话题不存在")
	}

	imagesJSON := "[]"
	if len(images) > 0 {
		b, _ := json.Marshal(images)
		imagesJSON = string(b)
	}

	// 解析IP地理位置
	location := ipgeo.GetInstance().GetLocation(ip)

	comment := &models.ForumComment{
		TopicID:     topicID,
		UserID:      userID,
		Content:     content,
		Images:      imagesJSON,
		ParentID:    parentID,
		ReplyToUser: replyToUser,
		IP:          ip,
		Location:    location,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// 更新话题评论数和最后评论时间
		now := time.Now()
		if err := tx.Model(&topic).Updates(map[string]interface{}{
			"comment_count":     gorm.Expr("comment_count + 1"),
			"last_comment_time": now,
			"last_comment_user": userID,
		}).Error; err != nil {
			return err
		}

		// 如果是回复，更新父评论的回复数
		if parentID > 0 {
			if err := tx.Model(&models.ForumComment{}).Where("id = ?", parentID).Update("reply_count", gorm.Expr("reply_count + 1")).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 填充用户信息
	s.fillCommentUser(comment)

	return comment, nil
}

// GetCommentList 获取评论列表
func (s *ForumService) GetCommentList(topicID uint64, page, pageSize int, currentUserID *uuid.UUID) ([]models.ForumComment, int64, error) {
	var comments []models.ForumComment
	var total int64

	// 只获取一级评论
	query := s.db.Model(&models.ForumComment{}).Where("topic_id = ? AND parent_id = 0 AND status = 0", topicID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	// 填充用户信息
	s.fillCommentUsers(comments)

	// 获取每个评论的回复（最多3条）
	for i := range comments {
		var replies []models.ForumComment
		s.db.Where("parent_id = ? AND status = 0", comments[i].ID).Order("created_at ASC").Limit(3).Find(&replies)
		s.fillCommentUsers(replies)
		s.fillReplyToNames(replies)
		comments[i].Replies = replies
	}

	// 填充点赞状态
	if currentUserID != nil {
		s.fillCommentLikeStatus(comments, *currentUserID)
	}

	return comments, total, nil
}

// GetCommentReplies 获取评论的回复
func (s *ForumService) GetCommentReplies(commentID uint64, page, pageSize int) ([]models.ForumComment, int64, error) {
	var replies []models.ForumComment
	var total int64

	query := s.db.Model(&models.ForumComment{}).Where("parent_id = ? AND status = 0", commentID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&replies).Error; err != nil {
		return nil, 0, err
	}

	s.fillCommentUsers(replies)
	s.fillReplyToNames(replies)

	return replies, total, nil
}

// DeleteComment 删除评论
func (s *ForumService) DeleteComment(id uint64, userID uuid.UUID, isAdmin bool) error {
	var comment models.ForumComment
	if err := s.db.First(&comment, id).Error; err != nil {
		return errors.New("评论不存在")
	}

	if !isAdmin && comment.UserID != userID {
		return errors.New("无权删除")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 软删除评论
		if err := tx.Model(&comment).Update("status", 1).Error; err != nil {
			return err
		}

		// 更新话题评论数
		if err := tx.Model(&models.ForumTopic{}).Where("id = ?", comment.TopicID).Update("comment_count", gorm.Expr("comment_count - 1")).Error; err != nil {
			return err
		}

		// 如果是回复，更新父评论的回复数
		if comment.ParentID > 0 {
			if err := tx.Model(&models.ForumComment{}).Where("id = ?", comment.ParentID).Update("reply_count", gorm.Expr("reply_count - 1")).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// ============= 点赞和收藏 =============

// LikeTopic 点赞话题
func (s *ForumService) LikeTopic(userID uuid.UUID, topicID uint64) (bool, error) {
	var existing models.ForumLike
	err := s.db.Where("user_id = ? AND entity_type = ? AND entity_id = ?", userID, "topic", topicID).First(&existing).Error

	if err == nil {
		// 已点赞，取消
		s.db.Delete(&existing)
		s.db.Model(&models.ForumTopic{}).Where("id = ?", topicID).Update("like_count", gorm.Expr("like_count - 1"))
		return false, nil
	}

	// 未点赞，添加
	like := &models.ForumLike{
		UserID:     userID,
		EntityType: "topic",
		EntityID:   topicID,
	}
	if err := s.db.Create(like).Error; err != nil {
		return false, err
	}
	s.db.Model(&models.ForumTopic{}).Where("id = ?", topicID).Update("like_count", gorm.Expr("like_count + 1"))
	return true, nil
}

// LikeComment 点赞评论
func (s *ForumService) LikeComment(userID uuid.UUID, commentID uint64) (bool, error) {
	var existing models.ForumLike
	err := s.db.Where("user_id = ? AND entity_type = ? AND entity_id = ?", userID, "comment", commentID).First(&existing).Error

	if err == nil {
		// 已点赞，取消
		s.db.Delete(&existing)
		s.db.Model(&models.ForumComment{}).Where("id = ?", commentID).Update("like_count", gorm.Expr("like_count - 1"))
		return false, nil
	}

	// 未点赞，添加
	like := &models.ForumLike{
		UserID:     userID,
		EntityType: "comment",
		EntityID:   commentID,
	}
	if err := s.db.Create(like).Error; err != nil {
		return false, err
	}
	s.db.Model(&models.ForumComment{}).Where("id = ?", commentID).Update("like_count", gorm.Expr("like_count + 1"))
	return true, nil
}

// FavoriteTopic 收藏话题
func (s *ForumService) FavoriteTopic(userID uuid.UUID, topicID uint64) (bool, error) {
	var existing models.ForumFavorite
	err := s.db.Where("user_id = ? AND topic_id = ?", userID, topicID).First(&existing).Error

	if err == nil {
		// 已收藏，取消
		s.db.Delete(&existing)
		s.db.Model(&models.ForumTopic{}).Where("id = ?", topicID).Update("favorite_count", gorm.Expr("favorite_count - 1"))
		return false, nil
	}

	// 未收藏，添加
	fav := &models.ForumFavorite{
		UserID:  userID,
		TopicID: topicID,
	}
	if err := s.db.Create(fav).Error; err != nil {
		return false, err
	}
	s.db.Model(&models.ForumTopic{}).Where("id = ?", topicID).Update("favorite_count", gorm.Expr("favorite_count + 1"))
	return true, nil
}

// GetMyFavorites 获取我的收藏
func (s *ForumService) GetMyFavorites(userID uuid.UUID, page, pageSize int) ([]models.ForumTopic, int64, error) {
	var favorites []models.ForumFavorite
	var total int64

	query := s.db.Model(&models.ForumFavorite{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&favorites).Error; err != nil {
		return nil, 0, err
	}

	// 获取话题
	topicIDs := make([]uint64, len(favorites))
	for i, f := range favorites {
		topicIDs[i] = f.TopicID
	}

	var topics []models.ForumTopic
	if len(topicIDs) > 0 {
		s.db.Where("id IN ?", topicIDs).Find(&topics)
		s.fillTopicUsers(topics)
	}

	return topics, total, nil
}

// GetMyTopics 获取我的话题
func (s *ForumService) GetMyTopics(userID uuid.UUID, page, pageSize int) ([]models.ForumTopic, int64, error) {
	var topics []models.ForumTopic
	var total int64

	query := s.db.Model(&models.ForumTopic{}).Where("user_id = ? AND status = ?", userID, models.TopicStatusNormal)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&topics).Error; err != nil {
		return nil, 0, err
	}

	s.fillTopicNodes(topics)
	return topics, total, nil
}

// ============= 辅助方法 =============

func (s *ForumService) fillTopicUser(topic *models.ForumTopic) {
	var user models.User
	if err := s.db.Select("id, username, nickname, avatar").First(&user, topic.UserID).Error; err == nil {
		topic.User = &user
	}
}

func (s *ForumService) fillTopicUsers(topics []models.ForumTopic) {
	if len(topics) == 0 {
		return
	}
	userIDs := make([]uuid.UUID, len(topics))
	for i, t := range topics {
		userIDs[i] = t.UserID
	}

	var users []models.User
	s.db.Select("id, username, nickname, avatar").Where("id IN ?", userIDs).Find(&users)

	userMap := make(map[uuid.UUID]*models.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	for i := range topics {
		if u, ok := userMap[topics[i].UserID]; ok {
			topics[i].User = u
		}
	}
}

func (s *ForumService) fillTopicNode(topic *models.ForumTopic) {
	var node models.ForumNode
	if err := s.db.First(&node, topic.NodeID).Error; err == nil {
		topic.Node = &node
	}
}

func (s *ForumService) fillTopicNodes(topics []models.ForumTopic) {
	if len(topics) == 0 {
		return
	}
	nodeIDs := make([]uint64, len(topics))
	for i, t := range topics {
		nodeIDs[i] = t.NodeID
	}

	var nodes []models.ForumNode
	s.db.Where("id IN ?", nodeIDs).Find(&nodes)

	nodeMap := make(map[uint64]*models.ForumNode)
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	for i := range topics {
		if n, ok := nodeMap[topics[i].NodeID]; ok {
			topics[i].Node = n
		}
	}
}

func (s *ForumService) fillTopicLikeStatus(topics []models.ForumTopic, userID uuid.UUID) {
	if len(topics) == 0 {
		return
	}
	topicIDs := make([]uint64, len(topics))
	for i, t := range topics {
		topicIDs[i] = t.ID
	}

	var likes []models.ForumLike
	s.db.Where("user_id = ? AND entity_type = ? AND entity_id IN ?", userID, "topic", topicIDs).Find(&likes)

	likeMap := make(map[uint64]bool)
	for _, l := range likes {
		likeMap[l.EntityID] = true
	}

	var favs []models.ForumFavorite
	s.db.Where("user_id = ? AND topic_id IN ?", userID, topicIDs).Find(&favs)

	favMap := make(map[uint64]bool)
	for _, f := range favs {
		favMap[f.TopicID] = true
	}

	for i := range topics {
		topics[i].IsLiked = likeMap[topics[i].ID]
		topics[i].IsFaved = favMap[topics[i].ID]
	}
}

func (s *ForumService) fillCommentUser(comment *models.ForumComment) {
	var user models.User
	if err := s.db.Select("id, username, nickname, avatar").First(&user, comment.UserID).Error; err == nil {
		comment.User = &user
	}
}

func (s *ForumService) fillCommentUsers(comments []models.ForumComment) {
	if len(comments) == 0 {
		return
	}
	userIDs := make([]uuid.UUID, len(comments))
	for i, c := range comments {
		userIDs[i] = c.UserID
	}

	var users []models.User
	s.db.Select("id, username, nickname, avatar").Where("id IN ?", userIDs).Find(&users)

	userMap := make(map[uuid.UUID]*models.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	for i := range comments {
		if u, ok := userMap[comments[i].UserID]; ok {
			comments[i].User = u
		}
	}
}

func (s *ForumService) fillReplyToNames(comments []models.ForumComment) {
	for i := range comments {
		if comments[i].ReplyToUser != nil {
			var user models.User
			if err := s.db.Select("nickname, username").First(&user, *comments[i].ReplyToUser).Error; err == nil {
				name := user.Nickname
				if name == "" {
					name = user.Username
				}
				comments[i].ReplyToName = name
			}
		}
	}
}

func (s *ForumService) fillCommentLikeStatus(comments []models.ForumComment, userID uuid.UUID) {
	if len(comments) == 0 {
		return
	}
	commentIDs := make([]uint64, len(comments))
	for i, c := range comments {
		commentIDs[i] = c.ID
	}

	var likes []models.ForumLike
	s.db.Where("user_id = ? AND entity_type = ? AND entity_id IN ?", userID, "comment", commentIDs).Find(&likes)

	likeMap := make(map[uint64]bool)
	for _, l := range likes {
		likeMap[l.EntityID] = true
	}

	for i := range comments {
		comments[i].IsLiked = likeMap[comments[i].ID]
	}
}

// ============= 管理员功能 =============

// AdminSetTopicTop 设置话题置顶
func (s *ForumService) AdminSetTopicTop(id uint64, isTop bool) error {
	return s.db.Model(&models.ForumTopic{}).Where("id = ?", id).Update("is_top", isTop).Error
}

// AdminSetTopicRecommend 设置话题推荐
func (s *ForumService) AdminSetTopicRecommend(id uint64, isRecommend bool) error {
	return s.db.Model(&models.ForumTopic{}).Where("id = ?", id).Update("is_recommend", isRecommend).Error
}

// AdminGetTopicList 管理员获取话题列表
func (s *ForumService) AdminGetTopicList(nodeID uint64, status int8, keyword string, page, pageSize int) ([]models.ForumTopic, int64, error) {
	var topics []models.ForumTopic
	var total int64

	query := s.db.Model(&models.ForumTopic{})
	if nodeID > 0 {
		query = query.Where("node_id = ?", nodeID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&topics).Error; err != nil {
		return nil, 0, err
	}

	s.fillTopicUsers(topics)
	s.fillTopicNodes(topics)

	return topics, total, nil
}
