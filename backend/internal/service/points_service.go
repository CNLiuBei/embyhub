// Package service 积分服务
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"feiniu-user-system/internal/database"
	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PointsService 积分服务
type PointsService struct {
	db *gorm.DB
}

// NewPointsService 创建积分服务
func NewPointsService(db *gorm.DB) *PointsService {
	return &PointsService{db: db}
}

// PointsConfig 积分配置
type PointsConfig struct {
	SignInBasePoints    int  `json:"sign_in_base_points"`    // 签到基础积分
	SignInContinueBonus int  `json:"sign_in_continue_bonus"` // 连续签到额外奖励
	SignInMaxBonus      int  `json:"sign_in_max_bonus"`      // 连续签到最大奖励（额外部分）
	RegisterPoints      int  `json:"register_points"`        // 注册奖励积分
	InvitePoints        int  `json:"invite_points"`          // 邀请奖励积分
	Enabled             bool `json:"enabled"`                // 是否启用积分系统
}

// GetDefaultConfig 获取默认配置
// 签到积分：基础5分 + 连续签到奖励（每天+1，最多+5）= 最高10分/天
func (s *PointsService) GetDefaultConfig() *PointsConfig {
	return &PointsConfig{
		SignInBasePoints:    5,  // 基础5分
		SignInContinueBonus: 1,  // 连续签到每天+1
		SignInMaxBonus:      5,  // 额外奖励最多5分，即最高10分/天
		RegisterPoints:      50, // 注册奖励50分
		InvitePoints:        30, // 邀请奖励30分
		Enabled:             true,
	}
}

// GetUserPoints 获取用户积分
func (s *PointsService) GetUserPoints(userID uuid.UUID) (int, error) {
	var user models.User
	if err := s.db.Select("points").First(&user, userID).Error; err != nil {
		return 0, errors.New("用户不存在")
	}
	return user.Points, nil
}

// AddPoints 增加积分
func (s *PointsService) AddPoints(userID uuid.UUID, points int, pointsType int8, remark string, relatedID string) error {
	if points <= 0 {
		return errors.New("积分必须大于0")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取用户当前积分
		var user models.User
		if err := tx.Select("id", "points").First(&user, userID).Error; err != nil {
			return errors.New("用户不存在")
		}

		pointsBefore := user.Points
		pointsAfter := pointsBefore + points

		// 更新用户积分
		if err := tx.Model(&user).Update("points", pointsAfter).Error; err != nil {
			return err
		}

		// 记录积分变动
		record := &models.PointsRecord{
			UserID:       userID,
			Type:         pointsType,
			Points:       points,
			PointsBefore: pointsBefore,
			PointsAfter:  pointsAfter,
			Remark:       remark,
			RelatedID:    relatedID,
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		// 清除缓存
		ctx := context.Background()
		database.DeleteCache(ctx, database.KeyUserInfo+userID.String())

		return nil
	})
}

// DeductPoints 扣除积分
func (s *PointsService) DeductPoints(userID uuid.UUID, points int, pointsType int8, remark string, relatedID string) error {
	if points <= 0 {
		return errors.New("积分必须大于0")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取用户当前积分
		var user models.User
		if err := tx.Select("id", "points").First(&user, userID).Error; err != nil {
			return errors.New("用户不存在")
		}

		if user.Points < points {
			return errors.New("积分不足")
		}

		pointsBefore := user.Points
		pointsAfter := pointsBefore - points

		// 更新用户积分
		if err := tx.Model(&user).Update("points", pointsAfter).Error; err != nil {
			return err
		}

		// 记录积分变动
		record := &models.PointsRecord{
			UserID:       userID,
			Type:         pointsType,
			Points:       -points,
			PointsBefore: pointsBefore,
			PointsAfter:  pointsAfter,
			Remark:       remark,
			RelatedID:    relatedID,
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		// 清除缓存
		ctx := context.Background()
		database.DeleteCache(ctx, database.KeyUserInfo+userID.String())

		return nil
	})
}

