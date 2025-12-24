// Package service 支付宝支付服务
package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"feiniu-user-system/internal/database"
	"feiniu-user-system/internal/models"

	"github.com/google/uuid"
	"github.com/smartwalle/alipay/v3"
	"gorm.io/gorm"
)

// AlipayService 支付宝服务
type AlipayService struct {
	db            *gorm.DB
	encryptionKey []byte
	client        *alipay.Client
}

// NewAlipayService 创建支付宝服务
func NewAlipayService(db *gorm.DB, encryptionKey string) *AlipayService {
	// 使用MD5生成32字节密钥
	hash := md5.Sum([]byte(encryptionKey))
	key := make([]byte, 32)
	copy(key[:16], hash[:])
	copy(key[16:], hash[:])

	return &AlipayService{
		db:            db,
		encryptionKey: key,
	}
}

// =====================================================
// 加密解密方法
// =====================================================

// EncryptSecret AES加密
func (s *AlipayService) EncryptSecret(plaintext string) (string, error) {
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

// DecryptSecret AES解密
func (s *AlipayService) DecryptSecret(ciphertext string) (string, error) {
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

// AlipayConfigResponse 配置响应
type AlipayConfigResponse struct {
	ID              uint   `json:"id"`
	AppID           string `json:"app_id"`
	AppPublicKey    string `json:"app_public_key"`
	AppPrivateKey   string `json:"app_private_key"`   // 脱敏显示
	AlipayPublicKey string `json:"alipay_public_key"`
	NotifyURL       string `json:"notify_url"`
	Enabled         bool   `json:"enabled"`
	IsProduction    bool   `json:"is_production"`
}

// GetConfig 获取配置
func (s *AlipayService) GetConfig() (*AlipayConfigResponse, error) {
	var config models.AlipayConfig
	err := s.db.First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AlipayConfigResponse{}, nil
		}
		return nil, err
	}

	// 解密私钥
	privateKey, _ := s.DecryptSecret(config.AppPrivateKeyEncrypted)

	return &AlipayConfigResponse{
		ID:              config.ID,
		AppID:           config.AppID,
		AppPublicKey:    config.AppPublicKey,
		AppPrivateKey:   privateKey,
		AlipayPublicKey: config.AlipayPublicKey,
		NotifyURL:       config.NotifyURL,
		Enabled:         config.Enabled,
		IsProduction:    config.IsProduction,
	}, nil
}

// SaveConfigRequest 保存配置请求
type AlipaySaveConfigRequest struct {
	AppID           string `json:"app_id" binding:"required"`
	AppPublicKey    string `json:"app_public_key"`
	AppPrivateKey   string `json:"app_private_key"`
	AlipayPublicKey string `json:"alipay_public_key" binding:"required"`
	NotifyURL       string `json:"notify_url" binding:"required"`
	Enabled         bool   `json:"enabled"`
	IsProduction    bool   `json:"is_production"`
}

// SaveConfig 保存配置
func (s *AlipayService) SaveConfig(req *AlipaySaveConfigRequest) error {
	var config models.AlipayConfig
	err := s.db.First(&config).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新配置
		if req.AppPrivateKey == "" {
			return errors.New("首次配置必须填写应用私钥")
		}

		privateKeyEnc, err := s.EncryptSecret(req.AppPrivateKey)
		if err != nil {
			return fmt.Errorf("加密私钥失败: %w", err)
		}

		config = models.AlipayConfig{
			AppID:                  req.AppID,
			AppPublicKey:           req.AppPublicKey,
			AppPrivateKeyEncrypted: privateKeyEnc,
			AlipayPublicKey:        req.AlipayPublicKey,
			NotifyURL:              req.NotifyURL,
			Enabled:                req.Enabled,
			IsProduction:           req.IsProduction,
		}
		return s.db.Create(&config).Error
	}

	if err != nil {
		return err
	}

	// 更新配置
	updates := map[string]interface{}{
		"app_id":            req.AppID,
		"app_public_key":    req.AppPublicKey,
		"alipay_public_key": req.AlipayPublicKey,
		"notify_url":        req.NotifyURL,
		"enabled":           req.Enabled,
		"is_production":     req.IsProduction,
	}

	if req.AppPrivateKey != "" && !strings.Contains(req.AppPrivateKey, "****") {
		privateKeyEnc, err := s.EncryptSecret(req.AppPrivateKey)
		if err != nil {
			return fmt.Errorf("加密私钥失败: %w", err)
		}
		updates["app_private_key_encrypted"] = privateKeyEnc
	}

	return s.db.Model(&config).Updates(updates).Error
}

