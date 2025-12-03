package task

import (
	"log"
	"time"

	"embyhub/internal/dao"
	"embyhub/internal/model"
	"embyhub/pkg/database"
	"embyhub/pkg/email"
)

// VipTask VIP到期处理任务
// 设计考虑：
// 1. 使用批量SQL更新，避免逐个用户处理
// 2. 分批处理，控制内存使用
// 3. 使用数据库索引加速查询
// 4. 支持十万级用户规模
type VipTask struct {
	interval  time.Duration
	batchSize int // 每批处理数量
	stopChan  chan struct{}
}

// NewVipTask 创建VIP任务
func NewVipTask(interval time.Duration) *VipTask {
	return &VipTask{
		interval:  interval,
		batchSize: 1000, // 每批处理1000个用户
		stopChan:  make(chan struct{}),
	}
}

// Start 启动VIP检查任务
func (t *VipTask) Start() {
	log.Printf("[VipTask] VIP到期检查任务已启动，间隔: %v", t.interval)

	// 启动时立即执行一次
	go t.run()

	// 定时执行
	ticker := time.NewTicker(t.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				t.run()
			case <-t.stopChan:
				ticker.Stop()
				log.Println("[VipTask] VIP检查任务已停止")
				return
			}
		}
	}()
}

// Stop 停止任务
func (t *VipTask) Stop() {
	close(t.stopChan)
}

// run 执行VIP检查
func (t *VipTask) run() {
	startTime := time.Now()
	log.Println("[VipTask] 开始检查VIP到期状态...")

	// 1. 批量更新过期VIP用户（高效：一条SQL搞定）
	expiredCount := t.processExpiredVip()

	// 2. 获取即将到期的VIP用户数量（用于统计/通知）
	expiringCount := t.countExpiringVip(3) // 3天内到期

	// 3. 发送VIP即将到期邮件提醒
	sentCount := t.sendExpiringReminders()

	duration := time.Since(startTime)
	log.Printf("[VipTask] VIP检查完成，耗时: %v，过期处理: %d，即将到期: %d，发送提醒: %d",
		duration, expiredCount, expiringCount, sentCount)
}

// processExpiredVip 批量处理过期VIP
// 使用单条SQL批量更新，高效处理大量数据
func (t *VipTask) processExpiredVip() int64 {
	now := time.Now()

	// 批量更新所有过期的VIP用户
	// 使用数据库原生批量更新，无需加载到内存
	result := database.DB.Exec(`
		UPDATE users 
		SET vip_level = 0, updated_at = ? 
		WHERE vip_level = 1 
		AND vip_expire_at IS NOT NULL 
		AND vip_expire_at < ?
	`, now, now)

	if result.Error != nil {
		log.Printf("[VipTask] 更新过期VIP失败: %v", result.Error)
		return 0
	}

	return result.RowsAffected
}

// countExpiringVip 统计即将到期的VIP数量
func (t *VipTask) countExpiringVip(days int) int64 {
	now := time.Now()
	expireTime := now.AddDate(0, 0, days)

	var count int64
	database.DB.Raw(`
		SELECT COUNT(*) FROM users 
		WHERE vip_level = 1 
		AND vip_expire_at IS NOT NULL 
		AND vip_expire_at > ? 
		AND vip_expire_at <= ?
	`, now, expireTime).Scan(&count)

	return count
}

// GetExpiredVipUserIDs 获取过期VIP用户ID（分批获取，用于需要额外处理的场景）
// 例如：发送通知、调用外部API等
func (t *VipTask) GetExpiredVipUserIDs(limit int, offset int) ([]int, error) {
	now := time.Now()
	var userIDs []int

	err := database.DB.Raw(`
		SELECT user_id FROM users 
		WHERE vip_level = 1 
		AND vip_expire_at IS NOT NULL 
		AND vip_expire_at < ?
		ORDER BY user_id
		LIMIT ? OFFSET ?
	`, now, limit, offset).Scan(&userIDs).Error

	return userIDs, err
}