// SignIn 签到
func (s *PointsService) SignIn(userID uuid.UUID) (*SignInResult, error) {
	today := time.Now().Format("2006-01-02")

	// 检查今天是否已签到
	var existRecord models.SignInRecord
	if err := s.db.Where("user_id = ? AND sign_date = ?", userID, today).First(&existRecord).Error; err == nil {
		return nil, errors.New("今天已经签到过了")
	}

	// 获取昨天的签到记录，计算连续签到天数
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var yesterdayRecord models.SignInRecord
	continueDays := 1
	if err := s.db.Where("user_id = ? AND sign_date = ?", userID, yesterday).First(&yesterdayRecord).Error; err == nil {
		continueDays = yesterdayRecord.ContinueDays + 1
	}

	// 计算积分
	config := s.GetDefaultConfig()
	points := config.SignInBasePoints
	bonus := (continueDays - 1) * config.SignInContinueBonus
	if bonus > config.SignInMaxBonus {
		bonus = config.SignInMaxBonus
	}
	points += bonus

	// 创建签到记录
	signRecord := &models.SignInRecord{
		UserID:       userID,
		SignDate:     today,
		Points:       points,
		ContinueDays: continueDays,
	}

	if err := s.db.Create(signRecord).Error; err != nil {
		return nil, err
	}

	// 增加积分
	remark := fmt.Sprintf("签到奖励（连续%d天）", continueDays)
	if err := s.AddPoints(userID, points, models.PointsTypeSignIn, remark, fmt.Sprintf("signin_%s", today)); err != nil {
		return nil, err
	}

	return &SignInResult{
		Points:       points,
		ContinueDays: continueDays,
		SignDate:     today,
	}, nil
}

// SignInResult 签到结果
type SignInResult struct {
	Points       int    `json:"points"`
	ContinueDays int    `json:"continue_days"`
	SignDate     string `json:"sign_date"`
}

// GetSignInStatus 获取签到状态
func (s *PointsService) GetSignInStatus(userID uuid.UUID) (*SignInStatus, error) {
	today := time.Now().Format("2006-01-02")

	// 检查今天是否已签到
	var todayRecord models.SignInRecord
	signedToday := s.db.Where("user_id = ? AND sign_date = ?", userID, today).First(&todayRecord).Error == nil

	// 获取连续签到天数
	continueDays := 0
	if signedToday {
		continueDays = todayRecord.ContinueDays
	} else {
		// 检查昨天是否签到
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		var yesterdayRecord models.SignInRecord
		if err := s.db.Where("user_id = ? AND sign_date = ?", userID, yesterday).First(&yesterdayRecord).Error; err == nil {
			continueDays = yesterdayRecord.ContinueDays
		}
	}

	// 获取本月签到天数
	monthStart := time.Now().Format("2006-01") + "-01"
	var monthCount int64
	s.db.Model(&models.SignInRecord{}).Where("user_id = ? AND sign_date >= ?", userID, monthStart).Count(&monthCount)

	// 获取本月签到日期列表
	var monthRecords []models.SignInRecord
	s.db.Where("user_id = ? AND sign_date >= ?", userID, monthStart).Order("sign_date").Find(&monthRecords)
	signDates := make([]string, len(monthRecords))
	for i, r := range monthRecords {
		signDates[i] = r.SignDate
	}

	return &SignInStatus{
		SignedToday:  signedToday,
		ContinueDays: continueDays,
		MonthCount:   int(monthCount),
		SignDates:    signDates,
	}, nil
}

// SignInStatus 签到状态
type SignInStatus struct {
	SignedToday  bool     `json:"signed_today"`
	ContinueDays int      `json:"continue_days"`
	MonthCount   int      `json:"month_count"`
	SignDates    []string `json:"sign_dates"`
}

