// Package service 私信服务
package service

import (
	"errors"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageService 私信服务
type MessageService struct {
	db *gorm.DB
}

// NewMessageService 创建私信服务
func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{db: db}
}

// CanSendMessage 检查是否可以发送私信
// 规则：1. 检查黑名单 2. 管理员可以私信任何人 3. 已有会话可以继续私信 4. 互相关注或单向关注可以私信
func (s *MessageService) CanSendMessage(fromUserID, toUserID uuid.UUID) (bool, string) {
	if fromUserID == toUserID {
		return false, "不能给自己发私信"
	}

	// 检查黑名单（双向）
	if blocked, reason := s.IsBlocked(fromUserID, toUserID); blocked {
		return false, reason
	}

	// 检查发送者是否是管理员
	var fromUser models.User
	if err := s.db.Select("role").First(&fromUser, fromUserID).Error; err == nil {
		if fromUser.Role >= 2 { // 管理员
			return true, ""
		}
	}

	// 检查是否已有会话（已经聊过天的可以继续聊）
	user1ID, user2ID := fromUserID, toUserID
	if user1ID.String() > user2ID.String() {
		user1ID, user2ID = user2ID, user1ID
	}
	var conv models.Conversation
	if err := s.db.Where("user1_id = ? AND user2_id = ?", user1ID, user2ID).First(&conv).Error; err == nil {
		return true, ""
	}

	// 检查对方是否关注了你（对方关注你，你可以回复）
	if s.IsFollowing(toUserID, fromUserID) {
		return true, ""
	}

	// 检查你是否关注了对方（你关注对方，可以发起私信）
	if s.IsFollowing(fromUserID, toUserID) {
		return true, ""
	}

	return false, "只能向关注你的人或你关注的人发送私信"
}

// SendMessage 发送私信
func (s *MessageService) SendMessage(fromUserID, toUserID uuid.UUID, content string, images []string) (*models.PrivateMessage, error) {
	// 检查私信权限
	canSend, reason := s.CanSendMessage(fromUserID, toUserID)
	if !canSend {
		return nil, errors.New(reason)
	}

	// 检查接收者是否存在
	var toUser models.User
	if err := s.db.First(&toUser, toUserID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	imagesJSON := "[]"
	if len(images) > 0 {
		// 简单处理，实际应该用json.Marshal
		imagesJSON = "[]"
	}

	msg := &models.PrivateMessage{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Content:    content,
		Images:     imagesJSON,
		Status:     models.MessageStatusUnread,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			return err
		}

		// 更新或创建会话
		return s.updateConversation(tx, fromUserID, toUserID, msg.ID)
	})

	if err != nil {
		return nil, err
	}

	// 填充发送者信息
	var fromUser models.User
	s.db.Select("id, username, nickname, avatar").First(&fromUser, fromUserID)
	msg.FromUser = &fromUser

	return msg, nil
}

// updateConversation 更新会话
// fromUserID: 发送者ID, toUserID: 接收者ID
func (s *MessageService) updateConversation(tx *gorm.DB, fromUserID, toUserID uuid.UUID, messageID uint64) error {
	// 确保存储时 user1ID < user2ID
	user1ID, user2ID := fromUserID, toUserID
	if user1ID.String() > user2ID.String() {
		user1ID, user2ID = user2ID, user1ID
	}

	var conv models.Conversation
	err := tx.Where("user1_id = ? AND user2_id = ?", user1ID, user2ID).First(&conv).Error

	now := time.Now()
	if err != nil {
		// 创建新会话
		conv = models.Conversation{
			User1ID:       user1ID,
			User2ID:       user2ID,
			LastMessageID: messageID,
			LastMessageAt: now,
		}
		// 设置未读数（接收者+1）
		// 接收者是 toUserID，判断 toUserID 是 user1 还是 user2
		if toUserID == user1ID {
			conv.User1Unread = 1
		} else {
			conv.User2Unread = 1
		}
		return tx.Create(&conv).Error
	}

	// 更新会话
	updates := map[string]interface{}{
		"last_message_id": messageID,
		"last_message_at": now,
	}

	// 增加接收者未读数
	// 接收者是 toUserID，判断 toUserID 是 user1 还是 user2
	if toUserID == conv.User1ID {
		updates["user1_unread"] = gorm.Expr("user1_unread + 1")
	} else {
		updates["user2_unread"] = gorm.Expr("user2_unread + 1")
	}

	return tx.Model(&conv).Updates(updates).Error
}