// GetDecryptedConfig 获取解密后的配置（内部使用）
func (s *AlipayService) GetDecryptedConfig() (*models.AlipayConfig, string, error) {
	var config models.AlipayConfig
	if err := s.db.First(&config).Error; err != nil {
		return nil, "", err
	}

	privateKey, err := s.DecryptSecret(config.AppPrivateKeyEncrypted)
	if err != nil {
		return nil, "", fmt.Errorf("解密私钥失败: %w", err)
	}

	return &config, privateKey, nil
}

// InitClient 初始化支付宝客户端
func (s *AlipayService) InitClient() (*alipay.Client, error) {
	config, privateKey, err := s.GetDecryptedConfig()
	if err != nil {
		return nil, err
	}

	if !config.Enabled {
		return nil, errors.New("支付宝支付未启用")
	}

	client, err := alipay.New(config.AppID, privateKey, config.IsProduction)
	if err != nil {
		return nil, fmt.Errorf("初始化支付宝客户端失败: %w", err)
	}

	// 加载支付宝公钥
	if err := client.LoadAliPayPublicKey(config.AlipayPublicKey); err != nil {
		return nil, fmt.Errorf("加载支付宝公钥失败: %w", err)
	}

	s.client = client
	return client, nil
}


// =====================================================
// 订单号生成
// =====================================================

var orderNoCounter uint64
var orderNoMutex sync.Mutex

// GenerateOrderNo 生成唯一订单号
// 格式: VIP{timestamp14}{counter4}{random2} = 23位
func GenerateOrderNo() string {
	orderNoMutex.Lock()
	orderNoCounter++
	counter := orderNoCounter
	orderNoMutex.Unlock()

	timestamp := time.Now().Format("20060102150405")
	
	// 使用计数器(4位) + 随机数(2位) 确保唯一性
	n, err := rand.Int(rand.Reader, big.NewInt(100))
	random := int64(0)
	if err == nil {
		random = n.Int64()
	}
	
	return fmt.Sprintf("VIP%s%04d%02d", timestamp, counter%10000, random)
}

// =====================================================
// 支付功能
// =====================================================

// PaymentResult 支付结果
type PaymentResult struct {
	OrderNo string `json:"order_no"`
	QRCode  string `json:"qr_code"`
	Amount  int64  `json:"amount"`
}

