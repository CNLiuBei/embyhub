// Package service 定时任务服务
package service

import (
	"log"
	"time"

	"feiniu-user-system/internal/models"

	"gorm.io/gorm"
)

// SchedulerService 定时任务服务
type SchedulerService struct {
	db *gorm.DB
}

// NewSchedulerService 创建定时任务服务
func NewSchedulerService(db *gorm.DB) *SchedulerService {
	return &SchedulerService{db: db}
}

// Start 启动所有定时任务
func (s *SchedulerService) Start() {
	log.Println("========================================")
	log.Println("🕐 定时任务服务启动")
	log.Println("========================================")

	// 立即执行一次
	s.CheckCardExpiry()
	s.UpdateBatchStats()
	s.CheckMemberExpiry()

	// 每小时检查一次卡密过期
	go s.runPeriodically("卡密过期检测", 1*time.Hour, s.CheckCardExpiry)

	// 每30分钟更新一次批次统计
	go s.runPeriodically("批次统计更新", 30*time.Minute, s.UpdateBatchStats)

	// 每天凌晨检查会员过期
	go s.runDaily("会员过期检测", 2, 0, s.CheckMemberExpiry)

	// 每周日凌晨清理过期数据
	go s.runWeekly("过期数据清理", time.Sunday, 3, 0, s.CleanExpiredData)

	// 每天凌晨4点清理过期用户
	go s.runDaily("过期用户清理", 4, 0, s.CleanExpiredUsers)

	// 每分钟检查一次积分自动赠送规则
	go s.runPeriodically("积分自动赠送", 1*time.Minute, s.ExecutePointsGiftRules)

	log.Println("✅ 所有定时任务已启动")
}

// runPeriodically 周期性执行任务（静默模式，不打印每次执行日志）
func (s *SchedulerService) runPeriodically(name string, interval time.Duration, task func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		// 静默执行，由具体任务决定是否打印日志
		task()
	}
}

// runDaily 每天定时执行
func (s *SchedulerService) runDaily(name string, hour, minute int, task func()) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}

		duration := next.Sub(now)
		log.Printf("📅 定时任务 [%s] 将在 %s 后执行 (下次执行时间: %s)",
			name, duration.Round(time.Minute), next.Format("2006-01-02 15:04:05"))

		time.Sleep(duration)
		log.Printf("⏰ 执行定时任务: %s", name)
		task()
	}
}

// runWeekly 每周定时执行
func (s *SchedulerService) runWeekly(name string, weekday time.Weekday, hour, minute int, task func()) {
	for {
		now := time.Now()
		daysUntil := (int(weekday) - int(now.Weekday()) + 7) % 7
		if daysUntil == 0 && (now.Hour() > hour || (now.Hour() == hour && now.Minute() >= minute)) {
			daysUntil = 7
		}

		next := now.AddDate(0, 0, daysUntil)
		next = time.Date(next.Year(), next.Month(), next.Day(), hour, minute, 0, 0, next.Location())

		duration := next.Sub(now)
		log.Printf("📅 定时任务 [%s] 将在 %s 后执行 (下次执行时间: %s)",
			name, duration.Round(time.Hour), next.Format("2006-01-02 15:04:05"))

		time.Sleep(duration)
		log.Printf("⏰ 执行定时任务: %s", name)
		task()
	}
}

