// Package service 统计服务
package service

import (
	"time"

	"feiniu-user-system/internal/models"

	"gorm.io/gorm"
)

// StatsService 统计服务
type StatsService struct {
	db *gorm.DB
}

// NewStatsService 创建统计服务
func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

// DashboardStats 仪表盘统计
type DashboardStats struct {
	TotalUsers    int64 `json:"total_users"`
	TodayRegister int64 `json:"today_register"`
	WeekRegister  int64 `json:"week_register"`
	MonthRegister int64 `json:"month_register"`
	TotalMembers  int64 `json:"total_members"`
	ActiveUsers   int64 `json:"active_users"`
}

// DailyStat 每日统计
type DailyStat struct {
	Date          string `json:"date"`
	RegisterCount int64  `json:"register_count"`
}

// GetUserStats 获取用户统计
func (s *StatsService) GetUserStats() *DashboardStats {
	var stats DashboardStats
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	// 总用户数
	s.db.Model(&models.User{}).Count(&stats.TotalUsers)

	// 今日注册
	s.db.Model(&models.User{}).Where("created_at >= ?", today).Count(&stats.TodayRegister)

	// 本周注册
	s.db.Model(&models.User{}).Where("created_at >= ?", weekAgo).Count(&stats.WeekRegister)

	// 本月注册
	s.db.Model(&models.User{}).Where("created_at >= ?", monthAgo).Count(&stats.MonthRegister)

	// 会员用户数
	s.db.Model(&models.User{}).Where("member_level > 0 AND (member_expire IS NULL OR member_expire > ?)", now).Count(&stats.TotalMembers)

	// 活跃用户（7天内登录）
	s.db.Model(&models.User{}).Where("last_login_at >= ?", weekAgo).Count(&stats.ActiveUsers)

	return &stats
}

// GetDailyStats 获取每日统计
func (s *StatsService) GetDailyStats(days int) []DailyStat {
	if days <= 0 {
		days = 7
	}
	if days > 30 {
		days = 30
	}

	results := make([]DailyStat, days)
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)

		var count int64
		s.db.Model(&models.User{}).Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).Count(&count)

		results[days-1-i] = DailyStat{
			Date:          startOfDay.Format("2006-01-02"),
			RegisterCount: count,
		}
	}

	return results
}

// VisitRanking 访问排行
type VisitRanking struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Visits   int64  `json:"visits"`
}

// GetVisitRanking 获取访问排行（基于登录次数）
func (s *StatsService) GetVisitRanking(limit int) []VisitRanking {
	if limit <= 0 {
		limit = 10
	}

	type result struct {
		UserID string
		Count  int64
	}

	var results []result
	today := time.Now().Truncate(24 * time.Hour)

	s.db.Model(&models.LoginLog{}).
		Select("user_id, COUNT(*) as count").
		Where("created_at >= ? AND status = 1", today).
		Group("user_id").
		Order("count DESC").
		Limit(limit).
		Scan(&results)

	rankings := make([]VisitRanking, 0, len(results))
	for i, r := range results {
		var user models.User
		if err := s.db.Select("username", "nickname").Where("id = ?", r.UserID).First(&user).Error; err == nil {
			name := user.Nickname
			if name == "" {
				name = user.Username
			}
			rankings = append(rankings, VisitRanking{
				Rank:     i + 1,
				Username: name,
				Visits:   r.Count,
			})
		}
	}

	return rankings
}