// GetPointsRecords 获取积分记录
func (s *PointsService) GetPointsRecords(userID uuid.UUID, page, pageSize int, pointsType *int8) ([]PointsRecordItem, int64, error) {
	var records []models.PointsRecord
	var total int64

	query := s.db.Model(&models.PointsRecord{}).Where("user_id = ?", userID)
	if pointsType != nil {
		query = query.Where("type = ?", *pointsType)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	items := make([]PointsRecordItem, len(records))
	for i, r := range records {
		items[i] = PointsRecordItem{
			ID:           r.ID,
			Type:         r.Type,
			TypeName:     r.GetTypeName(),
			Points:       r.Points,
			PointsBefore: r.PointsBefore,
			PointsAfter:  r.PointsAfter,
			Remark:       r.Remark,
			CreatedAt:    r.CreatedAt,
		}
	}

	return items, total, nil
}

// PointsRecordItem 积分记录项
type PointsRecordItem struct {
	ID           uint64    `json:"id"`
	Type         int8      `json:"type"`
	TypeName     string    `json:"type_name"`
	Points       int       `json:"points"`
	PointsBefore int       `json:"points_before"`
	PointsAfter  int       `json:"points_after"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetExchangeRules 获取兑换规则
func (s *PointsService) GetExchangeRules() ([]models.PointsExchangeRule, error) {
	var rules []models.PointsExchangeRule
	if err := s.db.Where("enabled = ?", true).Order("sort_order, points").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// ExchangePoints 积分兑换会员
func (s *PointsService) ExchangePoints(userID uuid.UUID, ruleID uint64) error {
	// 获取兑换规则
	var rule models.PointsExchangeRule
	if err := s.db.First(&rule, ruleID).Error; err != nil {
		return errors.New("兑换规则不存在")
	}

	if !rule.Enabled {
		return errors.New("该兑换规则已禁用")
	}

	// 检查用户积分
	var user models.User
	if err := s.db.Select("id", "points", "member_expire").First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.Points < rule.Points {
		return errors.New("积分不足")
	}

	// 扣除积分
	remark := fmt.Sprintf("兑换%s（%d天会员）", rule.Name, rule.MemberDays)
	if err := s.DeductPoints(userID, rule.Points, models.PointsTypeExchange, remark, fmt.Sprintf("exchange_%d", ruleID)); err != nil {
		return err
	}

	// 增加会员时长
	memberService := NewMemberService(s.db, nil)
	if err := memberService.AdminSetMember(userID, rule.MemberDays); err != nil {
		// 回滚积分
		s.AddPoints(userID, rule.Points, models.PointsTypeAdjust, "兑换失败退回", fmt.Sprintf("refund_%d", ruleID))
		return fmt.Errorf("增加会员时长失败: %v", err)
	}

	return nil
}

// AdminAdjustPoints 管理员调整积分
func (s *PointsService) AdminAdjustPoints(userID uuid.UUID, points int, remark string) error {
	if points == 0 {
		return errors.New("调整积分不能为0")
	}

	if points > 0 {
		return s.AddPoints(userID, points, models.PointsTypeAdjust, remark, "")
	}
	return s.DeductPoints(userID, -points, models.PointsTypeAdjust, remark, "")
}

// GetPointsStats 获取积分统计
func (s *PointsService) GetPointsStats() (*PointsStats, error) {
	var stats PointsStats

	// 总积分
	s.db.Model(&models.User{}).Select("COALESCE(SUM(points), 0)").Scan(&stats.TotalPoints)

	// 今日发放
	today := time.Now().Format("2006-01-02")
	s.db.Model(&models.PointsRecord{}).
		Where("DATE(created_at) = ? AND points > 0", today).
		Select("COALESCE(SUM(points), 0)").Scan(&stats.TodayIssued)

	// 今日消费
	s.db.Model(&models.PointsRecord{}).
		Where("DATE(created_at) = ? AND points < 0", today).
		Select("COALESCE(ABS(SUM(points)), 0)").Scan(&stats.TodayConsumed)

	// 今日签到人数
	s.db.Model(&models.SignInRecord{}).Where("sign_date = ?", today).Count(&stats.TodaySignIn)

	return &stats, nil
}

// PointsStats 积分统计
type PointsStats struct {
	TotalPoints   int   `json:"total_points"`
	TodayIssued   int   `json:"today_issued"`
	TodayConsumed int   `json:"today_consumed"`
	TodaySignIn   int64 `json:"today_sign_in"`
}

// CreateExchangeRule 创建兑换规则
func (s *PointsService) CreateExchangeRule(rule *models.PointsExchangeRule) error {
	return s.db.Create(rule).Error
}

// UpdateExchangeRule 更新兑换规则
func (s *PointsService) UpdateExchangeRule(id uint64, updates map[string]interface{}) error {
	return s.db.Model(&models.PointsExchangeRule{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteExchangeRule 删除兑换规则
func (s *PointsService) DeleteExchangeRule(id uint64) error {
	return s.db.Delete(&models.PointsExchangeRule{}, id).Error
}

// GetAllExchangeRules 获取所有兑换规则（管理员）
func (s *PointsService) GetAllExchangeRules() ([]models.PointsExchangeRule, error) {
	var rules []models.PointsExchangeRule
	if err := s.db.Order("sort_order, points").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// PointsRankingItem 积分排行项
type PointsRankingItem struct {
	Rank     int    `json:"rank"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Points   int    `json:"points"`
}

// GetPointsRanking 获取积分排行榜（分页）
func (s *PointsService) GetPointsRanking(page, pageSize int) ([]PointsRankingItem, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int64
	s.db.Model(&models.User{}).Count(&total)

	offset := (page - 1) * pageSize
	var users []models.User
	if err := s.db.Select("id, username, nickname, avatar, points").
		Order("points DESC, created_at ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	items := make([]PointsRankingItem, len(users))
	for i, u := range users {
		nickname := u.Nickname
		if nickname == "" {
			nickname = u.Username
		}
		items[i] = PointsRankingItem{
			Rank:     offset + i + 1,
			UserID:   u.ID.String(),
			Username: u.Username,
			Nickname: nickname,
			Avatar:   u.Avatar,
			Points:   u.Points,
		}
	}

	return items, total, nil
}

// GetUserPointsRank 获取用户积分排名
func (s *PointsService) GetUserPointsRank(userID uuid.UUID) (int, error) {
	var user models.User
	if err := s.db.Select("points").First(&user, userID).Error; err != nil {
		return 0, errors.New("用户不存在")
	}

	var rank int64
	s.db.Model(&models.User{}).Where("points > ?", user.Points).Count(&rank)

	return int(rank) + 1, nil
}

// GiftPointsResult 赠送积分结果
type GiftPointsResult struct {
	TotalUsers       int `json:"total_users"`
	SuccessCount     int `json:"success_count"`
	FailedCount      int `json:"failed_count"`
	NotificationSent int `json:"notification_sent"`
}

// GiftPointsRequest 赠送积分请求参数
type GiftPointsRequest struct {
	Points           int    `json:"points"`             // 赠送积分
	Remark           string `json:"remark"`             // 赠送原因
	TargetType       string `json:"target_type"`        // 目标类型: all, member, non_member, role
	MemberLevel      *int8  `json:"member_level"`       // 会员等级: 0普通 1月卡 2年卡
	Role             *int8  `json:"role"`               // 角色: 0普通用户 1会员 2管理员
	SendNotification bool   `json:"send_notification"`  // 是否发送站内通知
	NotificationTitle string `json:"notification_title"` // 通知标题
	NotificationBody  string `json:"notification_body"`  // 通知内容
	SendEmail        bool   `json:"send_email"`         // 是否发送邮件
	EmailTitle       string `json:"email_title"`        // 邮件标题
	EmailBody        string `json:"email_body"`         // 邮件内容
}

// GiftPointsToUsers 给指定条件的用户赠送积分
func (s *PointsService) GiftPointsToUsers(req *GiftPointsRequest) (*GiftPointsResult, []uuid.UUID, error) {
	if req.Points <= 0 {
		return nil, nil, errors.New("赠送积分必须大于0")
	}

	// 构建查询条件
	query := s.db.Select("id, email").Where("status = ?", 1)

	switch req.TargetType {
	case "member":
		// 会员用户（会员未过期）
		query = query.Where("member_expire IS NOT NULL AND member_expire > ?", time.Now())
		if req.MemberLevel != nil {
			query = query.Where("member_level = ?", *req.MemberLevel)
		}
	case "non_member":
		// 非会员用户
		query = query.Where("member_expire IS NULL OR member_expire <= ?", time.Now())
	case "role":
		// 按角色筛选
		if req.Role != nil {
			query = query.Where("role = ?", *req.Role)
		}
	// case "all" 默认不加额外条件
	}

	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return nil, nil, err
	}

	result := &GiftPointsResult{
		TotalUsers: len(users),
	}

	userIDs := make([]uuid.UUID, 0, len(users))

	// 逐个赠送积分
	for _, user := range users {
		if err := s.AddPoints(user.ID, req.Points, models.PointsTypeActivity, req.Remark, "gift_activity"); err != nil {
			result.FailedCount++
		} else {
			result.SuccessCount++
			userIDs = append(userIDs, user.ID)
		}
	}

	return result, userIDs, nil
}

// GiftPointsToAllUsers 给所有用户赠送积分（兼容旧接口）
func (s *PointsService) GiftPointsToAllUsers(points int, remark string) (*GiftPointsResult, error) {
	req := &GiftPointsRequest{
		Points:     points,
		Remark:     remark,
		TargetType: "all",
	}
	result, _, err := s.GiftPointsToUsers(req)
	return result, err
}

// GetAllUserEmails 获取所有正常用户的邮箱
func (s *PointsService) GetAllUserEmails() ([]string, error) {
	var users []models.User
	if err := s.db.Select("email").Where("status = ? AND email != ''", 1).Find(&users).Error; err != nil {
		return nil, err
	}

	emails := make([]string, 0, len(users))
	for _, u := range users {
		if u.Email != "" {
			emails = append(emails, u.Email)
		}
	}
	return emails, nil
}

// CreateBatchNotifications 批量创建通知
func (s *PointsService) CreateBatchNotifications(notifications []models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	return s.db.CreateInBatches(notifications, 100).Error
}

// GetUserCountByCondition 获取符合条件的用户数量
func (s *PointsService) GetUserCountByCondition(targetType string, memberLevel, role *int8) (int64, error) {
	query := s.db.Model(&models.User{}).Where("status = ?", 1)

	switch targetType {
	case "member":
		query = query.Where("member_expire IS NOT NULL AND member_expire > ?", time.Now())
		if memberLevel != nil {
			query = query.Where("member_level = ?", *memberLevel)
		}
	case "non_member":
		query = query.Where("member_expire IS NULL OR member_expire <= ?", time.Now())
	case "role":
		if role != nil {
			query = query.Where("role = ?", *role)
		}
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}


// ============= 自动赠送规则管理 =============

// GetGiftRules 获取所有自动赠送规则
func (s *PointsService) GetGiftRules() ([]models.PointsGiftRule, error) {
	var rules []models.PointsGiftRule
	if err := s.db.Order("created_at DESC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// GetGiftRule 获取单个规则
func (s *PointsService) GetGiftRule(id uint64) (*models.PointsGiftRule, error) {
	var rule models.PointsGiftRule
	if err := s.db.First(&rule, id).Error; err != nil {
		return nil, errors.New("规则不存在")
	}
	return &rule, nil
}

// CreateGiftRule 创建自动赠送规则
func (s *PointsService) CreateGiftRule(rule *models.PointsGiftRule) error {
	// 计算下次执行时间
	rule.NextExecuteAt = s.calculateNextExecuteTime(rule)
	return s.db.Create(rule).Error
}

// UpdateGiftRule 更新自动赠送规则
func (s *PointsService) UpdateGiftRule(id uint64, updates map[string]interface{}) error {
	var rule models.PointsGiftRule
	if err := s.db.First(&rule, id).Error; err != nil {
		return errors.New("规则不存在")
	}

	// 更新规则
	if err := s.db.Model(&rule).Updates(updates).Error; err != nil {
		return err
	}

	// 重新计算下次执行时间
	if err := s.db.First(&rule, id).Error; err != nil {
		return err
	}
	nextTime := s.calculateNextExecuteTime(&rule)
	return s.db.Model(&rule).Update("next_execute_at", nextTime).Error
}

// DeleteGiftRule 删除自动赠送规则
func (s *PointsService) DeleteGiftRule(id uint64) error {
	return s.db.Delete(&models.PointsGiftRule{}, id).Error
}

// ToggleGiftRule 切换规则启用状态
func (s *PointsService) ToggleGiftRule(id uint64) error {
	var rule models.PointsGiftRule
	if err := s.db.First(&rule, id).Error; err != nil {
		return errors.New("规则不存在")
	}
	return s.db.Model(&rule).Update("enabled", !rule.Enabled).Error
}

// calculateNextExecuteTime 计算下次执行时间
func (s *PointsService) calculateNextExecuteTime(rule *models.PointsGiftRule) *time.Time {
	now := time.Now()
	loc := now.Location()

	// 解析执行时间
	hour, minute := 8, 0 // 默认8:00
	if rule.ExecuteTime != "" {
		fmt.Sscanf(rule.ExecuteTime, "%d:%d", &hour, &minute)
	}

	var next time.Time

	switch rule.RuleType {
	case models.GiftRuleTypeDaily:
		// 每日执行
		next = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}

	case models.GiftRuleTypeWeekly:
		// 每周执行
		targetWeekday := time.Weekday(rule.ExecuteDay % 7) // 1-7 转换为 0-6
		if rule.ExecuteDay == 7 {
			targetWeekday = time.Sunday
		}
		daysUntil := int(targetWeekday) - int(now.Weekday())
		if daysUntil < 0 {
			daysUntil += 7
		}
		next = time.Date(now.Year(), now.Month(), now.Day()+daysUntil, hour, minute, 0, 0, loc)
		if next.Before(now) {
			next = next.Add(7 * 24 * time.Hour)
		}

	case models.GiftRuleTypeMonthly:
		// 每月执行
		targetDay := rule.ExecuteDay
		if targetDay < 1 {
			targetDay = 1
		}
		if targetDay > 28 {
			targetDay = 28 // 避免月末问题
		}
		next = time.Date(now.Year(), now.Month(), targetDay, hour, minute, 0, 0, loc)
		if next.Before(now) {
			next = next.AddDate(0, 1, 0)
		}

	case models.GiftRuleTypeYearly:
		// 每年执行
		targetMonth := time.Month(rule.ExecuteMonth)
		if targetMonth < 1 || targetMonth > 12 {
			targetMonth = 1
		}
		targetDay := rule.ExecuteDay
		if targetDay < 1 {
			targetDay = 1
		}
		if targetDay > 28 {
			targetDay = 28
		}
		next = time.Date(now.Year(), targetMonth, targetDay, hour, minute, 0, 0, loc)
		if next.Before(now) {
			next = next.AddDate(1, 0, 0)
		}

	default:
		next = now.Add(24 * time.Hour)
	}

	return &next
}

// ExecuteGiftRule 执行赠送规则
func (s *PointsService) ExecuteGiftRule(rule *models.PointsGiftRule) (*GiftPointsResult, error) {
	// 构建赠送请求
	req := &GiftPointsRequest{
		Points:            rule.Points,
		Remark:            fmt.Sprintf("自动赠送: %s", rule.Name),
		TargetType:        rule.TargetType,
		MemberLevel:       rule.MemberLevel,
		SendNotification:  rule.SendNotification,
		NotificationTitle: rule.NotificationTitle,
		NotificationBody:  rule.NotificationBody,
	}

	// 执行赠送
	result, userIDs, err := s.GiftPointsToUsers(req)
	if err != nil {
		return nil, err
	}

	// 发送站内通知
	if rule.SendNotification && len(userIDs) > 0 {
		title := rule.NotificationTitle
		if title == "" {
			title = "🎁 积分到账通知"
		}
		body := rule.NotificationBody
		if body == "" {
			body = fmt.Sprintf("恭喜您获得 %d 积分！来源：%s", rule.Points, rule.Name)
		}

		notifications := make([]models.Notification, len(userIDs))
		for i, uid := range userIDs {
			notifications[i] = models.Notification{
				UserID:  uid,
				Title:   title,
				Content: body,
				Type:    3, // 活动类型
			}
		}
		s.CreateBatchNotifications(notifications)
		result.NotificationSent = len(notifications)
	}

	// 记录执行日志
	log := &models.PointsGiftLog{
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		Points:       rule.Points,
		TotalUsers:   result.TotalUsers,
		SuccessCount: result.SuccessCount,
		FailedCount:  result.FailedCount,
		ExecuteAt:    time.Now(),
	}
	s.db.Create(log)

	// 更新规则执行时间
	now := time.Now()
	nextTime := s.calculateNextExecuteTime(rule)
	s.db.Model(rule).Updates(map[string]interface{}{
		"last_execute_at": now,
		"next_execute_at": nextTime,
	})

	return result, nil
}

// GetGiftLogs 获取赠送执行日志
func (s *PointsService) GetGiftLogs(ruleID uint64, page, pageSize int) ([]models.PointsGiftLog, int64, error) {
	var logs []models.PointsGiftLog
	var total int64

	query := s.db.Model(&models.PointsGiftLog{})
	if ruleID > 0 {
		query = query.Where("rule_id = ?", ruleID)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetPendingGiftRules 获取待执行的规则
func (s *PointsService) GetPendingGiftRules() ([]models.PointsGiftRule, error) {
	var rules []models.PointsGiftRule
	now := time.Now()
	if err := s.db.Where("enabled = ? AND next_execute_at <= ?", true, now).Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}