// CreatePayment 创建预支付订单
func (s *AlipayService) CreatePayment(userID uuid.UUID, planID uint) (*PaymentResult, error) {
	startTime := time.Now()

	// 1. 获取配置并初始化客户端
	config, _, err := s.GetDecryptedConfig()
	if err != nil {
		s.logAPICall("", models.AlipayLogActionCreate, "", "", models.AlipayLogStatusFailed, err.Error(), "", 0)
		return nil, errors.New("支付功能暂未开放")
	}

	if !config.Enabled {
		return nil, errors.New("支付功能暂未开放")
	}

	client, err := s.InitClient()
	if err != nil {
		s.logAPICall("", models.AlipayLogActionCreate, "", "", models.AlipayLogStatusFailed, err.Error(), "", 0)
		return nil, errors.New("支付服务初始化失败")
	}

	// 2. 查询套餐信息
	var plan models.VipPlan
	if err := s.db.First(&plan, planID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("套餐不存在")
		}
		return nil, err
	}

	if !plan.IsActive {
		return nil, errors.New("套餐已下架")
	}

	// 3. 生成订单号
	orderNo := GenerateOrderNo()

	// 4. 创建本地订单
	expireAt := time.Now().Add(30 * time.Minute) // 订单30分钟过期
	order := models.AlipayVipOrder{
		OrderNo:       orderNo,
		UserID:        userID,
		PlanID:        planID,
		PlanName:      plan.Name,
		Amount:        plan.Price,
		PaymentMethod: models.PaymentMethodAlipay,
		Status:        models.AlipayOrderStatusPending,
		ExpireAt:      &expireAt,
	}

	if err := s.db.Create(&order).Error; err != nil {
		s.logAPICall(orderNo, models.AlipayLogActionCreate, "", "", models.AlipayLogStatusFailed, err.Error(), "", 0)
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	// 5. 调用支付宝预创建接口
	var p = alipay.TradePreCreate{}
	p.NotifyURL = config.NotifyURL
	p.Subject = fmt.Sprintf("VIP会员-%s", plan.Name)
	p.OutTradeNo = orderNo
	p.TotalAmount = fmt.Sprintf("%.2f", float64(plan.Price)/100) // 分转元

	resp, err := client.TradePreCreate(context.Background(), p)
	if err != nil {
		// 更新订单状态为失败
		s.db.Model(&order).Update("status", models.AlipayOrderStatusFailed)
		s.logAPICall(orderNo, models.AlipayLogActionCreate, fmt.Sprintf("%+v", p), "", models.AlipayLogStatusFailed, err.Error(), "", time.Since(startTime).Milliseconds())
		return nil, errors.New("调用支付宝接口失败")
	}

	if !resp.IsSuccess() {
		s.db.Model(&order).Update("status", models.AlipayOrderStatusFailed)
		errMsg := fmt.Sprintf("%s: %s", resp.SubCode, resp.SubMsg)
		s.logAPICall(orderNo, models.AlipayLogActionCreate, fmt.Sprintf("%+v", p), fmt.Sprintf("%+v", resp), models.AlipayLogStatusFailed, errMsg, "", time.Since(startTime).Milliseconds())
		return nil, fmt.Errorf("支付宝返回错误: %s", resp.SubMsg)
	}

	// 6. 记录成功日志
	s.logAPICall(orderNo, models.AlipayLogActionCreate, fmt.Sprintf("%+v", p), fmt.Sprintf("%+v", resp), models.AlipayLogStatusSuccess, "", "", time.Since(startTime).Milliseconds())

	return &PaymentResult{
		OrderNo: orderNo,
		QRCode:  resp.QRCode,
		Amount:  plan.Price,
	}, nil
}

// =====================================================
// 异步通知处理
// =====================================================

