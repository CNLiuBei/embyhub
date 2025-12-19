// Package service 卡密服务
package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CardService 卡密服务
type CardService struct {
	db *gorm.DB
}

// NewCardService 创建卡密服务
func NewCardService(db *gorm.DB) *CardService {
	return &CardService{db: db}
}

// GenerateCardCode 生成随机卡密码（固定24位纯字符）
// prefix: 卡密前缀，如 "True"、"VIP" 等
// 格式: Prefix-XXXXXXXXXXXXXXXXXXXXXXXX (24位纯字符)
func GenerateCardCode(prefix string) string {
	// 如果没有前缀，默认使用 "True"
	if prefix == "" {
		prefix = "True"
	}

	// 固定24位
	const length = 24
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 去除容易混淆的字符 0O1I
	b := make([]byte, length)
	rand.Read(b)
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[int(b[i])%len(charset)]
	}

	// 格式: Prefix-XXXXXXXXXXXXXXXXXXXXXXXX
	return prefix + "-" + string(code)
}

// GenerateBatchNo 生成批次号
func GenerateBatchNo() string {
	return fmt.Sprintf("B%s", time.Now().Format("20060102150405"))
}

// CreateBatchRequest 创建批次请求
type CreateBatchRequest struct {
	CardType   int8       `json:"card_type" binding:"required,oneof=1 2 3 4"` // 1月卡 2季卡 3半年卡 4年卡
	Quantity   int        `json:"quantity" binding:"required,min=1,max=1000"`
	Duration   int        `json:"duration"`  // 有效天数，不填则根据类型自动设置
	ExpireAt   *time.Time `json:"expire_at"` // 卡密过期时间
	Remark     string     `json:"remark"`
	CodePrefix string     `json:"code_prefix"` // 卡密前缀，默认"True"
}

// CreateBatch 批量生成卡密
func (s *CardService) CreateBatch(adminID uuid.UUID, adminName string, req *CreateBatchRequest) (*models.CardBatch, []models.Card, error) {
	// 设置默认有效天数
	duration := req.Duration
	if duration == 0 {
		switch req.CardType {
		case 1:
			duration = 30 // 月卡30天
		case 2:
			duration = 90 // 季卡90天
		case 3:
			duration = 180 // 半年卡180天
		case 4:
			duration = 365 // 年卡365天
		default:
			duration = 30
		}
	}

	// 设置默认卡密前缀
	codePrefix := req.CodePrefix
	if codePrefix == "" {
		codePrefix = "True" // 默认前缀
	}

	// 创建批次
	batchNo := GenerateBatchNo()
	batch := &models.CardBatch{
		BatchNo:       batchNo,
		CardType:      req.CardType,
		Duration:      duration,
		Quantity:      req.Quantity,
		ExpireAt:      req.ExpireAt,
		CreatedBy:     adminID,
		CreatedByName: adminName,
		Remark:        req.Remark,
	}

	// 开启事务
	tx := s.db.Begin()

	if err := tx.Create(batch).Error; err != nil {
		tx.Rollback()
		return nil, nil, errors.New("创建批次失败")
	}

	// 生成卡密（固定24位纯字符）
	cards := make([]models.Card, req.Quantity)
	for i := 0; i < req.Quantity; i++ {
		cards[i] = models.Card{
			Code:      GenerateCardCode(codePrefix),
			BatchNo:   batchNo,
			CardType:  req.CardType,
			Duration:  duration,
			Status:    models.CardStatusUnused,
			ExpireAt:  req.ExpireAt,
			CreatedBy: adminID,
			Remark:    req.Remark,
		}
	}

	if err := tx.CreateInBatches(cards, 100).Error; err != nil {
		tx.Rollback()
		return nil, nil, errors.New("生成卡密失败")
	}

	tx.Commit()
	return batch, cards, nil
}

// RedeemRequest 兑换请求
type RedeemRequest struct {
	Code string `json:"code" binding:"required"`
}