// GetConversations 获取会话列表
func (s *MessageService) GetConversations(userID uuid.UUID, page, pageSize int) ([]models.Conversation, int64, error) {
	var conversations []models.Conversation
	var total int64

	query := s.db.Model(&models.Conversation{}).Where("user1_id = ? OR user2_id = ?", userID, userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("last_message_at DESC").Offset(offset).Limit(pageSize).Find(&conversations).Error; err != nil {
		return nil, 0, err
	}

	// 填充对方用户信息和最后一条消息
	for i := range conversations {
		// 确定对方用户ID
		otherUserID := conversations[i].User1ID
		if otherUserID == userID {
			otherUserID = conversations[i].User2ID
		}

		// 获取对方用户信息
		var otherUser models.User
		if err := s.db.Select("id, username, nickname, avatar").First(&otherUser, otherUserID).Error; err == nil {
			conversations[i].OtherUser = &otherUser
		}

		// 获取最后一条消息
		var lastMsg models.PrivateMessage
		if err := s.db.First(&lastMsg, conversations[i].LastMessageID).Error; err == nil {
			conversations[i].LastMessage = &lastMsg
		}

		// 设置未读数和静音状态
		if conversations[i].User1ID == userID {
			conversations[i].UnreadCount = conversations[i].User1Unread
			conversations[i].IsMuted = conversations[i].User1Muted
		} else {
			conversations[i].UnreadCount = conversations[i].User2Unread
			conversations[i].IsMuted = conversations[i].User2Muted
		}
	}

	return conversations, total, nil
}

// GetMessages 获取与某用户的消息列表
func (s *MessageService) GetMessages(userID, otherUserID uuid.UUID, page, pageSize int) ([]models.PrivateMessage, int64, error) {
	var messages []models.PrivateMessage
	var total int64

	query := s.db.Model(&models.PrivateMessage{}).Where(
		"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		userID, otherUserID, otherUserID, userID,
	)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	// 填充用户信息
	s.fillMessageUsers(messages)

	// 标记为已读
	s.MarkAsRead(userID, otherUserID)

	return messages, total, nil
}

// MarkAsRead 标记消息为已读
func (s *MessageService) MarkAsRead(userID, otherUserID uuid.UUID) error {
	// 更新消息状态
	now := time.Now()
	s.db.Model(&models.PrivateMessage{}).
		Where("from_user_id = ? AND to_user_id = ? AND status = ?", otherUserID, userID, models.MessageStatusUnread).
		Updates(map[string]interface{}{
			"status":  models.MessageStatusRead,
			"read_at": now,
		})

	// 更新会话未读数
	user1ID, user2ID := userID, otherUserID
	if user1ID.String() > user2ID.String() {
		user1ID, user2ID = user2ID, user1ID
	}

	var conv models.Conversation
	if err := s.db.Where("user1_id = ? AND user2_id = ?", user1ID, user2ID).First(&conv).Error; err == nil {
		if conv.User1ID == userID {
			s.db.Model(&conv).Update("user1_unread", 0)
		} else {
			s.db.Model(&conv).Update("user2_unread", 0)
		}
	}

	return nil
}

// GetUnreadCount 获取未读消息数
func (s *MessageService) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&models.PrivateMessage{}).
		Where("to_user_id = ? AND status = ?", userID, models.MessageStatusUnread).
		Count(&count).Error
	return count, err
}

// DeleteMessage 删除消息
func (s *MessageService) DeleteMessage(userID uuid.UUID, messageID uint64) error {
	var msg models.PrivateMessage
	if err := s.db.First(&msg, messageID).Error; err != nil {
		return errors.New("消息不存在")
	}

	if msg.FromUserID != userID && msg.ToUserID != userID {
		return errors.New("无权删除")
	}

	return s.db.Delete(&msg).Error
}