// HandleNotify 处理支付宝异步通知
func (s *AlipayService) HandleNotify(params url.Values) error {
	startTime := time.Now()
	orderNo := params.Get("out_trade_no")

	// 1. 初始化客户端并验证签名
	client, err := s.InitClient()
	if err != nil {
		s.logAPICall(orderNo, models.AlipayLogActionNotify, params.Encode(), "", models.AlipayLogStatusFailed, "初始化客户端失败: "+err.Error(), "", 0)
		return err
	}

	// 2. 验证签名
	if err := client.VerifySign(params); err != nil {
		s.logAPICall(orderNo, models.AlipayLogActionNotify, params.Encode(), "", models.AlipayLogStatusFailed, "签名验证失败: "+err.Error(), "", time.Since(startTime).Milliseconds())
		return errors.New("签名验证失败")
	}

	// 3. 解析通知
	noti, err := client.DecodeNotification(params)
	if err != nil {
		s.logAPICall(orderNo, models.AlipayLogActionNotify, params.Encode(), "", models.AlipayLogStatusFailed, "解析通知失败: "+err.Error(), "", time.Since(startTime).Milliseconds())
		return err
	}

	// 4. 查询订单
	var order models.AlipayVipOrder
	if err := s.db.Where("order_no = ?", noti.OutTradeNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logAPICall(orderNo, models.AlipayLogActionNotify, params.Encode(), "", models.AlipayLogStatusFailed, "订单不存在", "", time.Since(startTime).Milliseconds())
			return errors.New("订单不存在")
		}
		return err
	}

	// 5. 幂等检查：如果订单已处理，直接返回成功
	if order.Status == models.AlipayOrderStatusSuccess {
		s.logAPICall(orderNo, models.AlipayLogActionNotify, params.Encode(), "已处理(幂等)", models.AlipayLogStatusSuccess, "", "", time.Since(startTime).Milliseconds())
		return nil
	}

	// 6. 检查交易状态
	if noti.TradeStatus != "TRADE_SUCCESS" && noti.TradeStatus != "TRADE_FINISHED" {
		s.logAPICall(orderNo, models.AlipayLogActionNotify, params.Encode(), string(noti.TradeStatus), models.AlipayLogStatusSuccess, "", "", time.Since(startTime).Milliseconds())
		return nil // 非成功状态，不处理
	}

	// 7. 验证金额
	notifyAmount, _ := strconv.ParseFloat(noti.TotalAmount, 64)
	orderAmount := float64(order.Amount) / 100
	if notifyAmount != orderAmount {
		errMsg := fmt.Sprintf("金额不匹配: 通知金额=%.2f, 订单金额=%.2f", notifyAmount, orderAmount)
		s.logAPICall(orderNo, models.AlipayLogActionNotify, params.Encode(), "", models.AlipayLogStatusFailed, errMsg, "", time.Since(startTime).Milliseconds())
		return errors.New("金额不匹配")
	}

	// 8. 开启事务处理
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 9. 更新订单状态
	now := time.Now()
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":   models.AlipayOrderStatusSuccess,
		"trade_no": noti.TradeNo,
		"paid_at":  now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 10. 开通/续期VIP
	expireAt, isRenew, grantErr := s.grantVIP(tx, order.UserID, order.PlanID)
	if grantErr != nil {
		tx.Rollback()
		s.logAPICall(orderNo, models.AlipayLogActionNotify, params.Encode(), "", models.AlipayLogStatusFailed, "开通VIP失败: "+grantErr.Error(), "", time.Since(startTime).Milliseconds())
		return grantErr
	}

	// 查询套餐信息用于通知
	var plan models.VipPlan
	tx.First(&plan, order.PlanID)

	// 11. 创建余额流水记录
	balanceLog := models.BalanceLog{
		UserID:        order.UserID,
		ChangeAmount:  -order.Amount, // 负数表示支出
		BeforeBalance: 0,
		AfterBalance:  0,
		Type:          models.BalanceTypeAlipayVipPurchase,
		OrderNo:       order.OrderNo,
		Remark:        fmt.Sprintf("支付宝购买VIP: %s", order.PlanName),
	}
	if err := tx.Create(&balanceLog).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 12. 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 13. 发送VIP开通/续费通知（异步，不影响主流程）
	s.sendVipNotification(order.UserID, plan.Name, plan.DurationDays, expireAt, isRenew)

	s.logAPICall(orderNo, models.AlipayLogActionNotify, params.Encode(), "处理成功", models.AlipayLogStatusSuccess, "", "", time.Since(startTime).Milliseconds())
	return nil
}

// grantVIP 开通/续期VIP，返回新的到期时间和是否为续期
// 统一使用 users.member_expire 字段记录会员到期时间
func (s *AlipayService) grantVIP(tx *gorm.DB, userID uuid.UUID, planID uint) (expireAt time.Time, isRenew bool, err error) {
	// 查询套餐
	var plan models.VipPlan
	if err = tx.First(&plan, planID).Error; err != nil {
		return
	}

	// 查询用户当前会员状态（从 users 表）
	var user models.User
	if err = tx.Select("id, member_expire, member_level").Where("id = ?", userID).First(&user).Error; err != nil {
		return
	}

	now := time.Now()
	beforeExpire := user.MemberExpire // 记录变动前的到期时间

	// 计算新的到期时间
	if user.MemberExpire != nil && user.MemberExpire.After(now) {
		// VIP未过期，在现有基础上续期
		expireAt = user.MemberExpire.AddDate(0, 0, plan.DurationDays)
		isRenew = true
	} else {
		// VIP已过期或未开通，从当前时间开始
		expireAt = now.AddDate(0, 0, plan.DurationDays)
		isRenew = user.MemberExpire != nil // 如果之前有过期时间，说明是续期
	}

	// 更新 users 表的会员信息（唯一数据源）
	err = tx.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"member_expire": expireAt,
		"member_level":  models.MemberMonth, // 设置为月卡会员
		"role":          models.RoleMember,  // 升级为会员用户
		"status":        1,                  // 确保账户状态正常
	}).Error
	if err != nil {
		return
	}

	// 清除用户信息缓存，确保下次获取时读取最新数据
	ctx := context.Background()
	cacheKey := "user:info:" + userID.String()
	_ = database.DeleteCache(ctx, cacheKey)

	// 记录会员变动日志
	changeLog := models.MemberChangeLog{
		UserID:         userID,
		Source:         models.MemberSourceAlipay,
		ChangeDays:     plan.DurationDays,
		Amount:         plan.Price,
		BeforeExpireAt: beforeExpire,
		AfterExpireAt:  &expireAt,
		Remark:         fmt.Sprintf("支付宝购买VIP: %s", plan.Name),
	}
	err = tx.Create(&changeLog).Error

	return
}

