// Package service 闲管家虚拟货源对接服务
package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"feiniu-user-system/internal/models"

	"gorm.io/gorm"
)

// GoofishService 闲管家服务
type GoofishService struct {
	db            *gorm.DB
	encryptionKey []byte
}

// NewGoofishService 创建闲管家服务
func NewGoofishService(db *gorm.DB, encryptionKey string) *GoofishService {
	// 使用MD5生成32字节密钥
	hash := md5.Sum([]byte(encryptionKey))
	key := make([]byte, 32)
	copy(key[:16], hash[:])
	copy(key[16:], hash[:])

	return &GoofishService{
		db:            db,
		encryptionKey: key,
	}
}

// =====================================================
// 加密解密方法
// =====================================================

// encryptSecret AES加密
func (s *GoofishService) encryptSecret(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptSecret AES解密
func (s *GoofishService) decryptSecret(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// =====================================================
// 配置管理
// =====================================================

// GoofishConfigResponse 配置响应（脱敏）
type GoofishConfigResponse struct {
	ID         uint   `json:"id"`
	AppID      int64  `json:"app_id"`
	AppSecret  string `json:"app_secret"`  // 脱敏显示
	MchID      string `json:"mch_id"`
	MchSecret  string `json:"mch_secret"`  // 脱敏显示
	Enabled    bool   `json:"enabled"`
	GatewayURL string `json:"gateway_url"` // 接口网关地址
}

// GetConfig 获取配置（脱敏显示）
func (s *GoofishService) GetConfig() (*GoofishConfigResponse, error) {
	var config models.GoofishConfig
	err := s.db.First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &GoofishConfigResponse{}, nil
		}
		return nil, err
	}

	// 解密密钥（返回完整密钥，前端用Input.Password控制显示）
	appSecret, _ := s.decryptSecret(config.AppSecretEncrypted)
	mchSecret, _ := s.decryptSecret(config.MchSecretEncrypted)

	return &GoofishConfigResponse{
		ID:        config.ID,
		AppID:     config.AppID,
		AppSecret: appSecret,
		MchID:     config.MchID,
		MchSecret: mchSecret,
		Enabled:   config.Enabled,
	}, nil
}

// maskSecret 脱敏显示密钥
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// SaveConfigRequest 保存配置请求
type SaveConfigRequest struct {
	AppID     int64  `json:"app_id" binding:"required"`
	AppSecret string `json:"app_secret"` // 为空则不更新
	MchID     string `json:"mch_id" binding:"required"`
	MchSecret string `json:"mch_secret"` // 为空则不更新
	Enabled   bool   `json:"enabled"`
}

// SaveConfig 保存配置
func (s *GoofishService) SaveConfig(req *SaveConfigRequest) error {
	var config models.GoofishConfig
	err := s.db.First(&config).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新配置
		if req.AppSecret == "" || req.MchSecret == "" {
			return errors.New("首次配置必须填写完整的密钥")
		}

		appSecretEnc, err := s.encryptSecret(req.AppSecret)
		if err != nil {
			return fmt.Errorf("加密app_secret失败: %w", err)
		}

		mchSecretEnc, err := s.encryptSecret(req.MchSecret)
		if err != nil {
			return fmt.Errorf("加密mch_secret失败: %w", err)
		}

		config = models.GoofishConfig{
			AppID:              req.AppID,
			AppSecretEncrypted: appSecretEnc,
			MchID:              req.MchID,
			MchSecretEncrypted: mchSecretEnc,
			Enabled:            req.Enabled,
		}
		return s.db.Create(&config).Error
	}

	if err != nil {
		return err
	}

	// 更新配置
	updates := map[string]interface{}{
		"app_id":  req.AppID,
		"mch_id":  req.MchID,
		"enabled": req.Enabled,
	}

	if req.AppSecret != "" && !strings.Contains(req.AppSecret, "****") {
		appSecretEnc, err := s.encryptSecret(req.AppSecret)
		if err != nil {
			return fmt.Errorf("加密app_secret失败: %w", err)
		}
		updates["app_secret_encrypted"] = appSecretEnc
	}

	if req.MchSecret != "" && !strings.Contains(req.MchSecret, "****") {
		mchSecretEnc, err := s.encryptSecret(req.MchSecret)
		if err != nil {
			return fmt.Errorf("加密mch_secret失败: %w", err)
		}
		updates["mch_secret_encrypted"] = mchSecretEnc
	}

	return s.db.Model(&config).Updates(updates).Error
}

// GetDecryptedConfig 获取解密后的配置（内部使用）
func (s *GoofishService) GetDecryptedConfig() (*models.GoofishConfig, string, string, error) {
	var config models.GoofishConfig
	if err := s.db.First(&config).Error; err != nil {
		return nil, "", "", err
	}

	appSecret, err := s.decryptSecret(config.AppSecretEncrypted)
	if err != nil {
		return nil, "", "", fmt.Errorf("解密app_secret失败: %w", err)
	}

	mchSecret, err := s.decryptSecret(config.MchSecretEncrypted)
	if err != nil {
		return nil, "", "", fmt.Errorf("解密mch_secret失败: %w", err)
	}

	return &config, appSecret, mchSecret, nil
}


