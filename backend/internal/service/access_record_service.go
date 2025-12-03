package service

import (
	"embyhub/internal/dao"
	"embyhub/internal/model"
	"embyhub/pkg/database"
	"time"
)

type AccessRecordService struct {
	recordDAO *dao.AccessRecordDAO
	userDAO   *dao.UserDAO
}

func NewAccessRecordService() *AccessRecordService {
	return &AccessRecordService{
		recordDAO: dao.NewAccessRecordDAO(),
		userDAO:   dao.NewUserDAO(),
	}
}

// Create 创建访问记录
func (s *AccessRecordService) Create(req *model.AccessRecordCreateRequest) error {
	record := &model.AccessRecord{
		UserID:     req.UserID,
		Resource:   req.Resource,
		IPAddress:  req.IPAddress,
		DeviceInfo: req.DeviceInfo,
		AccessTime: time.Now(),
	}

	return s.recordDAO.Create(record)
}

// List 获取访问记录列表
func (s *AccessRecordService) List(req *model.AccessRecordListRequest) (*model.AccessRecordListResponse, error) {
	records, total, err := s.recordDAO.List(req)
	if err != nil {
		return nil, err
	}

	return &model.AccessRecordListResponse{
		Total: int(total),
		List:  records,
	}, nil
}

// GetStatistics 获取统计数据（带缓存）
func (s *AccessRecordService) GetStatistics() (*model.StatisticsResponse, error) {
	// 尝试从缓存获取
	if cached, err := Cache().GetStatistics(); err == nil {
		return cached, nil
	}

	// 总用户数
	totalUsers, err := s.userDAO.Count()
	if err != nil {
		return nil, err
	}

	// 活跃用户数（24小时内有访问）
	activeUsers, err := s.recordDAO.CountActiveUsers()
	if err != nil {
		return nil, err
	}

	// 今日访问次数
	todayAccess, err := s.recordDAO.CountToday()
	if err != nil {
		return nil, err
	}

	// Top5访问用户
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	topUsers, err := s.recordDAO.GetTopUsers(5, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// 访问趋势（最近7天）
	accessTrend, err := s.recordDAO.GetAccessTrend(7)
	if err != nil {
		return nil, err
	}

	// VIP统计
	vipStats := s.getVipStatistics()

	// 卡密统计
	cardKeyStats := s.getCardKeyStatistics()

	// 用户增长趋势
	userGrowth := s.getUserGrowthTrend(7)

	result := &model.StatisticsResponse{
		TotalUsers:   totalUsers,
		ActiveUsers:  activeUsers,
		TodayAccess:  todayAccess,
		TopUsers:     topUsers,
		AccessTrend:  accessTrend,
		UserGrowth:   userGrowth,
		VipStats:     vipStats,
		CardKeyStats: cardKeyStats,
	}

	// 存入缓存（1分钟）
	Cache().SetStatistics(result)

	return result, nil
}

// getVipStatistics 获取VIP统计
func (s *AccessRecordService) getVipStatistics() *model.VipStatistics {
	now := time.Now()
	stats := &model.VipStatistics{}

	// 有效VIP总数
	database.DB.Model(&model.User{}).
		Where("vip_level = 1 AND vip_expire_at > ?", now).
		Count(&stats.TotalVip)

	// 已过期VIP
	database.DB.Model(&model.User{}).
		Where("vip_level = 1 AND vip_expire_at <= ?", now).
		Count(&stats.ExpiredVip)

	// 3天内到期
	database.DB.Model(&model.User{}).
		Where("vip_level = 1 AND vip_expire_at > ? AND vip_expire_at <= ?", now, now.AddDate(0, 0, 3)).
		Count(&stats.Expiring3Day)

	// 7天内到期
	database.DB.Model(&model.User{}).
		Where("vip_level = 1 AND vip_expire_at > ? AND vip_expire_at <= ?", now, now.AddDate(0, 0, 7)).
		Count(&stats.Expiring7Day)

	return stats
}

// getCardKeyStatistics 获取卡密统计
func (s *AccessRecordService) getCardKeyStatistics() *model.CardKeyStatistics {
	stats := &model.CardKeyStatistics{}

	// 总数
	database.DB.Model(&model.CardKey{}).Count(&stats.TotalCards)

	// 未使用
	database.DB.Model(&model.CardKey{}).Where("status = 1").Count(&stats.UnusedCards)

	// 已使用
	database.DB.Model(&model.CardKey{}).Where("status = 2").Count(&stats.UsedCards)

	// 已禁用
	database.DB.Model(&model.CardKey{}).Where("status = 0").Count(&stats.DisabledCards)

	return stats
}

// getUserGrowthTrend 获取用户增长趋势
func (s *AccessRecordService) getUserGrowthTrend(days int) []*model.GrowthTrendItem {
	result := make([]*model.GrowthTrendItem, 0)
	now := time.Now()

	// 查询每天新增用户数
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		item := &model.GrowthTrendItem{
			Date: startOfDay.Format("01-02"),
		}

		// 当天新增用户
		database.DB.Model(&model.User{}).
			Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
			Count(&item.NewUsers)

		// 截止当天累计用户
		database.DB.Model(&model.User{}).
			Where("created_at < ?", endOfDay).
			Count(&item.TotalUsers)

		result = append(result, item)
	}

	return result
}
