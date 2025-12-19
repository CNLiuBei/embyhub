// Package service 积分卡密服务
package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PointsCardService 积分卡密服务
type PointsCardService struct {
	db            *gorm.DB
	pointsService *PointsService
}

// NewPointsCardService 创建积分卡密服务
func NewPointsCardService(db *gorm.DB, pointsService *PointsService) *PointsCardService {
	return &PointsCardService{db: db, pointsService: pointsService}
}

// generateCode 生成卡密码
// 格式: JF-XXXXXXXXXXXXXXXX (16位纯字符)
func generateCode() string {
	const length = 16
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 去除容易混淆的字符 0O1I
	b := make([]byte, length)
	rand.Read(b)
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[int(b[i])%len(charset)]
	}
	return "JF-" + string(code)
}

// generateBatchNo 生成批次号
func generateBatchNo() string {
	return fmt.Sprintf("PB%s", time.Now().Format("20060102150405"))
}

// CreateBatch 批量生成积分卡密
func (s *PointsCardService) CreateBatch(adminID uuid.UUID, adminName string, points, quantity int, remark string) (*models.PointsCardBatch, error) {
	if points <= 0 {
		return nil, errors.New("积分数量必须大于0")
	}
	if quantity <= 0 || quantity > 1000 {
		return nil, errors.New("生成数量必须在1-1000之间")
	}

	batchNo := generateBatchNo()

	// 创建批次
	batch := &models.PointsCardBatch{
		BatchNo:       batchNo,
		Points:        points,
		Quantity:      quantity,
		CreatedBy:     adminID,
		CreatedByName: adminName,
		Remark:        remark,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(batch).Error; err != nil {
			return err
		}

		// 批量生成卡密
		cards := make([]models.PointsCard, quantity)
		for i := 0; i < quantity; i++ {
			cards[i] = models.PointsCard{
				Code:      generateCode(),
				BatchNo:   batchNo,
				Points:    points,
				Status:    0,
				CreatedBy: adminID,
			}
		}

		return tx.CreateInBatches(cards, 100).Error
	})

	if err != nil {
		return nil, err
	}

	return batch, nil
}


// Redeem 兑换积分卡密
func (s *PointsCardService) Redeem(userID uuid.UUID, code string) (int, error) {
	var card models.PointsCard
	if err := s.db.Where("code = ?", code).First(&card).Error; err != nil {
		return 0, errors.New("卡密不存在")
	}

	if card.Status == 1 {
		return 0, errors.New("卡密已被使用")
	}
	if card.Status == 3 {
		return 0, errors.New("卡密已被禁用")
	}
	if card.ExpireAt != nil && card.ExpireAt.Before(time.Now()) {
		return 0, errors.New("卡密已过期")
	}

	// 更新卡密状态
	now := time.Now()
	if err := s.db.Model(&card).Updates(map[string]interface{}{
		"status":  1,
		"used_by": userID,
		"used_at": now,
	}).Error; err != nil {
		return 0, err
	}

	// 更新批次使用数量
	s.db.Model(&models.PointsCardBatch{}).Where("batch_no = ?", card.BatchNo).
		UpdateColumn("used_count", gorm.Expr("used_count + 1"))

	// 增加用户积分
	remark := fmt.Sprintf("卡密充值 %d 积分", card.Points)
	if err := s.pointsService.AddPoints(userID, card.Points, models.PointsTypeReward, remark, code); err != nil {
		return 0, err
	}

	return card.Points, nil
}

// GetBatchList 获取批次列表
func (s *PointsCardService) GetBatchList(page, pageSize int) ([]models.PointsCardBatch, int64, error) {
	var batches []models.PointsCardBatch
	var total int64

	s.db.Model(&models.PointsCardBatch{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := s.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&batches).Error; err != nil {
		return nil, 0, err
	}

	return batches, total, nil
}

// GetCardList 获取卡密列表
func (s *PointsCardService) GetCardList(page, pageSize int, batchNo string, status *int8, keyword string) ([]models.PointsCard, int64, error) {
	var cards []models.PointsCard
	var total int64

	query := s.db.Model(&models.PointsCard{})
	if batchNo != "" {
		query = query.Where("batch_no = ?", batchNo)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if keyword != "" {
		// 支持卡密模糊查询（不区分大小写）
		query = query.Where("UPPER(code) LIKE UPPER(?)", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&cards).Error; err != nil {
		return nil, 0, err
	}

	return cards, total, nil
}

// DisableCard 禁用卡密
func (s *PointsCardService) DisableCard(id uint64) error {
	return s.db.Model(&models.PointsCard{}).Where("id = ? AND status = 0", id).Update("status", 3).Error
}

// EnableCard 启用卡密
func (s *PointsCardService) EnableCard(id uint64) error {
	return s.db.Model(&models.PointsCard{}).Where("id = ? AND status = 3", id).Update("status", 0).Error
}

// GetStats 获取统计
func (s *PointsCardService) GetStats() (map[string]interface{}, error) {
	var totalBatches, totalCards, usedCards, unusedCards int64
	var totalPoints, usedPoints int64

	s.db.Model(&models.PointsCardBatch{}).Count(&totalBatches)
	s.db.Model(&models.PointsCard{}).Count(&totalCards)
	s.db.Model(&models.PointsCard{}).Where("status = 1").Count(&usedCards)
	s.db.Model(&models.PointsCard{}).Where("status = 0").Count(&unusedCards)
	s.db.Model(&models.PointsCard{}).Select("COALESCE(SUM(points), 0)").Scan(&totalPoints)
	s.db.Model(&models.PointsCard{}).Where("status = 1").Select("COALESCE(SUM(points), 0)").Scan(&usedPoints)

	return map[string]interface{}{
		"total_batches": totalBatches,
		"total_cards":   totalCards,
		"used_cards":    usedCards,
		"unused_cards":  unusedCards,
		"total_points":  totalPoints,
		"used_points":   usedPoints,
		"unused_points": totalPoints - usedPoints,
	}, nil
}

// ExportCards 导出卡密
func (s *PointsCardService) ExportCards(batchNo string) ([]models.PointsCard, error) {
	var cards []models.PointsCard
	if err := s.db.Where("batch_no = ?", batchNo).Find(&cards).Error; err != nil {
		return nil, err
	}
	return cards, nil
}

// DeleteBatch 删除批次（仅允许删除未使用的批次）
func (s *PointsCardService) DeleteBatch(batchNo string) error {
	// 检查是否有已使用的卡密
	var usedCount int64
	s.db.Model(&models.PointsCard{}).Where("batch_no = ? AND status = 1", batchNo).Count(&usedCount)
	if usedCount > 0 {
		return errors.New("该批次有已使用的卡密，无法删除")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除卡密
		if err := tx.Where("batch_no = ?", batchNo).Delete(&models.PointsCard{}).Error; err != nil {
			return err
		}
		// 删除批次
		if err := tx.Where("batch_no = ?", batchNo).Delete(&models.PointsCardBatch{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeleteCard 删除单张卡密（仅允许删除未使用的卡密）
func (s *PointsCardService) DeleteCard(id uint64) error {
	var card models.PointsCard
	if err := s.db.First(&card, id).Error; err != nil {
		return errors.New("卡密不存在")
	}
	if card.Status == 1 {
		return errors.New("已使用的卡密无法删除")
	}
	return s.db.Delete(&card).Error
}
