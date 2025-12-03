package service

import (
	"fmt"
	"math/rand"
	"time"

	"embyhub/internal/dao"
	"embyhub/internal/model"
	"embyhub/pkg/database"
	"embyhub/pkg/email"
)

// EmailService 邮件服务
type EmailService struct {
	configDAO *dao.SystemConfigDAO
	userDAO   *dao.UserDAO
}

// NewEmailService 创建邮件服务
func NewEmailService() *EmailService {
	return &EmailService{
		configDAO: dao.NewSystemConfigDAO(),
		userDAO:   dao.NewUserDAO(),
	}
}

// GetEmailClient 获取邮件客户端（公开方法）
func (s *EmailService) GetEmailClient() (*email.Client, error) {
	return s.getEmailClient()
}

// getEmailClient 获取邮件客户端
func (s *EmailService) getEmailClient() (*email.Client, error) {
	configMap, err := s.configDAO.BatchGet(email.GetConfigKeys())
	if err != nil {
		return nil, fmt.Errorf("获取邮件配置失败: %w", err)
	}

	// 根据provider检查配置
	provider := configMap["email_provider"]
	switch provider {
	case "aliyun":
		if configMap["aliyun_access_key_id"] == "" {
			return nil, fmt.Errorf("阿里云邮件未配置AccessKey")
		}
	case "resend":
		if configMap["resend_api_key"] == "" {
			return nil, fmt.Errorf("Resend未配置API Key")
		}
	case "aliyun_smtp", "smtp":
		if configMap["smtp_host"] == "" {
			return nil, fmt.Errorf("SMTP服务器未配置")
		}
	default:
		if configMap["smtp_host"] == "" {
			return nil, fmt.Errorf("SMTP服务器未配置")
		}
	}

	return email.NewClient(email.GetConfigFromDB(configMap)), nil
}

// generateCode 生成6位验证码
func (s *EmailService) generateCode() string {
	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < 6; i++ {
		code += fmt.Sprintf("%d", rand.Intn(10))
	}
	return code
}

// SendVerificationCode 发送验证码
func (s *EmailService) SendVerificationCode(req *model.SendCodeRequest) error {
	// 设置默认类型
	if req.Type == "" {
		req.Type = model.CodeTypeRegister
	}

	// 注册时检查邮箱是否已被使用
	if req.Type == model.CodeTypeRegister {
		exists, _ := s.userDAO.ExistsByEmail(req.Email)
		if exists {
			return fmt.Errorf("该邮箱已被注册")
		}
	}

	// 检查是否频繁发送（1分钟内只能发送1次）
	var recentCode model.EmailCode
	err := database.DB.Where("email = ? AND type = ? AND created_at > ?",
		req.Email, req.Type, time.Now().Add(-1*time.Minute)).
		Order("created_at DESC").First(&recentCode).Error
	if err == nil {
		return fmt.Errorf("发送过于频繁，请1分钟后重试")
	}

	// 生成验证码
	code := s.generateCode()

	// 保存到数据库
	emailCode := &model.EmailCode{
		Email:     req.Email,
		Code:      code,
		Type:      req.Type,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10分钟有效
	}
	if err := database.DB.Create(emailCode).Error; err != nil {
		return fmt.Errorf("保存验证码失败: %w", err)
	}

	// 发送邮件
	client, err := s.getEmailClient()
	if err != nil {
		return err
	}

	if err := client.SendVerificationCode(req.Email, code); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}

// VerifyCode 验证验证码
func (s *EmailService) VerifyCode(email, code, codeType string) error {
	var emailCode model.EmailCode
	err := database.DB.Where("email = ? AND code = ? AND type = ? AND used = false AND expires_at > ?",
		email, code, codeType, time.Now()).
		Order("created_at DESC").First(&emailCode).Error
	if err != nil {
		return fmt.Errorf("验证码无效或已过期")
	}

	// 标记为已使用
	database.DB.Model(&emailCode).Update("used", true)

	return nil
}

// SendPasswordResetCode 发送密码重置验证码
func (s *EmailService) SendPasswordResetCode(emailAddr string) error {
	// 检查邮箱是否存在
	exists, _ := s.userDAO.ExistsByEmail(emailAddr)
	if !exists {
		return fmt.Errorf("该邮箱未注册")
	}

	// 检查发送频率
	var recentCode model.EmailCode
	err := database.DB.Where("email = ? AND type = ? AND created_at > ?",
		emailAddr, model.CodeTypeResetPassword, time.Now().Add(-1*time.Minute)).
		Order("created_at DESC").First(&recentCode).Error
	if err == nil {
		return fmt.Errorf("发送过于频繁，请1分钟后重试")
	}

	// 生成验证码
	code := s.generateCode()

	// 保存到数据库
	emailCode := &model.EmailCode{
		Email:     emailAddr,
		Code:      code,
		Type:      model.CodeTypeResetPassword,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := database.DB.Create(emailCode).Error; err != nil {
		return fmt.Errorf("保存验证码失败: %w", err)
	}

	// 发送邮件
	client, err := s.getEmailClient()
	if err != nil {
		return err
	}

	return client.SendPasswordResetCode(emailAddr, code)
}

// ResetPassword 重置密码
func (s *EmailService) ResetPassword(emailAddr, code, newPassword string) error {
	// 验证验证码
	if err := s.VerifyCode(emailAddr, code, model.CodeTypeResetPassword); err != nil {
		return err
	}

	// 获取用户
	user, err := s.userDAO.GetByEmail(emailAddr)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 更新密码
	authService := NewAuthService()
	return authService.ChangePassword(user.UserID, newPassword)
}

// TestEmailConfig 测试邮件配置
func (s *EmailService) TestEmailConfig(to string) error {
	client, err := s.getEmailClient()
	if err != nil {
		return err
	}

	subject := "邮件配置测试"
	body := `
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
    <h2 style="color: #52c41a;">✅ 邮件配置成功！</h2>
    <p>恭喜，您的SMTP邮件服务配置正确。</p>
    <p style="color: #999; font-size: 12px;">此邮件由Emby用户管理系统发送</p>
</body>
</html>
`
	return client.Send(to, subject, body)
}
