// Package service IP黑名单服务
package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IPBlacklistService IP黑名单服务
type IPBlacklistService struct {
	db    *gorm.DB
	cache map[string]*models.IPBlacklist
	mu    sync.RWMutex
}

// NewIPBlacklistService 创建IP黑名单服务
func NewIPBlacklistService(db *gorm.DB) *IPBlacklistService {
	s := &IPBlacklistService{
		db:    db,
		cache: make(map[string]*models.IPBlacklist),
	}
	// 启动时加载缓存
	s.loadCache()
	// 启动定期清理
	go s.cleanExpired()
	return s
}

// loadCache 加载缓存
func (s *IPBlacklistService) loadCache() {
	var list []models.IPBlacklist
	s.db.Find(&list)

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range list {
		s.cache[list[i].IP] = &list[i]
	}
}

// cleanExpired 定期清理过期记录
func (s *IPBlacklistService) cleanExpired() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 清理数据库中的过期记录
		s.db.Where("expire_at IS NOT NULL AND expire_at < ?", time.Now()).Delete(&models.IPBlacklist{})
		// 重新加载缓存
		s.loadCache()
	}
}

// IsBlocked 检查IP是否被封禁
func (s *IPBlacklistService) IsBlocked(ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, exists := s.cache[ip]
	if !exists {
		return false
	}

	// 检查是否过期
	if record.ExpireAt != nil && record.ExpireAt.Before(time.Now()) {
		return false
	}

	return true
}

// Add 添加IP到黑名单
type AddIPRequest struct {
	IP       string `json:"ip" binding:"required"`
	Reason   string `json:"reason"`
	Duration int    `json:"duration"` // 封禁时长（分钟），0表示永久
}

func (s *IPBlacklistService) Add(adminID uuid.UUID, req *AddIPRequest) error {
	// 检查是否已存在
	var count int64
	s.db.Model(&models.IPBlacklist{}).Where("ip = ?", req.IP).Count(&count)
	if count > 0 {
		return errors.New("该IP已在黑名单中")
	}

	record := &models.IPBlacklist{
		IP:        req.IP,
		Reason:    req.Reason,
		CreatedBy: adminID,
	}

	if req.Duration > 0 {
		expireAt := time.Now().Add(time.Duration(req.Duration) * time.Minute)
		record.ExpireAt = &expireAt
	}

	if err := s.db.Create(record).Error; err != nil {
		return errors.New("添加失败")
	}

	// 更新缓存
	s.mu.Lock()
	s.cache[req.IP] = record
	s.mu.Unlock()

	return nil
}

// Remove 从黑名单移除IP
func (s *IPBlacklistService) Remove(ip string) error {
	result := s.db.Where("ip = ?", ip).Delete(&models.IPBlacklist{})
	if result.RowsAffected == 0 {
		return errors.New("IP不在黑名单中")
	}

	// 更新缓存
	s.mu.Lock()
	delete(s.cache, ip)
	s.mu.Unlock()

	return nil
}

// GetList 获取黑名单列表
func (s *IPBlacklistService) GetList(page, pageSize int) ([]models.IPBlacklist, int64, error) {
	var list []models.IPBlacklist
	var total int64

	s.db.Model(&models.IPBlacklist{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := s.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// AutoBlock 自动拉黑IP（系统检测到攻击行为时调用）
// 如果是第二次或更多次被封禁，则永久封禁
func (s *IPBlacklistService) AutoBlock(ip string, reason string, durationMinutes int) (bool, error) {
	// 检查是否已存在于当前黑名单
	s.mu.RLock()
	_, exists := s.cache[ip]
	s.mu.RUnlock()
	if exists {
		return false, nil // 已在黑名单中
	}

	// 查询历史封禁记录（包括已过期的）
	var historyCount int64
	s.db.Unscoped().Model(&models.IPBlacklist{}).Where("ip = ?", ip).Count(&historyCount)

	// 如果之前被封禁过，这次直接永久封禁
	isPermanent := historyCount > 0
	blockCount := int(historyCount) + 1

	record := &models.IPBlacklist{
		IP:         ip,
		Reason:     reason,
		BlockCount: blockCount,
	}

	if isPermanent {
		// 永久封禁，不设置过期时间
		record.Reason = fmt.Sprintf("%s（累计第%d次违规，永久封禁）", reason, blockCount)
	} else if durationMinutes > 0 {
		expireAt := time.Now().Add(time.Duration(durationMinutes) * time.Minute)
		record.ExpireAt = &expireAt
	}

	if err := s.db.Create(record).Error; err != nil {
		return isPermanent, err
	}

	// 更新缓存
	s.mu.Lock()
	s.cache[ip] = record
	s.mu.Unlock()

	return isPermanent, nil
}

// GetBlockCount 获取IP的历史封禁次数
func (s *IPBlacklistService) GetBlockCount(ip string) int {
	var count int64
	s.db.Unscoped().Model(&models.IPBlacklist{}).Where("ip = ?", ip).Count(&count)
	return int(count)
}
