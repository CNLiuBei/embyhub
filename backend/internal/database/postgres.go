// Package database 数据库连接管理
package database

import (
	"fmt"
	"log"
	"time"

	"feiniu-user-system/internal/config"
	"feiniu-user-system/internal/models"
	"feiniu-user-system/pkg/emby"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var appConfig *config.Config

// SetConfig 设置应用配置（在InitPostgres之前调用）
func SetConfig(cfg *config.Config) {
	appConfig = cfg
}

// InitPostgres 初始化PostgreSQL连接
func InitPostgres(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 自动迁移表结构
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 初始化超级管理员
	if err := initSuperAdmin(); err != nil {
		return fmt.Errorf("初始化超级管理员失败: %w", err)
	}

	// 初始化积分兑换规则
	if err := initPointsExchangeRules(); err != nil {
		return fmt.Errorf("初始化积分兑换规则失败: %w", err)
	}

	// 初始化论坛节点
	if err := initForumNodes(); err != nil {
		return fmt.Errorf("初始化论坛节点失败: %w", err)
	}

	return nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.LoginLog{},
		&models.UserDevice{},
		&models.WatchHistory{},
		&models.Favorite{},
		&models.Notification{},
		&models.OperationLog{},
		&models.MemberOrder{},
		&models.Card{},
		&models.CardBatch{},
		&models.Setting{},
		// VIP相关表
		&models.VipPlan{},
		&models.UserVip{},
		&models.VipOrder{},
		&models.BalanceLog{},
		// 公告表
		&models.Announcement{},
		// IP黑名单表
		&models.IPBlacklist{},
		// 邀请码表
		&models.InviteCode{},
		&models.InviteRecord{},
		// 积分系统表
		&models.PointsRecord{},
		&models.SignInRecord{},
		&models.PointsExchangeRule{},
		// 积分卡密表
		&models.PointsCard{},
		&models.PointsCardBatch{},
		// 积分自动赠送规则表
		&models.PointsGiftRule{},
		&models.PointsGiftLog{},
		// 论坛系统表
		&models.ForumNode{},
		&models.ForumTopic{},
		&models.ForumComment{},
		&models.ForumLike{},
		&models.ForumFavorite{},
		&models.TopicView{},
		// 私信系统表
		&models.PrivateMessage{},
		&models.Conversation{},
		&models.UserFollow{},
		&models.UserBlacklist{},
	)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// initSuperAdmin 初始化超级管理员账号（系统唯一）
func initSuperAdmin() error {
	// 检查是否已存在超级管理员
	var count int64
	if err := DB.Model(&models.User{}).Where("role = ?", models.RoleSuperAdmin).Count(&count).Error; err != nil {
		return err
	}

	// 如果已存在超级管理员，跳过初始化
	if count > 0 {
		return nil
	}

	// 检查是否存在admin用户
	var existingAdmin models.User
	err := DB.Where("username = ?", "admin").First(&existingAdmin).Error

	if err == nil {
		// admin用户已存在，升级为超级管理员
		if err := DB.Model(&existingAdmin).Updates(map[string]interface{}{
			"role":          models.RoleSuperAdmin,
			"member_expire": time.Now().AddDate(100, 0, 0),
			"nickname":      "超级管理员",
		}).Error; err != nil {
			return fmt.Errorf("升级admin为超级管理员失败: %w", err)
		}
		log.Printf("✅ 已将现有admin账号升级为超级管理员！")

		// 同步到Emby
		if appConfig != nil && appConfig.Emby.Enabled {
			go syncToEmby(existingAdmin.Username, "admin123")
		}

		return nil
	}

	// admin用户不存在，创建新的超级管理员
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	superAdmin := &models.User{
		ID:           uuid.New(),
		Username:     "admin",
		Password:     string(hashedPassword),
		Nickname:     "超级管理员",
		Email:        "admin@emby.local",
		Status:       1,
		Role:         models.RoleSuperAdmin,
		MemberExpire: timePtr(time.Now().AddDate(100, 0, 0)), // 长期会员
	}

	if err := DB.Create(superAdmin).Error; err != nil {
		return fmt.Errorf("创建超级管理员失败: %w", err)
	}

	log.Printf("✅ 超级管理员初始化成功！")
	log.Printf("   账号: %s", superAdmin.Username)
	log.Printf("   密码: admin123")
	log.Printf("   ⚠️  请登录后立即修改密码！")

	// 同步创建Emby用户
	if appConfig != nil && appConfig.Emby.Enabled {
		go syncToEmby(superAdmin.Username, "admin123")
	}

	return nil
}

// syncToEmby 异步同步用户到Emby
func syncToEmby(username, password string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Emby同步异常: %v", r)
		}
	}()

	log.Println("========================================")
	log.Printf("开始同步用户到Emby: %s", username)
	log.Printf("目标服务: %s", appConfig.Emby.BaseURL)

	client := emby.NewClient(&emby.Config{
		BaseURL:   appConfig.Emby.BaseURL,
		AdminUser: appConfig.Emby.AdminUser,
		AdminPass: appConfig.Emby.AdminPass,
	})

	_, err := client.CreateUser(username, password)
	if err != nil {
		log.Printf("⚠️  同步失败: %v", err)
		log.Println("   提示: 您可以稍后手动在Emby中创建该用户")
	} else {
		log.Printf("✅ 同步成功！用户 %s 已创建到Emby", username)
	}
	log.Println("========================================")
}