// Redeem 用户兑换卡密
func (s *CardService) Redeem(userID uuid.UUID, code string) (*models.MemberOrder, error) {
	// 格式化卡密码(去除空格和横杠统一处理)
	code = formatCardCode(code)

	// 查找卡密
	var card models.Card
	if err := s.db.Where("code = ?", code).First(&card).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("卡密不存在")
		}
		return nil, err
	}

	// 检查卡密状态
	if card.Status == models.CardStatusUsed {
		return nil, errors.New("卡密已被使用")
	}
	if card.Status == models.CardStatusExpired {
		return nil, errors.New("卡密已过期")
	}
	if card.Status == models.CardStatusDisable {
		return nil, errors.New("卡密已被禁用")
	}

	// 检查卡密是否过期
	if card.ExpireAt != nil && card.ExpireAt.Before(time.Now()) {
		s.db.Model(&card).Update("status", models.CardStatusExpired)
		return nil, errors.New("卡密已过期")
	}

	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 计算会员到期时间
	now := time.Now()
	expireTime := now.AddDate(0, 0, card.Duration)

	// 如果用户已是会员且未过期，则累加时间
	if user.MemberExpire != nil && user.MemberExpire.After(now) {
		expireTime = user.MemberExpire.AddDate(0, 0, card.Duration)
	}

	// 开启事务
	tx := s.db.Begin()

	// 更新卡密状态
	usedAt := time.Now()
	if err := tx.Model(&card).Updates(map[string]interface{}{
		"status":  models.CardStatusUsed,
		"used_by": userID,
		"used_at": usedAt,
	}).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("兑换失败")
	}

	// 更新批次使用数量
	tx.Model(&models.CardBatch{}).Where("batch_no = ?", card.BatchNo).
		UpdateColumn("used_count", gorm.Expr("used_count + ?", 1))

	// 升级为会员用户并更新会员到期时间
	// 同时恢复账户状态（如果账户被禁用）
	if err := tx.Model(&user).Updates(map[string]interface{}{
		"role":          models.RoleMember,  // 升级为会员用户
		"member_level":  models.MemberMonth, // 设置会员等级
		"member_expire": expireTime,
		"status":        1, // 恢复账户状态为正常
	}).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("更新会员信息失败")
	}

	// 创建兑换记录
	order := &models.MemberOrder{
		OrderNo:    fmt.Sprintf("R%s%06d", time.Now().Format("20060102150405"), card.ID%1000000),
		UserID:     userID,
		CardID:     card.ID,
		CardCode:   card.Code,
		Level:      0, // 不再使用会员等级
		Duration:   card.Duration,
		Status:     1, // 已兑换
		RedeemTime: now,
		ExpireTime: expireTime,
	}

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("创建订单失败")
	}

	tx.Commit()
	return order, nil
}

// 格式化卡密码
func formatCardCode(code string) string {
	// 移除空格
	result := ""
	for _, c := range code {
		if c != ' ' {
			result += string(c)
		}
	}
	return result
}

// GetCardList 获取卡密列表(管理员)
type CardListRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
	BatchNo  string `form:"batch_no"`
	CardType int8   `form:"card_type"`
	Status   *int8  `form:"status"`
	Code     string `form:"code"`
}

// CardWithUser 带用户名的卡密响应
type CardWithUser struct {
	models.Card
	UsedByName string `json:"used_by_name"` // 使用者用户名
}

func (s *CardService) GetCardList(req *CardListRequest) ([]CardWithUser, int64, error) {
	var cards []models.Card
	var total int64

	query := s.db.Model(&models.Card{})

	if req.BatchNo != "" {
		query = query.Where("batch_no = ?", req.BatchNo)
	}
	if req.CardType > 0 {
		query = query.Where("card_type = ?", req.CardType)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Code != "" {
		query = query.Where("code LIKE ?", "%"+req.Code+"%")
	}

	query.Count(&total)

	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&cards).Error; err != nil {
		return nil, 0, err
	}

	// 获取使用者用户名
	result := make([]CardWithUser, len(cards))
	userIDs := make([]uuid.UUID, 0)
	for _, card := range cards {
		if card.UsedBy != nil {
			userIDs = append(userIDs, *card.UsedBy)
		}
	}

	userMap := make(map[uuid.UUID]string)
	if len(userIDs) > 0 {
		var users []models.User
		s.db.Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			userMap[u.ID] = u.Username
		}
	}

	for i, card := range cards {
		result[i] = CardWithUser{Card: card}
		if card.UsedBy != nil {
			if name, ok := userMap[*card.UsedBy]; ok {
				result[i].UsedByName = name
			}
		}
	}

	return result, total, nil
}