// =====================================================
// 签名验证
// =====================================================

// VerifySign 验证闲管家签名
// 签名规则: md5("app_id,app_secret,bodyMd5,timestamp,mch_id,mch_secret")
func (s *GoofishService) VerifySign(mchID, timestamp, sign, body string) (int, error) {
	// 获取配置
	config, appSecret, mchSecret, err := s.GetDecryptedConfig()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 1000, errors.New("商户不存在")
		}
		return 500, fmt.Errorf("获取配置失败: %w", err)
	}

	// 检查是否启用
	if !config.Enabled {
		return 1001, errors.New("商户不可用")
	}

	// 验证mch_id
	if mchID != config.MchID {
		return 1000, errors.New("商户不存在")
	}

	// 验证时间戳（5分钟有效期）
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return 408, errors.New("时间戳格式错误")
	}

	now := time.Now().Unix()
	if now-ts > 300 || ts-now > 300 {
		return 408, errors.New("时间戳已超过有效期")
	}

	// 计算签名
	// bodyMd5 = md5(body)
	bodyMd5 := md5Hash(body)

	// signStr = "app_id,app_secret,bodyMd5,timestamp,mch_id,mch_secret"
	signStr := fmt.Sprintf("%d,%s,%s,%s,%s,%s",
		config.AppID, appSecret, bodyMd5, timestamp, mchID, mchSecret)
	
	// 调试日志
	fmt.Printf("[DEBUG] 签名验证 - body: %s\n", body)
	fmt.Printf("[DEBUG] 签名验证 - bodyMd5: %s\n", bodyMd5)
	fmt.Printf("[DEBUG] 签名验证 - signStr: %s\n", signStr)
	expectedSign := md5Hash(signStr)
	fmt.Printf("[DEBUG] 签名验证 - expectedSign: %s, receivedSign: %s\n", expectedSign, sign)

	// 验证签名
	if !strings.EqualFold(sign, expectedSign) {
		return 401, errors.New("签名错误")
	}

	return 0, nil
}

// md5Hash 计算MD5哈希
func md5Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// =====================================================
// 商品管理
// =====================================================