// timePtr 返回时间指针
func timePtr(t time.Time) *time.Time {
	return &t
}

// initPointsExchangeRules 初始化积分兑换规则
// 天卡1000积分为基准，时间越长越划算
func initPointsExchangeRules() error {
	// 删除旧的"体验卡"规则（已改名为天卡）
	DB.Where("name = ?", "体验卡").Delete(&models.PointsExchangeRule{})

	// 默认兑换规则（天卡1000积分为基准，时间越长越划算）
	defaultRules := []models.PointsExchangeRule{
		{Name: "天卡", Points: 1000, MemberDays: 1, Description: "1天会员", SortOrder: 1, Enabled: true},
		{Name: "周卡", Points: 5000, MemberDays: 7, Description: "7天会员", SortOrder: 2, Enabled: true},
		{Name: "月卡", Points: 15000, MemberDays: 30, Description: "30天会员", SortOrder: 3, Enabled: true},
		{Name: "季卡", Points: 36000, MemberDays: 90, Description: "90天会员", SortOrder: 4, Enabled: true},
		{Name: "半年卡", Points: 60000, MemberDays: 180, Description: "180天会员", SortOrder: 5, Enabled: true},
		{Name: "年卡", Points: 100000, MemberDays: 365, Description: "365天会员", SortOrder: 6, Enabled: true},
	}

	for _, rule := range defaultRules {
		var existing models.PointsExchangeRule
		result := DB.Where("name = ?", rule.Name).First(&existing)
		if result.Error != nil {
			if err := DB.Create(&rule).Error; err != nil {
				return fmt.Errorf("创建兑换规则失败: %w", err)
			}
			log.Printf("✅ 创建积分兑换规则: %s", rule.Name)
		} else {
			if err := DB.Model(&existing).Updates(map[string]interface{}{
				"points":      rule.Points,
				"member_days": rule.MemberDays,
				"description": rule.Description,
				"sort_order":  rule.SortOrder,
			}).Error; err != nil {
				return fmt.Errorf("更新兑换规则失败: %w", err)
			}
		}
	}

	return nil
}


// initForumNodes 初始化论坛节点
func initForumNodes() error {
	// 默认论坛节点
	defaultNodes := []models.ForumNode{
		{Name: "影视推荐", Description: "分享好看的电影、电视剧", Icon: "🎬", SortOrder: 1, Status: 1},
		{Name: "资源求助", Description: "找不到的资源可以在这里求助", Icon: "🔍", SortOrder: 2, Status: 1},
		{Name: "技术交流", Description: "播放器、客户端等技术问题讨论", Icon: "💻", SortOrder: 3, Status: 1},
		{Name: "站务公告", Description: "网站公告和通知", Icon: "📢", SortOrder: 4, Status: 1},
		{Name: "闲聊灌水", Description: "随便聊聊，放松一下", Icon: "☕", SortOrder: 5, Status: 1},
	}

	for _, node := range defaultNodes {
		var existing models.ForumNode
		result := DB.Where("name = ?", node.Name).First(&existing)
		if result.Error != nil {
			if err := DB.Create(&node).Error; err != nil {
				return fmt.Errorf("创建论坛节点失败: %w", err)
			}
			log.Printf("✅ 创建论坛节点: %s", node.Name)
		}
	}

	return nil
}
