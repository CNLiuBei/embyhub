package dao

import (
	"embyhub/internal/model"
	"embyhub/pkg/database"
)

type CardKeyDAO struct{}

func NewCardKeyDAO() *CardKeyDAO {
	return &CardKeyDAO{}
}

// Create 创建卡密
func (d *CardKeyDAO) Create(cardKey *model.CardKey) error {
	return database.DB.Create(cardKey).Error
}

// BatchCreate 批量创建卡密
func (d *CardKeyDAO) BatchCreate(cardKeys []*model.CardKey) error {
	return database.DB.Create(&cardKeys).Error
}

// GetByID 根据ID获取卡密
func (d *CardKeyDAO) GetByID(id int) (*model.CardKey, error) {
	var cardKey model.CardKey
	err := database.DB.Preload("UsedByUser").Preload("CreatedByUser").
		Where("id = ?", id).First(&cardKey).Error
	if err != nil {
		return nil, err
	}
	return &cardKey, nil
}

// GetByCode 根据卡密码获取卡密
func (d *CardKeyDAO) GetByCode(cardCode string) (*model.CardKey, error) {
	var cardKey model.CardKey
	err := database.DB.Preload("UsedByUser").
		Where("card_code = ?", cardCode).First(&cardKey).Error
	if err != nil {
		return nil, err
	}
	return &cardKey, nil
}

// Update 更新卡密
func (d *CardKeyDAO) Update(cardKey *model.CardKey) error {
	return database.DB.Save(cardKey).Error
}

// Delete 删除卡密
func (d *CardKeyDAO) Delete(id int) error {
	return database.DB.Delete(&model.CardKey{}, id).Error
}

// BatchDelete 批量删除卡密
func (d *CardKeyDAO) BatchDelete(ids []int) error {
	return database.DB.Delete(&model.CardKey{}, ids).Error
}

// List 获取卡密列表
func (d *CardKeyDAO) List(req *model.CardKeyListRequest) ([]*model.CardKey, int64, error) {
	var cardKeys []*model.CardKey
	var total int64

	query := database.DB.Model(&model.CardKey{})

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 类型筛选
	if req.CardType != nil {
		query = query.Where("card_type = ?", *req.CardType)
	}

	// 关键词搜索
	if req.Keyword != "" {
		query = query.Where("card_code LIKE ? OR remark LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%")
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
	err := query.Preload("UsedByUser").Preload("CreatedByUser").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&cardKeys).Error

	return cardKeys, total, err
}

// CountByStatus 统计各状态数量
func (d *CardKeyDAO) CountByStatus() (map[int]int64, error) {
	type Result struct {
		Status int
		Count  int64
	}
	var results []Result

	err := database.DB.Model(&model.CardKey{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[int]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}