// GetExpiringVipUsers 获取即将到期的VIP用户（分批获取，用于发送提醒）
type ExpiringVipUser struct {
	UserID      int       `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	VipExpireAt time.Time `json:"vip_expire_at"`
	DaysLeft    int       `json:"days_left"`
}

func (t *VipTask) GetExpiringVipUsers(days int, limit int, offset int) ([]ExpiringVipUser, error) {
	now := time.Now()
	expireTime := now.AddDate(0, 0, days)
	var users []ExpiringVipUser

	err := database.DB.Raw(`
		SELECT 
			user_id, 
			username, 
			email, 
			vip_expire_at,
			EXTRACT(DAY FROM (vip_expire_at - ?))::int as days_left
		FROM users 
		WHERE vip_level = 1 
		AND vip_expire_at IS NOT NULL 
		AND vip_expire_at > ? 
		AND vip_expire_at <= ?
		ORDER BY vip_expire_at
		LIMIT ? OFFSET ?
	`, now, now, expireTime, limit, offset).Scan(&users).Error

	return users, err
}

// ProcessWithCallback 分批处理过期用户并执行回调
// 适用于需要对每个用户执行额外操作的场景（如发送通知）
func (t *VipTask) ProcessWithCallback(callback func(userIDs []int) error) error {
	offset := 0
	for {
		userIDs, err := t.GetExpiredVipUserIDs(t.batchSize, offset)
		if err != nil {
			return err
		}

		if len(userIDs) == 0 {
			break
		}

		// 执行回调
		if err := callback(userIDs); err != nil {
			log.Printf("[VipTask] 回调处理失败: %v", err)
		}

		offset += t.batchSize

		// 防止死循环，最多处理100批（10万用户）
		if offset >= t.batchSize*100 {
			log.Println("[VipTask] 达到最大处理批次限制")
			break
		}
	}

	return nil
}

// VipStatistics VIP统计信息
type VipStatistics struct {
	TotalVip     int64 `json:"total_vip"`      // VIP总数
	ExpiredToday int64 `json:"expired_today"`  // 今日过期
	Expiring3Day int64 `json:"expiring_3_day"` // 3天内到期
	Expiring7Day int64 `json:"expiring_7_day"` // 7天内到期
}

// GetVipStatistics 获取VIP统计信息
func (t *VipTask) GetVipStatistics() (*VipStatistics, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	stats := &VipStatistics{}

	// 有效VIP总数
	database.DB.Raw(`
		SELECT COUNT(*) FROM users 
		WHERE vip_level = 1 AND vip_expire_at > ?
	`, now).Scan(&stats.TotalVip)

	// 今日到期
	database.DB.Raw(`
		SELECT COUNT(*) FROM users 
		WHERE vip_level = 1 
		AND vip_expire_at > ? AND vip_expire_at <= ?
	`, now, today).Scan(&stats.ExpiredToday)

	// 3天内到期
	database.DB.Raw(`
		SELECT COUNT(*) FROM users 
		WHERE vip_level = 1 
		AND vip_expire_at > ? AND vip_expire_at <= ?
	`, now, now.AddDate(0, 0, 3)).Scan(&stats.Expiring3Day)

	// 7天内到期
	database.DB.Raw(`
		SELECT COUNT(*) FROM users 
		WHERE vip_level = 1 
		AND vip_expire_at > ? AND vip_expire_at <= ?
	`, now, now.AddDate(0, 0, 7)).Scan(&stats.Expiring7Day)

	return stats, nil
}

// sendExpiringReminders 发送VIP即将到期提醒邮件
// 每天只发送一次（3天内到期的用户）
func (t *VipTask) sendExpiringReminders() int {
	// 获取邮件配置
	configDAO := dao.NewSystemConfigDAO()
	configMap, err := configDAO.BatchGet(email.GetConfigKeys())
	if err != nil {
		return 0
	}
	// 根据provider检查配置
	provider := configMap["email_provider"]
	switch provider {
	case "aliyun":
		if configMap["aliyun_access_key_id"] == "" {
			return 0
		}
	case "resend":
		if configMap["resend_api_key"] == "" {
			return 0
		}
	case "aliyun_smtp", "smtp":
		if configMap["smtp_host"] == "" {
			return 0
		}
	default:
		if configMap["smtp_host"] == "" {
			return 0
		}
	}

	emailClient := email.NewClient(email.GetConfigFromDB(configMap))

	// 获取3天内到期且有邮箱的用户
	now := time.Now()
	expireTime := now.AddDate(0, 0, 3)

	var users []model.User
	database.DB.Where("vip_level = 1 AND email != '' AND vip_expire_at > ? AND vip_expire_at <= ?",
		now, expireTime).Find(&users)

	sentCount := 0
	for _, user := range users {
		// 计算剩余天数
		daysLeft := int(user.VipExpireAt.Sub(now).Hours() / 24)
		if daysLeft < 1 {
			daysLeft = 1
		}

		// 发送VIP到期提醒邮件
		expireDate := user.VipExpireAt.Format("2006年01月02日 15:04")
		if err := emailClient.SendVipExpiringEmail(user.Email, user.Username, expireDate, daysLeft); err != nil {
			log.Printf("[VipTask] 发送VIP到期提醒失败 user=%s email=%s: %v", user.Username, user.Email, err)
		} else {
			sentCount++
			log.Printf("[VipTask] 发送VIP到期提醒成功 user=%s email=%s days_left=%d", user.Username, user.Email, daysLeft)
		}
	}

	return sentCount
}