// fillMessageUsers 填充消息用户信息
func (s *MessageService) fillMessageUsers(messages []models.PrivateMessage) {
	if len(messages) == 0 {
		return
	}

	userIDs := make(map[uuid.UUID]bool)
	for _, m := range messages {
		userIDs[m.FromUserID] = true
		userIDs[m.ToUserID] = true
	}

	ids := make([]uuid.UUID, 0, len(userIDs))
	for id := range userIDs {
		ids = append(ids, id)
	}

	var users []models.User
	s.db.Select("id, username, nickname, avatar").Where("id IN ?", ids).Find(&users)

	userMap := make(map[uuid.UUID]*models.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	for i := range messages {
		if u, ok := userMap[messages[i].FromUserID]; ok {
			messages[i].FromUser = u
		}
		if u, ok := userMap[messages[i].ToUserID]; ok {
			messages[i].ToUser = u
		}
	}
}

// ============= 用户关注 =============

// FollowUser 关注用户
func (s *MessageService) FollowUser(userID, followID uuid.UUID) (bool, error) {
	if userID == followID {
		return false, errors.New("不能关注自己")
	}

	// 检查被关注者是否存在
	var user models.User
	if err := s.db.First(&user, followID).Error; err != nil {
		return false, errors.New("用户不存在")
	}

	var existing models.UserFollow
	err := s.db.Where("user_id = ? AND follow_id = ?", userID, followID).First(&existing).Error

	if err == nil {
		// 已关注，取消
		s.db.Delete(&existing)
		return false, nil
	}

	// 未关注，添加
	follow := &models.UserFollow{
		UserID:   userID,
		FollowID: followID,
	}
	if err := s.db.Create(follow).Error; err != nil {
		return false, err
	}
	return true, nil
}

// IsFollowing 是否已关注
func (s *MessageService) IsFollowing(userID, followID uuid.UUID) bool {
	var count int64
	s.db.Model(&models.UserFollow{}).Where("user_id = ? AND follow_id = ?", userID, followID).Count(&count)
	return count > 0
}

// GetFollowings 获取关注列表
func (s *MessageService) GetFollowings(userID uuid.UUID, page, pageSize int) ([]models.User, int64, error) {
	var follows []models.UserFollow
	var total int64

	query := s.db.Model(&models.UserFollow{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	followIDs := make([]uuid.UUID, len(follows))
	for i, f := range follows {
		followIDs[i] = f.FollowID
	}

	var users []models.User
	if len(followIDs) > 0 {
		s.db.Select("id, username, nickname, avatar, bio").Where("id IN ?", followIDs).Find(&users)
	}

	return users, total, nil
}

// GetFollowers 获取粉丝列表
func (s *MessageService) GetFollowers(userID uuid.UUID, page, pageSize int) ([]models.User, int64, error) {
	var follows []models.UserFollow
	var total int64

	query := s.db.Model(&models.UserFollow{}).Where("follow_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	followerIDs := make([]uuid.UUID, len(follows))
	for i, f := range follows {
		followerIDs[i] = f.UserID
	}

	var users []models.User
	if len(followerIDs) > 0 {
		s.db.Select("id, username, nickname, avatar, bio").Where("id IN ?", followerIDs).Find(&users)
	}

	return users, total, nil
}

// GetFollowStats 获取关注统计
func (s *MessageService) GetFollowStats(userID uuid.UUID) (followings, followers int64) {
	s.db.Model(&models.UserFollow{}).Where("user_id = ?", userID).Count(&followings)
	s.db.Model(&models.UserFollow{}).Where("follow_id = ?", userID).Count(&followers)
	return
}

// SearchUsers 搜索用户（用于发起私信）
func (s *MessageService) SearchUsers(currentUserID uuid.UUID, keyword string, limit int) ([]models.User, error) {
	var users []models.User
	if keyword == "" {
		return users, nil
	}
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	// 搜索用户名或昵称，排除自己和禁用用户
	err := s.db.Select("id, username, nickname, avatar").
		Where("id != ?", currentUserID).
		Where("status != ?", 2). // 排除禁用用户（status=2）
		Where("username ILIKE ? OR nickname ILIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Limit(limit).
		Find(&users).Error

	return users, err
}

// ============= 消息撤回 =============

// RecallMessage 撤回消息（5分钟内可撤回）
func (s *MessageService) RecallMessage(userID uuid.UUID, messageID uint64) error {
	var msg models.PrivateMessage
	if err := s.db.First(&msg, messageID).Error; err != nil {
		return errors.New("消息不存在")
	}

	// 只能撤回自己发送的消息
	if msg.FromUserID != userID {
		return errors.New("只能撤回自己发送的消息")
	}

	// 检查是否已撤回
	if msg.Status == models.MessageStatusRecalled {
		return errors.New("消息已撤回")
	}

	// 检查时间限制（5分钟内）
	if time.Since(msg.CreatedAt) > time.Duration(models.MessageRecallTimeLimit)*time.Minute {
		return errors.New("超过撤回时间限制")
	}

	// 更新消息状态和内容
	return s.db.Model(&msg).Updates(map[string]interface{}{
		"status":  models.MessageStatusRecalled,
		"content": "[消息已撤回]",
	}).Error
}

// ============= 黑名单功能 =============

// BlockUser 拉黑用户
func (s *MessageService) BlockUser(userID, blockedID uuid.UUID, reason string) error {
	if userID == blockedID {
		return errors.New("不能拉黑自己")
	}

	// 检查被拉黑用户是否存在
	var user models.User
	if err := s.db.First(&user, blockedID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 检查是否已拉黑
	var existing models.UserBlacklist
	if err := s.db.Where("user_id = ? AND blocked_id = ?", userID, blockedID).First(&existing).Error; err == nil {
		return errors.New("已在黑名单中")
	}

	// 添加到黑名单
	blacklist := &models.UserBlacklist{
		UserID:    userID,
		BlockedID: blockedID,
		Reason:    reason,
	}
	return s.db.Create(blacklist).Error
}

// UnblockUser 取消拉黑
func (s *MessageService) UnblockUser(userID, blockedID uuid.UUID) error {
	result := s.db.Where("user_id = ? AND blocked_id = ?", userID, blockedID).Delete(&models.UserBlacklist{})
	if result.RowsAffected == 0 {
		return errors.New("用户不在黑名单中")
	}
	return result.Error
}

// GetBlacklist 获取黑名单列表
func (s *MessageService) GetBlacklist(userID uuid.UUID, page, pageSize int) ([]models.UserBlacklist, int64, error) {
	var blacklist []models.UserBlacklist
	var total int64

	query := s.db.Model(&models.UserBlacklist{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&blacklist).Error; err != nil {
		return nil, 0, err
	}

	// 填充被拉黑用户信息
	blockedIDs := make([]uuid.UUID, len(blacklist))
	for i, b := range blacklist {
		blockedIDs[i] = b.BlockedID
	}

	if len(blockedIDs) > 0 {
		var users []models.User
		s.db.Select("id, username, nickname, avatar").Where("id IN ?", blockedIDs).Find(&users)

		userMap := make(map[uuid.UUID]*models.User)
		for i := range users {
			userMap[users[i].ID] = &users[i]
		}

		for i := range blacklist {
			if u, ok := userMap[blacklist[i].BlockedID]; ok {
				blacklist[i].BlockedUser = u
			}
		}
	}

	return blacklist, total, nil
}

// IsBlocked 检查是否被拉黑（双向检查）
func (s *MessageService) IsBlocked(userID, otherUserID uuid.UUID) (bool, string) {
	var count int64
	// 检查对方是否拉黑了我
	s.db.Model(&models.UserBlacklist{}).Where("user_id = ? AND blocked_id = ?", otherUserID, userID).Count(&count)
	if count > 0 {
		return true, "对方已将你拉黑"
	}

	// 检查我是否拉黑了对方
	s.db.Model(&models.UserBlacklist{}).Where("user_id = ? AND blocked_id = ?", userID, otherUserID).Count(&count)
	if count > 0 {
		return true, "你已将对方拉黑"
	}

	return false, ""
}

// ============= 会话静音 =============

// MuteConversation 静音/取消静音会话
func (s *MessageService) MuteConversation(userID, otherUserID uuid.UUID) (bool, error) {
	// 确保user1ID < user2ID
	user1ID, user2ID := userID, otherUserID
	if user1ID.String() > user2ID.String() {
		user1ID, user2ID = user2ID, user1ID
	}

	var conv models.Conversation
	if err := s.db.Where("user1_id = ? AND user2_id = ?", user1ID, user2ID).First(&conv).Error; err != nil {
		return false, errors.New("会话不存在")
	}

	// 切换静音状态
	var newMuted bool
	if userID == conv.User1ID {
		newMuted = !conv.User1Muted
		s.db.Model(&conv).Update("user1_muted", newMuted)
	} else {
		newMuted = !conv.User2Muted
		s.db.Model(&conv).Update("user2_muted", newMuted)
	}

	return newMuted, nil
}

// GetConversationMuteStatus 获取会话静音状态
func (s *MessageService) GetConversationMuteStatus(userID, otherUserID uuid.UUID) bool {
	user1ID, user2ID := userID, otherUserID
	if user1ID.String() > user2ID.String() {
		user1ID, user2ID = user2ID, user1ID
	}

	var conv models.Conversation
	if err := s.db.Where("user1_id = ? AND user2_id = ?", user1ID, user2ID).First(&conv).Error; err != nil {
		return false
	}

	if userID == conv.User1ID {
		return conv.User1Muted
	}
	return conv.User2Muted
}