// sendVipNotification 发送VIP开通/续费通知（邮件）
func (s *AlipayService) sendVipNotification(userID uuid.UUID, planName string, days int, expireAt time.Time, isRenew bool) {
	// 查询用户信息
	var user models.User
	if err := s.db.Select("email, username, nickname").First(&user, userID).Error; err != nil {
		return
	}

	// 获取邮件服务
	emailService := GetEmailServiceFromDB(s.db)
	if emailService == nil || !emailService.IsEnabled() {
		return
	}

	// 发送邮件通知
	username := user.Nickname
	if username == "" {
		username = user.Username
	}
	expireDateStr := expireAt.Format("2006-01-02 15:04")
	
	go func() {
		_ = emailService.SendMemberActivated(user.Email, username, days, expireDateStr, isRenew)
	}()
}


// =====================================================
// 订单查询
// =====================================================

// OrderStatus 订单状态响应
type OrderStatus struct {
	OrderNo     string     `json:"order_no"`
	TradeNo     string     `json:"trade_no"`
	TradeStatus string     `json:"trade_status"`
	Amount      int64      `json:"amount"`
	PlanName    string     `json:"plan_name"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	PaidAt      *time.Time `json:"paid_at"`
}

// QueryOrderStatus 查询订单状态
func (s *AlipayService) QueryOrderStatus(orderNo string, userID uuid.UUID) (*OrderStatus, error) {
	var order models.AlipayVipOrder
	query := s.db.Where("order_no = ?", orderNo)
	
	// 如果提供了userID，验证订单归属
	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		return nil, err
	}

	result := &OrderStatus{
		OrderNo:   order.OrderNo,
		TradeNo:   order.TradeNo,
		Amount:    order.Amount,
		PlanName:  order.PlanName,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
		PaidAt:    order.PaidAt,
	}

	// 如果订单是待支付状态且超过1分钟，主动查询支付宝（作为回调的补充）
	if order.Status == models.AlipayOrderStatusPending && time.Since(order.CreatedAt) > 60*time.Second {
		s.syncOrderFromAlipay(&order)
		// 重新查询订单
		s.db.Where("order_no = ?", orderNo).First(&order)
		result.Status = order.Status
		result.TradeNo = order.TradeNo
		result.PaidAt = order.PaidAt
	}

	// 映射交易状态
	switch order.Status {
	case models.AlipayOrderStatusPending:
		result.TradeStatus = "WAIT_BUYER_PAY"
	case models.AlipayOrderStatusSuccess:
		result.TradeStatus = "TRADE_SUCCESS"
	case models.AlipayOrderStatusFailed:
		result.TradeStatus = "TRADE_CLOSED"
	case models.AlipayOrderStatusClosed:
		result.TradeStatus = "TRADE_CLOSED"
	}

	return result, nil
}

// syncOrderFromAlipay 从支付宝同步订单状态（作为异步通知的补充）
func (s *AlipayService) syncOrderFromAlipay(order *models.AlipayVipOrder) {
	startTime := time.Now()
	orderNo := order.OrderNo

	// 幂等检查：如果订单已处理，直接返回
	if order.Status == models.AlipayOrderStatusSuccess {
		return
	}

	client, err := s.InitClient()
	if err != nil {
		return
	}

	var p = alipay.TradeQuery{}
	p.OutTradeNo = orderNo

	resp, err := client.TradeQuery(context.Background(), p)
	if err != nil {
		s.logAPICall(orderNo, models.AlipayLogActionQuery, fmt.Sprintf("%+v", p), "", models.AlipayLogStatusFailed, err.Error(), "", time.Since(startTime).Milliseconds())
		return
	}

	s.logAPICall(orderNo, models.AlipayLogActionQuery, fmt.Sprintf("%+v", p), fmt.Sprintf("%+v", resp), models.AlipayLogStatusSuccess, "", "", time.Since(startTime).Milliseconds())

	if !resp.IsSuccess() {
		return
	}

	// 根据支付宝返回状态更新本地订单
	switch resp.TradeStatus {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		// 支付成功，开启事务处理
		tx := s.db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		now := time.Now()
		// 更新订单状态
		if err := tx.Model(order).Updates(map[string]interface{}{
			"status":   models.AlipayOrderStatusSuccess,
			"trade_no": resp.TradeNo,
			"paid_at":  now,
		}).Error; err != nil {
			tx.Rollback()
			return
		}

		// 开通/续期VIP
		expireAt, isRenew, grantErr := s.grantVIP(tx, order.UserID, order.PlanID)
		if grantErr != nil {
			tx.Rollback()
			s.logAPICall(orderNo, models.AlipayLogActionQuery, "", "", models.AlipayLogStatusFailed, "开通VIP失败: "+grantErr.Error(), "", 0)
			return
		}

		// 查询套餐信息用于通知
		var plan models.VipPlan
		tx.First(&plan, order.PlanID)

		// 创建余额流水记录
		balanceLog := models.BalanceLog{
			UserID:        order.UserID,
			ChangeAmount:  -order.Amount,
			BeforeBalance: 0,
			AfterBalance:  0,
			Type:          models.BalanceTypeAlipayVipPurchase,
			OrderNo:       order.OrderNo,
			Remark:        fmt.Sprintf("支付宝购买VIP: %s", order.PlanName),
		}
		if err := tx.Create(&balanceLog).Error; err != nil {
			tx.Rollback()
			return
		}

		if err := tx.Commit().Error; err != nil {
			return
		}

		// 发送VIP开通/续费通知（异步，不影响主流程）
		s.sendVipNotification(order.UserID, plan.Name, plan.DurationDays, expireAt, isRenew)

	case "TRADE_CLOSED":
		// 交易关闭
		s.db.Model(order).Update("status", models.AlipayOrderStatusClosed)
	}
}

// OrderListRequest 订单列表请求
type AlipayOrderListRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	Status   string `form:"status" json:"status"`
}

// OrderListResponse 订单列表响应
type AlipayOrderListResponse struct {
	List  []OrderStatus `json:"list"`
	Total int64         `json:"total"`
}

// GetOrderList 获取用户订单列表
func (s *AlipayService) GetOrderList(userID uuid.UUID, req *AlipayOrderListRequest) (*AlipayOrderListResponse, error) {
	var orders []models.AlipayVipOrder
	var total int64

	query := s.db.Model(&models.AlipayVipOrder{}).Where("user_id = ?", userID)

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	query.Count(&total)

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, err
	}

	list := make([]OrderStatus, len(orders))
	for i, o := range orders {
		list[i] = OrderStatus{
			OrderNo:   o.OrderNo,
			TradeNo:   o.TradeNo,
			Amount:    o.Amount,
			PlanName:  o.PlanName,
			Status:    o.Status,
			CreatedAt: o.CreatedAt,
			PaidAt:    o.PaidAt,
		}
	}

	return &AlipayOrderListResponse{
		List:  list,
		Total: total,
	}, nil
}

// =====================================================
// VIP套餐查询
// =====================================================

// GetVipPlans 获取可用的VIP套餐列表
func (s *AlipayService) GetVipPlans() ([]models.VipPlan, error) {
	var plans []models.VipPlan
	if err := s.db.Where("is_active = ?", true).Order("price ASC").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

// GetAllVipPlans 获取所有VIP套餐列表（管理后台用）
func (s *AlipayService) GetAllVipPlans() ([]models.VipPlan, error) {
	var plans []models.VipPlan
	if err := s.db.Order("price ASC").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

// CreateVipPlan 创建VIP套餐
func (s *AlipayService) CreateVipPlan(name string, price int64, durationDays int, description string) (*models.VipPlan, error) {
	plan := &models.VipPlan{
		Name:         name,
		Price:        price,
		DurationDays: durationDays,
		Description:  description,
		IsActive:     true,
	}
	if err := s.db.Create(plan).Error; err != nil {
		return nil, err
	}
	return plan, nil
}

// UpdateVipPlan 更新VIP套餐
func (s *AlipayService) UpdateVipPlan(id uint, name string, price int64, durationDays int, description string, isActive bool) (*models.VipPlan, error) {
	var plan models.VipPlan
	if err := s.db.First(&plan, id).Error; err != nil {
		return nil, err
	}
	
	plan.Name = name
	plan.Price = price
	plan.DurationDays = durationDays
	plan.Description = description
	plan.IsActive = isActive
	
	if err := s.db.Save(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

// DeleteVipPlan 删除VIP套餐
func (s *AlipayService) DeleteVipPlan(id uint) error {
	return s.db.Delete(&models.VipPlan{}, id).Error
}

// ToggleVipPlanStatus 切换VIP套餐状态
func (s *AlipayService) ToggleVipPlanStatus(id uint) (*models.VipPlan, error) {
	var plan models.VipPlan
	if err := s.db.First(&plan, id).Error; err != nil {
		return nil, err
	}
	
	plan.IsActive = !plan.IsActive
	if err := s.db.Save(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

// =====================================================
// 日志记录
// =====================================================

// logAPICall 记录API调用日志
func (s *AlipayService) logAPICall(orderNo, action, request, response, status, errorMsg, clientIP string, duration int64) {
	log := models.AlipayLog{
		OrderNo:      orderNo,
		Action:       action,
		RequestBody:  request,
		ResponseBody: response,
		Status:       status,
		ErrorMsg:     errorMsg,
		ClientIP:     clientIP,
		Duration:     duration,
	}
	s.db.Create(&log)
}

// GetAPILogs 获取API日志列表
func (s *AlipayService) GetAPILogs(page, pageSize int, orderNo string) ([]models.AlipayLog, int64, error) {
	var logs []models.AlipayLog
	var total int64

	query := s.db.Model(&models.AlipayLog{})

	if orderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+orderNo+"%")
	}

	query.Count(&total)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// =====================================================
// 测试连接
// =====================================================

// =====================================================
// 会员变动记录
// =====================================================

// MemberChangeLogItem 会员变动记录项
type MemberChangeLogItem struct {
	ID             uint64     `json:"id"`
	Source         string     `json:"source"`
	SourceName     string     `json:"source_name"`
	OrderNo        string     `json:"order_no"`
	ChangeDays     int        `json:"change_days"`
	Amount         int64      `json:"amount"`
	BeforeExpireAt *time.Time `json:"before_expire_at"`
	AfterExpireAt  *time.Time `json:"after_expire_at"`
	Remark         string     `json:"remark"`
	CreatedAt      time.Time  `json:"created_at"`
}

// MemberChangeLogResponse 会员变动记录响应
type MemberChangeLogResponse struct {
	List  []MemberChangeLogItem `json:"list"`
	Total int64                 `json:"total"`
}

// GetMemberChangeLogs 获取用户会员变动记录
func (s *AlipayService) GetMemberChangeLogs(userID uuid.UUID, c interface{ Query(string) string }) (*MemberChangeLogResponse, error) {
	var logs []models.MemberChangeLog
	var total int64

	query := s.db.Model(&models.MemberChangeLog{}).Where("user_id = ?", userID)
	query.Count(&total)

	// 分页
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, err
	}

	list := make([]MemberChangeLogItem, len(logs))
	for i, log := range logs {
		list[i] = MemberChangeLogItem{
			ID:             log.ID,
			Source:         log.Source,
			SourceName:     log.GetSourceName(),
			OrderNo:        log.OrderNo,
			ChangeDays:     log.ChangeDays,
			Amount:         log.Amount,
			BeforeExpireAt: log.BeforeExpireAt,
			AfterExpireAt:  log.AfterExpireAt,
			Remark:         log.Remark,
			CreatedAt:      log.CreatedAt,
		}
	}

	return &MemberChangeLogResponse{
		List:  list,
		Total: total,
	}, nil
}

// =====================================================
// 测试连接
// =====================================================

// TestConnection 测试支付宝连接
func (s *AlipayService) TestConnection() error {
	client, err := s.InitClient()
	if err != nil {
		return err
	}

	// 使用查询接口测试连接（查询一个不存在的订单）
	var p = alipay.TradeQuery{}
	p.OutTradeNo = "TEST_CONNECTION_" + time.Now().Format("20060102150405")

	_, err = client.TradeQuery(context.Background(), p)
	// 即使订单不存在，只要能正常调用接口就说明连接正常
	// 支付宝会返回 ACQ.TRADE_NOT_EXIST 错误，这是正常的
	if err != nil && !strings.Contains(err.Error(), "TRADE_NOT_EXIST") {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	return nil
}
