package dao

import (
	"embyhub/internal/model"
	"embyhub/pkg/database"
	"time"
)

type AccessRecordDAO struct{}

func NewAccessRecordDAO() *AccessRecordDAO {
	return &AccessRecordDAO{}
}

// Create 创建访问记录
func (d *AccessRecordDAO) Create(record *model.AccessRecord) error {
	return database.DB.Create(record).Error
}

// GetByID 根据ID获取访问记录
func (d *AccessRecordDAO) GetByID(recordID int64) (*model.AccessRecord, error) {
	var record model.AccessRecord
	err := database.DB.Preload("User").Where("record_id = ?", recordID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// List 获取访问记录列表
func (d *AccessRecordDAO) List(req *model.AccessRecordListRequest) ([]*model.AccessRecord, int64, error) {
	var records []*model.AccessRecord
	var total int64

	query := database.DB.Model(&model.AccessRecord{})

	// 用户筛选
	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 时间范围筛选
	if !req.StartTime.IsZero() {
		query = query.Where("access_time >= ?", req.StartTime)
	}
	if !req.EndTime.IsZero() {
		query = query.Where("access_time <= ?", req.EndTime)
	}

	// 资源筛选
	if req.Resource != "" {
		query = query.Where("resource LIKE ?", "%"+req.Resource+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 查询列表
	err := query.Preload("User").
		Order("access_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&records).Error

	return records, total, err
}

// CountToday 统计今日访问次数
func (d *AccessRecordDAO) CountToday() (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := database.DB.Model(&model.AccessRecord{}).
		Where("access_time >= ?", today).
		Count(&count).Error
	return count, err
}

// CountByDateRange 统计时间范围内的访问次数
func (d *AccessRecordDAO) CountByDateRange(startTime, endTime time.Time) (int64, error) {
	var count int64
	err := database.DB.Model(&model.AccessRecord{}).
		Where("access_time >= ? AND access_time <= ?", startTime, endTime).
		Count(&count).Error
	return count, err
}

// GetTopUsers 获取访问次数最多的用户
func (d *AccessRecordDAO) GetTopUsers(limit int, startTime, endTime time.Time) ([]*model.TopUserItem, error) {
	var items []*model.TopUserItem

	query := database.DB.Table("access_records").
		Select("access_records.user_id, users.username, COUNT(*) as access_count").
		Joins("LEFT JOIN users ON access_records.user_id = users.user_id").
		Group("access_records.user_id, users.username").
		Order("access_count DESC").
		Limit(limit)

	if !startTime.IsZero() {
		query = query.Where("access_records.access_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("access_records.access_time <= ?", endTime)
	}

	err := query.Scan(&items).Error
	return items, err
}

// GetAccessTrend 获取访问趋势数据
func (d *AccessRecordDAO) GetAccessTrend(days int) ([]*model.AccessTrendItem, error) {
	var items []*model.AccessTrendItem

	// 使用PostgreSQL的日期计算方式
	err := database.DB.Raw(`
		SELECT DATE(access_time) as date, COUNT(*) as count 
		FROM access_records 
		WHERE access_time >= CURRENT_DATE - INTERVAL '1 day' * ?
		GROUP BY DATE(access_time) 
		ORDER BY date ASC
	`, days).Scan(&items).Error

	return items, err
}

// DeleteOldRecords 删除旧访问记录
func (d *AccessRecordDAO) DeleteOldRecords(days int) error {
	return database.DB.Exec(`
		DELETE FROM access_records 
		WHERE access_time < CURRENT_TIMESTAMP - INTERVAL '1 day' * ?
	`, days).Error
}

// CountActiveUsers 统计活跃用户数（24小时内有访问记录）
func (d *AccessRecordDAO) CountActiveUsers() (int64, error) {
	var count int64
	err := database.DB.Raw(`
		SELECT COUNT(DISTINCT user_id) 
		FROM access_records 
		WHERE access_time >= CURRENT_TIMESTAMP - INTERVAL '24 hours'
	`).Scan(&count).Error
	return count, err
}