// GetBatchList 获取批次列表
func (s *CardService) GetBatchList(page, pageSize int) ([]models.CardBatch, int64, error) {
	var batches []models.CardBatch
	var total int64

	s.db.Model(&models.CardBatch{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := s.db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&batches).Error; err != nil {
		return nil, 0, err
	}

	return batches, total, nil
}

// DisableCard 禁用卡密
func (s *CardService) DisableCard(cardID uint64) error {
	result := s.db.Model(&models.Card{}).Where("id = ? AND status = ?", cardID, models.CardStatusUnused).
		Update("status", models.CardStatusDisable)
	if result.RowsAffected == 0 {
		return errors.New("卡密不存在或已被使用")
	}
	return result.Error
}

// EnableCard 启用卡密
func (s *CardService) EnableCard(cardID uint64) error {
	result := s.db.Model(&models.Card{}).Where("id = ? AND status = ?", cardID, models.CardStatusDisable).
		Update("status", models.CardStatusUnused)
	if result.RowsAffected == 0 {
		return errors.New("卡密不存在或状态不正确")
	}
	return result.Error
}

// DeleteCard 删除卡密
func (s *CardService) DeleteCard(cardID uint64) error {
	// 检查卡密是否存在
	var card models.Card
	if err := s.db.First(&card, cardID).Error; err != nil {
		return errors.New("卡密不存在")
	}

	// 只允许删除未使用或已禁用的卡密
	if card.Status == models.CardStatusUsed {
		return errors.New("已使用的卡密不能删除")
	}

	// 删除卡密
	if err := s.db.Delete(&card).Error; err != nil {
		return errors.New("删除失败")
	}

	return nil
}

// GetUserRedeemHistory 获取用户兑换记录
func (s *CardService) GetUserRedeemHistory(userID uuid.UUID, page, pageSize int) ([]models.MemberOrder, int64, error) {
	var orders []models.MemberOrder
	var total int64

	query := s.db.Model(&models.MemberOrder{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// ExportCards 导出卡密(返回卡密码列表)
func (s *CardService) ExportCards(batchNo string) ([]string, error) {
	var cards []models.Card
	if err := s.db.Where("batch_no = ?", batchNo).Find(&cards).Error; err != nil {
		return nil, err
	}

	codes := make([]string, len(cards))
	for i, card := range cards {
		codes[i] = card.Code
	}
	return codes, nil
}

// RenewByCardRequest 使用卡密续费请求（公开接口，无需登录）
type RenewByCardRequest struct {
	Account string `json:"account" binding:"required"` // 账号或邮箱
	Code    string `json:"code" binding:"required"`    // 卡密码
}

// RenewByCard 禁用用户使用卡密续费（公开接口，无需登录）
// 允许被禁用的用户通过账号+卡密来续费并恢复账户
func (s *CardService) RenewByCard(account, code string) (*models.MemberOrder, error) {
	// 格式化卡密码
	code = formatCardCode(code)

	// 查找用户
	var user models.User
	if err := s.db.Where("username = ? OR email = ?", account, account).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 查找卡密
	var card models.Card
	if err := s.db.Where("code = ?", code).First(&card).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("卡密不存在")
		}
		return nil, err
	}

	// 检查卡密状态
	if card.Status == models.CardStatusUsed {
		return nil, errors.New("卡密已被使用")
	}
	if card.Status == models.CardStatusExpired {
		return nil, errors.New("卡密已过期")
	}
	if card.Status == models.CardStatusDisable {
		return nil, errors.New("卡密已被禁用")
	}

	// 检查卡密是否过期
	if card.ExpireAt != nil && card.ExpireAt.Before(time.Now()) {
		s.db.Model(&card).Update("status", models.CardStatusExpired)
		return nil, errors.New("卡密已过期")
	}

	// 计算会员到期时间
	now := time.Now()
	expireTime := now.AddDate(0, 0, card.Duration)

	// 如果用户已是会员且未过期，则累加时间
	if user.MemberExpire != nil && user.MemberExpire.After(now) {
		expireTime = user.MemberExpire.AddDate(0, 0, card.Duration)
	}

	// 开启事务
	tx := s.db.Begin()

	// 更新卡密状态
	usedAt := time.Now()
	if err := tx.Model(&card).Updates(map[string]interface{}{
		"status":  models.CardStatusUsed,
		"used_by": user.ID,
		"used_at": usedAt,
	}).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("兑换失败")
	}

	// 更新批次使用数量
	tx.Model(&models.CardBatch{}).Where("batch_no = ?", card.BatchNo).
		UpdateColumn("used_count", gorm.Expr("used_count + ?", 1))

	// 升级为会员用户，恢复账户状态
	if err := tx.Model(&user).Updates(map[string]interface{}{
		"role":          models.RoleMember,  // 升级为会员用户
		"member_level":  models.MemberMonth, // 设置会员等级
		"member_expire": expireTime,
		"status":        1, // 恢复账户状态为正常
	}).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("更新会员信息失败")
	}

	// 创建兑换记录
	order := &models.MemberOrder{
		OrderNo:    fmt.Sprintf("R%s%06d", time.Now().Format("20060102150405"), card.ID%1000000),
		UserID:     user.ID,
		CardID:     card.ID,
		CardCode:   card.Code,
		Level:      0,
		Duration:   card.Duration,
		Status:     1, // 已兑换
		RedeemTime: now,
		ExpireTime: expireTime,
	}

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("创建订单失败")
	}

	// 发送站内续费成功通知
	notification := &models.Notification{
		UserID:  user.ID,
		Title:   "续费成功 - 账户已恢复",
		Content: fmt.Sprintf("恭喜您续费成功！会员有效期至%s，账户已恢复正常，请重新登录。", expireTime.Format("2006-01-02")),
		Type:    2,
	}
	tx.Create(notification)

	tx.Commit()

	// 发送邮件通知
	if user.Email != "" {
		go func() {
			emailSvc := GetEmailServiceFromDB(s.db)
			if emailSvc == nil {
				return
			}
			if err := emailSvc.SendMemberActivated(user.Email, user.Nickname, card.Duration, expireTime.Format("2006-01-02"), true); err != nil {
				log.Printf("发送续费成功邮件失败: %v", err)
			}
		}()
	}

	return order, nil
}

// GetCardStats 获取卡密统计
type CardStats struct {
	TotalCards    int64 `json:"total_cards"`
	UnusedCards   int64 `json:"unused_cards"`
	UsedCards     int64 `json:"used_cards"`
	ExpiredCards  int64 `json:"expired_cards"`
	DisabledCards int64 `json:"disabled_cards"`
	TotalBatches  int64 `json:"total_batches"`
}

func (s *CardService) GetCardStats() (*CardStats, error) {
	stats := &CardStats{}

	s.db.Model(&models.Card{}).Count(&stats.TotalCards)
	s.db.Model(&models.Card{}).Where("status = ?", models.CardStatusUnused).Count(&stats.UnusedCards)
	s.db.Model(&models.Card{}).Where("status = ?", models.CardStatusUsed).Count(&stats.UsedCards)
	s.db.Model(&models.Card{}).Where("status = ?", models.CardStatusExpired).Count(&stats.ExpiredCards)
	s.db.Model(&models.Card{}).Where("status = ?", models.CardStatusDisable).Count(&stats.DisabledCards)
	s.db.Model(&models.CardBatch{}).Count(&stats.TotalBatches)

	return stats, nil
}