// GoodsListRequest 商品列表请求
type GoodsListRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	Status   *int8  `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
}

// GetGoodsList 获取商品列表（管理后台）
func (s *GoofishService) GetGoodsList(req *GoodsListRequest) ([]models.GoofishGoods, int64, error) {
	var goods []models.GoofishGoods
	var total int64

	query := s.db.Model(&models.GoofishGoods{})

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Keyword != "" {
		query = query.Where("goods_no LIKE ? OR goods_name LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	query.Count(&total)

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&goods).Error; err != nil {
		return nil, 0, err
	}

	return goods, total, nil
}

// GoofishGoodsListResponse 闲管家商品列表响应
type GoofishGoodsListResponse struct {
	GoodsNo    string `json:"goods_no"`
	GoodsName  string `json:"goods_name"`
	GoodsType  int    `json:"goods_type"` // 固定为2（卡密）
	Price      int64  `json:"price"`
	Stock      int32  `json:"stock"`      // int32 类型
	Status     int8   `json:"status"`
	UpdateTime int32  `json:"update_time"` // 更新时间戳
}

// GetGoodsListForAPI 获取商品列表（闲管家API）
func (s *GoofishService) GetGoodsListForAPI(goodsType int, page, pageSize int) ([]GoofishGoodsListResponse, int64, error) {
	var goods []models.GoofishGoods
	var total int64

	query := s.db.Model(&models.GoofishGoods{}).Where("status = ?", models.GoofishGoodsStatusOnSale)
	query.Count(&total)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&goods).Error; err != nil {
		return nil, 0, err
	}

	// 转换为API响应格式
	result := make([]GoofishGoodsListResponse, len(goods))
	for i, g := range goods {
		stock, _ := s.CalculateStock(g.CardType)
		result[i] = GoofishGoodsListResponse{
			GoodsNo:    g.GoodsNo,
			GoodsName:  g.GoodsName,
			GoodsType:  models.GoofishGoodsTypeKami,
			Price:      g.Price,
			Stock:      int32(stock),
			Status:     g.Status,
			UpdateTime: int32(g.UpdatedAt.Unix()),
		}
	}

	return result, total, nil
}

// GoofishGoodsDetailResponse 闲管家商品详情响应
type GoofishGoodsDetailResponse struct {
	GoodsNo    string `json:"goods_no"`
	GoodsName  string `json:"goods_name"`
	GoodsType  int    `json:"goods_type"`
	Price      int64  `json:"price"`
	Stock      int32  `json:"stock"`
	Status     int8   `json:"status"`
	UpdateTime int32  `json:"update_time"`
}

// GetGoodsDetail 获取商品详情
func (s *GoofishService) GetGoodsDetail(goodsNo string) (*GoofishGoodsDetailResponse, error) {
	var goods models.GoofishGoods
	if err := s.db.Where("goods_no = ?", goodsNo).First(&goods).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	stock, _ := s.CalculateStock(goods.CardType)

	return &GoofishGoodsDetailResponse{
		GoodsNo:    goods.GoodsNo,
		GoodsName:  goods.GoodsName,
		GoodsType:  models.GoofishGoodsTypeKami,
		Price:      goods.Price,
		Stock:      int32(stock),
		Status:     goods.Status,
		UpdateTime: int32(goods.UpdatedAt.Unix()),
	}, nil
}

// GetGoodsDetailWithTime 获取商品详情（带更新时间，用于闲管家API）
func (s *GoofishService) GetGoodsDetailWithTime(goodsNo string) (*GoofishGoodsDetailResponse, error) {
	return s.GetGoodsDetail(goodsNo)
}

// CalculateStock 计算商品库存（对应card_type的未使用卡密数量）
func (s *GoofishService) CalculateStock(cardType int8) (int64, error) {
	var count int64
	err := s.db.Model(&models.Card{}).
		Where("card_type = ? AND status = ?", cardType, models.CardStatusUnused).
		Count(&count).Error
	return count, err
}

// CreateGoodsRequest 创建商品请求
type CreateGoodsRequest struct {
	GoodsNo         string `json:"goods_no" binding:"required"`
	GoodsName       string `json:"goods_name" binding:"required"`
	CardType        int8   `json:"card_type" binding:"required,oneof=1 2 3 4"`
	Price           int64  `json:"price" binding:"required,min=1"`
	Status          int8   `json:"status"`
	AutoGenerate    bool   `json:"auto_generate"`
	CardPrefix      string `json:"card_prefix"`
	Duration        int    `json:"duration"`
	MaxAutoGenerate int    `json:"max_auto_generate"`
}

// CreateGoods 创建商品映射
func (s *GoofishService) CreateGoods(req *CreateGoodsRequest) error {
	// 检查商品编码是否已存在
	var count int64
	s.db.Model(&models.GoofishGoods{}).Where("goods_no = ?", req.GoodsNo).Count(&count)
	if count > 0 {
		return errors.New("商品编码已存在")
	}

	// 设置默认值
	status := req.Status
	if status == 0 {
		status = models.GoofishGoodsStatusOnSale
	}

	duration := req.Duration
	if duration == 0 {
		switch req.CardType {
		case 1:
			duration = 30
		case 2:
			duration = 90
		case 3:
			duration = 180
		case 4:
			duration = 365
		}
	}

	maxAutoGen := req.MaxAutoGenerate
	if maxAutoGen == 0 {
		maxAutoGen = 10
	}

	goods := models.GoofishGoods{
		GoodsNo:         req.GoodsNo,
		GoodsName:       req.GoodsName,
		CardType:        req.CardType,
		Price:           req.Price,
		Status:          status,
		AutoGenerate:    req.AutoGenerate,
		CardPrefix:      req.CardPrefix,
		Duration:        duration,
		MaxAutoGenerate: maxAutoGen,
	}

	return s.db.Create(&goods).Error
}

// UpdateGoodsRequest 更新商品请求
type UpdateGoodsRequest struct {
	GoodsName       string `json:"goods_name"`
	CardType        int8   `json:"card_type"`
	Price           int64  `json:"price"`
	Status          int8   `json:"status"`
	AutoGenerate    *bool  `json:"auto_generate"`
	CardPrefix      string `json:"card_prefix"`
	Duration        int    `json:"duration"`
	MaxAutoGenerate int    `json:"max_auto_generate"`
}

// UpdateGoods 更新商品映射
func (s *GoofishService) UpdateGoods(id uint, req *UpdateGoodsRequest) error {
	var goods models.GoofishGoods
	if err := s.db.First(&goods, id).Error; err != nil {
		return errors.New("商品不存在")
	}

	updates := make(map[string]interface{})
	if req.GoodsName != "" {
		updates["goods_name"] = req.GoodsName
	}
	if req.CardType > 0 {
		updates["card_type"] = req.CardType
	}
	if req.Price > 0 {
		updates["price"] = req.Price
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}
	if req.AutoGenerate != nil {
		updates["auto_generate"] = *req.AutoGenerate
	}
	if req.CardPrefix != "" {
		updates["card_prefix"] = req.CardPrefix
	}
	if req.Duration > 0 {
		updates["duration"] = req.Duration
	}
	if req.MaxAutoGenerate > 0 {
		updates["max_auto_generate"] = req.MaxAutoGenerate
	}

	return s.db.Model(&goods).Updates(updates).Error
}

// DeleteGoods 删除商品映射
func (s *GoofishService) DeleteGoods(id uint) error {
	result := s.db.Delete(&models.GoofishGoods{}, id)
	if result.RowsAffected == 0 {
		return errors.New("商品不存在")
	}
	return result.Error
}

// AutoGenerateGoodsResponse 自动生成商品响应
type AutoGenerateGoodsResponse struct {
	Created int      `json:"created"` // 新创建数量
	Skipped int      `json:"skipped"` // 跳过数量（已存在）
	Details []string `json:"details"` // 详情
}

// AutoGenerateGoods 根据系统中的卡密类型自动生成商品映射
func (s *GoofishService) AutoGenerateGoods() (*AutoGenerateGoodsResponse, error) {
	// 定义卡密类型配置
	cardTypes := []struct {
		CardType int8
		Name     string
		GoodsNo  string
		Duration int
		Price    int64 // 默认价格（分）
	}{
		{1, "月卡", "CARD_MONTH", 30, 1000},
		{2, "季卡", "CARD_QUARTER", 90, 2500},
		{3, "半年卡", "CARD_HALFYEAR", 180, 4500},
		{4, "年卡", "CARD_YEAR", 365, 8000},
	}

	resp := &AutoGenerateGoodsResponse{
		Details: make([]string, 0),
	}

	for _, ct := range cardTypes {
		// 检查该类型的卡密是否存在库存
		var stockCount int64
		s.db.Model(&models.Card{}).Where("card_type = ? AND status = ?", ct.CardType, models.CardStatusUnused).Count(&stockCount)

		// 检查商品映射是否已存在
		var existingGoods models.GoofishGoods
		err := s.db.Where("goods_no = ?", ct.GoodsNo).First(&existingGoods).Error
		
		if err == nil {
			// 已存在，跳过
			resp.Skipped++
			resp.Details = append(resp.Details, fmt.Sprintf("%s(%s): 已存在，跳过", ct.Name, ct.GoodsNo))
			continue
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("查询商品失败: %w", err)
		}

		// 创建新商品映射
		goods := models.GoofishGoods{
			GoodsNo:         ct.GoodsNo,
			GoodsName:       ct.Name,
			CardType:        ct.CardType,
			Price:           ct.Price,
			Status:          models.GoofishGoodsStatusOnSale,
			AutoGenerate:    true,  // 默认启用自动生成
			CardPrefix:      "GF",
			Duration:        ct.Duration,
			MaxAutoGenerate: 10,
		}

		if err := s.db.Create(&goods).Error; err != nil {
			return nil, fmt.Errorf("创建商品映射失败: %w", err)
		}

		resp.Created++
		resp.Details = append(resp.Details, fmt.Sprintf("%s(%s): 创建成功，库存%d", ct.Name, ct.GoodsNo, stockCount))
	}

	return resp, nil
}

// =====================================================
// 订单处理（核心功能）
// =====================================================

// CreateKamiOrderRequest 创建卡密订单请求
type CreateKamiOrderRequest struct {
	OrderNo     string `json:"order_no" binding:"required"`
	GoodsNo     string `json:"goods_no" binding:"required"`
	BuyQuantity int    `json:"buy_quantity" binding:"required,min=1"`
	MaxAmount   int64  `json:"max_amount"`
	NotifyURL   string `json:"notify_url"`
	BizOrderNo  string `json:"biz_order_no"`
	ProductID   int64  `json:"product_id"`
	ProductSku  int64  `json:"product_sku"`
	ItemID      int64  `json:"item_id"`
}

// CardItem 卡密项
type CardItem struct {
	CardNo  string `json:"card_no,omitempty"`
	CardPwd string `json:"card_pwd"`
}

// CreateKamiOrderResponse 创建卡密订单响应
type CreateKamiOrderResponse struct {
	OrderNo     string     `json:"order_no"`
	OutOrderNo  string     `json:"out_order_no"`
	OrderStatus int8       `json:"order_status"`
	OrderAmount int64      `json:"order_amount"`
	OrderTime   int64      `json:"order_time"`
	EndTime     int64      `json:"end_time,omitempty"`
	CardItems   []CardItem `json:"card_items,omitempty"`
	Remark      string     `json:"remark,omitempty"`
}

// CreateKamiOrder 创建卡密订单（核心方法）
func (s *GoofishService) CreateKamiOrder(req *CreateKamiOrderRequest, clientIP string) (*CreateKamiOrderResponse, int, error) {
	// 1. 幂等检查：检查订单是否已存在
	var existingOrder models.GoofishOrder
	err := s.db.Where("order_no = ?", req.OrderNo).First(&existingOrder).Error
	if err == nil {
		// 订单已存在，返回已有结果（幂等）
		var orderCards []models.GoofishOrderCard
		s.db.Where("order_no = ?", req.OrderNo).Find(&orderCards)

		cardItems := make([]CardItem, len(orderCards))
		for i, oc := range orderCards {
			cardItems[i] = CardItem{CardPwd: oc.CardCode}
		}

		return &CreateKamiOrderResponse{
			OrderNo:     existingOrder.OrderNo,
			OutOrderNo:  existingOrder.OutOrderNo,
			OrderStatus: existingOrder.OrderStatus,
			OrderAmount: existingOrder.OrderAmount,
			OrderTime:   existingOrder.CreatedAt.Unix(),
			EndTime:     existingOrder.UpdatedAt.Unix(),
			CardItems:   cardItems,
		}, 0, nil
	}

	// 2. 查询商品映射
	var goods models.GoofishGoods
	if err := s.db.Where("goods_no = ?", req.GoodsNo).First(&goods).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 1100, errors.New("商品不存在")
		}
		return nil, 500, err
	}

	// 检查商品状态
	if goods.Status != models.GoofishGoodsStatusOnSale {
		return nil, 1101, errors.New("商品不可用")
	}

	// 3. 计算订单金额
	orderAmount := goods.Price * int64(req.BuyQuantity)

	// 检查最大金额限制
	if req.MaxAmount > 0 && orderAmount > req.MaxAmount {
		return nil, 1202, errors.New("下单金额低于成本价")
	}

	// 4. 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 5. 获取或生成卡密
	cards, err := s.getOrGenerateCards(tx, &goods, req.BuyQuantity)
	if err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "库存不足") {
			return nil, 1102, err
		}
		return nil, 500, err
	}

	// 6. 生成本系统订单号
	outOrderNo := fmt.Sprintf("GF%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)

	// 7. 创建订单记录
	cardCodes := make([]string, len(cards))
	for i, c := range cards {
		cardCodes[i] = c.Code
	}
	cardCodesJSON, _ := json.Marshal(cardCodes)

	order := models.GoofishOrder{
		OrderNo:     req.OrderNo,
		OutOrderNo:  outOrderNo,
		BizOrderNo:  req.BizOrderNo,
		GoodsNo:     req.GoodsNo,
		GoodsName:   goods.GoodsName,
		BuyQuantity: req.BuyQuantity,
		OrderAmount: orderAmount,
		OrderStatus: models.GoofishOrderStatusSuccess,
		CardCodes:   string(cardCodesJSON),
		ClientIP:    clientIP,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, 500, fmt.Errorf("创建订单失败: %w", err)
	}

	// 8. 创建发货记录（防重复）
	for _, card := range cards {
		orderCard := models.GoofishOrderCard{
			OrderNo:  req.OrderNo,
			CardID:   card.ID,
			CardCode: card.Code,
			GoodsNo:  req.GoodsNo,
		}
		if err := tx.Create(&orderCard).Error; err != nil {
			tx.Rollback()
			return nil, 500, fmt.Errorf("创建发货记录失败: %w", err)
		}
	}

	// 9. 更新卡密状态为已使用
	cardIDs := make([]uint64, len(cards))
	for i, c := range cards {
		cardIDs[i] = c.ID
	}

	if err := tx.Model(&models.Card{}).Where("id IN ?", cardIDs).Updates(map[string]interface{}{
		"status":  models.CardStatusUsed,
		"used_at": time.Now(),
		"remark":  fmt.Sprintf("闲管家订单: %s", req.OrderNo),
	}).Error; err != nil {
		tx.Rollback()
		return nil, 500, fmt.Errorf("更新卡密状态失败: %w", err)
	}

	// 10. 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, 500, fmt.Errorf("提交事务失败: %w", err)
	}

	// 11. 构建响应
	cardItems := make([]CardItem, len(cards))
	for i, c := range cards {
		cardItems[i] = CardItem{CardPwd: c.Code}
	}

	now := time.Now()
	return &CreateKamiOrderResponse{
		OrderNo:     req.OrderNo,
		OutOrderNo:  outOrderNo,
		OrderStatus: models.GoofishOrderStatusSuccess,
		OrderAmount: orderAmount,
		OrderTime:   now.Unix(),
		EndTime:     now.Unix(),
		CardItems:   cardItems,
	}, 0, nil
}

// getOrGenerateCards 获取或生成卡密
func (s *GoofishService) getOrGenerateCards(tx *gorm.DB, goods *models.GoofishGoods, quantity int) ([]models.Card, error) {
	// 查询可用卡密
	var cards []models.Card
	err := tx.Where("card_type = ? AND status = ?", goods.CardType, models.CardStatusUnused).
		Order("id ASC").
		Limit(quantity).
		Find(&cards).Error
	if err != nil {
		return nil, err
	}

	// 库存充足
	if len(cards) >= quantity {
		return cards[:quantity], nil
	}

	// 库存不足，检查是否启用自动生成
	if !goods.AutoGenerate {
		return nil, errors.New("商品库存不足")
	}

	// 计算需要生成的数量
	needGenerate := quantity - len(cards)
	if needGenerate > goods.MaxAutoGenerate {
		return nil, fmt.Errorf("商品库存不足，需要生成%d张，超过单次最大生成数量%d", needGenerate, goods.MaxAutoGenerate)
	}

	// 生成新卡密
	newCards, err := s.generateCards(tx, goods, needGenerate)
	if err != nil {
		return nil, err
	}

	// 合并已有卡密和新生成的卡密
	cards = append(cards, newCards...)
	return cards, nil
}

// generateCards 生成新卡密
func (s *GoofishService) generateCards(tx *gorm.DB, goods *models.GoofishGoods, quantity int) ([]models.Card, error) {
	// 生成批次号
	batchNo := fmt.Sprintf("AUTO%s", time.Now().Format("20060102150405"))

	// 设置卡密前缀
	prefix := goods.CardPrefix
	if prefix == "" {
		prefix = "GF" // 默认前缀
	}

	// 生成卡密
	cards := make([]models.Card, quantity)
	for i := 0; i < quantity; i++ {
		cards[i] = models.Card{
			Code:     GenerateCardCode(prefix),
			BatchNo:  batchNo,
			CardType: goods.CardType,
			Duration: goods.Duration,
			Status:   models.CardStatusUnused,
			Remark:   "闲管家自动生成",
		}
	}

	// 批量保存
	if err := tx.CreateInBatches(cards, 100).Error; err != nil {
		return nil, fmt.Errorf("生成卡密失败: %w", err)
	}

	// 重新查询获取ID
	var savedCards []models.Card
	if err := tx.Where("batch_no = ?", batchNo).Find(&savedCards).Error; err != nil {
		return nil, err
	}

	return savedCards, nil
}


// =====================================================
// 订单查询
// =====================================================

// GetOrderDetail 获取订单详情
func (s *GoofishService) GetOrderDetail(orderNo string) (*models.GoofishOrder, []models.GoofishOrderCard, error) {
	var order models.GoofishOrder
	if err := s.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var orderCards []models.GoofishOrderCard
	s.db.Where("order_no = ?", orderNo).Find(&orderCards)

	return &order, orderCards, nil
}

// OrderDetailResponse 订单详情响应（闲管家API）
type OrderDetailResponse struct {
	OrderType   int32      `json:"order_type"`             // 订单类型：2=卡密订单
	OrderNo     string     `json:"order_no"`               // 管家订单号
	OutOrderNo  string     `json:"out_order_no"`           // 货源订单号
	OrderStatus int8       `json:"order_status"`           // 订单状态
	OrderAmount int64      `json:"order_amount"`           // 订单金额（分）
	GoodsNo     string     `json:"goods_no"`               // 商品编码
	GoodsName   string     `json:"goods_name"`             // 商品名称
	BuyQuantity int32      `json:"buy_quantity"`           // 购买数量
	OrderTime   int32      `json:"order_time"`             // 下单时间
	EndTime     int32      `json:"end_time,omitempty"`     // 完结时间
	CardItems   []CardItem `json:"card_items,omitempty"`   // 卡密列表
	Remark      string     `json:"remark,omitempty"`       // 订单备注
}

// GetOrderDetailForAPI 获取订单详情（闲管家API）
func (s *GoofishService) GetOrderDetailForAPI(orderNo string, outOrderNo string) (*OrderDetailResponse, error) {
	var order models.GoofishOrder
	var err error
	
	// 优先使用 order_no 查询，否则使用 out_order_no
	if orderNo != "" {
		err = s.db.Where("order_no = ?", orderNo).First(&order).Error
	} else if outOrderNo != "" {
		err = s.db.Where("out_order_no = ?", outOrderNo).First(&order).Error
	} else {
		return nil, nil
	}
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var orderCards []models.GoofishOrderCard
	s.db.Where("order_no = ?", order.OrderNo).Find(&orderCards)

	cardItems := make([]CardItem, len(orderCards))
	for i, oc := range orderCards {
		cardItems[i] = CardItem{CardPwd: oc.CardCode}
	}

	resp := &OrderDetailResponse{
		OrderType:   2, // 卡密订单
		OrderNo:     order.OrderNo,
		OutOrderNo:  order.OutOrderNo,
		OrderStatus: order.OrderStatus,
		OrderAmount: order.OrderAmount,
		GoodsNo:     order.GoodsNo,
		GoodsName:   order.GoodsName,
		BuyQuantity: int32(order.BuyQuantity),
		OrderTime:   int32(order.CreatedAt.Unix()),
		CardItems:   cardItems,
		Remark:      order.Remark,
	}

	if order.OrderStatus == models.GoofishOrderStatusSuccess || order.OrderStatus == models.GoofishOrderStatusFailed {
		resp.EndTime = int32(order.UpdatedAt.Unix())
	}

	return resp, nil
}

// OrderListRequest 订单列表请求
type OrderListRequest struct {
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
	OrderNo    string `form:"order_no" json:"order_no"`
	GoodsNo    string `form:"goods_no" json:"goods_no"`
	BizOrderNo string `form:"biz_order_no" json:"biz_order_no"`
	Status     *int8  `form:"status" json:"status"`
	StartTime  string `form:"start_time" json:"start_time"`
	EndTime    string `form:"end_time" json:"end_time"`
}

// GetOrderList 获取订单列表（管理后台）
func (s *GoofishService) GetOrderList(req *OrderListRequest) ([]models.GoofishOrder, int64, error) {
	var orders []models.GoofishOrder
	var total int64

	query := s.db.Model(&models.GoofishOrder{})

	if req.OrderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+req.OrderNo+"%")
	}
	if req.GoodsNo != "" {
		query = query.Where("goods_no = ?", req.GoodsNo)
	}
	if req.BizOrderNo != "" {
		query = query.Where("biz_order_no LIKE ?", "%"+req.BizOrderNo+"%")
	}
	if req.Status != nil {
		query = query.Where("order_status = ?", *req.Status)
	}
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	query.Count(&total)

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// =====================================================
// API日志
// =====================================================

// LogAPICall 记录API调用
func (s *GoofishService) LogAPICall(log *models.GoofishLog) error {
	return s.db.Create(log).Error
}

// LogListRequest 日志列表请求
type LogListRequest struct {
	Page      int    `form:"page" json:"page"`
	PageSize  int    `form:"page_size" json:"page_size"`
	Endpoint  string `form:"endpoint" json:"endpoint"`
	StartTime string `form:"start_time" json:"start_time"`
	EndTime   string `form:"end_time" json:"end_time"`
}

// GetAPILogs 获取API日志列表
func (s *GoofishService) GetAPILogs(req *LogListRequest) ([]models.GoofishLog, int64, error) {
	var logs []models.GoofishLog
	var total int64

	query := s.db.Model(&models.GoofishLog{})

	if req.Endpoint != "" {
		query = query.Where("endpoint LIKE ?", "%"+req.Endpoint+"%")
	}
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	query.Count(&total)

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// CleanOldLogs 清理30天前的日志
func (s *GoofishService) CleanOldLogs() (int64, error) {
	threshold := time.Now().AddDate(0, 0, -30)
	result := s.db.Where("created_at < ?", threshold).Delete(&models.GoofishLog{})
	return result.RowsAffected, result.Error
}


// =====================================================
// 商品订阅管理
// =====================================================

// SubscribeGoodsRequest 订阅商品变更通知请求
type SubscribeGoodsRequest struct {
	GoodsType int    `json:"goods_type" binding:"required,oneof=1 2"`
	GoodsNo   string `json:"goods_no" binding:"required"`
	Token     string `json:"token"`
	NotifyURL string `json:"notify_url"`
}

// SubscribeGoods 订阅商品变更通知
func (s *GoofishService) SubscribeGoods(req *SubscribeGoodsRequest) error {
	// 检查商品是否存在
	var goods models.GoofishGoods
	if err := s.db.Where("goods_no = ?", req.GoodsNo).First(&goods).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("商品不存在")
		}
		return err
	}

	// 检查是否已订阅（同一商品+token组合）
	var existing models.GoofishGoodsSubscription
	err := s.db.Where("goods_no = ? AND token = ?", req.GoodsNo, req.Token).First(&existing).Error
	if err == nil {
		// 已存在，更新
		return s.db.Model(&existing).Updates(map[string]interface{}{
			"notify_url":     req.NotifyURL,
			"subscribe_time": time.Now(),
		}).Error
	}

	// 创建新订阅
	subscription := models.GoofishGoodsSubscription{
		GoodsType:     req.GoodsType,
		GoodsNo:       req.GoodsNo,
		Token:         req.Token,
		NotifyURL:     req.NotifyURL,
		SubscribeTime: time.Now(),
	}

	return s.db.Create(&subscription).Error
}

// UnsubscribeGoodsRequest 取消订阅请求
type UnsubscribeGoodsRequest struct {
	GoodsType int    `json:"goods_type" binding:"required,oneof=1 2"`
	GoodsNo   string `json:"goods_no" binding:"required"`
	Token     string `json:"token" binding:"required"`
}

// UnsubscribeGoods 取消商品变更通知订阅
func (s *GoofishService) UnsubscribeGoods(req *UnsubscribeGoodsRequest) error {
	result := s.db.Where("goods_no = ? AND token = ?", req.GoodsNo, req.Token).
		Delete(&models.GoofishGoodsSubscription{})
	
	if result.RowsAffected == 0 {
		return errors.New("订阅不存在")
	}
	return result.Error
}

// SubscriptionListRequest 订阅列表请求
type SubscriptionListRequest struct {
	GoodsType int    `json:"goods_type"`
	GoodsNo   string `json:"goods_no"`
	PageNo    int    `json:"page_no"`
	PageSize  int    `json:"page_size"`
}

// SubscriptionItem 订阅项
type SubscriptionItem struct {
	GoodsType     int    `json:"goods_type"`
	GoodsNo       string `json:"goods_no"`
	SubscribeTime int32  `json:"subscribe_time"`
	Token         string `json:"token,omitempty"`
	NotifyURL     string `json:"notify_url,omitempty"`
}

// GetSubscriptionList 获取商品订阅列表
func (s *GoofishService) GetSubscriptionList(req *SubscriptionListRequest) ([]SubscriptionItem, int64, error) {
	var subscriptions []models.GoofishGoodsSubscription
	var total int64

	query := s.db.Model(&models.GoofishGoodsSubscription{})

	if req.GoodsType > 0 {
		query = query.Where("goods_type = ?", req.GoodsType)
	}
	if req.GoodsNo != "" {
		query = query.Where("goods_no = ?", req.GoodsNo)
	}

	query.Count(&total)

	pageNo := req.PageNo
	if pageNo < 1 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (pageNo - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&subscriptions).Error; err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	result := make([]SubscriptionItem, len(subscriptions))
	for i, sub := range subscriptions {
		result[i] = SubscriptionItem{
			GoodsType:     sub.GoodsType,
			GoodsNo:       sub.GoodsNo,
			SubscribeTime: int32(sub.SubscribeTime.Unix()),
			Token:         sub.Token,
			NotifyURL:     sub.NotifyURL,
		}
	}

	return result, total, nil
}

// =====================================================
// 商品变更通知（主动推送到闲管家）
// =====================================================

// GoodsChangeItem 商品变更项
type GoodsChangeItem struct {
	GoodsNo    string `json:"goods_no"`
	GoodsType  int    `json:"goods_type"`
	Price      int64  `json:"price,omitempty"`
	Stock      int32  `json:"stock,omitempty"`
	Status     int8   `json:"status"`
	ChangeTime int32  `json:"change_time"`
}

// NotifyGoodsChange 通知商品变更到所有订阅者
func (s *GoofishService) NotifyGoodsChange(goodsNo string) error {
	// 获取商品信息
	var goods models.GoofishGoods
	if err := s.db.Where("goods_no = ?", goodsNo).First(&goods).Error; err != nil {
		return fmt.Errorf("商品不存在: %w", err)
	}

	// 计算库存
	stock, _ := s.CalculateStock(goods.CardType)

	// 获取该商品的所有订阅
	var subscriptions []models.GoofishGoodsSubscription
	if err := s.db.Where("goods_no = ?", goodsNo).Find(&subscriptions).Error; err != nil {
		return fmt.Errorf("查询订阅失败: %w", err)
	}

	if len(subscriptions) == 0 {
		return nil // 没有订阅者，无需通知
	}

	// 构建通知数据
	changeItem := GoodsChangeItem{
		GoodsNo:    goods.GoodsNo,
		GoodsType:  models.GoofishGoodsTypeKami,
		Price:      goods.Price,
		Stock:      int32(stock),
		Status:     goods.Status,
		ChangeTime: int32(time.Now().Unix()),
	}

	// 获取配置用于签名
	config, appSecret, mchSecret, err := s.GetDecryptedConfig()
	if err != nil {
		return fmt.Errorf("获取配置失败: %w", err)
	}

	// 通知每个订阅者
	for _, sub := range subscriptions {
		if sub.NotifyURL == "" {
			continue
		}
		go s.sendGoodsChangeNotify(config, appSecret, mchSecret, sub, changeItem)
	}

	return nil
}

// sendGoodsChangeNotify 发送商品变更通知
func (s *GoofishService) sendGoodsChangeNotify(config *models.GoofishConfig, appSecret, mchSecret string, sub models.GoofishGoodsSubscription, item GoodsChangeItem) {
	// 构建请求体
	body := map[string]interface{}{
		"items": []GoodsChangeItem{item},
	}
	bodyBytes, _ := json.Marshal(body)
	bodyStr := string(bodyBytes)

	// 计算签名
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	bodyMd5 := md5Hash(bodyStr)
	signStr := fmt.Sprintf("%d,%s,%s,%s,%s,%s",
		config.AppID, appSecret, bodyMd5, timestamp, config.MchID, mchSecret)
	sign := md5Hash(signStr)

	// 构建URL
	notifyURL := sub.NotifyURL
	if sub.Token != "" {
		// 如果有token，替换URL中的{token}或追加到路径
		if strings.Contains(notifyURL, "{token}") {
			notifyURL = strings.Replace(notifyURL, "{token}", sub.Token, 1)
		} else if !strings.HasSuffix(notifyURL, "/") {
			notifyURL = notifyURL + "/" + sub.Token
		} else {
			notifyURL = notifyURL + sub.Token
		}
	}

	// 添加签名参数
	if strings.Contains(notifyURL, "?") {
		notifyURL = fmt.Sprintf("%s&mch_id=%s&timestamp=%s&sign=%s", notifyURL, config.MchID, timestamp, sign)
	} else {
		notifyURL = fmt.Sprintf("%s?mch_id=%s&timestamp=%s&sign=%s", notifyURL, config.MchID, timestamp, sign)
	}

	// 发送HTTP请求
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", notifyURL, strings.NewReader(bodyStr))
	if err != nil {
		fmt.Printf("[ERROR] 创建商品变更通知请求失败: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] 发送商品变更通知失败: %v, URL: %s\n", err, notifyURL)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("[INFO] 商品变更通知已发送: goods_no=%s, url=%s, status=%d\n", item.GoodsNo, notifyURL, resp.StatusCode)
}

// NotifyAllGoodsChange 通知所有商品变更（用于批量同步）
func (s *GoofishService) NotifyAllGoodsChange() error {
	// 获取所有在架商品
	var goodsList []models.GoofishGoods
	if err := s.db.Where("status = ?", models.GoofishGoodsStatusOnSale).Find(&goodsList).Error; err != nil {
		return err
	}

	for _, goods := range goodsList {
		if err := s.NotifyGoodsChange(goods.GoodsNo); err != nil {
			fmt.Printf("[WARN] 通知商品变更失败: %s, %v\n", goods.GoodsNo, err)
		}
	}

	return nil
}