// CheckCardExpiry 检查并标记过期卡密
func (s *SchedulerService) CheckCardExpiry() {
	log.Println("开始检查卡密过期状态...")

	result := s.db.Model(&models.Card{}).
		Where("status = ? AND expire_at IS NOT NULL AND expire_at < ?",
			models.CardStatusUnused, time.Now()).
		Update("status", models.CardStatusExpired)

	if result.Error != nil {
		log.Printf("❌ 检查卡密过期失败: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("✅ 标记了 %d 张过期卡密", result.RowsAffected)
	} else {
		log.Println("ℹ️  没有需要标记的过期卡密")
	}
}

// UpdateBatchStats 更新批次统计数据
func (s *SchedulerService) UpdateBatchStats() {
	log.Println("开始更新批次统计...")

	var batches []models.CardBatch
	if err := s.db.Find(&batches).Error; err != nil {
		log.Printf("❌ 查询批次失败: %v", err)
		return
	}

	updated := 0
	for _, batch := range batches {
		var usedCount int64
		s.db.Model(&models.Card{}).
			Where("batch_no = ? AND status = ?", batch.BatchNo, models.CardStatusUsed).
			Count(&usedCount)

		if int(usedCount) != batch.UsedCount {
			s.db.Model(&models.CardBatch{}).
				Where("id = ?", batch.ID).
				Update("used_count", usedCount)
			updated++
		}
	}

	if updated > 0 {
		log.Printf("✅ 更新了 %d 个批次的统计数据", updated)
	} else {
		log.Println("ℹ️  批次统计数据已是最新")
	}
}

// CheckMemberExpiry 检查会员过期
// 调用 MemberService.CheckMemberExpire() 统一处理会员过期逻辑
// 包含：降级会员等级、禁用账户、发送站内通知、发送邮件通知
func (s *SchedulerService) CheckMemberExpiry() {
	log.Println("开始检查会员过期状态...")

	// 使用 MemberService 统一处理会员过期逻辑
	memberService := NewMemberService(s.db, nil)
	if err := memberService.CheckMemberExpire(); err != nil {
		log.Printf("❌ 检查会员过期失败: %v", err)
		return
	}

	log.Println("✅ 会员过期检测完成")
}

// CleanExpiredData 清理过期数据
func (s *SchedulerService) CleanExpiredData() {
	log.Println("开始清理过期数据...")

	// 清理旧的登录日志（保留3个月）
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	result := s.db.Where("created_at < ?", threeMonthsAgo).
		Delete(&models.LoginLog{})

	if result.Error != nil {
		log.Printf("❌ 清理登录日志失败: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("✅ 清理了 %d 条旧登录日志", result.RowsAffected)
	}

	// 清理旧的已读通知（保留1个月）
	oneMonthAgo := time.Now().AddDate(0, -1, 0)
	result = s.db.Where("is_read = ? AND created_at < ?", true, oneMonthAgo).
		Delete(&models.Notification{})

	if result.Error != nil {
		log.Printf("❌ 清理通知失败: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("✅ 清理了 %d 条旧通知", result.RowsAffected)
	}

	log.Println("✅ 过期数据清理完成")
}

// CleanExpiredUsers 清理过期用户
// 清理条件：会员已过期超过指定天数 且 长时间未登录的用户
func (s *SchedulerService) CleanExpiredUsers() {
	log.Println("开始清理过期用户...")

	// 获取清理设置
	settingService := NewSettingService(s.db)
	cleanupSettings, err := settingService.GetUserCleanupSettings()
	if err != nil {
		log.Printf("❌ 获取用户清理设置失败: %v", err)
		return
	}

	if !cleanupSettings.Enabled {
		log.Println("ℹ️  用户清理功能未启用")
		return
	}

	now := time.Now()
	inactiveThreshold := now.AddDate(0, 0, -cleanupSettings.InactiveDays)
	expiredThreshold := now.AddDate(0, 0, -cleanupSettings.ExpiredDays)

	// 查找需要清理的用户
	// 条件：1. 非管理员 2. 会员已过期超过指定天数 3. 最后登录时间超过指定天数
	var usersToDelete []models.User
	query := s.db.Where("role < ?", models.RoleAdmin). // 非管理员
		Where("member_expire IS NOT NULL AND member_expire < ?", expiredThreshold). // 会员过期超过阈值
		Where("(last_login_at IS NULL OR last_login_at < ?)", inactiveThreshold) // 长时间未登录

	if err := query.Find(&usersToDelete).Error; err != nil {
		log.Printf("❌ 查询过期用户失败: %v", err)
		return
	}

	if len(usersToDelete) == 0 {
		log.Println("ℹ️  没有需要清理的过期用户")
		return
	}

	log.Printf("📋 找到 %d 个需要清理的过期用户", len(usersToDelete))

	// 获取媒体适配器（用于删除Emby账号）
	var mediaAdapter interface {
		DeleteUser(username string) error
	}
	if cleanupSettings.DeleteEmbyAccount {
		mediaAdapter = GetMediaAdapterFromDB(s.db)
	}

	deletedCount := 0
	for _, user := range usersToDelete {
		log.Printf("🗑️  清理用户: %s (ID: %s, 会员过期: %v, 最后登录: %v)",
			user.Username, user.ID, user.MemberExpire, user.LastLoginAt)

		// 删除用户关联数据
		if err := s.deleteUserRelatedData(user.ID); err != nil {
			log.Printf("❌ 删除用户 %s 关联数据失败: %v", user.Username, err)
			continue
		}

		// 删除Emby账号
		if mediaAdapter != nil && user.Username != "" {
			if err := mediaAdapter.DeleteUser(user.Username); err != nil {
				log.Printf("⚠️  删除用户 %s 的Emby账号失败: %v", user.Username, err)
			} else {
				log.Printf("✅ 删除用户 %s 的Emby账号成功", user.Username)
			}
		}

		// 硬删除用户（不是软删除）
		if err := s.db.Unscoped().Delete(&user).Error; err != nil {
			log.Printf("❌ 删除用户 %s 失败: %v", user.Username, err)
			continue
		}

		deletedCount++
	}

	log.Printf("✅ 过期用户清理完成，共清理 %d 个用户", deletedCount)
}

// deleteUserRelatedData 删除用户关联数据
func (s *SchedulerService) deleteUserRelatedData(userID interface{}) error {
	// 删除登录日志
	if err := s.db.Where("user_id = ?", userID).Delete(&models.LoginLog{}).Error; err != nil {
		log.Printf("删除登录日志失败: %v", err)
	}

	// 删除用户设备
	if err := s.db.Where("user_id = ?", userID).Delete(&models.UserDevice{}).Error; err != nil {
		log.Printf("删除用户设备失败: %v", err)
	}

	// 删除观影记录
	if err := s.db.Where("user_id = ?", userID).Delete(&models.WatchHistory{}).Error; err != nil {
		log.Printf("删除观影记录失败: %v", err)
	}

	// 删除收藏
	if err := s.db.Where("user_id = ?", userID).Delete(&models.Favorite{}).Error; err != nil {
		log.Printf("删除收藏失败: %v", err)
	}

	// 删除通知
	if err := s.db.Where("user_id = ?", userID).Delete(&models.Notification{}).Error; err != nil {
		log.Printf("删除通知失败: %v", err)
	}

	// 删除余额记录
	if err := s.db.Where("user_id = ?", userID).Delete(&models.BalanceRecord{}).Error; err != nil {
		log.Printf("删除余额记录失败: %v", err)
	}

	// 删除会员订单
	if err := s.db.Where("user_id = ?", userID).Delete(&models.MemberOrder{}).Error; err != nil {
		log.Printf("删除会员订单失败: %v", err)
	}

	// 删除邀请记录（作为被邀请人）
	if err := s.db.Where("invitee_id = ?", userID).Delete(&models.InviteRecord{}).Error; err != nil {
		log.Printf("删除邀请记录失败: %v", err)
	}

	return nil
}




// ExecutePointsGiftRules 执行积分自动赠送规则
func (s *SchedulerService) ExecutePointsGiftRules() {
	pointsService := NewPointsService(s.db)

	// 获取待执行的规则
	rules, err := pointsService.GetPendingGiftRules()
	if err != nil {
		log.Printf("❌ 获取待执行的赠送规则失败: %v", err)
		return
	}

	if len(rules) == 0 {
		// 没有待执行的规则，静默返回，不打印日志
		return
	}

	log.Printf("📋 发现 %d 个待执行的积分赠送规则", len(rules))

	for _, rule := range rules {
		log.Printf("⏰ 执行积分自动赠送规则: %s (ID: %d)", rule.Name, rule.ID)

		result, err := pointsService.ExecuteGiftRule(&rule)
		if err != nil {
			log.Printf("❌ 执行规则 %s 失败: %v", rule.Name, err)
			continue
		}

		log.Printf("✅ 规则 %s 执行完成: 目标用户 %d, 成功 %d, 失败 %d, 通知 %d",
			rule.Name, result.TotalUsers, result.SuccessCount, result.FailedCount, result.NotificationSent)
	}
}
